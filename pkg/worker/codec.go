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

	payload := json.RawMessage(msg.Payload)
	return &Event{
		Provider:   env.Provider,
		Type:       env.Name,
		Topic:      topic,
		Metadata:   metadata,
		Payload:    payload,
		Normalized: env.Data,
	}, nil
}
