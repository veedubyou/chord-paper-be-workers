package worker

import (
	"chord-paper-be-workers/src/application/jobs/job_router"
	"chord-paper-be-workers/src/lib/cerr"

	"github.com/apex/log"

	"github.com/streadway/amqp"
)

type MessageChannel interface {
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Close() error
}

type QueueWorker struct {
	channel   MessageChannel
	jobRouter job_router.JobRouter
	queueName string
}

func NewQueueWorker(channel MessageChannel, queueName string, jobRouter job_router.JobRouter) QueueWorker {
	return QueueWorker{
		channel:   channel,
		queueName: queueName,
		jobRouter: jobRouter,
	}
}

func NewQueueWorkerFromConnection(conn *amqp.Connection, queueName string, jobRouter job_router.JobRouter) (QueueWorker, error) {
	rabbitChannel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return QueueWorker{}, cerr.Wrap(err).Error("Failed to get channel")
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
		return QueueWorker{}, cerr.Wrap(err).Error("Failed to declare queue")
	}

	return NewQueueWorker(rabbitChannel, queue.Name, jobRouter), nil
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
		return cerr.Field("queue_name", q.queueName).
			Wrap(err).Error("Failed to start consuming from channel")
	}

	for message := range messageStream {
		logger := log.WithField("message_type", message.Type)
		logger.Info("Handling message")
		err := q.jobRouter.HandleMessage(message)
		if err != nil {
			err = cerr.Field("message_type", message.Type).
				Wrap(err).Error("Failed to process message")

			cerr.Log(err)

			if err = message.Nack(false, false); err != nil {
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
