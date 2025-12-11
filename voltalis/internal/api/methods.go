package api

import "fmt"

func (c *Client) GetMe() (*User, error) {
	var u User
	err := c.get("/api/account/me", &u)
	return &u, err
}

func (c *Client) GetAppliances() ([]Appliance, error) {
	var apps []Appliance
	err := c.get(fmt.Sprintf("/api/site/%d/managed-appliance", c.SiteID), &apps)
	return apps, err
}

func (c *Client) GetAppliance(applianceID int) (*Appliance, error) {
	var app Appliance
	err := c.get(fmt.Sprintf("/api/site/%d/managed-appliance/%d", c.SiteID, applianceID), &app)
	return &app, err
}

func (c *Client) GetConsumptionRealtime() (*Consumption, error) {
	var cons Consumption
	err := c.get(fmt.Sprintf("/api/site/%d/consumption/realtime", c.SiteID), &cons)
	return &cons, err
}

func (c *Client) GetManualSettings() ([]ManualSetting, error) {
	var settings []ManualSetting
	err := c.get(fmt.Sprintf("/api/site/%d/manualsetting", c.SiteID), &settings)
	return settings, err
}

func (c *Client) EnableQuickSetting(qsID int, enabled bool) error {
	body := map[string]bool{"enabled": enabled}
	return c.put(fmt.Sprintf("/api/site/%d/quicksettings/%d/enable", c.SiteID, qsID), body, nil)
}

func (c *Client) GetPrograms() ([]Program, error) {
	var programs []Program
	err := c.get(fmt.Sprintf("/api/site/%d/programming/program", c.SiteID), &programs)
	return programs, err
}

// UpdateProgram met à jour un programme (activation/désactivation)
func (c *Client) UpdateProgram(programID int, request UpdateProgramRequest) error {
	return c.put(fmt.Sprintf("/api/site/%d/programming/program/%d", c.SiteID, programID), request, nil)
}

// GetQuickSettings récupère la liste des quicksettings disponibles
func (c *Client) GetQuickSettings() ([]QuickSettings, error) {
	var qs []QuickSettings
	err := c.get(fmt.Sprintf("/api/site/%d/quicksettings", c.SiteID), &qs)
	return qs, err
}

// UpdateQuickSettings met à jour un quicksetting complet
func (c *Client) UpdateQuickSettings(qsID int, qs QuickSettings) error {
	return c.put(fmt.Sprintf("/api/site/%d/quicksettings/%d", c.SiteID, qsID), qs, nil)
}

// UpdateManualSetting met à jour un réglage manuel pour un radiateur spécifique
func (c *Client) UpdateManualSetting(manualSettingID int, request UpdateManualSettingRequest) error {
	return c.put(fmt.Sprintf("/api/site/%d/manualsetting/%d", c.SiteID, manualSettingID), request, nil)
}

// CreateManualSetting crée un nouveau réglage manuel pour un radiateur
func (c *Client) CreateManualSetting(request UpdateManualSettingRequest) (*ManualSetting, error) {
	var result ManualSetting
	err := c.post(fmt.Sprintf("/api/site/%d/manualsetting", c.SiteID), request, &result)
	return &result, err
}
