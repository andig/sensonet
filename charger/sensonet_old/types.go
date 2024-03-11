package sensonet_old

const PVUSESTRATEGY_HOTWATER_THEN_HEATING string = "hotwater_then_heating"
const PVUSESTRATEGY_HOTWATER string = "hotwater"
const PVUSESTRATEGY_HEATING string = "heating"
const OPERATIONMODE_TIME_CONTROLLED string = "TIME_CONTROLLED"
const QUICKMODE_HOTWATER string = "Hotwater Boost"
const QUICKMODE_HEATING string = "Heating Quick Veto"

const TOKEN_URL string = "https://smart.vaillant.com/mobile/api/v4/account/authentication/v1/token/new"
const AUTH_URL string = "https://smart.vaillant.com/mobile/api/v4/account/authentication/v1/authenticate"
const FACILITIES_URL string = "https://smart.vaillant.com/mobile/api/v4/facilities"
const SYSTEM_URL string = "/systemcontrol/tli/v1"
const LIVEREPORT_URL string = "/livereport/v1"
const HOTWATERBOOST_URL string = "/systemcontrol/tli/v1/dhw/configuration/hotwater_boost"
const ZONEQUICKVETO_URL string = "/systemcontrol/tli/v1/zones/Control_ZO%01d/configuration/quick_veto"

// VR921 Types

type Vr921RelevantDataZonesStruct struct {
	ID                     string
	Name                   string
	ActiveFunction         string
	Enabled                bool
	CurrentDesiredSetpoint float64
	OperationMode          string
	CurrentQuickmode       string
	QuickVeto              struct {
		ExpiresAt           string
		TemperatureSetpoint float64
	}
	InsideTemperature float64
}

type Vr921RelevantDataStruct struct {
	SerialNumber string
	Timestamp    int64
	//PvMode          int64
	//PvModeActive    int64
	//PvModeTimestamp int64

	Status struct {
		Datetime           string
		OutsideTemperature float64
	}

	Hotwater struct {
		HotwaterTemperatureSetpoint float64
		OperationMode               string
		HotwaterSystemState         string // to store the meta info state for system
		HotwaterSystemTimestamp     int64  // to store the meta info timestamp for system
		HotwaterLiveTemperature     float64
		HotwaterLiveState           string // to store the meta info state for live report
		HotwaterLiveTimestamp       int64  // to store the meta info timestamp for live report
		CurrentQuickmode            string
	}

	Zones []Vr921RelevantDataZonesStruct
}

type FacilitiesListStruct struct {
	Body struct {
		FacilitiesList []struct {
			Capabilities       []string `json:"capabilities"`
			FirmwareVersion    string   `json:"firmwareVersion"`
			Name               string   `json:"name"`
			NetworkInformation struct {
				MacAddressEthernet        string `json:"macAddressEthernet"`
				MacAddressWifiAccessPoint string `json:"macAddressWifiAccessPoint"`
				MacAddressWifiClient      string `json:"macAddressWifiClient"`
			} `json:"networkInformation"`
			ResponsibleCountryCode string `json:"responsibleCountryCode"`
			SerialNumber           string `json:"serialNumber"`
			SupportedBrand         string `json:"supportedBrand"`
		} `json:"facilitiesList"`
	} `json:"body"`
	Meta struct{} `json:"meta"`
}

type SystemStruct struct {
	Body struct {
		Configuration struct {
			ManualCooling struct {
				EndDate   string `json:"end_date"`
				StartDate string `json:"start_date"`
			} `json:"manual_cooling"`
		} `json:"configuration"`
		Dhw struct {
			Circulation struct {
				Configuration struct {
					OperationMode string `json:"operation_mode"`
				} `json:"configuration"`
				Timeprogram struct {
					Friday []struct {
						EndTime   string `json:"end_time"`
						StartTime string `json:"start_time"`
					} `json:"friday"`
					Monday []struct {
						EndTime   string `json:"end_time"`
						StartTime string `json:"start_time"`
					} `json:"monday"`
					Saturday []struct {
						EndTime   string `json:"end_time"`
						StartTime string `json:"start_time"`
					} `json:"saturday"`
					Sunday []struct {
						EndTime   string `json:"end_time"`
						StartTime string `json:"start_time"`
					} `json:"sunday"`
					Thursday []struct {
						EndTime   string `json:"end_time"`
						StartTime string `json:"start_time"`
					} `json:"thursday"`
					Tuesday []struct {
						EndTime   string `json:"end_time"`
						StartTime string `json:"start_time"`
					} `json:"tuesday"`
					Wednesday []struct {
						EndTime   string `json:"end_time"`
						StartTime string `json:"start_time"`
					} `json:"wednesday"`
				} `json:"timeprogram"`
			} `json:"circulation"`
			Configuration struct {
				CurrentQuickmode string `json:"current_quickmode"`
				Away             struct {
					EndDatetime   string `json:"end_datetime"`
					StartDatetime string `json:"start_datetime"`
				} `json:"away"`
			} `json:"configuration"`
			Hotwater struct {
				Configuration struct {
					HotwaterTemperatureSetpoint float64 `json:"hotwater_temperature_setpoint"`
					OperationMode               string  `json:"operation_mode"`
				} `json:"configuration"`
				Timeprogram struct {
					Friday []struct {
						EndTime   string `json:"end_time"`
						StartTime string `json:"start_time"`
					} `json:"friday"`
					Monday []struct {
						EndTime   string `json:"end_time"`
						StartTime string `json:"start_time"`
					} `json:"monday"`
					Saturday []struct {
						EndTime   string `json:"end_time"`
						StartTime string `json:"start_time"`
					} `json:"saturday"`
					Sunday []struct {
						EndTime   string `json:"end_time"`
						StartTime string `json:"start_time"`
					} `json:"sunday"`
					Thursday []struct {
						EndTime   string `json:"end_time"`
						StartTime string `json:"start_time"`
					} `json:"thursday"`
					Tuesday []struct {
						EndTime   string `json:"end_time"`
						StartTime string `json:"start_time"`
					} `json:"tuesday"`
					Wednesday []struct {
						EndTime   string `json:"end_time"`
						StartTime string `json:"start_time"`
					} `json:"wednesday"`
				} `json:"timeprogram"`
			} `json:"hotwater"`
		} `json:"dhw"`
		Status struct {
			Datetime           string  `json:"datetime"`
			OutsideTemperature float64 `json:"outside_temperature"`
		} `json:"status"`
		Zones []struct {
			ID            string `json:"_id"`
			Configuration struct {
				ActiveFunction string `json:"active_function"`
				Away           struct {
					EndDatetime         string  `json:"end_datetime"`
					StartDatetime       string  `json:"start_datetime"`
					TemperatureSetpoint float64 `json:"temperature_setpoint"`
				} `json:"away"`
				CurrentDesiredSetpoint float64 `json:"current_desired_setpoint"`
				CurrentQuickmode       string  `json:"current_quickmode"`
				QuickVeto              struct {
					ExpiresAt           string  `json:"expires_at"`
					TemperatureSetpoint float64 `json:"temperature_setpoint"`
				} `json:"quick_veto"`
				EcoMode             bool    `json:"eco_mode"`
				Enabled             bool    `json:"enabled"`
				Humidity            float64 `json:"humidity"`
				InsideTemperature   float64 `json:"inside_temperature"`
				ManualCoolingActive bool    `json:"manual_cooling_active"`
				Name                string  `json:"name"`
			} `json:"configuration"`
			Heating struct {
				Configuration struct {
					ManualModeTemperatureSetpoint float64 `json:"manual_mode_temperature_setpoint"`
					OperationMode                 string  `json:"operation_mode"`
					SetbackTemperatureSetpoint    float64 `json:"setback_temperature_setpoint"`
				} `json:"configuration"`
				Timeprogram struct {
					Friday []struct {
						EndTime   string  `json:"end_time"`
						Setpoint  float64 `json:"setpoint"`
						StartTime string  `json:"start_time"`
					} `json:"friday"`
					Monday []struct {
						EndTime   string  `json:"end_time"`
						Setpoint  float64 `json:"setpoint"`
						StartTime string  `json:"start_time"`
					} `json:"monday"`
					Saturday []struct {
						EndTime   string  `json:"end_time"`
						Setpoint  float64 `json:"setpoint"`
						StartTime string  `json:"start_time"`
					} `json:"saturday"`
					Sunday []struct {
						EndTime   string  `json:"end_time"`
						Setpoint  float64 `json:"setpoint"`
						StartTime string  `json:"start_time"`
					} `json:"sunday"`
					Thursday []struct {
						EndTime   string  `json:"end_time"`
						Setpoint  float64 `json:"setpoint"`
						StartTime string  `json:"start_time"`
					} `json:"thursday"`
					Tuesday []struct {
						EndTime   string  `json:"end_time"`
						Setpoint  float64 `json:"setpoint"`
						StartTime string  `json:"start_time"`
					} `json:"tuesday"`
					Wednesday []struct {
						EndTime   string  `json:"end_time"`
						Setpoint  float64 `json:"setpoint"`
						StartTime string  `json:"start_time"`
					} `json:"wednesday"`
				} `json:"timeprogram"`
			} `json:"heating"`
		} `json:"zones"`
	} `json:"body"`
	Meta struct {
		ResourceState []struct {
			Link struct {
				Rel          string `json:"rel"`
				ResourceLink string `json:"resourceLink"`
			} `json:"link"`
			State     string `json:"state"`
			Timestamp int64  `json:"timestamp"`
		} `json:"resourceState"`
	} `json:"meta"`
}

type LiveReportStruct struct {
	Body struct {
		Devices []struct {
			ID      string `json:"_id"`
			Name    string `json:"name"`
			Reports []struct {
				ID                       string  `json:"_id"`
				AssociatedDeviceFunction string  `json:"associated_device_function"`
				MeasurementCategory      string  `json:"measurement_category"`
				Name                     string  `json:"name"`
				Unit                     string  `json:"unit"`
				Value                    float64 `json:"value"`
			} `json:"reports"`
		} `json:"devices"`
	} `json:"body"`
	Meta struct {
		ResourceState []struct {
			Link struct {
				Rel          string `json:"rel"`
				ResourceLink string `json:"resourceLink"`
			} `json:"link"`
			State     string `json:"state"`
			Timestamp int64  `json:"timestamp"`
		} `json:"resourceState"`
	} `json:"meta"`
}

type ZoneConfigurationStruct struct {
	Body struct {
		ActiveFunction string `json:"active_function"`
		Away           struct {
			EndDatetime         string  `json:"end_datetime"`
			StartDatetime       string  `json:"start_datetime"`
			TemperatureSetpoint float64 `json:"temperature_setpoint"`
		} `json:"away"`
		EcoMode             bool   `json:"eco_mode"`
		Enabled             bool   `json:"enabled"`
		ManualCoolingActive bool   `json:"manual_cooling_active"`
		Name                string `json:"name"`
	} `json:"body"`
	Meta struct {
		ResourceState []struct {
			Link struct {
				Rel          string `json:"rel"`
				ResourceLink string `json:"resourceLink"`
			} `json:"link"`
			State     string `json:"state"`
			Timestamp int64  `json:"timestamp"`
		} `json:"resourceState"`
	} `json:"meta"`
}
