package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/cleanair/incident/internal/usecase"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitConsumer struct {
	conn    *amqp.Connection
	ch      *amqp.Channel
	handler *usecase.IncidentHandler
}

func NewRabbitConsumer(amqpURL string, handler *usecase.IncidentHandler) (*RabbitConsumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Setup exchange and queues
	err = ch.ExchangeDeclare("flight_events", "topic", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare("incident_events", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(q.Name, "flight.incident.#", "flight_events", false, nil)
	if err != nil {
		return nil, err
	}

	return &RabbitConsumer{conn: conn, ch: ch, handler: handler}, nil
}

func (c *RabbitConsumer) Start(ctx context.Context) error {
	msgs, err := c.ch.Consume("incident_events", "", false, false, false, false, nil)
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
				var evt usecase.IncidentEvent
				if err := json.Unmarshal(msg.Body, &evt); err == nil {
					if handleErr := c.handler.HandleIncident(evt); handleErr == nil {
						msg.Ack(false)
					} else {
						log.Printf("Failed to process incident: %v", handleErr)
						msg.Nack(false, false)
					}
				} else {
					log.Printf("Failed to parse incident json: %v", err)
					msg.Nack(false, false)
				}
			}
		}
	}()

	return nil
}
