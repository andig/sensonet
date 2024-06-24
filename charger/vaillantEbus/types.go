package vaillantEbus

const PVUSESTRATEGY_HOTWATER_THEN_HEATING string = "hotwater_then_heating"
const PVUSESTRATEGY_HOTWATER string = "hotwater"
const PVUSESTRATEGY_HEATING string = "heating"
const OPERATIONMODE_AUTO string = "auto"
const QUICKMODE_HOTWATER string = "Hotwater Boost"
const QUICKMODE_HEATING string = "Heating Quick Veto"
const QUICKMODE_NOTHING string = "Charger running idle"

const EBUSDREAD_STATUS_TIME = "vdatetime"
const EBUSDREAD_STATUS_OUTSIDETEMPERATURE = "outsidetemp"
const EBUSDREAD_STATUS_SYSTEMFLOWTEMPERATUE = "SystemFlowTemp"
const EBUSDREAD_STATUS_WATERPRESSURE = "WaterPressure"
const EBUSDREAD_STATUS_CURRENTCONSUMEDPOWER = "CurrentConsumedPower"
const EBUSDREAD_STATUS_STATUS01 = "Status01"
const EBUSDREAD_STATUS_STATE = "State"
const EBUSDREAD_HOTWATER_OPMODE = "HwcOpMode"
const EBUSDREAD_HOTWATER_TEMPDESIRED = "HwcTempDesired"
const EBUSDREAD_HOTWATER_STORAGETEMP = "HwcStorageTemp"
const EBUSDREAD_HOTWATER_SFMODE = "HwcSFMode"
const EBUSDREAD_ZONE_NAME = "Shortname"                              //To be added by the zone prefix
const EBUSDREAD_ZONE_ACTUALROOMTEMPDESIRED = "ActualRoomTempDesired" //To be added by the zone prefix
const EBUSDREAD_ZONE_OPMODE = "OpMode"                               //To be added by the zone prefix
const EBUSDREAD_ZONE_SFMODE = "SFMode"                               //To be added by the zone prefix
const EBUSDREAD_ZONE_ROOMTEMP = "RoomTemp"                           //To be added by the zone prefix
const EBUSDREAD_ZONE_QUICKVETOTEMP = "QuickVetoTemp"                 //To be added by the zone prefix
const EBUSDREAD_ZONE_QUICKVETOENDDATE = "QVEndDate"                  //To be added by the zone prefix
const EBUSDREAD_ZONE_QUICKVETOENDTIME = "QVEndTime"                  //To be added by the zone prefix
const EBUSDREAD_ZONE_QUICKVETODURATION = "QVDuration"                //To be added by the zone prefix

//To be deleted
//const AUTH_BASE_URL string = "https://identity.vaillant-group.com/auth/realms"
//const LOGIN_URL string = AUTH_BASE_URL + "/%s/login-actions/authenticate"
//const TOKEN_URL string = AUTH_BASE_URL + "/%s/protocol/openid-connect/token"
//const AUTH_URL string = AUTH_BASE_URL + "/%s/protocol/openid-connect/auth"
//const API_URL_BASE string = "https://api.vaillant-group.com/service-connected-control/end-user-app-api/v1"

//const CLIENT_ID string = "myvaillant"

//const HOTWATERBOOST_URL string = "/systems/%s/tli/domestic-hot-water/%01d/boost"
//const ZONEQUICKVETO_URL string = "/systems/%s/tli/zones/%01d/quick-veto"

// Types fpr Vaillant data

type VaillantRelDataZonesStruct struct {
	Index                 int
	Name                  string
	ActualRoomTempDesired float64
	OpMode                string
	SFMode                string
	QuickVetoTemp         float64
	QuickVetoEndTime      string
	QuickVetoEndDate      string
	InsideTemperature     float64
	RoomTemp              float64
}

type VaillantRelDataHeatCircuitsStruct struct {
	Index                 int
	ActualFlowTempDesired float64
	FlowTemp              float64
	Status                string
}

/*type VaillantRelDataScansStruct struct {
	Index int
	id    string
}*/

type VaillantRelDataStruct struct {
	//SerialNumber string
	//Timestamp    int64
	//PvMode          int64
	//PvModeActive    int64
	//PvModeTimestamp int64

	Status struct {
		Time                  string
		SensorData1           string
		SensorData2           string
		OutsideTemperature    float64
		SystemFlowTemperature float64
		WaterPressure         float64
		ControllerForSFMode   string
		CurrentConsumedPower  float64
		Status01              string
		State                 string
	}

	Hotwater struct {
		Index          int // necessary?
		HwcTempDesired float64
		HwcOpMode      string
		HwcStorageTemp float64
		HwcSFMode      string
	}

	Zones []VaillantRelDataZonesStruct

	HeatCircuits []VaillantRelDataHeatCircuitsStruct

	//	Scans []VaillantRelDataScansStruct
}
