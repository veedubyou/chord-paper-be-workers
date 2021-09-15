package main

import (
	"chord-paper-be-workers/src/application/jobs/job_message"
	"chord-paper-be-workers/src/application/jobs/start"
	"encoding/json"
	"os"

	"github.com/streadway/amqp"
)

func main() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		panic("Can't get rabbitmq url")
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	rabbitChannel, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer rabbitChannel.Close()

	queue, err := rabbitChannel.QueueDeclare(
		"chord-paper-dev",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		panic(err)
	}

	startJobParams := start.JobParams{
		TrackIdentifier: job_message.TrackIdentifier{
			TrackListID: "ad2fca6d-8c32-4030-86c0-8b5339347253",
			TrackID:     "440a7737-bcda-4761-ae89-d85880f4bce3",
		},
	}

	jobBody, err := json.Marshal(startJobParams)

	if err != nil {
		panic(err)
	}

	job := amqp.Publishing{Type: start.JobType, Body: jobBody}

	job.DeliveryMode = amqp.Persistent
	job.ContentType = "application/json"

	err = rabbitChannel.Publish("", queue.Name, true, false, job)

	if err != nil {
		panic(err)
	}
}
