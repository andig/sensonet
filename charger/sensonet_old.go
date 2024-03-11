package charger

import (
	"fmt"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/charger/sensonet_old"
	"github.com/evcc-io/evcc/util"
)

// Sensonet charger implementation
type Sensonet_old struct {
	conn *sensonet_old.Switch
	*switchSocket
}

func init() {
	registry.Add("sensonet_old", NewSensonetOldFromConfig)
}

// NewSensonetFromConfig creates a Sensonet charger from generic config
func NewSensonetOldFromConfig(other map[string]interface{}) (api.Charger, error) {
	var cc struct {
		embed         `mapstructure:",squash"`
		User          string
		Password      string
		PvUseStrategy string
		HeatingZone   int32
		Phases        int32
		//		HeatingVetoDuration      int32
		HeatingTemperatureOffset float64
		StandbyPower             float64
	}

	if err := util.DecodeOther(other, &cc); err != nil {
		return nil, err
	}
	//WW Setting api.feature "Heating(4)" manually, because it does not work from the evcc.yaml file
	//WW Can be deleted, if it works from the config file
	cc.embed.Features_ = append(cc.embed.Features_, 4)

	return NewSensonetOld(cc.embed, cc.User, cc.Password, cc.PvUseStrategy, cc.HeatingZone, cc.Phases, cc.HeatingTemperatureOffset, cc.StandbyPower)
}

// NewSensonet creates Sensonet charger
func NewSensonetOld(embed embed, user, password, pvUseStrategy string, heatingZone, phases int32, heatingTemperatureOffset, standbypower float64) (*Sensonet_old, error) {
	conn, err := sensonet_old.NewConnection(user, password, pvUseStrategy, heatingZone, phases, heatingTemperatureOffset)
	if err != nil {
		return nil, err
	}

	c := &Sensonet_old{
		conn: sensonet_old.NewSwitch(conn),
	}

	c.switchSocket = NewSwitchSocket(&embed, c.Enabled, c.conn.CurrentPower, standbypower)

	return c, nil
}

// Enabled implements the api.Charger interface
func (c *Sensonet_old) Enabled() (bool, error) {
	return c.conn.Enabled()
}

// Enable implements the api.Charger interface
func (c *Sensonet_old) Enable(enable bool) error {
	err := c.conn.Enable(enable)
	if err != nil {
		return err
	}

	enabled, err := c.Enabled()
	switch {
	case err != nil:
		return err
	case enable != enabled:
		onoff := map[bool]string{true: "on", false: "off"}
		return fmt.Errorf("switch %s failed", onoff[enable])
	default:
		return nil
	}

}

func (c *Sensonet_old) Phases() int {
	return c.conn.Phases()
}

// Status implements the api.ChargeState interface
func (c *Sensonet_old) Status() (api.ChargeStatus, error) {
	status, err := c.conn.Status()
	if err != nil {
		return api.StatusA, err
	}
	return status, err
}

func (c *Sensonet_old) ModeText() string {
	switch c.conn.CurrentQuickmode() {
	case sensonet_old.QUICKMODE_HOTWATER:
		return " (Hotwater Boost active)"
	case sensonet_old.QUICKMODE_HEATING:
		return " (Heating Quick Veto active)"
	}
	return " (No Quick Mode active; hotwater temperature shown)"
}
