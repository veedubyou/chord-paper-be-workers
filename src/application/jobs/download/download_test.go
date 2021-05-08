package download_test

import (
	"chord-paper-be-workers/src/application/cloud_storage/entity/entityfakes"
	"chord-paper-be-workers/src/application/integration_test/dummy"

	"chord-paper-be-workers/src/application/executor/executorfakes"
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
		dummyTrackStore  *dummy.TrackStore
		fakeFileStore    *entityfakes.FakeFileStore
		fakePublisher    *publishfakes.FakePublisher
		fakeExecutor     *executorfakes.FakeExecutor
		handler          download.JobHandler

		message []byte
	)

	BeforeEach(func() {
		message = nil
		youtubeDLBinPath = "/bin/youtube-dl"
		bucketName = "bucket-head"
		dummyTrackStore = &dummy.TrackStore{}
		fakeFileStore = &entityfakes.FakeFileStore{}
		fakePublisher = &publishfakes.FakePublisher{}
		fakeExecutor = &executorfakes.FakeExecutor{}

		youtubeDownloader, err := downloader.NewYoutubeDLer(youtubeDLBinPath, ".", fakeFileStore, fakeExecutor)
		Expect(err).NotTo(HaveOccurred())
		trackDownloader := downloader.NewTrackDownloader(youtubeDownloader, dummyTrackStore, bucketName)
		handler = download.NewJobHandler(trackDownloader, fakePublisher)
	})

	Describe("Well formed message", func() {
		var job download.JobParams
		BeforeEach(func() {
			job = download.JobParams{
				TrackListID: "tracklistID",
				TrackID:     "trackID",
			}
		})

		Describe("", func() {})
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
