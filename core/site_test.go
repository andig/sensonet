package core

import (
	"testing"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/util"
	"github.com/stretchr/testify/assert"
)

func TestSitePower(t *testing.T) {
	tc := []struct {
		maxGrid, grid, battery, site float64
	}{
		{0, 0, 0, 0},    // silent night
		{0, 0, 1, 1},    // battery discharging
		{0, 0, -1, -1},  // battery charging -> negative result cannot occur in reality
		{0, 1, 0, 1},    // grid import
		{0, 1, 1, 2},    // grid import + battery discharging
		{0, -1, 0, -1},  // grid export
		{0, -1, -1, -2}, // grid export + battery charging
		{0, 1, -1, 0},   // grid import + battery charging -> should not happen
		{0.5, 1, -1, 1}, // grid import + DC battery charging
	}

	log := util.NewLogger("foo")

	for _, tc := range tc {
		res := sitePower(log, tc.maxGrid, tc.grid, tc.battery, 0)
		if res != tc.site {
			t.Errorf("sitePower wanted %.f, got %.f", tc.site, res)
		}
	}
}

func TestGreenShare(t *testing.T) {
	tc := []struct {
		title                                                 string
		grid, pv, battery, home, lp                           float64
		greenShareTotal, greenShareHome, greenShareLoadpoints float64
	}{
		{
			"half grid, half pv, green home",
			1000, 1000, 0, 1000, 1000,
			0.5, 1, 0,
		},
		{
			"half grid, half pv, no home",
			1000, 1000, 0, 0, 2000,
			0.5, 1, 0.5,
		},
		{
			"half grid, half pv, no lp",
			2500, 2500, 0, 5000, 0,
			0.5, 0.5, 0,
		},
		{
			"full pv",
			0, 5000, 0, 1000, 4000,
			1, 1, 1,
		},
		{
			"full grid",
			5000, 0, 0, 1000, 4000,
			0, 0, 0,
		},
		{
			"half grid, half battery, green home",
			1000, 0, 1000, 1000, 1000,
			0.5, 1, 0,
		},
		{
			"half grid, half battery, no home",
			1000, 0, 1000, 0, 2000,
			0.5, 1, 0.5,
		},
		{
			"half grid, half battery, no lp",
			1000, 0, 1000, 2000, 0,
			0.5, 0.5, 0,
		},
		{
			"full pv, pv export",
			-5000, 10000, 0, 1000, 4000,
			1, 1, 1,
		},
		{
			"full pv, pv export, no lp",
			-5000, 10000, 0, 5000, 0,
			1, 1, 1,
		},
		{
			"full pv, pv export, battery charge",
			-2500, 10000, -2500, 1000, 4000,
			1, 1, 1,
		},
		{
			"full grid, battery charge",
			3000, 0, -1000, 1000, 1000,
			0, 0, 0,
		},
		{
			"full grid, battery charge, no lp",
			2000, 0, -1000, 1000, 0,
			0, 0, 0,
		},
		{
			"half grid, half pv, battery charge, no lp",
			1000, 1000, -1000, 1000, 0,
			0.5, 1, 0,
		},
		{
			"half grid, half pv, battery charge, home, lp",
			1000, 1000, -1000, 500, 500,
			0.5, 1, 0,
		},
		{
			"pv ac limited, battery charge & grid import",
			1000, 3000, -1000, 1000, 2000,
			0.75, 1, 0.5,
		},
	}

	for _, tc := range tc {
		t.Log(tc.title)

		s := &Site{
			gridPower:    tc.grid,
			pvPower:      tc.pv,
			batteryPower: tc.battery,
		}

		totalPower := tc.grid + tc.pv + max(0, tc.battery)
		greenShareTotal := s.greenShare(0, totalPower)
		if greenShareTotal != tc.greenShareTotal {
			t.Errorf("greenShareTotal wanted %.3f, got %.3f", tc.greenShareTotal, greenShareTotal)
		}
		greenShareHome := s.greenShare(0, tc.home)
		if greenShareHome != tc.greenShareHome {
			t.Errorf("greenShareHome wanted %.3f, got %.3f", tc.greenShareHome, greenShareHome)
		}
		greenShareLoadpoints := s.greenShare(tc.home+max(0, -tc.battery), totalPower)
		if greenShareLoadpoints != tc.greenShareLoadpoints {
			t.Errorf("greenShareLoadpoints wanted %.3f, got %.3f", tc.greenShareLoadpoints, greenShareLoadpoints)
		}
	}
}

func TestRequiredBatteryMode(t *testing.T) {
	tc := []struct {
		gridChargeActive bool
		mode, res        api.BatteryMode
	}{
		{false, api.BatteryUnknown, api.BatteryUnknown}, // ignore
		{false, api.BatteryNormal, api.BatteryUnknown},  // ignore
		{false, api.BatteryHold, api.BatteryNormal},
		{false, api.BatteryCharge, api.BatteryNormal},

		{true, api.BatteryUnknown, api.BatteryCharge},
		{true, api.BatteryNormal, api.BatteryCharge},
		{true, api.BatteryHold, api.BatteryCharge},
		{true, api.BatteryCharge, api.BatteryUnknown}, // ignore
	}

	for _, tc := range tc {
		t.Logf("%+v", tc)

		s := &Site{
			batteryMode: tc.mode,
		}

		res := s.requiredBatteryMode(tc.gridChargeActive, api.Rate{})
		assert.Equal(t, tc.res, res, "expected %s, got %s", tc.res, res)
	}
}
