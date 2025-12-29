package worker

import (
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill/message"
)

// Codec is an interface for decoding messages from a message broker into an Event.
type Codec interface {
	// Decode transforms a Watermill message into an Event.
	Decode(topic string, msg *message.Message) (*Event, error)
}

// DefaultCodec is the default implementation of the Codec interface.
// It decodes a JSON payload into an Event.
type DefaultCodec struct{}

// envelope is used to unmarshal the basic event properties.
type envelope struct {
	Provider string                 `json:"provider"`
	Name     string                 `json:"name"`
	Data     map[string]interface{} `json:"data"`
}

// Decode unmarshals a Watermill message into an Event.
func (DefaultCodec) Decode(topic string, msg *message.Message) (*Event, error) {
	var env envelope
	if err := json.Unmarshal(msg.Payload, &env); err != nil {
		return nil, err
	}

	metadata := make(map[string]string, len(msg.Metadata))
	for key, value := range msg.Metadata {
		metadata[key] = value
	}

	provider := env.Provider
	if provider == "" {
		provider = msg.Metadata.Get("provider")
	}
	eventName := env.Name
	if eventName == "" {
		eventName = msg.Metadata.Get("event")
	}

	normalized := env.Data
	if normalized == nil {
		var raw interface{}
		if err := json.Unmarshal(msg.Payload, &raw); err == nil {
			if object, ok := raw.(map[string]interface{}); ok {
				normalized = object
			}
		}
	}

	payload := json.RawMessage(msg.Payload)
	return &Event{
		Provider:   provider,
		Type:       eventName,
		Topic:      topic,
		Metadata:   metadata,
		Payload:    payload,
		Normalized: normalized,
	}, nil
}
