// Package mqtt provides MQTT client support for remote control of Immich Kiosk.
//
// On connect it publishes Home Assistant MQTT Discovery messages so the kiosk
// appears automatically in Home Assistant. Each named client (e.g. ?client=living-room)
// gets its own HA device with Next and Previous buttons, in addition to a global
// device that controls all connected clients at once.
//
// MQTT topics:
//
//	<prefix>/command          – send command to ALL connected clients
//	<prefix>/<client>/command – send command to a specific client only
//	<prefix>/status           – availability (online/offline) for global device
//	<prefix>/<client>/status  – availability for a specific client
package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"charm.land/log/v2"
	pahomqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/damongolding/immich-kiosk/internal/config"
)

// Command represents a navigation command received via MQTT.
type Command string

const (
	CommandNext     Command = "next"
	CommandPrevious Command = "previous"
)

// Handler is called when a valid command is received.
// target is the client name the command is directed at, or "" for all clients.
type Handler func(cmd Command, target string)

// Client wraps the paho MQTT client.
type Client struct {
	client      pahomqtt.Client
	handlers    []Handler
	handlersMu  sync.RWMutex
	topic       string // global command topic
	availTopic  string // global availability topic
	topicPrefix string
}

// haDevice is the Home Assistant device block.
type haDevice struct {
	Identifiers  []string `json:"identifiers"`
	Name         string   `json:"name"`
	Model        string   `json:"model"`
	Manufacturer string   `json:"manufacturer"`
}

// haButtonConfig is the HA MQTT Discovery payload for a button entity.
type haButtonConfig struct {
	Name              string   `json:"name"`
	UniqueID          string   `json:"unique_id"`
	CommandTopic      string   `json:"command_topic"`
	PayloadPress      string   `json:"payload_press"`
	AvailabilityTopic string   `json:"availability_topic"`
	Device            haDevice `json:"device"`
}

// New creates and connects a new MQTT client.
func New(ctx context.Context, settings config.KioskSettings) (*Client, error) {
	broker := settings.MqttBroker
	if broker == "" {
		return nil, fmt.Errorf("mqtt_broker is not configured")
	}

	prefix := strings.TrimRight(settings.MqttTopicPrefix, "/")

	c := &Client{
		topic:       prefix + "/command",
		availTopic:  prefix + "/status",
		topicPrefix: prefix,
	}

	opts := pahomqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, settings.MqttPort))
	opts.SetClientID(settings.MqttClientID)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(10 * time.Second)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetWill(c.availTopic, "offline", 1, true)

	if settings.MqttUsername != "" {
		opts.SetUsername(settings.MqttUsername)
	}
	if settings.MqttPassword != "" {
		opts.SetPassword(settings.MqttPassword)
	}

	opts.SetOnConnectHandler(func(client pahomqtt.Client) {
		log.Warn("MQTT connected", "broker", broker)

		client.Publish(c.availTopic, 1, true, "online")

		// Global discovery (controls all clients at once)
		c.publishDiscovery(client, settings.MqttClientID, "", "Immich Kiosk (all)")

		// Subscribe to global commands: immich-kiosk/command
		if t := client.Subscribe(c.topic, 1, c.messageHandler); t.Wait() && t.Error() != nil {
			log.Error("MQTT subscribe failed", "topic", c.topic, "err", t.Error())
		} else {
			log.Warn("MQTT subscribed", "topic", c.topic)
		}

		// Subscribe to per-client commands: immich-kiosk/+/command
		wildcardTopic := prefix + "/+/command"
		if t := client.Subscribe(wildcardTopic, 1, c.messageHandler); t.Wait() && t.Error() != nil {
			log.Error("MQTT subscribe failed", "topic", wildcardTopic, "err", t.Error())
		} else {
			log.Warn("MQTT subscribed", "topic", wildcardTopic)
		}
	})

	opts.SetConnectionLostHandler(func(client pahomqtt.Client, err error) {
		log.Warn("MQTT connection lost", "err", err)
	})

	opts.SetReconnectingHandler(func(client pahomqtt.Client, opts *pahomqtt.ClientOptions) {
		log.Warn("MQTT reconnecting...")
	})

	client := pahomqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("mqtt connect: %w", err)
	}

	c.client = client

	go func() {
		<-ctx.Done()
		log.Warn("MQTT disconnecting")
		c.client.Publish(c.availTopic, 1, true, "offline")
		time.Sleep(200 * time.Millisecond)
		client.Disconnect(500)
	}()

	return c, nil
}

// isValidClientName returns true if clientName is safe to use as an MQTT
// topic segment (non-empty, no '/', '#', or '+' characters).
func isValidClientName(name string) bool {
	if name == "" {
		return false
	}
	for _, ch := range name {
		if ch == '/' || ch == '#' || ch == '+' {
			return false
		}
	}
	return true
}

// PublishClientDiscovery publishes HA MQTT Discovery for a named client
// (e.g. "living-room"). Called automatically when a new client connects via SSE.
func (c *Client) PublishClientDiscovery(clientName string) {
	if clientName == "_global" {
		return
	}
	if !isValidClientName(clientName) {
		log.Warn("MQTT invalid client name, skipping discovery", "client", clientName)
		return
	}
	uniqueID := "immich-kiosk-" + clientName
	displayName := "Immich Kiosk - " + clientName
	c.publishDiscovery(c.client, uniqueID, clientName, displayName)
	log.Warn("MQTT client discovery published", "client", clientName)
}

// publishDiscovery publishes HA MQTT Discovery for Next and Previous buttons.
// clientName="" means the global device (all clients).
func (c *Client) publishDiscovery(client pahomqtt.Client, uniqueIDPrefix, clientName, deviceName string) {
	var commandTopic string
	if clientName == "" {
		commandTopic = c.topic
	} else {
		commandTopic = fmt.Sprintf("%s/%s/command", c.topicPrefix, clientName)
	}
	// All entities share the global availability topic — only the server
	// can reliably report online/offline (via LWT), not individual browsers.
	availTopic := c.availTopic

	device := haDevice{
		Identifiers:  []string{uniqueIDPrefix},
		Name:         deviceName,
		Model:        "Immich Kiosk",
		Manufacturer: "immich-kiosk",
	}

	buttons := []struct{ name, payload string }{
		{"Next image", "next"},
		{"Previous image", "previous"},
	}

	for _, b := range buttons {
		discoveryTopic := fmt.Sprintf("homeassistant/button/%s/%s/config", uniqueIDPrefix, b.payload)
		cfg := haButtonConfig{
			Name:              b.name,
			UniqueID:          uniqueIDPrefix + "-" + b.payload,
			CommandTopic:      commandTopic,
			PayloadPress:      b.payload,
			AvailabilityTopic: availTopic,
			Device:            device,
		}
		payload, err := json.Marshal(cfg)
		if err != nil {
			log.Error("MQTT discovery marshal failed", "err", err)
			continue
		}
		client.Publish(discoveryTopic, 1, true, payload)
	}
}

// AddHandler registers a handler called when a command is received.
func (c *Client) AddHandler(h Handler) {
	c.handlersMu.Lock()
	c.handlers = append(c.handlers, h)
	c.handlersMu.Unlock()
}

// messageHandler processes incoming MQTT messages and determines target client.
func (c *Client) messageHandler(_ pahomqtt.Client, msg pahomqtt.Message) {
	topic := msg.Topic()
	payload := strings.TrimSpace(strings.ToLower(string(msg.Payload())))
	log.Debug("MQTT message received", "topic", topic, "payload", payload)

	// Determine target client from topic structure:
	// prefix/command          → target = "" (all)
	// prefix/clientname/command → target = "clientname"
	var target string
	withoutPrefix := strings.TrimPrefix(topic, c.topicPrefix+"/")
	if withoutPrefix != "command" {
		// format is "<clientname>/command"
		target = strings.TrimSuffix(withoutPrefix, "/command")
	}

	var cmd Command
	switch payload {
	case string(CommandNext):
		cmd = CommandNext
	case string(CommandPrevious):
		cmd = CommandPrevious
	default:
		log.Warn("MQTT unknown command", "payload", payload)
		return
	}

	c.handlersMu.RLock()
	handlers := c.handlers
	c.handlersMu.RUnlock()
	for _, h := range handlers {
		h(cmd, target)
	}
}
