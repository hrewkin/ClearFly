package consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/cleanair/incident/internal/usecase"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitPublisher fans out incident-driven notifications onto the
// notification_events queue, where the notification microservice picks
// them up.
type RabbitPublisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewRabbitPublisher dials RabbitMQ and declares the notification topic.
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

// Publish satisfies usecase.NotificationPublisher.
func (p *RabbitPublisher) Publish(ctx context.Context, n usecase.Notification) error {
	if p == nil || p.ch == nil {
		return nil
	}
	body, err := json.Marshal(n)
	if err != nil {
		return err
	}
	pubCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := p.ch.PublishWithContext(pubCtx, "notification_events", "notify."+n.Type, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	}); err != nil {
		log.Printf("incident publisher: failed to publish: %v", err)
		return err
	}
	return nil
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
