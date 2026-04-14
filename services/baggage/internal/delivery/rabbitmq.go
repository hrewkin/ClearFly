package delivery

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/cleanair/baggage/internal/repository"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type BaggageEvent struct {
	ID          string `json:"id"`
	PassengerID string `json:"passenger_id"`
	Status      string `json:"status"`
	Location    string `json:"location"`
}

type RabbitConsumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	repo repository.Repository
}

func NewRabbitConsumer(amqpURL string, repo repository.Repository) (*RabbitConsumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Declare exchange and queue
	err = ch.ExchangeDeclare("baggage_events", "topic", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare("baggage_scanned", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(q.Name, "baggage.#", "baggage_events", false, nil)
	if err != nil {
		return nil, err
	}

	return &RabbitConsumer{conn: conn, ch: ch, repo: repo}, nil
}

func (c *RabbitConsumer) Start(ctx context.Context) error {
	msgs, err := c.ch.Consume("baggage_scanned", "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.ch.Close()
				c.conn.Close()
				return
			case msg := <-msgs:
				var evt BaggageEvent
				if err := json.Unmarshal(msg.Body, &evt); err == nil {
					log.Printf("Received baggage event: %+v", evt)
					// Upsert to DB, parsing uuids would be better but let's assume valid
					b, err := ParseEvent(&evt)
					if err == nil {
						c.repo.Upsert(context.Background(), b)
						msg.Ack(false)
					} else {
						msg.Nack(false, false)
					}
				} else {
					msg.Nack(false, false)
				}
			}
		}
	}()

	return nil
}

func ParseEvent(evt *BaggageEvent) (*repository.BaggageStatus, error) {
	id, err := uuid.Parse(evt.ID)
	if err != nil {
		return nil, err
	}
	pid, err := uuid.Parse(evt.PassengerID)
	if err != nil {
		return nil, err
	}

	return &repository.BaggageStatus{
		ID:          id,
		PassengerID: pid,
		Status:      evt.Status,
		Location:    evt.Location,
		UpdatedAt:   time.Now(),
	}, nil
}
