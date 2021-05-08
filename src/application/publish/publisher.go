package publish

import (
	"chord-paper-be-workers/src/lib/werror"

	"github.com/streadway/amqp"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

var _ Publisher = RabbitMQPublisher{}

//counterfeiter:generate . Publisher
type Publisher interface {
	Publish(msg amqp.Publishing) error
}

func NewRabbitMQPublisher(conn *amqp.Connection, queueName string) (RabbitMQPublisher, error) {
	channel, err := conn.Channel()
	if err != nil {
		return RabbitMQPublisher{}, werror.WrapError("Failed to create rabbit channel", err)
	}

	return RabbitMQPublisher{
		channel:   channel,
		queueName: queueName,
	}, nil
}

type RabbitMQPublisher struct {
	channel   *amqp.Channel
	queueName string
}

func (r RabbitMQPublisher) Publish(msg amqp.Publishing) error {
	msg.ContentType = "application/json"
	msg.DeliveryMode = amqp.Persistent
	return r.channel.Publish("", r.queueName, true, false, msg)
}
