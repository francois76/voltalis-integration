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

func (c *Client) EnableQuickSetting(qsID int, enabled bool) error {
	body := map[string]bool{"enabled": enabled}
	return c.put(fmt.Sprintf("/api/site/%d/quicksettings/%d/enable", c.SiteID, qsID), body, nil)
}
