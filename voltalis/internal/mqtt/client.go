package mqtt

import (
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/francois76/voltalis-integration/voltalis/internal/state"
)

type Client struct {
	mqtt.Client
	stateMutex    sync.Mutex
	stateTopicMap map[SetTopic]string // possède la dernière valeur set par HA sur chaque topic
	StateManager  *StateManager       // machine à état de plus haut niveau ne renvoyant à l'exterieur que les données à renvoyer à voltalis
}

func InitClient(broker string, clientID string, password string) (*Client, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID)
	if password != "" {
		opts = opts.SetPassword(password)
	}
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	stateManager := NewStateManager()
	stateManager.UpdateState(state.ResourceState{
		ControllerState: state.ControllerState{},
		HeaterState:     map[int64]state.HeaterState{},
	})
	return &Client{
		Client:        client,
		stateTopicMap: make(map[SetTopic]string),
		StateManager:  stateManager,
	}, nil
}

func (c *Client) GetTopicState(topic SetTopic) string {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()
	return c.stateTopicMap[topic]
}

func (c *Client) BuildHeaterCommands(id int64) HeaterCommandPayload {
	return HeaterCommandPayload{
		ModeCommandTopic:        NewHeaterTopic[SetTopic](id, "mode"),
		PresetModeCommandTopic:  NewHeaterTopic[SetTopic](id, "preset_mode"),
		TemperatureCommandTopic: NewHeaterTopic[SetTopic](id, "temp"),
	}
}
