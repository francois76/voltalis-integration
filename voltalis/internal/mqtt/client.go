package mqtt

import mqtt "github.com/eclipse/paho.mqtt.golang"

type Client struct {
	mqtt.Client
}

func InitClient(broker string, clientID string) (*Client, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return &Client{Client: client}, nil
}
