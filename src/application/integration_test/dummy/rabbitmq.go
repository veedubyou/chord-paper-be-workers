package dummy

import (
	"chord-paper-be-workers/src/application/publish"
	"chord-paper-be-workers/src/application/worker"

	"github.com/streadway/amqp"
)

var _ publish.Publisher = &RabbitMQ{}
var _ worker.MessageChannel = &RabbitMQ{}

type RabbitMQ struct {
	Unavailable    bool
	MessageChannel chan amqp.Delivery
}

func NewRabbitMQ() *RabbitMQ {
	return &RabbitMQ{
		Unavailable:    false,
		MessageChannel: make(chan amqp.Delivery, 100),
	}
}

func (r *RabbitMQ) Publish(msg amqp.Publishing) error {
	if r.Unavailable {
		return NetworkFailure
	}

	r.MessageChannel <- amqp.Delivery{
		ContentType:     msg.ContentType,
		ContentEncoding: msg.ContentEncoding,
		DeliveryMode:    msg.DeliveryMode,
		Timestamp:       msg.Timestamp,
		Type:            msg.Type,
		Body:            msg.Body,
	}
	return nil
}

func (r *RabbitMQ) Consume(_ string, _ string, _ bool, _ bool, _ bool, _ bool, _ amqp.Table) (<-chan amqp.Delivery, error) {
	if r.Unavailable {
		return nil, NetworkFailure
	}

	return r.MessageChannel, nil
}

func (r *RabbitMQ) Close() error {
	return nil
}
