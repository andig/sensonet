package sensonet_old

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/provider"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/request"
)

// Connection is the Sensonet connection
type Connection struct {
	*request.Helper
	log           *util.Logger
	token         string
	serialNumber  string
	pvUseStrategy string
	heatingZone   int32
	phases        int32
	//	heatingVetoDuration      int32
	heatingTemperatureOffset float64
	statusCache              provider.Cacheable[Vr921RelevantDataStruct]
	cache                    time.Duration
	currentQuickmode         string
	quickmodeStarted         int64
	onoff                    bool
}

// Global variable SensoNetConn is used to make data available in vehicle vks (not needed without vehicle vks)
var SensoNetOldConn *Connection

// NewConnection creates a new Sensonet device connection.
func NewConnection(user, password, pvUseStrategy string, heatingZone, phases int32, heatingTemperatureOffset float64) (*Connection, error) {

	log := util.NewLogger("sensonet")
	client := request.NewHelper(log)

	conn := &Connection{
		Helper: client,
	}
	conn.cache = 5 * time.Minute
	conn.pvUseStrategy = pvUseStrategy
	conn.heatingZone = heatingZone
	conn.phases = phases
	//	conn.heatingVetoDuration = heatingVetoDuration
	conn.heatingTemperatureOffset = heatingTemperatureOffset
	conn.log = log
	conn.currentQuickmode = ""
	conn.quickmodeStarted = time.Now().Unix()
	SensoNetOldConn = conn //this is not needed without vehicle vks

	var err error
	conn.Client.Jar, err = cookiejar.New(nil)
	if err != nil {
		err = fmt.Errorf("could not reset cookie jar. error: %s", err)
		return conn, err
	}
	conn.token, err = getToken(conn, user, password)
	if err != nil {
		err = fmt.Errorf("could not get token. error: %s", err)
		return conn, err
	}
	conn.log.DEBUG.Printf("New Token: %s\n", conn.token)
	err = authenticate(conn, user, conn.token)
	if err != nil {
		err = fmt.Errorf("could not authenticate. error: %s", err)
		return conn, err
	}
	conn.serialNumber, err = getAndExtractFacilitiesList(conn)
	if err != nil {
		err = fmt.Errorf("could not get serial number. error: %s", err)
		return conn, err
	}
	conn.statusCache = provider.ResettableCached(func() (Vr921RelevantDataStruct, error) {
		var res Vr921RelevantDataStruct
		err := getCurrentSystem(conn, &res)
		if err == nil {
			err = getCurrentLiveReport(conn, &res)
		}
		return res, err
	}, conn.cache)

	return conn, nil
}

func (c *Connection) reset() {
	c.statusCache.Reset()
	//c.meterCache.Reset()
}

func getSensonetHttpHeader() http.Header {
	// Returns the http header for http requests to sensonet
	return http.Header{
		"content-type":        {"application/json; charset=UTF-8"},
		"Accept-Encoding":     {"gzip"},
		"Accept":              {"application/json"},
		"Vaillant-Mobile-App": {"multiMATIC v2.1.45 b389 (Android)"},
		"Cache-Control":       {"no-cache"},
		"Pragma":              {"no-cache"},
	}
}

// VR921 functions
func getToken(c *Connection, user, pasword string) (string, error) {
	params, err := json.Marshal(map[string]string{
		"smartphoneId": "pymultiMATIC",
		"username":     user,
		"password":     pasword,
	})
	if err != nil {
		err = fmt.Errorf("error while json.Marshal. Error: %s", err)
		return "", err
	}

	urlnewtoken := TOKEN_URL
	req, err1 := http.NewRequest("POST", urlnewtoken, bytes.NewBuffer(params))
	if err1 != nil {
		err1 = fmt.Errorf("client: could not create request: %s", err1)
		return "", err1
	}
	req.Header = getSensonetHttpHeader()
	var data map[string]map[string]interface{}
	err = c.DoJSON(req, &data)
	if err != nil {
		err = fmt.Errorf("could not get token. error: %s", err)
		return "", err
	}
	return data["body"]["authToken"].(string), nil
}

func authenticate(c *Connection, user, token string) error {
	params, err := json.Marshal(map[string]string{
		"smartphoneId": "pymultiMATIC",
		"username":     user,
		"authToken":    token,
	})
	if err != nil {
		err = fmt.Errorf("error while json.Marshal. error: %s", err)
		return err
	}

	urlauthenticate := AUTH_URL
	req, err1 := http.NewRequest("POST", urlauthenticate, bytes.NewBuffer(params))
	if err1 != nil {
		err1 = fmt.Errorf("client: could not create request: %s", err1)
		return err1
	}
	req.Header = getSensonetHttpHeader()
	var resp []byte
	resp, err = c.DoBody(req)
	if err != nil {
		err = fmt.Errorf("could not authenticate. error: %s", err)
		c.log.DEBUG.Printf("Response: %s\n", resp)
		return err
	}
	return err
}

func getAndExtractFacilitiesList(c *Connection) (string, error) {
	urlFacilities := FACILITIES_URL
	var res FacilitiesListStruct
	err := c.GetJSON(urlFacilities, &res)
	if err != nil {
		err = fmt.Errorf("error getting facilities list: %s", err)
		return "", err
	}
	serialNumber := res.Body.FacilitiesList[0].SerialNumber
	return serialNumber, err
}

func getAndExtractSystem(c *Connection, relData *Vr921RelevantDataStruct) error {
	urlSystem := FACILITIES_URL + "/" + c.serialNumber + SYSTEM_URL
	var system *SystemStruct
	err := c.GetJSON(urlSystem, &system)
	if err != nil {
		err = fmt.Errorf("error getting system: %s", err)
		return err
	}
	relData.Timestamp = time.Now().Unix()
	relData.Status.Datetime = system.Body.Status.Datetime
	relData.Status.OutsideTemperature = system.Body.Status.OutsideTemperature
	relData.Hotwater.HotwaterTemperatureSetpoint = system.Body.Dhw.Hotwater.Configuration.HotwaterTemperatureSetpoint
	relData.Hotwater.OperationMode = system.Body.Dhw.Hotwater.Configuration.OperationMode
	if relData.Hotwater.CurrentQuickmode != system.Body.Dhw.Configuration.CurrentQuickmode {
		relData.Hotwater.CurrentQuickmode = system.Body.Dhw.Configuration.CurrentQuickmode
		if system.Body.Dhw.Configuration.CurrentQuickmode == "HOTWATER_BOOST" {
			c.currentQuickmode = QUICKMODE_HOTWATER
			c.quickmodeStarted = time.Now().Unix()
			c.onoff = true
		}
	}
	for _, systemResource := range system.Meta.ResourceState {
		//looking for "dhw/configuration" in link
		if strings.Contains(systemResource.Link.ResourceLink, "dhw/configuration") {
			relData.Hotwater.HotwaterSystemState = systemResource.State
			relData.Hotwater.HotwaterSystemTimestamp = systemResource.Timestamp
		}
	}

	//Extract information for all zones
	if len(relData.Zones) == 0 {
		relData.Zones = make([]Vr921RelevantDataZonesStruct, 0)
	}
	for i, systemBodyZone := range system.Body.Zones {
		if len(relData.Zones) <= i {
			//If relData.Zones array is not big enough, new elements are appended, especially at first ExtractSystem call
			//At the moment, relData.Zones is not shortened, if later GetSystem calls returns less system.Body.Zones
			zone := Vr921RelevantDataZonesStruct{}
			relData.Zones = append(relData.Zones, zone)
		}
		relData.Zones[i].Name = systemBodyZone.Configuration.Name
		relData.Zones[i].ID = systemBodyZone.ID
		relData.Zones[i].ActiveFunction = systemBodyZone.Configuration.ActiveFunction
		relData.Zones[i].Enabled = systemBodyZone.Configuration.Enabled
		relData.Zones[i].OperationMode = systemBodyZone.Heating.Configuration.OperationMode
		relData.Zones[i].CurrentDesiredSetpoint = systemBodyZone.Configuration.CurrentDesiredSetpoint
		relData.Zones[i].CurrentQuickmode = systemBodyZone.Configuration.CurrentQuickmode
		relData.Zones[i].QuickVeto.ExpiresAt = systemBodyZone.Configuration.QuickVeto.ExpiresAt
		relData.Zones[i].QuickVeto.TemperatureSetpoint = systemBodyZone.Configuration.QuickVeto.TemperatureSetpoint
		relData.Zones[i].InsideTemperature = systemBodyZone.Configuration.InsideTemperature
		if relData.Zones[i].CurrentQuickmode != "" {
			c.currentQuickmode = QUICKMODE_HEATING
			c.quickmodeStarted = time.Now().Unix()
			c.onoff = true
		}
	}
	return err
}

func getAndExtractLiveReport(c *Connection, relData *Vr921RelevantDataStruct) error {
	urlLiveReport := FACILITIES_URL + "/" + c.serialNumber + LIVEREPORT_URL
	var liveReport *LiveReportStruct
	err := c.GetJSON(urlLiveReport, &liveReport)
	if err != nil {
		err = fmt.Errorf("error getting system: %s", err)
		return err
	}
	relData.Timestamp = time.Now().Unix()
	for deviceNo, liveReportDevice := range liveReport.Body.Devices {
		//looking for device ID Control_DHW
		if liveReportDevice.ID == "Control_DHW" {
			for _, liveReportDeviceReport := range liveReportDevice.Reports {
				//looking for report ID DomesticHotWaterTankTemperature
				if liveReportDeviceReport.ID == "DomesticHotWaterTankTemperature" {
					relData.Hotwater.HotwaterLiveTemperature = liveReportDeviceReport.Value
				}
			}
			relData.Hotwater.HotwaterLiveState = liveReport.Meta.ResourceState[deviceNo].State
			relData.Hotwater.HotwaterLiveTimestamp = liveReport.Meta.ResourceState[deviceNo].Timestamp
		}
	}
	return err
}

func getCurrentSystem(c *Connection, relData *Vr921RelevantDataStruct) error {
	lastOperationMode := ""
	nbOfIdenticalVals := 0
	var err error
	for (relData.Hotwater.HotwaterSystemState != "SYNCED") && (nbOfIdenticalVals < 3) {
		err = getAndExtractSystem(c, relData)
		if err != nil {
			return err
		}
		if relData.Hotwater.OperationMode != lastOperationMode {
			lastOperationMode = relData.Hotwater.OperationMode
		} else {
			nbOfIdenticalVals = nbOfIdenticalVals + 1
		}
	}
	return err
}

func getCurrentLiveReport(c *Connection, relData *Vr921RelevantDataStruct) error {
	lastMeasuredTemp := 0.00
	nbOfIdenticalVals := 0
	var err error
	for (relData.Hotwater.HotwaterLiveState != "SYNCED") && (nbOfIdenticalVals < 3) {
		err = getAndExtractLiveReport(c, relData)
		if err != nil {
			return err
		}
		if relData.Hotwater.HotwaterLiveTemperature != lastMeasuredTemp {
			lastMeasuredTemp = relData.Hotwater.HotwaterLiveTemperature
		} else {
			nbOfIdenticalVals = nbOfIdenticalVals + 1
		}
	}
	return err
}

func (d *Connection) Phases() int {
	return int(d.phases)
}

func (d *Connection) CurrentQuickmode() string {
	return d.currentQuickmode
}

// CurrentTemp is called bei Soc
func (d *Connection) CurrentTemp() (float64, error) {
	res, err := d.statusCache.Get()
	if err != nil {
		d.log.ERROR.Println("Switch.CurrentTemp. Error: ", err)
		return 0, err
	}
	if d.CurrentQuickmode() == QUICKMODE_HEATING {
		currentTemp := 5.0
		for _, z := range res.Zones {
			if currentTemp == 5.0 && z.InsideTemperature > currentTemp {
				currentTemp = z.InsideTemperature
			}
			if z.ID == fmt.Sprintf("Control_ZO%01d", d.heatingZone) && z.InsideTemperature != 0.0 {
				currentTemp = z.InsideTemperature
			}
		}
		return currentTemp, nil
	}
	return float64(res.Hotwater.HotwaterLiveTemperature), nil
}

// TargetTemp is called bei TargetSoc
func (d *Connection) TargetTemp() (float64, error) {
	res, err := d.statusCache.Get()
	if err != nil {
		d.log.ERROR.Println("Switch.TargetTemp. Error: ", err)
		return 0, err
	}
	if d.CurrentQuickmode() == QUICKMODE_HEATING {
		for _, z := range res.Zones {
			if z.ID == fmt.Sprintf("Control_ZO%01d", d.heatingZone) {
				return float64(z.QuickVeto.TemperatureSetpoint), nil
			}
		}
	}
	return float64(res.Hotwater.HotwaterTemperatureSetpoint), nil
}

func (d *Connection) Status() (api.ChargeStatus, error) {
	status := api.StatusA // disconnected
	if d.CurrentQuickmode() != "" {
		status = api.StatusC
	} else {
		whichQuickMode, err := d.WhichQuickMode()
		if err != nil {
			err = fmt.Errorf("error while computing which quick mode to start: %s", err)
			return status, err
		}
		if whichQuickMode > 0 {
			status = api.StatusB
		}
	}
	return status, nil
}

/* Not needed anymore
// GetVr921 is used by vehicle sensonet_vehicle
func (d *Connection) GetVr921() (*Vr921RelevantDataStruct, error) {
	res, err := d.statusCache.Get()
	if err != nil {
		return &res, err
	}
	return &res, err
}*/

// This function checks the operation mode of heating and hotwater and the hotwater live temperature
// and returns, which quick mode should be started, when evcc sends an "Enable"
func (c *Connection) WhichQuickMode() (int, error) {
	//c := sh.Connection
	res, err := c.statusCache.Get()
	if err != nil {
		err = fmt.Errorf("could not read status cache before hotwater boost: %s", err)
		return 0, err
	}
	c.log.DEBUG.Println("PV Use Strategy = ", c.pvUseStrategy)
	c.log.DEBUG.Printf("Checking if hot water boost possible. Operation Mode = %s, temperature setpoint= %f, live temperature= %f", res.Hotwater.OperationMode, res.Hotwater.HotwaterTemperatureSetpoint, res.Hotwater.HotwaterLiveTemperature)
	hotWaterBoostPossible := false
	if res.Hotwater.HotwaterLiveTemperature <= res.Hotwater.HotwaterTemperatureSetpoint-5 &&
		res.Hotwater.OperationMode == OPERATIONMODE_TIME_CONTROLLED {
		hotWaterBoostPossible = true
	}

	heatingQuickVetoPossible := false
	for _, z := range res.Zones {
		if z.ID == fmt.Sprintf("Control_ZO%01d", c.heatingZone) {
			c.log.DEBUG.Printf("Checking if heating quick veto possible. Operation Mode = %s", z.OperationMode)
			if z.OperationMode == OPERATIONMODE_TIME_CONTROLLED {
				heatingQuickVetoPossible = true
			}
		}
	}

	whichQuickMode := 0
	switch c.pvUseStrategy {
	case PVUSESTRATEGY_HOTWATER:
		if hotWaterBoostPossible {
			whichQuickMode = 1
		} else {
			c.log.DEBUG.Println("PV Use Strategy = hotwater, but hotwater boost not possible")
		}
	case PVUSESTRATEGY_HEATING:
		if heatingQuickVetoPossible {
			whichQuickMode = 2
		} else {
			c.log.DEBUG.Println("PV Use Strategy = heating, but heating quick veto not possible")
		}
	case PVUSESTRATEGY_HOTWATER_THEN_HEATING:
		if hotWaterBoostPossible {
			whichQuickMode = 1
		} else {
			if heatingQuickVetoPossible {
				whichQuickMode = 2
			} else {
				c.log.DEBUG.Println("PV Use Strategy = hotwater_then_heating, but both not possible")
			}
		}
	}
	return whichQuickMode, err
}
