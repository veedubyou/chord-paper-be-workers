package integration_test_test

import (
	"bytes"
	"chord-paper-be-workers/src/application/integration_test/dummy"
	"chord-paper-be-workers/src/application/jobs/download"
	"chord-paper-be-workers/src/application/jobs/download/downloader"
	"chord-paper-be-workers/src/application/jobs/save_stems_to_db"
	"chord-paper-be-workers/src/application/jobs/split"
	"chord-paper-be-workers/src/application/jobs/split/splitter"
	"chord-paper-be-workers/src/application/jobs/split/splitter/file_splitter"
	"chord-paper-be-workers/src/application/tracks/entity"
	"chord-paper-be-workers/src/application/worker"
	"context"

	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("IntegrationTest", func() {
	var (
		tracklistID       string
		trackID           string
		originalURL       string
		originalTrackData []byte
		bucketName        string

		rabbitMQ          *dummy.RabbitMQ
		fileStore         *dummy.FileStore
		trackStore        *dummy.TrackStore
		youtubeDLExecutor *dummy.YoutubeDLExecutor
		spleeterExecutor  *dummy.SpleeterExecutor

		queueWorker worker.QueueWorker
		run         func()
	)

	BeforeEach(func() {
		By("Assigning data to variables", func() {
			tracklistID = "track-list-ID"
			trackID = "track-ID"
			originalURL = "https://third-party/jams.mp3"
			originalTrackData = []byte("cool-jamz")
			bucketName = "bucket-head"
		})

		By("Instantiating all dummies", func() {
			rabbitMQ = dummy.NewRabbitMQ()
			fileStore = dummy.NewDummyFileStore()
			trackStore = dummy.NewDummyTrackStore()
			youtubeDLExecutor = dummy.NewDummyYoutubeDLExecutor()
			spleeterExecutor = dummy.NewDummySpleeterExecutor()
		})

		By("Setting up the track store", func() {
			track := entity.SplitStemTrack{
				BaseTrack: entity.BaseTrack{
					TrackType: entity.SplitFourStemsType,
				},
				OriginalURL: originalURL,
			}
			err := trackStore.SetTrack(context.Background(), tracklistID, trackID, track)
			Expect(err).NotTo(HaveOccurred())
		})

		By("Setting up the youtubeDL executor", func() {
			youtubeDLExecutor.AddURL(originalURL, originalTrackData)
		})

		handlers := []worker.MessageHandler{}

		By("Creating the download job handler", func() {
			youtubedler, err := downloader.NewYoutubeDLer("/whatever/youtube-dl", workingDir, fileStore, youtubeDLExecutor)
			Expect(err).NotTo(HaveOccurred())
			trackDownloader := downloader.NewTrackDownloader(youtubedler, trackStore, bucketName)
			handler := download.NewJobHandler(trackDownloader, rabbitMQ)
			handlers = append(handlers, handler)
		})

		By("Creating the split job handler", func() {
			localFileSplitter, err := file_splitter.NewLocalFileSplitter(workingDir, "/whatever/spleeter", spleeterExecutor)
			Expect(err).NotTo(HaveOccurred())
			remoteFileSplitter, err := file_splitter.NewRemoteFileSplitter(workingDir, fileStore, localFileSplitter)
			Expect(err).NotTo(HaveOccurred())
			trackSplitter := splitter.NewTrackSplitter(remoteFileSplitter, trackStore, bucketName)
			handler := split.NewJobHandler(trackSplitter, rabbitMQ)
			handlers = append(handlers, handler)
		})

		By("Creating the save stems to DB job handler", func() {
			handler := save_stems_to_db.NewJobHandler(trackStore)
			handlers = append(handlers, handler)
		})

		By("Instantiating the worker", func() {
			queueWorker = worker.NewQueueWorker(rabbitMQ, "test-queue", handlers)
		})

		By("Setting up the run routine", func() {
			run = func() {
				go func() {
					err := queueWorker.Start()
					Expect(err).NotTo(HaveOccurred())
				}()

				message, err := download.CreateJobMessage(tracklistID, trackID)
				Expect(err).NotTo(HaveOccurred())
				err = rabbitMQ.Publish(message)
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})

	It("gets 3 acks", func() {
		run()

		Eventually(func() int {
			return rabbitMQ.AckCounter
		}).Should(Equal(3))
	})

	It("gets no nacks", func() {
		run()

		Consistently(func() int {
			return rabbitMQ.NackCounter
		}).Should(Equal(0))
	})

	It("uploads the data and converts the track", func() {
		run()

		Eventually(func() bool {
			track, err := trackStore.GetTrack(context.Background(), tracklistID, trackID)
			if err != nil {
				return false
			}

			stemTrack, ok := track.(entity.StemTrack)
			if !ok {
				return false
			}

			if stemTrack.TrackType != entity.FourStemsType {
				return false
			}

			if len(stemTrack.StemURLs) != 4 {
				return false
			}

			for stemName, stemURL := range stemTrack.StemURLs {
				contents, err := fileStore.GetFile(context.Background(), stemURL)
				if err != nil {
					return false
				}

				expectedContent := []byte(string(originalTrackData) + "-" + stemName)
				if bytes.Compare(contents, expectedContent) != 0 {
					return false
				}
			}

			return true
		}).Should(BeTrue())
	})
})
