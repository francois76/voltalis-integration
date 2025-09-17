package mqtt

import (
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	mqtt.Client
	stateMutex sync.Mutex
	stateMap   map[SetTopic]string // clé = topic, value = dernière valeur lue
}

func InitClient(broker string, clientID string) (*Client, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return &Client{
		Client:   client,
		stateMap: make(map[SetTopic]string),
	}, nil
}

func (c *Client) GetState(topic SetTopic) string {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()
	return c.stateMap[topic]
}
