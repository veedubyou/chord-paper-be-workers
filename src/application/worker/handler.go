package worker

type MessageHandler interface {
	JobType() string
	HandleMessage(message []byte) error
}
