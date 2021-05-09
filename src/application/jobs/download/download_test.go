package download_test

import (
	"chord-paper-be-workers/src/application/cloud_storage/store"
	"chord-paper-be-workers/src/application/integration_test/dummy"
	"chord-paper-be-workers/src/application/jobs/split"
	"chord-paper-be-workers/src/application/tracks/entity"
	"context"
	"fmt"

	"chord-paper-be-workers/src/application/jobs/download"
	"chord-paper-be-workers/src/application/jobs/download/downloader"
	"chord-paper-be-workers/src/application/publish/publishfakes"
	"encoding/json"

	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Download Job Handler", func() {
	var (
		youtubeDLBinPath string
		bucketName       string

		dummyTrackStore *dummy.TrackStore
		dummyFileStore  *dummy.FileStore
		dummyExecutor   *dummy.YoutubeDLExecutor
		fakePublisher   *publishfakes.FakePublisher

		handler download.JobHandler

		message           []byte
		originalURL       string
		originalTrackData []byte

		tracklistID string
		trackID     string
	)

	BeforeEach(func() {
		By("Initializing all variables", func() {
			message = nil
			youtubeDLBinPath = "/bin/youtube-dl"
			bucketName = "bucket-head"

			tracklistID = "tracklist-id"
			trackID = "track-id"
			originalURL = "https://some-third-party/coolsong.mp3"
			originalTrackData = []byte("cool_jamz")

			dummyTrackStore = dummy.NewDummyTrackStore()
			dummyFileStore = dummy.NewDummyFileStore()
			dummyExecutor = dummy.NewDummyYoutubeDLExecutor()

			fakePublisher = &publishfakes.FakePublisher{}
		})

		By("Setting up the dummy track store data", func() {
			err := dummyTrackStore.SetTrack(context.Background(), tracklistID, trackID, entity.SplitStemTrack{
				BaseTrack: entity.BaseTrack{
					TrackType: entity.SplitFourStemsType,
				},
				OriginalURL: originalURL,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		By("Setting up the dummy executor", func() {
			dummyExecutor.AddURL(originalURL, originalTrackData)
		})

		By("Instantiating the handler", func() {
			youtubeDownloader, err := downloader.NewYoutubeDLer(youtubeDLBinPath, workingDir, dummyFileStore, dummyExecutor)
			Expect(err).NotTo(HaveOccurred())
			trackDownloader := downloader.NewTrackDownloader(youtubeDownloader, dummyTrackStore, bucketName)
			handler = download.NewJobHandler(trackDownloader, fakePublisher)
		})
	})

	Describe("Well formed message", func() {
		var job download.JobParams
		BeforeEach(func() {
			job = download.JobParams{
				TrackListID: tracklistID,
				TrackID:     trackID,
			}

			var err error
			message, err = json.Marshal(job)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("Happy path", func() {
			var err error
			var expectedSavedURL string

			BeforeEach(func() {
				err = handler.HandleMessage(message)
				expectedSavedURL = fmt.Sprintf("%s/%s/%s/%s/original/original.mp3", store.GOOGLE_STORAGE_HOST, bucketName, tracklistID, trackID)
			})

			It("doesn't return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("saved the track to the file store", func() {
				contents, err := dummyFileStore.GetFile(context.Background(), expectedSavedURL)
				Expect(err).NotTo(HaveOccurred())
				Expect(contents).To(Equal(originalTrackData))
			})

			It("publishes the next job", func() {
				msg := fakePublisher.PublishArgsForCall(0)
				Expect(msg.Type).To(Equal(split.JobType))

				var splitJob split.JobParams
				err := json.Unmarshal(msg.Body, &splitJob)
				Expect(err).NotTo(HaveOccurred())
				Expect(splitJob.TrackListID).To(Equal(tracklistID))
				Expect(splitJob.TrackID).To(Equal(trackID))
				Expect(splitJob.SavedOriginalURL).To(Equal(expectedSavedURL))
			})
		})

		Describe("Can't reach track store", func() {
			BeforeEach(func() {
				dummyTrackStore.Unavailable = true
			})

			It("returns an error", func() {
				err := handler.HandleMessage(message)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Poorly formed message", func() {
		BeforeEach(func() {
			job := download.JobParams{
				TrackListID: "tracklistID",
			}

			var err error
			message, err = json.Marshal(job)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns error", func() {
			err := handler.HandleMessage(message)
			Expect(err).To(HaveOccurred())
		})
	})
})
