package application

import (
	filestore "chord-paper-be-workers/src/application/cloud_storage/store"
	"chord-paper-be-workers/src/application/executor"
	"chord-paper-be-workers/src/application/jobs/save_stems_to_db"
	"chord-paper-be-workers/src/application/jobs/split"
	"chord-paper-be-workers/src/application/jobs/split/splitter"
	"chord-paper-be-workers/src/application/jobs/split/splitter/file_splitter"
	"chord-paper-be-workers/src/application/jobs/transfer"
	"chord-paper-be-workers/src/application/jobs/transfer/download"
	"chord-paper-be-workers/src/application/publish"
	trackstore "chord-paper-be-workers/src/application/tracks/store"
	"chord-paper-be-workers/src/application/worker"
	"chord-paper-be-workers/src/lib/cerr"
	"chord-paper-be-workers/src/lib/env"
	"fmt"
	"os"

	"github.com/streadway/amqp"
)

func getEnvOrPanic(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("No env variable found for key %s", key))
	}

	return val
}

func ensureOk(err error) {
	if err != nil {
		panic(err)
	}
}

type App struct {
	worker worker.QueueWorker
}

func NewApp() App {
	rabbitURL := getEnvOrPanic("RABBITMQ_URL")
	consumerConn, err := amqp.Dial(rabbitURL)
	ensureOk(err)
	producerConn, err := amqp.Dial(rabbitURL)
	ensureOk(err)

	return App{
		worker: newWorker(consumerConn, producerConn),
	}
}

func (a *App) Start() error {
	err := a.worker.Start()
	if err != nil {
		return cerr.Wrap(err).Error("Failed to start worker")
	}

	return nil
}

func newWorker(consumerConn *amqp.Connection, producerConn *amqp.Connection) worker.QueueWorker {
	publisher := newPublisher(producerConn)
	queueWorker, err := worker.NewQueueWorkerFromConnection(
		consumerConn,
		queueName(),
		[]worker.MessageHandler{
			newDownloadJobHandler(publisher),
			newSplitJobHandler(publisher),
			newSaveToDBJobHandler(),
		})
	ensureOk(err)
	return queueWorker
}

func queueName() string {
	return getEnvOrPanic("RABBITMQ_QUEUE_NAME")
}

func newPublisher(conn *amqp.Connection) publish.RabbitMQPublisher {
	publisher, err := publish.NewRabbitMQPublisher(conn, queueName())
	ensureOk(err)
	return publisher
}

func newGoogleFileStore() filestore.GoogleFileStore {
	jsonKey := getEnvOrPanic("GOOGLE_CLOUD_KEY")

	fileStore, err := filestore.NewGoogleFileStore(jsonKey)
	ensureOk(err)
	return fileStore
}

func newDownloadJobHandler(publisher publish.Publisher) transfer.JobHandler {
	youtubeDLBinPath := getEnvOrPanic("YOUTUBEDL_BIN_PATH")
	workingDir := getEnvOrPanic("YOUTUBEDL_WORKING_DIR_PATH")
	err := os.MkdirAll(workingDir, os.ModePerm)
	ensureOk(err)

	youtubedler := download.NewYoutubeDLer(youtubeDLBinPath, executor.BinaryFileExecutor{})
	genericdler := download.NewGenericDLer()

	selectdler := download.NewSelectDLer(youtubedler, genericdler)

	trackStore := trackstore.NewDynamoDBTrackStore(env.Get())
	bucketName := getEnvOrPanic("GOOGLE_CLOUD_STORAGE_BUCKET_NAME")
	trackDownloader, err := transfer.NewTrackTransferrer(selectdler, trackStore, newGoogleFileStore(), bucketName, workingDir)
	ensureOk(err)

	return transfer.NewJobHandler(trackDownloader, publisher)
}

func newSplitJobHandler(publisher publish.Publisher) split.JobHandler {
	workingDir := getEnvOrPanic("SPLEETER_WORKING_DIR_PATH")
	spleeterBinPath := getEnvOrPanic("SPLEETER_BIN_PATH")
	err := os.MkdirAll(workingDir, os.ModePerm)
	ensureOk(err)

	localUsecase, err := file_splitter.NewLocalFileSplitter(workingDir, spleeterBinPath, executor.BinaryFileExecutor{})
	ensureOk(err)

	googleFileStore := newGoogleFileStore()
	remoteUsecase, err := file_splitter.NewRemoteFileSplitter(workingDir, googleFileStore, localUsecase)
	ensureOk(err)

	trackStore := trackstore.NewDynamoDBTrackStore(env.Get())
	songSplitUsecase := splitter.NewTrackSplitter(remoteUsecase, trackStore, "chord-paper-tracks")

	return split.NewJobHandler(songSplitUsecase, publisher)
}

func newSaveToDBJobHandler() save_stems_to_db.JobHandler {
	trackStore := trackstore.NewDynamoDBTrackStore(env.Get())
	return save_stems_to_db.NewJobHandler(trackStore)
}
