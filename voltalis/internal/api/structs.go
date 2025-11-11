package api

import "time"

type User struct {
	ID          int     `json:"id"`
	Firstname   string  `json:"firstname"`
	Lastname    string  `json:"lastname"`
	Email       string  `json:"email"`
	Phones      []Phone `json:"phones"`
	DefaultSite Site    `json:"defaultSite"`
}

type Phone struct {
	PhoneType   *string `json:"phoneType"`
	PhoneNumber string  `json:"phoneNumber"`
	IsDefault   bool    `json:"isDefault"`
}

type Site struct {
	ID         int    `json:"id"`
	Address    string `json:"address"`
	PostalCode string `json:"postalCode"`
	City       string `json:"city"`
	Country    string `json:"country"`
}

type Appliance struct {
	ID              int         `json:"id"`
	Name            string      `json:"name"`
	ApplianceType   string      `json:"applianceType"`
	ModulatorType   string      `json:"modulatorType"`
	AvailableModes  []string    `json:"availableModes"`
	VoltalisVersion string      `json:"voltalisVersion"`
	Programming     Programming `json:"programming"`
	HeatingLevel    int         `json:"heatingLevel"`
}

// Structure pour le champ programming
type Programming struct {
	ProgType           string  `json:"progType"`
	ProgName           string  `json:"progName"`
	IDManualSetting    *int    `json:"idManualSetting"`
	IsOn               bool    `json:"isOn"`
	UntilFurtherNotice *bool   `json:"untilFurtherNotice"`
	Mode               string  `json:"mode"`
	IDPlanning         int     `json:"idPlanning"`
	EndDate            *string `json:"endDate"`
	TemperatureTarget  float64 `json:"temperatureTarget"`
	DefaultTemperature float64 `json:"defaultTemperature"`
}

type ManualSetting struct {
	ID                 int        `json:"id"`
	Enabled            bool       `json:"enabled"`
	IDAppliance        int        `json:"idAppliance"`
	ApplianceName      string     `json:"applianceName"`
	ApplianceType      string     `json:"applianceType"`
	UntilFurtherNotice bool       `json:"untilFurtherNotice"`
	IsOn               bool       `json:"isOn"`
	Mode               string     `json:"mode"`
	HeatingLevel       int        `json:"heatingLevel"`
	EndDate            *time.Time `json:"endDate"`
	TemperatureTarget  float64    `json:"temperatureTarget"`
}

type Consumption struct {
	AggregationStepInSeconds int `json:"aggregationStepInSeconds"`
	Consumptions             []struct {
		StepTimestampInUtc     time.Time `json:"stepTimestampInUtc"`
		TotalConsumptionInWh   float64   `json:"totalConsumptionInWh"`
		TotalConsumptionInCurr float64   `json:"totalConsumptionInCurrency"`
	} `json:"consumptions"`
}

type Program struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}
