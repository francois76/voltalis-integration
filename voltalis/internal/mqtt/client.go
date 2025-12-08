package mqtt

import (
	"log/slog"
	"sync"
	"time"

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
		SetClientID(clientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetMaxReconnectInterval(30 * time.Second).
		SetKeepAlive(30 * time.Second).
		SetPingTimeout(10 * time.Second).
		SetCleanSession(false). // Garde les subscriptions côté broker après reconnexion
		SetConnectionLostHandler(func(client mqtt.Client, err error) {
			slog.Error("Connexion MQTT perdue", "error", err)
		}).
		SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
			slog.Warn("Tentative de reconnexion MQTT...")
		}).
		SetOnConnectHandler(func(client mqtt.Client) {
			slog.Info("Connexion MQTT établie")
		})

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

func (c *Client) buildClimateCommands(id int64) ClimateCommandPayload {
	return ClimateCommandPayload{
		ModeCommandTopic:        NewHeaterTopic[SetTopic](id, "mode"),
		PresetModeCommandTopic:  NewHeaterTopic[SetTopic](id, "preset_mode"),
		TemperatureCommandTopic: NewHeaterTopic[SetTopic](id, "temp"),
	}
}

func (c *Client) buildClimateStates(id int64) ClimateStatePayload {
	return ClimateStatePayload{
		ModeStateTopic:          NewHeaterTopic[GetTopic](id, "mode"),
		PresetModeStateTopic:    NewHeaterTopic[GetTopic](id, "preset_mode"),
		TemperatureStateTopic:   NewHeaterTopic[GetTopic](id, "temp"),
		CommandTopic:            NewHeaterTopic[GetTopic](id, "set"),
		CurrentTemperatureTopic: NewHeaterTopic[GetTopic](id, "current_temp"),
	}
}

func (c *Client) BuildControllerCommandTopic() ControllerSetTopics {
	return ControllerSetTopics{
		Mode:     getPayloadSelectMode(CONTROLLER_DEVICE, PRESET_SELECT_CONTROLLER...).CommandTopic,
		Duration: getPayloadSelectDuration(CONTROLLER_DEVICE).CommandTopic,
		Program:  getPayloadSelectProgram().CommandTopic,
	}
}

func (c *Client) BuildHeaterCommandTopic(id int64) HeaterSetTopics {
	climate := c.buildClimateCommands(id)
	durationPayload := getPayloadSelectDuration(buildDeviceInfo(id, ""))
	return HeaterSetTopics{
		Mode:           climate.ModeCommandTopic,
		PresetMode:     climate.PresetModeCommandTopic,
		Temperature:    climate.TemperatureCommandTopic,
		SingleDuration: durationPayload.CommandTopic,
	}
}
