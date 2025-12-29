package worker

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

func MiddlewareFromWatermill(m message.HandlerMiddleware) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, evt *Event) error {
			msg := message.NewMessage(watermill.NewUUID(), evt.Payload)
			if evt.Metadata != nil {
				msg.Metadata = message.Metadata{}
				for key, value := range evt.Metadata {
					msg.Metadata[key] = value
				}
			}
			wrapped := m(func(_ *message.Message) ([]*message.Message, error) {
				return nil, next(ctx, evt)
			})
			_, err := wrapped(msg)
			return err
		}
	}
}
