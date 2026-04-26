package delivery

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/cleanair/notification/internal/usecase"
	amqp "github.com/rabbitmq/amqp091-go"
)

// IncomingEvent matches the payload published by booking & incident services
// to the notification_events queue.
type IncomingEvent struct {
	Type        string                 `json:"type"`
	PassengerID string                 `json:"passenger_id"`
	FlightID    string                 `json:"flight_id"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Channels    []string               `json:"channels"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

// RabbitConsumer drains the notification_events queue and persists
// notifications into the in-memory store.
type RabbitConsumer struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	store *usecase.Store
}

// NewRabbitConsumer dials RabbitMQ and binds to the notification_events
// exchange/queue.
func NewRabbitConsumer(amqpURL string, store *usecase.Store) (*RabbitConsumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := ch.ExchangeDeclare("notification_events", "topic", true, false, false, false, nil); err != nil {
		return nil, err
	}
	if _, err := ch.QueueDeclare("notification_events", true, false, false, false, nil); err != nil {
		return nil, err
	}
	if err := ch.QueueBind("notification_events", "notify.#", "notification_events", false, nil); err != nil {
		return nil, err
	}
	return &RabbitConsumer{conn: conn, ch: ch, store: store}, nil
}

// Start launches a goroutine that consumes events and writes them to the store.
func (c *RabbitConsumer) Start(ctx context.Context) error {
	msgs, err := c.ch.Consume("notification_events", "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				_ = c.ch.Close()
				_ = c.conn.Close()
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				var evt IncomingEvent
				if err := json.Unmarshal(msg.Body, &evt); err != nil {
					log.Printf("notification: unmarshal failed: %v", err)
					_ = msg.Nack(false, false)
					continue
				}
				channels := evt.Channels
				if len(channels) == 0 {
					channels = []string{"PUSH"}
				}
				for _, ch := range channels {
					c.store.Add(usecase.Notification{
						PassengerID: evt.PassengerID,
						FlightID:    evt.FlightID,
						Type:        evt.Type,
						Title:       firstNonEmpty(evt.Title, evt.Type),
						Content:     evt.Content,
						Channel:     strings.ToUpper(ch),
					})
				}
				_ = msg.Ack(false)
			}
		}
	}()
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
