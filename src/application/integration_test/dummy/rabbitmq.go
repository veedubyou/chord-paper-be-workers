package dummy

import (
	"chord-paper-be-workers/src/application/publish"
	"chord-paper-be-workers/src/application/worker"

	"github.com/streadway/amqp"
)

var _ publish.Publisher = &RabbitMQ{}
var _ worker.MessageChannel = &RabbitMQ{}
var _ amqp.Acknowledger = RabbitMQAcknowledger{}

type RabbitMQ struct {
	AckCounter     int
	NackCounter    int
	Unavailable    bool
	MessageChannel chan amqp.Delivery
}

type RabbitMQAcknowledger struct {
	ack  func()
	nack func()
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

	acknowledger := RabbitMQAcknowledger{
		ack: func() {
			r.AckCounter++
		},
		nack: func() {
			r.NackCounter++
		},
	}

	r.MessageChannel <- amqp.Delivery{
		Acknowledger:    acknowledger,
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

func (r RabbitMQAcknowledger) Ack(tag uint64, multiple bool) error {
	r.ack()
	return nil
}
func (r RabbitMQAcknowledger) Nack(tag uint64, multiple bool, requeue bool) error {
	r.nack()
	return nil
}
func (r RabbitMQAcknowledger) Reject(tag uint64, requeue bool) error {
	r.nack()
	return nil
}
