package sensonet

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/provider"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/request"
	"github.com/thanhpk/randstr"
)

// Connection is the Sensonet connection
type Connection struct {
	*request.Helper
	log            *util.Logger
	user           string
	password       string
	realm          string
	tokenRes       TokenRequestStruct
	code           string
	codeVerifier   string
	tokenExpiresAt time.Time
	systemId       string
	pvUseStrategy  string
	heatingZone    int
	phases         int
	//	heatingVetoDuration      int32
	heatingTemperatureOffset float64
	statusCache              provider.Cacheable[Vr921RelevantDataStruct]
	cache                    time.Duration
	currentQuickmode         string
	quickmodeStarted         time.Time
	quickmodeStopped         time.Time
	onoff                    bool
	quickVetoSetPoint        float32
	quickVetoExpiresAt       string
}

// Global variable SensoNetConn is used to make data available in vehicle vks (not needed without vehicle vks)
var SensoNetConn *Connection

// NewConnection creates a new Sensonet device connection.
func NewConnection(user, password, realm, pvUseStrategy string, heatingZone, phases int, heatingTemperatureOffset float64) (*Connection, error) {
	log := util.NewLogger("sensonet")
	client := request.NewHelper(log)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	conn := &Connection{
		Helper: client,
	}
	conn.cache = 90 * time.Second
	conn.user = user
	conn.password = password
	conn.realm = realm
	conn.pvUseStrategy = pvUseStrategy
	conn.heatingZone = heatingZone
	conn.phases = phases
	//	conn.heatingVetoDuration = heatingVetoDuration
	conn.heatingTemperatureOffset = heatingTemperatureOffset
	conn.log = log
	conn.currentQuickmode = ""
	conn.quickmodeStarted = time.Now()
	SensoNetConn = conn //this is not needed without vehicle sensonet_vehicle

	var err error
	conn.Client.Jar, err = cookiejar.New(nil)
	if err != nil {
		err = fmt.Errorf("could not reset cookie jar. error: %s", err)
		return conn, err
	}

	err = conn.loginAndGetToken()
	if err != nil {
		//err = fmt.Errorf("could not login and get token. error: %s", err)
		return conn, err
	}

	conn.systemId, err = conn.getHomes()
	if err != nil {
		err = fmt.Errorf("could not get systemId. error: %s", err)
		return conn, err
	}

	conn.statusCache = provider.ResettableCached(func() (Vr921RelevantDataStruct, error) {
		var res Vr921RelevantDataStruct
		err := conn.getSystem(&res)
		/*if err == nil {
			err = conn.getCurrentLiveReport(&res)
		}*/
		return res, err
	}, conn.cache)

	return conn, nil
}

func (c *Connection) loginAndGetToken() error {
	var err error
	c.code, c.codeVerifier, err = c.getCode()
	if err != nil {
		//err = fmt.Errorf("could not get code. error: %s", err)
		return err
	}
	c.log.DEBUG.Printf("New Code: %s\n", c.code)
	c.tokenRes, err = c.getToken()
	if err != nil {
		//err = fmt.Errorf("could not get token. error: %s", err)
		return err
	}
	c.log.DEBUG.Println("Got new Token:")
	//c.log.DEBUG.Printf("New Token: %s\n", c.token)
	c.tokenExpiresAt = time.Now().Add(time.Duration(c.tokenRes.ExpiresIn * int64(time.Second)))
	c.log.DEBUG.Printf("Token expires at: %02d:%02d:%02d", c.tokenExpiresAt.Hour(), c.tokenExpiresAt.Minute(), c.tokenExpiresAt.Second())
	return err
}

func (c *Connection) reset() {
	c.statusCache.Reset()
	//c.meterCache.Reset()
}

func (c *Connection) getSensonetHttpHeader() http.Header {
	// Returns the http header for http requests to sensonet
	return http.Header{
		"Authorization":             {"Bearer " + c.tokenRes.AccessToken},
		"x-app-identifier":          {"VAILLANT"},
		"Accept-Language":           {"en-GB"},
		"Accept":                    {"application/json, text/plain, */*"},
		"x-client-locale":           {"en-GB"},
		"x-idm-identifier":          {"KEYCLOAK"},
		"ocp-apim-subscription-key": {"1e0a2f3511fb4c5bbb1c7f9fedd20b1c"},
		"User-Agent":                {"okhttp/4.9.2"},
		"Connection":                {"keep-alive"},
		//"Content-Type":              {"application/json"},
	}
}

func generateCode() (string, string) {
	codeVerifier := randstr.String(128)
	sha2 := sha256.New()
	io.WriteString(sha2, codeVerifier)
	codeChallenge := base64.RawURLEncoding.EncodeToString(sha2.Sum(nil))
	return codeVerifier, codeChallenge
}

func computeLoginUrl(loginHtlm, realm string) string {
	loginUrl := fmt.Sprintf(LOGIN_URL, realm)
	index1 := strings.Index(loginHtlm, "authenticate?")
	index2 := strings.Index(loginHtlm[index1:], "\"")
	loginUrl = loginUrl + loginHtlm[index1+12:index1+index2]
	/*result = re.search(fmt.Sprintf(LOGIN_URL, realm)+ r"\?([^\"]*)",
		login_html,
	)*/
	return html.UnescapeString(loginUrl)
}

func (c *Connection) getCode() (string, string, error) {
	codeVerifier, codeChallenge := generateCode()
	code := ""
	auth_querystring := url.Values{}
	auth_querystring.Set("response_type", "code")
	auth_querystring.Set("client_id", CLIENT_ID)
	auth_querystring.Set("code", "code_challenge")
	auth_querystring.Set("redirect_uri", "enduservaillant.page.link://login")
	auth_querystring.Set("code_challenge_method", "S256")
	auth_querystring.Set("code_challenge", codeChallenge)

	urlnewcode := fmt.Sprintf(AUTH_URL, c.realm) + "?" + auth_querystring.Encode()

	req, err1 := http.NewRequest("GET", urlnewcode, nil)
	if err1 != nil {
		err1 = fmt.Errorf("client: could not create request: %s", err1)
		return "", "", err1
	}

	resp, err := c.Do(req)
	if err != nil {
		err = fmt.Errorf("could not get code. error: %s", err)
		return "", "", err
	}
	defer resp.Body.Close()
	loginHtml, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("could not read response body. error: %s", err)
		return "", "", err
	}
	//c.log.DEBUG.Printf("Got login html %s", loginHtml)
	if val, ok := resp.Header["Location"]; ok {
		parsedUrl, _ := url.Parse(val[0])
		code = parsedUrl.Query()["code"][0]
	}
	if code != "" {
		c.log.DEBUG.Println("Got code from first http request: ", code)
		return code, codeVerifier, err
	}

	loginUrl := computeLoginUrl(string(loginHtml), c.realm)
	if loginUrl == "" {
		err = api.ErrTimeout
		err = fmt.Errorf("could not compute login url. error: %s", err)
		return "", "", err
	}
	c.log.DEBUG.Printf("Got login url %s", loginUrl)

	params := url.Values{}
	params.Set("username", c.user)
	params.Set("password", c.password)
	params.Set("credentialId", "")
	req1, err3 := http.NewRequest("POST", loginUrl, strings.NewReader(params.Encode()))
	if err3 != nil {
		err3 = fmt.Errorf("getCode: could not create request: %s", err3)
		return "", "", err3
	}
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = c.Do(req1)
	if err != nil {
		err = fmt.Errorf("could not get code. error: %s", err)
		return "", "", err
	}
	if val, ok := resp.Header["Location"]; ok {
		parsedUrl, _ := url.Parse(val[0])
		code = parsedUrl.Query()["code"][0]
		return code, codeVerifier, nil
	}
	err = api.ErrMissingCredentials
	err = fmt.Errorf("could not get code from second http request. error: %s", err)
	return "", "", err
}

func (c *Connection) getToken() (TokenRequestStruct, error) {
	var tokenRes TokenRequestStruct
	params := url.Values{}
	params.Set("grant_type", "authorization_code")
	params.Set("client_id", CLIENT_ID)
	params.Set("code", c.code)
	params.Set("code_verifier", c.codeVerifier)
	params.Set("redirect_uri", "enduservaillant.page.link://login")

	urlnewtoken := fmt.Sprintf(TOKEN_URL, c.realm)
	req1, err := http.NewRequest("POST", urlnewtoken, strings.NewReader(params.Encode()))
	if err != nil {
		err = fmt.Errorf("getToken: could not create request: %s", err)
		return tokenRes, err
	}
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	err = c.DoJSON(req1, &tokenRes)
	if err != nil {
		err = fmt.Errorf("could not get token. error: %s", err)
		return tokenRes, err
	}
	return tokenRes, nil
}

func (c *Connection) refreshToken() error {
	var tokenRes TokenRequestStruct
	params := url.Values{}
	params.Set("grant_type", "refresh_token")
	params.Set("client_id", CLIENT_ID)
	params.Set("refresh_token", c.tokenRes.RefreshToken)

	urlnewtoken := fmt.Sprintf(TOKEN_URL, c.realm)
	req1, err := http.NewRequest("POST", urlnewtoken, strings.NewReader(params.Encode()))
	if err != nil {
		err = fmt.Errorf("refreshToken: could not create request: %s", err)
		return err
	}
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	err = c.DoJSON(req1, &tokenRes)
	if err != nil {
		c.log.INFO.Printf("Call of refreshToken not susccessful. Error: %s TryingloginAndGetToken", err)
		err = c.loginAndGetToken()
		if err != nil {
			err = fmt.Errorf("could not relogin in refresh token. error: %s", err)
		} else {
			c.log.INFO.Println("Relogin successful")
		}
		return err
	}
	c.tokenRes = tokenRes
	return nil
}

func (c *Connection) getHomes() (string, error) {
	urlGetHomes := API_URL_BASE + "/homes" // hier muss ggf. noch was ergänzt werden
	req, err := http.NewRequest("GET", urlGetHomes, nil)
	if err != nil {
		err = fmt.Errorf("error getting homes list: %s", err)
		return "", err
	}
	req.Header = c.getSensonetHttpHeader()
	var res HomesStruct
	err = c.DoJSON(req, &res)
	if err != nil {
		err = fmt.Errorf("error getting homes list: %s", err)
		return "", err
	}
	systemId := res[0].SystemID
	return systemId, err
}

func (c *Connection) getSystem(relData *Vr921RelevantDataStruct) error {
	//controlIdentifier := c.getControlIdentifier()
	systemUrl := API_URL_BASE + fmt.Sprintf("/systems/%s/tli", c.systemId)
	req, err := http.NewRequest("GET", systemUrl, nil)
	if err != nil {
		err = fmt.Errorf("error perparing request for systems: %s", err)
		return err
	}
	req.Header = c.getSensonetHttpHeader()
	var system SystemStruct
	err = c.DoJSON(req, &system)
	if err != nil {
		c.log.ERROR.Println("Error getting system. Error:", err)
		c.log.INFO.Println("Trying to refresh token")
		err = c.refreshToken()
		if err != nil {
			err = fmt.Errorf("could not refresh token. error: %s", err)
			return err
		}
		//d.log.DEBUG.Println("Refresh token successful")
		c.tokenExpiresAt = time.Now().Add(time.Duration(c.tokenRes.ExpiresIn * int64(time.Second)))
		c.log.DEBUG.Printf("Refreshed token expires at: %02d:%02d:%02d", c.tokenExpiresAt.Hour(), c.tokenExpiresAt.Minute(), c.tokenExpiresAt.Second())
		req.Header = c.getSensonetHttpHeader()
		err = c.DoJSON(req, &system)
		if err != nil {
			err = fmt.Errorf("error getting systems: %s", err)
			return err
		}
	}

	relData.Timestamp = time.Now().Unix()
	relData.Status.OutsideTemperature = system.State.System.OutdoorTemperature
	relData.Status.SystemFlowTemperature = system.State.System.SystemFlowTemperature
	relData.Hotwater.HotwaterTemperatureSetpoint = system.Configuration.Dhw[0].TappingSetpoint
	relData.Hotwater.HotwaterLiveTemperature = system.State.Dhw[0].CurrentDhwTemperature
	relData.Hotwater.OperationMode = system.Configuration.Dhw[0].OperationModeDhw
	relData.Hotwater.Index = system.Configuration.Dhw[0].Index
	relData.Hotwater.CurrentQuickmode = system.State.Dhw[0].CurrentSpecialFunction
	/*if system.State.Dhw[0].CurrentSpecialFunction == "CYLINDER_BOOST" && (c.currentQuickmode != "") {
		c.currentQuickmode = QUICKMODE_HOTWATER
		c.quickmodeStarted = time.Now()
		c.onoff = true
	}*/

	//Extract information for all zones
	//c.quickVetoSetPoint = 0.0 // Reset c.quickVetoSetPoint
	if len(relData.Zones) == 0 {
		relData.Zones = make([]Vr921RelevantDataZonesStruct, 0)
	}
	for i, systemStateZone := range system.State.Zones {
		if len(relData.Zones) <= i {
			//If relData.Zones array is not big enough, new elements are appended, especially at first ExtractSystem call
			//At the moment, relData.Zones is not shortened, if later GetSystem calls returns less system.Body.Zones
			zone := Vr921RelevantDataZonesStruct{}
			relData.Zones = append(relData.Zones, zone)
		}
		// Looking for the matching system configuration zone
		found := -1
		for j, systemConfigurationZone := range system.Configuration.Zones {
			if systemStateZone.Index == systemConfigurationZone.Index {
				found = j
				break
			}
		}
		if found < 0 {
			c.log.ERROR.Println("System.State.Zones[] und System.Configuration.Zones[] do not match")
			return err
		}
		systemConfigurationZone := system.Configuration.Zones[found]

		// Looking for the matching system state circuit
		found = -1
		for j, systemStateCircuit := range system.State.Circuits {
			if systemStateZone.Index == systemStateCircuit.Index {
				found = j
				break
			}
		}
		if found < 0 {
			c.log.ERROR.Println("System.State.Zones[] und System.State.Circuits[] do not match")
			return err
		}
		systemStateCircuit := system.State.Circuits[found]

		relData.Zones[i].Name = systemConfigurationZone.General.Name
		relData.Zones[i].Index = systemStateZone.Index
		//relData.Zones[i].ActiveFunction = systemStateZone.Configuration.ActiveFunction
		//relData.Zones[i].Enabled = systemStateZone.Configuration.Enabled
		relData.Zones[i].OperationMode = systemConfigurationZone.Heating.OperationModeHeating
		relData.Zones[i].CurrentDesiredSetpoint = systemStateZone.DesiredRoomTemperatureSetpoint
		relData.Zones[i].CurrentQuickmode = systemStateZone.CurrentSpecialFunction
		if systemStateZone.CurrentSpecialFunction == "QUICK_VETO" {
			//relData.Zones[i].QuickVeto.ExpiresAt = systemStateZone.Configuration.QuickVeto.ExpiresAt
			relData.Zones[i].QuickVeto.TemperatureSetpoint = systemStateZone.DesiredRoomTemperatureSetpoint
			c.quickVetoSetPoint = float32(systemStateZone.DesiredRoomTemperatureSetpoint)
		} else {
			relData.Zones[i].QuickVeto.ExpiresAt = ""
			relData.Zones[i].QuickVeto.TemperatureSetpoint = 0
		}
		relData.Zones[i].InsideTemperature = systemStateZone.CurrentRoomTemperature
		relData.Zones[i].CurrentCircuitFlowTemperature = systemStateCircuit.CurrentCircuitFlowTemperature
		/*if (relData.Zones[i].CurrentQuickmode != "NONE") && (c.currentQuickmode != "") {
			if c.currentQuickmode != QUICKMODE_HEATING {
				c.currentQuickmode = QUICKMODE_HEATING
				c.quickmodeStarted = time.Now()
				c.onoff = true
			}
		}*/
	}
	//Added by WW: This block is used during development to analyse the system report return from the Vaillant portal
	level := util.WWlogLevelForArea("sensonet").String()
	if level == "DEBUG" || level == "TRACE" {
		c.log.DEBUG.Println("Writing debug information to files debug_sensonet_system.txt and debug_sensonet_reldata.txt")
		fo, ioerr := os.Create("debug_sensonet_system.txt")
		if ioerr != nil {
			c.log.ERROR.Println("Error creating debug_sensonet_system.txt. Error:", ioerr)
			return err
		} else {
			bytes, _ := json.MarshalIndent(system, "", "  ")
			_, ioerr = fo.Write(bytes)
			if ioerr != nil {
				c.log.ERROR.Println("Error writing in debug_sensonet_system.txt. Error:", ioerr)
				return err
			}
			fo.Close()
		}
		fo, ioerr = os.Create("debug_sensonet_reldata.txt")
		if ioerr != nil {
			c.log.ERROR.Println("Error creating debug_sensonet_reldata.txt. Error:", ioerr)
			return err
		} else {
			bytes, _ := json.MarshalIndent(relData, "", "  ")
			_, ioerr = fo.Write(bytes)
			if ioerr != nil {
				c.log.ERROR.Println("Error writing in debug_sensonet_reldata.txt. Error:", ioerr)
				return err
			}
			fo.Close()
		}
	}
	//Added by WW: End of block

	c.log.INFO.Println("New system information read from myVaillant portal.")
	return err
}

func (d *Connection) Phases() int {
	return d.phases
}

func (d *Connection) CurrentQuickmode() string {
	return d.currentQuickmode
}

func (d *Connection) QuickVetoExpiresAt() string {
	return d.quickVetoExpiresAt
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
			if z.Index == d.heatingZone && z.InsideTemperature != 0.0 {
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
			if z.Index == d.heatingZone {
				return float64(z.QuickVeto.TemperatureSetpoint), nil
			}
		}
		return float64(d.quickVetoSetPoint), nil
	}
	return float64(res.Hotwater.HotwaterTemperatureSetpoint), nil
}

// CheckPVUseStrategy is called bei vaillant-ebus_vehicle.Soc()
func (d *Connection) CheckPVUseStrategy(vehicleStrategy string) error {
	if d.pvUseStrategy != vehicleStrategy {
		d.log.INFO.Printf("Changing PVUseStrategy of charger from '%s' to '%s'", d.pvUseStrategy, vehicleStrategy)
		d.pvUseStrategy = vehicleStrategy
	}
	return nil
}

func (d *Connection) Status() (api.ChargeStatus, error) {
	status := api.StatusB
	if time.Now().After(d.tokenExpiresAt) {
		status = api.StatusA // disconnected
	}
	if d.CurrentQuickmode() != "" {
		status = api.StatusC
	}
	return status, nil
}

// This function checks the operation mode of heating and hotwater and the hotwater live temperature
// and returns, which quick mode should be started, when evcc sends an "Enable"
func (c *Connection) WhichQuickMode() (int, error) {
	res, err := c.statusCache.Get()
	if err != nil {
		err = fmt.Errorf("could not read status cache before hotwater boost: %s", err)
		return 0, err
	}
	//c.log.DEBUG.Println("PV Use Strategy = ", c.pvUseStrategy)
	c.log.DEBUG.Printf("Checking if hot water boost possible. Operation Mode = %s, temperature setpoint= %02.2f, live temperature= %02.2f", res.Hotwater.OperationMode, res.Hotwater.HotwaterTemperatureSetpoint, res.Hotwater.HotwaterLiveTemperature)
	hotWaterBoostPossible := false
	// For pvUseStrategy='hotwater', a hotwater boost is possible when hotwater storage temperature is less than the temperature setpoint.
	// For other pvUseStrategies, a hotwater boost is possible when hotwater storage temperature is less than the temperature setpoint minus 5°C
	addOn := -5.0
	if c.pvUseStrategy == PVUSESTRATEGY_HOTWATER {
		addOn = 0.0
	}
	if res.Hotwater.HotwaterLiveTemperature < res.Hotwater.HotwaterTemperatureSetpoint+addOn &&
		res.Hotwater.OperationMode == OPERATIONMODE_TIME_CONTROLLED {
		hotWaterBoostPossible = true
	}
	heatingQuickVetoPossible := false
	for _, z := range res.Zones {
		if z.Index == c.heatingZone {
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
