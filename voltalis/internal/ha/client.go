package ha

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Client struct {
	token  string
	client *http.Client
}

func NewClient() *Client {
	token := os.Getenv("SUPERVISOR_TOKEN")
	if token == "" {
		panic("No supervisor token found")
	}
	return &Client{
		token:  token,
		client: &http.Client{},
	}
}

func (c *Client) PublishState(entityID string, state string, attributes map[string]any) {
	url := fmt.Sprintf("http://supervisor/core/api/states/%s", entityID)

	payload := map[string]any{
		"state":      state,
		"attributes": attributes,
	}

	data, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		fmt.Println("Error publishing state:", err)
		return
	}
	defer resp.Body.Close()
}
