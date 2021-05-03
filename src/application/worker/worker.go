package worker

import (
	"chord-paper-be-workers/src/lib/werror"

	"github.com/apex/log"

	"github.com/streadway/amqp"
)

type MessageChannel interface {
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Close() error
}

type QueueWorker struct {
	channel         MessageChannel
	messageHandlers []MessageHandler
	queueName       string
}

func NewQueueWorker(channel MessageChannel, queueName string, messageHandlers []MessageHandler) QueueWorker {
	return QueueWorker{
		channel:         channel,
		queueName:       queueName,
		messageHandlers: messageHandlers,
	}
}

func NewQueueWorkerFromConnection(conn *amqp.Connection, queueName string, messageHandlers []MessageHandler) (QueueWorker, error) {
	rabbitChannel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return QueueWorker{}, werror.WrapError("Failed to get channel", err)
	}

	queue, err := rabbitChannel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		_ = rabbitChannel.Close()
		return QueueWorker{}, werror.WrapError("Failed to declare queue", err)
	}

	return NewQueueWorker(rabbitChannel, queue.Name, messageHandlers), nil
}

func (q *QueueWorker) Start() error {
	log.Info("Starting worker")

	defer q.channel.Close()

	messageStream, err := q.channel.Consume(
		q.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return werror.WrapError("Failed to start consuming from channel", err)
	}

	for message := range messageStream {
		logger := log.WithField("message_type", message.Type)
		logger.Info("Handling message")
		err := q.handleMessage(message)
		if err != nil {
			logger.WithField("err", err).Error("Failed to process messages")
			if err = message.Nack(false, true); err != nil {
				logger.Error("Failed to nack message")
			}
		} else {
			logger.Info("Successfully processed message")
			if err = message.Ack(false); err != nil {
				logger.Error("Failed to ack message")
			}
		}
	}

	return nil
}

func (q *QueueWorker) handleMessage(message amqp.Delivery) error {
	for _, handler := range q.messageHandlers {
		if message.Type == handler.JobType() {
			return handler.HandleMessage(message.Body)
		}
	}

	return werror.WrapError("No appropriate message handler found", nil)
}
