package application

import (
	"chord-paper-be-workers/src/application/cloud_storage/store"
	"chord-paper-be-workers/src/application/jobs/download"
	"chord-paper-be-workers/src/application/jobs/download/downloader"
	"chord-paper-be-workers/src/application/jobs/save_stems_to_db"
	trackstore "chord-paper-be-workers/src/application/jobs/save_stems_to_db/store"
	"chord-paper-be-workers/src/application/jobs/split"
	"chord-paper-be-workers/src/application/jobs/split/splitter"
	"chord-paper-be-workers/src/application/jobs/split/splitter/file_splitter"
	"chord-paper-be-workers/src/application/publish"
	"chord-paper-be-workers/src/application/worker"
	"fmt"
	"os"
	"strconv"

	"github.com/apex/log"

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
	workers []worker.QueueWorker
}

func NewApp() App {
	rabbitURL := getEnvOrPanic("RABBITMQ_URL")
	consumerConn, err := amqp.Dial(rabbitURL)
	ensureOk(err)
	producerConn, err := amqp.Dial(rabbitURL)
	ensureOk(err)

	workers := []worker.QueueWorker{}
	numWorkers := getNumWorkers()
	for i := 0; i < numWorkers; i++ {
		workers = append(workers, newWorker(consumerConn, producerConn))
	}

	return App{
		workers: workers,
	}
}

func (a *App) Start() {
	for _, queueWorker := range a.workers {
		go func() {
			err := queueWorker.Start()
			if err != nil {
				log.Error("Failed to start worker!")
			}
		}()
	}
}

func getNumWorkers() int {
	numWorkersStr := getEnvOrPanic("NUM_WORKERS")
	numWorkers, err := strconv.Atoi(numWorkersStr)
	ensureOk(err)
	return numWorkers
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

func newGoogleFileStore() store.GoogleFileStore {
	jsonKey := getEnvOrPanic("GOOGLE_CLOUD_KEY")

	fileStore, err := store.NewGoogleFileStore(jsonKey)
	ensureOk(err)
	return fileStore
}

func newDownloadJobHandler(publisher publish.Publisher) download.JobHandler {
	youtubeDLBinPath := getEnvOrPanic("YOUTUBEDL_BIN_PATH")
	workingDir := getEnvOrPanic("YOUTUBEDL_WORKING_DIR_PATH")
	err := os.MkdirAll(workingDir, os.ModePerm)
	ensureOk(err)

	dler, err := downloader.NewYoutubeDLer(youtubeDLBinPath, workingDir, newGoogleFileStore())
	ensureOk(err)

	bucketName := getEnvOrPanic("GOOGLE_CLOUD_STORAGE_BUCKET_NAME")
	trackDownloader := downloader.NewTrackDownloader(dler, bucketName)

	return download.NewJobHandler(trackDownloader, publisher)
}

func newSplitJobHandler(publisher publish.Publisher) split.JobHandler {
	workingDir := getEnvOrPanic("SPLEETER_WORKING_DIR_PATH")
	spleeterBinPath := getEnvOrPanic("SPLEETER_BIN_PATH")
	err := os.MkdirAll(workingDir, os.ModePerm)
	ensureOk(err)

	localUsecase, err := file_splitter.NewLocalFileSplitter(workingDir, spleeterBinPath)
	ensureOk(err)

	googleFileStore := newGoogleFileStore()
	remoteUsecase, err := file_splitter.NewRemoteFileSplitter(workingDir, googleFileStore, localUsecase)
	ensureOk(err)

	songSplitUsecase := splitter.NewTrackSplitter(remoteUsecase, "chord-paper-tracks")

	return split.NewJobHandler(songSplitUsecase, publisher)
}

func newSaveToDBJobHandler() save_stems_to_db.JobHandler {
	trackStore := trackstore.NewDynamoDBTrackStore()
	return save_stems_to_db.NewJobHandler(trackStore)
}
