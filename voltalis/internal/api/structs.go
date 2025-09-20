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
	ID             int      `json:"id"`
	Name           string   `json:"name"`
	ApplianceType  string   `json:"applianceType"`
	AvailableModes []string `json:"availableModes"`
	HeatingLevel   int      `json:"heatingLevel"`
}

type Consumption struct {
	AggregationStepInSeconds int `json:"aggregationStepInSeconds"`
	Consumptions             []struct {
		StepTimestampInUtc     time.Time `json:"stepTimestampInUtc"`
		TotalConsumptionInWh   float64   `json:"totalConsumptionInWh"`
		TotalConsumptionInCurr float64   `json:"totalConsumptionInCurrency"`
	} `json:"consumptions"`
}
