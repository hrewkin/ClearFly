package delivery

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/cleanair/booking/internal/usecase"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitPublisher publishes notification events to the notification_events
// queue. It implements usecase.NotificationPublisher.
type RabbitPublisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewRabbitPublisher dials RabbitMQ and declares the exchange/queue used by
// the notification microservice.
//
// On failure, returns a *RabbitPublisher with nil channel so callers can
// proceed without notifications. This keeps booking responsive even when
// RabbitMQ is unavailable.
func NewRabbitPublisher(amqpURL string) (*RabbitPublisher, error) {
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
	return &RabbitPublisher{conn: conn, ch: ch}, nil
}

// Publish sends a notification event to the bus.
func (p *RabbitPublisher) Publish(ctx context.Context, evt usecase.NotificationEvent) error {
	if p == nil || p.ch == nil {
		return nil
	}
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	pubCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	err = p.ch.PublishWithContext(pubCtx, "notification_events", "notify."+evt.Type, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
	if err != nil {
		log.Printf("notification publish failed: %v", err)
	}
	return err
}

// Close releases the AMQP connection.
func (p *RabbitPublisher) Close() {
	if p == nil {
		return
	}
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
}
