package sensonet_old

import (
	"bytes"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/evcc-io/evcc/util/request"
)

type Switch struct {
	*Connection
}

func NewSwitch(conn *Connection) *Switch {
	res := &Switch{
		Connection: conn,
	}

	return res
}

// Enabled implements the api.Charger interface
func (sh *Switch) Enabled() (bool, error) {
	d := sh.Connection
	res, err := d.statusCache.Get()
	if err != nil {
		d.log.ERROR.Println("Switch.Enabled. Error: ", err)
		return false, err
	}
	d.log.DEBUG.Println("Switch.Enabled. d.currentQuickmode:", d.currentQuickmode, "started at:", time.Unix(d.quickmodeStarted, 0))
	if d.currentQuickmode == QUICKMODE_HOTWATER {
		d.log.DEBUG.Println("Switch.Enabled. Hotwater quick mode:", res.Hotwater.CurrentQuickmode)
		if res.Hotwater.CurrentQuickmode == "" {
			d.log.DEBUG.Println("Switch.Enabled. res.Hotwater.CurrentQuickmode should be active but is off")
			//return false, nil
		}
	}
	if d.currentQuickmode == QUICKMODE_HEATING {
		for _, z := range res.Zones {
			if z.ID == fmt.Sprintf("Control_ZO%01d", d.heatingZone) {
				d.log.DEBUG.Println("Switch.Enabled. Zone quick mode:", z.CurrentQuickmode, ", Temperature Setpoint:", z.QuickVeto.TemperatureSetpoint, ", Expires at:", z.QuickVeto.ExpiresAt)
				if z.CurrentQuickmode == "" {
					d.log.DEBUG.Println("Switch.Enabled. z.CurrentQuickmode should be active but is off")
					//return false, nil
				}
			}
		}
	}
	return d.onoff, nil
}

// Enable implements the api.Charger interface
func (sh *Switch) Enable(enable bool) error {
	var err error

	d := sh.Connection
	//Reset status cache and get new reports from sensonet before starting or stopping quick modes
	d.reset()
	res, err := d.statusCache.Get()
	if err != nil {
		err = fmt.Errorf("could not read status cache before hotwater boost: %s", err)
		return err
	}
	if enable {
		whichQuickMode, err := d.WhichQuickMode()
		if err != nil {
			err = fmt.Errorf("error while computing which quick mode to start: %s", err)
			return err
		}

		switch whichQuickMode {
		case 1:
			err = sh.startHotWaterBoost()
			if err == nil {
				d.currentQuickmode = QUICKMODE_HOTWATER
				d.quickmodeStarted = time.Now().Unix()
				d.log.DEBUG.Println("Starting quick mode (hotwater boost)", res.Hotwater.CurrentQuickmode)
			}
		case 2:
			err = sh.startZoneQuickVeto(&res)
			if err == nil {
				d.currentQuickmode = QUICKMODE_HEATING
				d.quickmodeStarted = time.Now().Unix()
				d.log.DEBUG.Println("Starting zone quick veto")
			}
		default:
			d.log.DEBUG.Println("Enable called but no quick mode possible")
		}
	} else {
		switch d.currentQuickmode {
		case QUICKMODE_HOTWATER:
			err = sh.stopHotWaterBoost()
			if err == nil {
				d.log.DEBUG.Println("Stopping Quick Mode", res.Hotwater.CurrentQuickmode)
			}
		case QUICKMODE_HEATING:
			err = sh.stopZoneQuickVeto()
			if err == nil {
				d.log.DEBUG.Println("Stopping Zone Quick Veto")
			}
		default:
			d.log.DEBUG.Println("Nothing to do, no quick mode active")
		}
		d.currentQuickmode = ""
		d.quickmodeStarted = time.Now().Unix()
	}
	//Reset status cache and get new reports from sensonet
	d.reset()
	res, err1 := d.statusCache.Get()
	if err1 != nil {
		err1 = fmt.Errorf("could not get current live and system report after starting/stopping quick modes: %s", err1)
		return err1
	}
	d.log.DEBUG.Println("Switch.Enable. Hotwater quick mode:", res.Hotwater.CurrentQuickmode)
	d.onoff = enable
	return err
}

// CurrentPower implements the api.Meter interface
func (sh *Switch) CurrentPower() (float64, error) {
	var power float64

	d := sh.Connection
	d.log.DEBUG.Println("Switch.CurrentPower", d.currentQuickmode, time.Unix(d.quickmodeStarted, 0))

	// Returns dummy values for CurrentPower if called
	if d.onoff {
		power = 3000.0
	} else {
		power = 0.0
	}
	if d.currentQuickmode == QUICKMODE_HEATING {
		power = 1500.0
	}
	return power, nil
}

func (sh *Switch) startHotWaterBoost() error {
	c := sh.Connection
	urlHotwaterBoost := FACILITIES_URL + "/" + c.serialNumber + HOTWATERBOOST_URL
	req, err := http.NewRequest("PUT", urlHotwaterBoost, bytes.NewBuffer(nil))
	if err != nil {
		err = fmt.Errorf("client: could not create request: %s", err)
		return err
	}
	req.Header = getSensonetHttpHeader()
	var resp []byte
	resp, err = c.DoBody(req)
	if err != nil {
		err = fmt.Errorf("could not start hotwater boost. Error: %s", err)
		c.log.DEBUG.Printf("Response: %s\n", resp)
		return err
	}
	return err
}

func (sh *Switch) stopHotWaterBoost() error {
	c := sh.Connection
	urlHotwaterBoost := FACILITIES_URL + "/" + c.serialNumber + HOTWATERBOOST_URL
	req, err := http.NewRequest("DELETE", urlHotwaterBoost, bytes.NewBuffer(nil))
	if err != nil {
		err = fmt.Errorf("client: could not create request: %s", err)
		return err
	}
	req.Header = getSensonetHttpHeader()
	var resp []byte
	resp, err = c.DoBody(req)
	if err != nil {
		err = fmt.Errorf("could not stop hotwater boost. Error: %s", err)
		c.log.DEBUG.Printf("Response: %s\n", resp)
		return err
	}
	return err
}

func (sh *Switch) startZoneQuickVeto(relData *Vr921RelevantDataStruct) error {
	c := sh.Connection
	urlZoneQuickVeto := FACILITIES_URL + "/" + c.serialNumber + fmt.Sprintf(ZONEQUICKVETO_URL, c.heatingZone)
	temperatureSetpoint := 0.0
	for _, z := range relData.Zones {
		if z.ID == fmt.Sprintf("Control_ZO%01d", c.heatingZone) {
			temperatureSetpoint = z.CurrentDesiredSetpoint
		}
	}
	if temperatureSetpoint == 0.0 {
		c.log.ERROR.Printf("Could not detect current desired setpoint for zone: Control_ZO%01d. Probably inactive due to time program.\n", c.heatingZone)
		temperatureSetpoint = 20.0
	}
	// Temperature Setpoint for quick veto is rounded to 0.5 °C
	vetoSetpoint := float32(math.Round(2*(temperatureSetpoint+c.heatingTemperatureOffset)) / 2.0)
	data := map[string]float32{
		"temperature_setpoint": float32(vetoSetpoint),
		"duration":             float32(0.5), // duration for quick veto is 0,5 hours
		//		"duration":             float32(c.heatingVetoDuration) / 30.0,
	}
	req, err := http.NewRequest("PUT", urlZoneQuickVeto, request.MarshalJSON(data))
	if err != nil {
		err = fmt.Errorf("client: could not create request: %s", err)
		return err
	}
	req.Header = getSensonetHttpHeader()
	var resp []byte
	c.log.DEBUG.Printf("Sending PUT request to: %s\n", urlZoneQuickVeto)
	resp, err = c.DoBody(req)
	if err != nil {
		err = fmt.Errorf("could not start quick veto. Error: %s", err)
		c.log.DEBUG.Printf("Response: %s\n", resp)
		return err
	}
	return err
}

func (sh *Switch) stopZoneQuickVeto() error {
	c := sh.Connection
	urlZoneQuickVeto := FACILITIES_URL + "/" + c.serialNumber + fmt.Sprintf(ZONEQUICKVETO_URL, c.heatingZone)
	req, err := http.NewRequest("DELETE", urlZoneQuickVeto, bytes.NewBuffer(nil))
	if err != nil {
		err = fmt.Errorf("client: could not create request: %s", err)
		return err
	}
	req.Header = getSensonetHttpHeader()
	var resp []byte
	c.log.DEBUG.Printf("Sending DELETE request to: %s\n", urlZoneQuickVeto)
	resp, err = c.DoBody(req)
	if err != nil {
		err = fmt.Errorf("could not stop quick veto. Error: %s", err)
		c.log.DEBUG.Printf("Response: %s\n", resp)
		return err
	}
	return err
}
