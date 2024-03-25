package sensonet

import "time"

const PVUSESTRATEGY_HOTWATER_THEN_HEATING string = "hotwater_then_heating"
const PVUSESTRATEGY_HOTWATER string = "hotwater"
const PVUSESTRATEGY_HEATING string = "heating"
const OPERATIONMODE_TIME_CONTROLLED string = "TIME_CONTROLLED"
const QUICKMODE_HOTWATER string = "Hotwater Boost"
const QUICKMODE_HEATING string = "Heating Quick Veto"

const AUTH_BASE_URL string = "https://identity.vaillant-group.com/auth/realms"
const LOGIN_URL string = AUTH_BASE_URL + "/%s/login-actions/authenticate"
const TOKEN_URL string = AUTH_BASE_URL + "/%s/protocol/openid-connect/token"
const AUTH_URL string = AUTH_BASE_URL + "/%s/protocol/openid-connect/auth"
const API_URL_BASE string = "https://api.vaillant-group.com/service-connected-control/end-user-app-api/v1"

// const SYSTEM_URL string = "/systemcontrol/tli/v1"
// const FACILITIES_URL string = "not to be used"
// const LIVEREPORT_URL string = "/livereport/v1"
const CLIENT_ID string = "myvaillant"

const HOTWATERBOOST_URL string = "/systems/%s/tli/domestic-hot-water/%01d/boost"
const ZONEQUICKVETO_URL string = "/systems/%s/tli/zones/%01d/quick-veto"

//zone_quick_veto_url= API_URL_BASE+'/systems/9d2cc41b-12fd-47cc-b351-35d7e11fb1ab'+'/tli/zones/%01d/quick-veto'

type TokenRequestStruct struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

// VR921 Types

type Vr921RelevantDataZonesStruct struct {
	Index int
	Name  string
	//ActiveFunction         string
	//Enabled                bool
	CurrentDesiredSetpoint float64
	OperationMode          string
	CurrentQuickmode       string
	QuickVeto              struct {
		ExpiresAt           string
		TemperatureSetpoint float64
	}
	InsideTemperature             float64
	CurrentCircuitFlowTemperature float64
}

type Vr921RelevantDataStruct struct {
	SerialNumber string
	Timestamp    int64
	//PvMode          int64
	//PvModeActive    int64
	//PvModeTimestamp int64

	Status struct {
		Datetime              string
		OutsideTemperature    float64
		SystemFlowTemperature float64
	}

	Hotwater struct {
		Index                       int
		HotwaterTemperatureSetpoint float64
		OperationMode               string
		HotwaterLiveTemperature     float64
		CurrentQuickmode            string
	}

	Zones []Vr921RelevantDataZonesStruct
}

type HomesStruct []struct {
	HomeName string `json:"homeName"`
	Address  struct {
		Street      string `json:"street"`
		Extension   any    `json:"extension"`
		City        string `json:"city"`
		PostalCode  string `json:"postalCode"`
		CountryCode string `json:"countryCode"`
	} `json:"address"`
	SerialNumber    string `json:"serialNumber"`
	SystemID        string `json:"systemId"`
	ProductMetadata struct {
		ProductType    string `json:"productType"`
		ProductionYear string `json:"productionYear"`
		ProductionWeek string `json:"productionWeek"`
		ArticleNumber  string `json:"articleNumber"`
	} `json:"productMetadata"`
	State               string    `json:"state"`
	MigrationState      string    `json:"migrationState"`
	MigrationFinishedAt time.Time `json:"migrationFinishedAt"`
	OnlineState         string    `json:"onlineState"`
	Firmware            struct {
		Version        string `json:"version"`
		UpdateEnabled  bool   `json:"updateEnabled"`
		UpdateRequired bool   `json:"updateRequired"`
	} `json:"firmware"`
	Nomenclature       string `json:"nomenclature"`
	Cag                bool   `json:"cag"`
	CountryCode        string `json:"countryCode"`
	ProductInformation string `json:"productInformation"`
	FirmwareVersion    string `json:"firmwareVersion"`
}

type SystemStruct struct {
	State struct {
		System struct {
			OutdoorTemperature           float64 `json:"outdoorTemperature"`
			OutdoorTemperatureAverage24H float64 `json:"outdoorTemperatureAverage24h"`
			SystemFlowTemperature        float64 `json:"systemFlowTemperature"`
			SystemWaterPressure          float64 `json:"systemWaterPressure"`
			EnergyManagerState           string  `json:"energyManagerState"`
			SystemOff                    bool    `json:"systemOff"`
		} `json:"system"`
		Zones []struct {
			Index                                 int     `json:"index"`
			DesiredRoomTemperatureSetpointHeating float64 `json:"desiredRoomTemperatureSetpointHeating"`
			DesiredRoomTemperatureSetpoint        float64 `json:"desiredRoomTemperatureSetpoint"`
			CurrentRoomTemperature                float64 `json:"currentRoomTemperature,omitempty"`
			CurrentRoomHumidity                   float64 `json:"currentRoomHumidity,omitempty"`
			CurrentSpecialFunction                string  `json:"currentSpecialFunction"`
			HeatingState                          string  `json:"heatingState"`
		} `json:"zones"`
		Circuits []struct {
			Index                         int     `json:"index"`
			CircuitState                  string  `json:"circuitState"`
			CurrentCircuitFlowTemperature float64 `json:"currentCircuitFlowTemperature,omitempty"`
			HeatingCircuitFlowSetpoint    float64 `json:"heatingCircuitFlowSetpoint"`
			CalculatedEnergyManagerState  string  `json:"calculatedEnergyManagerState"`
		} `json:"circuits"`
		Dhw []struct {
			Index                  int     `json:"index"`
			CurrentSpecialFunction string  `json:"currentSpecialFunction"`
			CurrentDhwTemperature  float64 `json:"currentDhwTemperature"`
		} `json:"dhw"`
	} `json:"state"`
	Properties struct {
		System struct {
			ControllerType                     string  `json:"controllerType"`
			SystemScheme                       int     `json:"systemScheme"`
			BackupHeaterType                   string  `json:"backupHeaterType"`
			BackupHeaterAllowedFor             string  `json:"backupHeaterAllowedFor"`
			ModuleConfigurationVR71            int     `json:"moduleConfigurationVR71"`
			EnergyProvidePowerCutBehavior      string  `json:"energyProvidePowerCutBehavior"`
			SmartPhotovoltaicBufferOffset      float64 `json:"smartPhotovoltaicBufferOffset"`
			ExternalEnergyManagementActivation bool    `json:"externalEnergyManagementActivation"`
		} `json:"system"`
		Zones []struct {
			Index                  int    `json:"index"`
			IsActive               bool   `json:"isActive"`
			ZoneBinding            string `json:"zoneBinding"`
			IsCoolingAllowed       bool   `json:"isCoolingAllowed"`
			AssociatedCircuitIndex int    `json:"associatedCircuitIndex"`
		} `json:"zones"`
		Circuits []struct {
			Index                    int    `json:"index"`
			MixerCircuitTypeExternal string `json:"mixerCircuitTypeExternal"`
			HeatingCircuitType       string `json:"heatingCircuitType"`
		} `json:"circuits"`
		Dhw []struct {
			Index       int     `json:"index"`
			MinSetpoint float64 `json:"minSetpoint"`
			MaxSetpoint float64 `json:"maxSetpoint"`
		} `json:"dhw"`
	} `json:"properties"`
	Configuration struct {
		System struct {
			ContinuousHeatingStartSetpoint float64 `json:"continuousHeatingStartSetpoint"`
			AlternativePoint               float64 `json:"alternativePoint"`
			HeatingCircuitBivalencePoint   float64 `json:"heatingCircuitBivalencePoint"`
			DhwBivalencePoint              float64 `json:"dhwBivalencePoint"`
			AdaptiveHeatingCurve           bool    `json:"adaptiveHeatingCurve"`
			DhwMaximumLoadingTime          int     `json:"dhwMaximumLoadingTime"`
			DhwHysteresis                  float64 `json:"dhwHysteresis"`
			DhwFlowSetpointOffset          float64 `json:"dhwFlowSetpointOffset"`
			ContinuousHeatingRoomSetpoint  float64 `json:"continuousHeatingRoomSetpoint"`
			HybridControlStrategy          string  `json:"hybridControlStrategy"`
			MaxFlowSetpointHpError         float64 `json:"maxFlowSetpointHpError"`
			DhwMaximumTemperature          float64 `json:"dhwMaximumTemperature"`
			MaximumPreheatingTime          int     `json:"maximumPreheatingTime"`
			ParalellTankLoadingAllowed     bool    `json:"paralellTankLoadingAllowed"`
		} `json:"system"`
		Zones []struct {
			Index   int `json:"index"`
			General struct {
				Name                 string    `json:"name"`
				HolidayStartDateTime time.Time `json:"holidayStartDateTime"`
				HolidayEndDateTime   time.Time `json:"holidayEndDateTime"`
				HolidaySetpoint      float64   `json:"holidaySetpoint"`
			} `json:"general"`
			Heating struct {
				OperationModeHeating      string  `json:"operationModeHeating"`
				SetBackTemperature        float64 `json:"setBackTemperature"`
				ManualModeSetpointHeating float64 `json:"manualModeSetpointHeating"`
				TimeProgramHeating        struct {
					MetaInfo struct {
						MinSlotsPerDay          int  `json:"minSlotsPerDay"`
						MaxSlotsPerDay          int  `json:"maxSlotsPerDay"`
						SetpointRequiredPerSlot bool `json:"setpointRequiredPerSlot"`
					} `json:"metaInfo"`
					Monday []struct {
						StartTime int     `json:"startTime"`
						EndTime   int     `json:"endTime"`
						Setpoint  float64 `json:"setpoint"`
					} `json:"monday"`
					Tuesday []struct {
						StartTime int     `json:"startTime"`
						EndTime   int     `json:"endTime"`
						Setpoint  float64 `json:"setpoint"`
					} `json:"tuesday"`
					Wednesday []struct {
						StartTime int     `json:"startTime"`
						EndTime   int     `json:"endTime"`
						Setpoint  float64 `json:"setpoint"`
					} `json:"wednesday"`
					Thursday []struct {
						StartTime int     `json:"startTime"`
						EndTime   int     `json:"endTime"`
						Setpoint  float64 `json:"setpoint"`
					} `json:"thursday"`
					Friday []struct {
						StartTime int     `json:"startTime"`
						EndTime   int     `json:"endTime"`
						Setpoint  float64 `json:"setpoint"`
					} `json:"friday"`
					Saturday []struct {
						StartTime int     `json:"startTime"`
						EndTime   int     `json:"endTime"`
						Setpoint  float64 `json:"setpoint"`
					} `json:"saturday"`
					Sunday []struct {
						StartTime int     `json:"startTime"`
						EndTime   int     `json:"endTime"`
						Setpoint  float64 `json:"setpoint"`
					} `json:"sunday"`
				} `json:"timeProgramHeating"`
			} `json:"heating"`
		} `json:"zones"`
		Circuits []struct {
			Index                                  int     `json:"index"`
			HeatingCurve                           float64 `json:"heatingCurve"`
			HeatingFlowTemperatureMinimumSetpoint  float64 `json:"heatingFlowTemperatureMinimumSetpoint"`
			HeatingFlowTemperatureMaximumSetpoint  float64 `json:"heatingFlowTemperatureMaximumSetpoint"`
			HeatDemandLimitedByOutsideTemperature  float64 `json:"heatDemandLimitedByOutsideTemperature"`
			HeatingCircuitFlowSetpointExcessOffset float64 `json:"heatingCircuitFlowSetpointExcessOffset"`
			SetBackModeEnabled                     bool    `json:"setBackModeEnabled"`
			RoomTemperatureControlMode             string  `json:"roomTemperatureControlMode"`
		} `json:"circuits"`
		Dhw []struct {
			Index                int       `json:"index"`
			OperationModeDhw     string    `json:"operationModeDhw"`
			TappingSetpoint      float64   `json:"tappingSetpoint"`
			HolidayStartDateTime time.Time `json:"holidayStartDateTime"`
			HolidayEndDateTime   time.Time `json:"holidayEndDateTime"`
			TimeProgramDhw       struct {
				MetaInfo struct {
					MinSlotsPerDay          int  `json:"minSlotsPerDay"`
					MaxSlotsPerDay          int  `json:"maxSlotsPerDay"`
					SetpointRequiredPerSlot bool `json:"setpointRequiredPerSlot"`
				} `json:"metaInfo"`
				Monday []struct {
					StartTime int `json:"startTime"`
					EndTime   int `json:"endTime"`
				} `json:"monday"`
				Tuesday []struct {
					StartTime int `json:"startTime"`
					EndTime   int `json:"endTime"`
				} `json:"tuesday"`
				Wednesday []struct {
					StartTime int `json:"startTime"`
					EndTime   int `json:"endTime"`
				} `json:"wednesday"`
				Thursday []struct {
					StartTime int `json:"startTime"`
					EndTime   int `json:"endTime"`
				} `json:"thursday"`
				Friday []struct {
					StartTime int `json:"startTime"`
					EndTime   int `json:"endTime"`
				} `json:"friday"`
				Saturday []struct {
					StartTime int `json:"startTime"`
					EndTime   int `json:"endTime"`
				} `json:"saturday"`
				Sunday []struct {
					StartTime int `json:"startTime"`
					EndTime   int `json:"endTime"`
				} `json:"sunday"`
			} `json:"timeProgramDhw"`
			TimeProgramCirculationPump struct {
				MetaInfo struct {
					MinSlotsPerDay          int  `json:"minSlotsPerDay"`
					MaxSlotsPerDay          int  `json:"maxSlotsPerDay"`
					SetpointRequiredPerSlot bool `json:"setpointRequiredPerSlot"`
				} `json:"metaInfo"`
				Monday []struct {
					StartTime int `json:"startTime"`
					EndTime   int `json:"endTime"`
				} `json:"monday"`
				Tuesday []struct {
					StartTime int `json:"startTime"`
					EndTime   int `json:"endTime"`
				} `json:"tuesday"`
				Wednesday []struct {
					StartTime int `json:"startTime"`
					EndTime   int `json:"endTime"`
				} `json:"wednesday"`
				Thursday []struct {
					StartTime int `json:"startTime"`
					EndTime   int `json:"endTime"`
				} `json:"thursday"`
				Friday []struct {
					StartTime int `json:"startTime"`
					EndTime   int `json:"endTime"`
				} `json:"friday"`
				Saturday []struct {
					StartTime int `json:"startTime"`
					EndTime   int `json:"endTime"`
				} `json:"saturday"`
				Sunday []struct {
					StartTime int `json:"startTime"`
					EndTime   int `json:"endTime"`
				} `json:"sunday"`
			} `json:"timeProgramCirculationPump"`
		} `json:"dhw"`
	} `json:"configuration"`
}
