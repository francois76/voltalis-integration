package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
	SiteID     int
}

func NewClient(baseURL, login, password string) (*Client, error) {
	c := &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}

	if err := c.login(login, password); err != nil {
		return nil, err
	}
	me, err := c.GetMe()
	if err != nil {
		return nil, err
	}
	c.SiteID = me.DefaultSite.ID
	return c, nil
}

func (c *Client) login(login, password string) error {
	reqBody := map[string]string{
		"login":    login,
		"password": password,
	}
	b, _ := json.Marshal(reqBody)
	resp, err := c.HTTPClient.Post(c.BaseURL+"/auth/login", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: %s", resp.Status)
	}

	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	c.Token = out.Token
	return nil
}

func (c *Client) get(path string, out interface{}) error {
	req, _ := http.NewRequest("GET", c.BaseURL+path, nil)
	req.Header.Set("Authorization", "Bearer "+c.Token)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) put(path string, body interface{}, out interface{}) error {
	b, _ := json.Marshal(body)
	slog.Debug("API PUT request", "path", path, "body", string(b))
	req, _ := http.NewRequest("PUT", c.BaseURL+path, bytes.NewBuffer(b))
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *Client) post(path string, body interface{}, out interface{}) error {
	b, _ := json.Marshal(body)
	slog.Debug("API POST request", "path", path, "body", string(b))
	req, _ := http.NewRequest("POST", c.BaseURL+path, bytes.NewBuffer(b))
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
