package vaillantEbus

import (
	"fmt"
	"math"
	"time"
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
	var err error
	d := sh.Connection

	err = d.getSystem(&d.relData, false)
	if err != nil {
		d.log.ERROR.Println("Switch.Enabled. Error: ", err)
		return false, err
	}
	d.log.DEBUG.Println("Status last read from ebusd at:", d.relData.Status.Time)
	if d.currentQuickmode != "" {
		d.log.DEBUG.Println("In Switch.Enabled: Connection.currentQuickmode:", d.currentQuickmode, "started at:", (d.quickmodeStarted).Format("2006-01-02 15:04:05"))
	} else {
		d.log.DEBUG.Println("In Switch.Enabled: Connection.currentQuickmode not set. Timestamp:", (d.quickmodeStarted).Format("2006-01-02 15:04:05"))
		if (d.relData.Hotwater.HwcSFMode != "still necessary?") && (d.relData.Hotwater.HwcSFMode != "auto") {
			d.log.DEBUG.Println("In Switch.Enabled: d.relData.Hotwater.HwcSFMode should be inactive but is on")
			if d.quickmodeStarted.Add(1 * time.Minute).Before(time.Now()) {
				// When the reported hotwater.CurrentQuickmode is not "Regular" more then 5 minutes after the end of the charge session (or the start of evcc),
				// this means that the heat pump is in hotwater boost
				d.currentQuickmode = QUICKMODE_HOTWATER
				d.quickmodeStarted = time.Now()
				d.onoff = true
			}
		}
		for _, z := range d.relData.Zones {
			if z.Index == d.heatingZone {
				d.log.DEBUG.Println("In Switch.Enabled: Zone quick mode:", z.SFMode, ", Temperature Setpoint:", z.QuickVetoTemp, "(", d.quickVetoSetPoint, "), Expires at:", d.quickVetoExpiresAt)
				if (z.SFMode != "still necessary?") && (z.SFMode != "auto") {
					d.log.DEBUG.Println("In Switch.Enabled: z.CurrentQuickmode should be inactive but is on")
					if d.quickmodeStarted.Add(1 * time.Minute).Before(time.Now()) {
						// When the reported z.CurrentQuickmode is not "NONE" more then 5 minutes after the end of a charge session (or the start of evcc),
						// this means that the zone quick veto startet by other means as evcc
						d.currentQuickmode = QUICKMODE_HEATING
						d.quickmodeStarted = time.Now()
						d.onoff = true
					}
				}
			}
		}
	}
	switch d.currentQuickmode {
	case QUICKMODE_HOTWATER:
		d.log.DEBUG.Println("In Switch.Enabled: Hotwater quick mode:", d.relData.Hotwater.HwcSFMode)
		if (d.relData.Hotwater.HwcSFMode == "still necessary?") || (d.relData.Hotwater.HwcSFMode == "auto") {
			d.log.DEBUG.Println("In Switch.Enabled: res.Hotwater.CurrentQuickmode should be active but is off")
			if d.quickmodeStarted.Add(1 * time.Minute).Before(time.Now()) {
				// When the reported hotwater.CurrentQuickmode has changed to "Regular" more then 5 minutes after the beginning of the charge session,
				// this means that the heat pump has stopped the hotwater boost itself
				d.currentQuickmode = ""
				d.quickmodeStopped = time.Now()
				d.onoff = false
			}
		}
	case QUICKMODE_HEATING:
		for _, z := range d.relData.Zones {
			if z.Index == d.heatingZone {
				d.log.DEBUG.Println("In Switch.Enabled: Zone quick mode:", z.SFMode, ", Temperature Setpoint:", z.QuickVetoTemp, "(", d.quickVetoSetPoint, "), Expires at:", d.quickVetoExpiresAt)
				if (z.SFMode == "still necessary?") || (z.SFMode == "auto") {
					d.log.DEBUG.Println("In Switch.Enabled: z.CurrentQuickmode should be active but is off")
					if d.quickmodeStarted.Add(1 * time.Minute).Before(time.Now()) {
						// When the reported z.CurrentQuickmode has changed to "NONE" more then 5 minutes after the beginning of the charge session,
						// this means that the zone quick veto ended or was stopped by other means as evcc
						d.currentQuickmode = ""
						d.quickmodeStopped = time.Now()
						d.onoff = false
					}
				}
			}
		}
	case QUICKMODE_NOTHING:
		if d.quickmodeStarted.Add(10 * time.Minute).Before(time.Now()) {
			d.log.DEBUG.Println("Idle charge mode for more than 10 minutes. Turning it off")
			d.currentQuickmode = ""
			d.quickmodeStopped = time.Now()
			d.onoff = false
		}
	case "":
		//Nothing to do
	default:
		d.log.ERROR.Println("Unknown quick mode in case statement:", d.currentQuickmode)
	}
	return d.onoff, nil
}

// Enable implements the api.Charger interface
func (sh *Switch) Enable(enable bool) error {
	var err error
	var whichQuickMode int
	d := sh.Connection
	//get new reports from ebusd before starting or stopping quick modes
	/*err := d.getSystem(&d.relData, true)
	if err != nil {
		err = fmt.Errorf("could not read status cache before hotwater boost: %s", err)
		return err
	}*/
	if d.currentQuickmode == "" && d.quickmodeStopped.After(d.quickmodeStarted) && d.quickmodeStopped.Add(2*time.Minute).After(time.Now()) {
		enable = false
	}

	if enable {
		whichQuickMode, err = d.WhichQuickMode()
		if err != nil {
			err = fmt.Errorf("error while computing which quick mode to start: %s", err)
			return err
		}

		switch whichQuickMode {
		case 1:
			if d.currentQuickmode == QUICKMODE_HEATING && d.relData.Zones[0].SFMode == "load" {
				//if zone quick veto active, then stop it
				err = sh.stopZoneQuickVeto()
				if err == nil {
					d.log.DEBUG.Println("Stopping Zone Quick Veto")
				}
			}
			err = sh.startHotWaterBoost()
			if err == nil {
				d.currentQuickmode = QUICKMODE_HOTWATER
				d.quickmodeStarted = time.Now()
				d.log.DEBUG.Println("Starting Hotwater Boost")
			}
		case 2:
			if d.currentQuickmode == QUICKMODE_HOTWATER && d.relData.Hotwater.HwcSFMode == "load" {
				//if hotwater boost active, then stop it
				err = sh.stopHotWaterBoost()
				if err == nil {
					d.log.DEBUG.Println("Stopping Hotwater Boost")
				}
			}
			err = sh.startZoneQuickVeto()
			if err == nil {
				d.currentQuickmode = QUICKMODE_HEATING
				d.quickmodeStarted = time.Now()
				d.log.DEBUG.Println("Starting Zone Quick Veto")
			}
		default:
			if d.currentQuickmode == QUICKMODE_HOTWATER && d.relData.Hotwater.HwcSFMode == "load" {
				//if hotwater boost active, then stop it
				err = sh.stopHotWaterBoost()
				if err == nil {
					d.log.DEBUG.Println("Stopping Hotwater Boost")
				}
			}
			if d.currentQuickmode == QUICKMODE_HEATING && d.relData.Zones[0].SFMode == "load" {
				//if zone quick veto active, then stop it
				err = sh.stopZoneQuickVeto()
				if err == nil {
					d.log.DEBUG.Println("Stopping Zone Quick Veto")
				}
			}
			d.currentQuickmode = QUICKMODE_NOTHING
			d.quickmodeStarted = time.Now()
			d.log.DEBUG.Println("Enable called but no quick mode possible. Starting idle mode")
			//d.log.INFO.Println("Enable called but no quick mode possible")
		}
	} else {
		switch d.currentQuickmode {
		case QUICKMODE_HOTWATER:
			err = sh.stopHotWaterBoost()
			if err == nil {
				d.log.DEBUG.Println("Stopping Hotwater Boost")
			}
		case QUICKMODE_HEATING:
			err = sh.stopZoneQuickVeto()
			if err == nil {
				d.log.DEBUG.Println("Stopping Zone Quick Veto")
			}
		case QUICKMODE_NOTHING:
			d.log.DEBUG.Println("Stopping idle quick mode")
		default:
			d.log.DEBUG.Println("Nothing to do, no quick mode active")
		}
		d.currentQuickmode = ""
	}
	if err == nil {
		d.onoff = enable
		err = d.getSFMode(&d.relData) //Update SFMode for hotwater and heating zone
	}
	return err
}

// CurrentPower implements the api.Meter interface
// Those are just dummy values. For eal values, an energy meter like Shelly 3EM is necessary
func (sh *Switch) CurrentPower() (float64, error) {
	d := sh.Connection
	power := d.relData.Status.CurrentConsumedPower * 1000

	d.log.DEBUG.Println("Switch.CurrentPower", d.currentQuickmode, "Power:", power)
	return power, nil
}

func (sh *Switch) startHotWaterBoost() error {
	c := sh.Connection
	message := " -c " + c.relData.Status.ControllerForSFMode + " " + EBUSDREAD_HOTWATER_SFMODE + " load"
	err := c.ebusdWrite(message)
	if err != nil {
		err = fmt.Errorf("could not start hotwater boost. Error: %s", err)
		return err
	}
	return err
}

func (sh *Switch) stopHotWaterBoost() error {
	c := sh.Connection
	message := " -c " + c.relData.Status.ControllerForSFMode + " " + EBUSDREAD_HOTWATER_SFMODE + " auto"
	err := c.ebusdWrite(message)
	if err != nil {
		err = fmt.Errorf("could not stop hotwater boost. Error: %s", err)
		return err
	}
	return err
}

func (sh *Switch) startZoneQuickVeto() error {
	c := sh.Connection
	i := 0 //Index for relData.zones[]
	zonePrefix := fmt.Sprintf("z%01d", c.heatingZone)
	temperatureSetpoint := c.relData.Zones[i].ActualRoomTempDesired
	if temperatureSetpoint == 0.0 {
		c.log.ERROR.Printf("Could not detect current desired setpoint for zone: %01d. Probably inactive due to time program.\n", c.heatingZone)
		temperatureSetpoint = 20.0
	}
	// Temperature Setpoint for quick veto is rounded to 0.5 °C
	vetoSetpoint := float32(math.Round(2*(temperatureSetpoint+c.heatingTemperatureOffset)) / 2.0)
	vetoDuration := float32(0.5)
	//Hier muss noch die Quic-Veto-Dauer gesetzt werden
	message := " -c " + c.relData.Status.ControllerForSFMode + " " + zonePrefix + EBUSDREAD_ZONE_QUICKVETOTEMP + fmt.Sprintf(" %2.1f", vetoSetpoint)
	err := c.ebusdWrite(message)
	if err != nil {
		err = fmt.Errorf("could not start zone quick veto. Error: %s", err)
		return err
	}
	message = " -c " + c.relData.Status.ControllerForSFMode + " " + zonePrefix + EBUSDREAD_ZONE_SFMODE + " load"
	err = c.ebusdWrite(message)
	if err != nil {
		err = fmt.Errorf("could not start zone quick veto. Error: %s", err)
		return err
	}
	c.quickVetoSetPoint = vetoSetpoint
	c.quickVetoExpiresAt = (time.Now().Add(time.Duration(int64(vetoDuration*60) * int64(time.Minute)))).Format("15:04")
	return err
}

func (sh *Switch) stopZoneQuickVeto() error {
	c := sh.Connection
	zonePrefix := fmt.Sprintf("z%01d", c.heatingZone)
	message := " -c " + c.relData.Status.ControllerForSFMode + " " + zonePrefix + EBUSDREAD_ZONE_SFMODE + " auto"
	err := c.ebusdWrite(message)
	if err != nil {
		err = fmt.Errorf("could not stop zone quick veto. Error: %s", err)
		return err
	}
	//Hier müssen ggf. noch Ergänzungen hin
	c.quickVetoSetPoint = 0
	c.quickVetoExpiresAt = ""
	return err
}
