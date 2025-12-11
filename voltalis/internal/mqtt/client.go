package mqtt

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/francois76/voltalis-integration/voltalis/internal/state"
)

// subscriptionInfo stocke les infos nécessaires pour se réabonner après reconnexion
type subscriptionInfo struct {
	topic   string
	handler mqtt.MessageHandler
}

type Client struct {
	mqtt.Client
	stateMutex    sync.Mutex
	stateTopicMap map[SetTopic]string // possède la dernière valeur set par HA sur chaque topic
	StateManager  *StateManager       // machine à état de plus haut niveau ne renvoyant à l'exterieur que les données à renvoyer à voltalis

	// Pour la gestion des réabonnements après reconnexion
	subscriptionsMutex sync.Mutex
	subscriptions      []subscriptionInfo
	hasConnectedOnce   atomic.Bool // true après la première connexion réussie avec subscriptions
}

func InitClient(broker string, clientID string, password string) (*Client, error) {
	// Créer le wrapper client d'abord
	c := &Client{
		stateTopicMap: make(map[SetTopic]string),
		subscriptions: make([]subscriptionInfo, 0),
	}

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetMaxReconnectInterval(30 * time.Second).
		SetKeepAlive(30 * time.Second).
		SetPingTimeout(10 * time.Second).
		SetCleanSession(false).
		SetConnectionLostHandler(func(client mqtt.Client, err error) {
			slog.Error("Connexion MQTT perdue", "error", err)
		}).
		SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
			slog.Warn("Tentative de reconnexion MQTT...")
		}).
		SetOnConnectHandler(func(client mqtt.Client) {
			// Ne réabonner que si on a déjà eu une première connexion avec des subscriptions
			if c.hasConnectedOnce.Load() {
				slog.Info("Reconnexion MQTT établie, réabonnement aux topics...")
				c.resubscribeAll()
			} else {
				slog.Info("Connexion MQTT établie")
			}
		})

	if password != "" {
		opts = opts.SetPassword(password)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	c.Client = client

	stateManager := NewStateManager()
	stateManager.UpdateState(state.ResourceState{
		ControllerState: state.ControllerState{},
		HeaterState:     map[int64]state.HeaterState{},
	})
	c.StateManager = stateManager

	return c, nil
}

// resubscribeAll réabonne à tous les topics après une reconnexion
func (c *Client) resubscribeAll() {
	c.subscriptionsMutex.Lock()
	subs := make([]subscriptionInfo, len(c.subscriptions))
	copy(subs, c.subscriptions)
	c.subscriptionsMutex.Unlock()

	for _, sub := range subs {
		slog.Debug("Réabonnement au topic", "topic", sub.topic)
		token := c.Client.Subscribe(sub.topic, 0, sub.handler)
		if token.Wait() && token.Error() != nil {
			slog.Error("Échec du réabonnement", "topic", sub.topic, "error", token.Error())
		}
	}
	slog.Info("Réabonnement terminé", "count", len(subs))
}

// registerSubscription enregistre une subscription pour pouvoir se réabonner après reconnexion
func (c *Client) registerSubscription(topic string, handler mqtt.MessageHandler) {
	c.subscriptionsMutex.Lock()
	defer c.subscriptionsMutex.Unlock()
	c.subscriptions = append(c.subscriptions, subscriptionInfo{topic: topic, handler: handler})
}

// MarkSubscriptionsComplete marque que toutes les subscriptions initiales sont faites
func (c *Client) MarkSubscriptionsComplete() {
	c.hasConnectedOnce.Store(true)
	slog.Info("Subscriptions MQTT initialisées", "count", len(c.subscriptions))
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

func (c *Client) BuildControllerStateTopic() ControllerGetTopics {
	return ControllerGetTopics{
		Mode:     getPayloadSelectMode(CONTROLLER_DEVICE, PRESET_SELECT_CONTROLLER...).StateTopic,
		Duration: getPayloadSelectDuration(CONTROLLER_DEVICE).StateTopic,
		Program:  getPayloadSelectProgram().StateTopic,
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

func (c *Client) BuildHeaterStateTopic(id int64) HeaterGetTopics {
	climate := c.buildClimateStates(id)
	durationPayload := getPayloadSelectDuration(buildDeviceInfo(id, ""))
	return HeaterGetTopics{
		Mode:           climate.ModeStateTopic,
		PresetMode:     climate.PresetModeStateTopic,
		Temperature:    climate.TemperatureStateTopic,
		SingleDuration: durationPayload.StateTopic,
		Action:         NewHeaterTopic[GetTopic](id, "action"),
	}
}
