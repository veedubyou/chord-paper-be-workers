package split_test

import (
	"chord-paper-be-workers/src/application/cloud_storage/store"
	"chord-paper-be-workers/src/application/integration_test/dummy"
	"chord-paper-be-workers/src/application/jobs/save_stems_to_db"
	"chord-paper-be-workers/src/application/jobs/split"
	"chord-paper-be-workers/src/application/jobs/split/splitter"
	"chord-paper-be-workers/src/application/jobs/split/splitter/file_splitter"
	"chord-paper-be-workers/src/application/publish/publishfakes"
	"chord-paper-be-workers/src/application/tracks/entity"
	"context"
	"encoding/json"
	"fmt"

	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Split handler", func() {
	var (
		bucketName string

		dummyTrackStore *dummy.TrackStore
		dummyFileStore  *dummy.FileStore
		dummyExecutor   *dummy.SpleeterExecutor
		fakePublisher   *publishfakes.FakePublisher

		handler split.JobHandler

		message           []byte
		savedOriginalURL  string
		remoteURLBase     string
		originalTrackData []byte

		tracklistID string
		trackID     string
		trackType   entity.TrackType
	)

	BeforeEach(func() {
		By("Assigning all the variables data", func() {
			tracklistID = "tracklist-ID"
			trackID = "track-ID"
			trackType = entity.InvalidType
			bucketName = "bucket-head"

			remoteURLBase = fmt.Sprintf("%s/%s/%s/%s", store.GOOGLE_STORAGE_HOST, bucketName, tracklistID, trackID)
			savedOriginalURL = fmt.Sprintf("%s/original/original.mp3", remoteURLBase)
			originalTrackData = []byte("cool_jamz")
		})

		By("Instantiating all mocks", func() {
			dummyTrackStore = dummy.NewDummyTrackStore()
			dummyFileStore = dummy.NewDummyFileStore()
			dummyExecutor = dummy.NewDummySpleeterExecutor()
			fakePublisher = &publishfakes.FakePublisher{}
		})

		By("Setting up file on the file store", func() {
			err := dummyFileStore.WriteFile(context.Background(), savedOriginalURL, originalTrackData)
			Expect(err).NotTo(HaveOccurred())
		})

		By("Instantiating the handler", func() {
			localSplitter, err := file_splitter.NewLocalFileSplitter(workingDir, "/somewhere/spleeter", dummyExecutor)
			Expect(err).NotTo(HaveOccurred())

			remoteSplitter, err := file_splitter.NewRemoteFileSplitter(workingDir, dummyFileStore, localSplitter)
			Expect(err).NotTo(HaveOccurred())

			trackSplitter := splitter.NewTrackSplitter(remoteSplitter, dummyTrackStore, bucketName)
			handler = split.NewJobHandler(trackSplitter, fakePublisher)
		})
	})

	JustBeforeEach(func() {
		Expect(trackType).NotTo(Equal(entity.InvalidType))

		prevUnavailable := dummyTrackStore.Unavailable
		dummyTrackStore.Unavailable = false

		err := dummyTrackStore.SetTrack(context.Background(), tracklistID, trackID, entity.SplitStemTrack{
			BaseTrack: entity.BaseTrack{
				TrackType: trackType,
			},
			OriginalURL: "https://whocares",
		})

		dummyTrackStore.Unavailable = prevUnavailable

		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Well formed message", func() {
		BeforeEach(func() {
			job := split.JobParams{
				TrackListID:      tracklistID,
				TrackID:          trackID,
				SavedOriginalURL: savedOriginalURL,
			}

			var err error
			message, err = json.Marshal(job)
			Expect(err).NotTo(HaveOccurred())

			// just setting something for now so that other paths
			// don't run into an error
			trackType = entity.TwoStemsType
		})

		Describe("Happy path", func() {
			var (
				err error

				expectedPublishedStemURLs map[string]string
				expectedStemFileContent   map[string][]byte

				expectPublishNextJob = func() {
					msg := fakePublisher.PublishArgsForCall(0)
					Expect(msg.Type).To(Equal(save_stems_to_db.JobType))

					var saveDBJob save_stems_to_db.JobParams
					err := json.Unmarshal(msg.Body, &saveDBJob)
					Expect(err).NotTo(HaveOccurred())
					Expect(saveDBJob.TrackListID).To(Equal(tracklistID))
					Expect(saveDBJob.TrackID).To(Equal(trackID))
					Expect(saveDBJob.StemURLS).To(Equal(expectedPublishedStemURLs))
				}

				expectUploadedStemFiles = func() {
					Expect(expectedStemFileContent).NotTo(BeEmpty())
					for stemURL, stemFileContent := range expectedStemFileContent {
						storedBytes, err := dummyFileStore.GetFile(context.Background(), stemURL)
						Expect(err).NotTo(HaveOccurred())
						Expect(storedBytes).To(Equal(stemFileContent))
					}
				}
			)

			BeforeEach(func() {
				err = nil
			})

			JustBeforeEach(func() {
				err = handler.HandleMessage(message)
			})

			Describe("2stems", func() {
				BeforeEach(func() {
					trackType = entity.SplitTwoStemsType

					vocalsURL := remoteURLBase + "/2stems/vocals.mp3"
					accompanimentURL := remoteURLBase + "/2stems/accompaniment.mp3"

					expectedPublishedStemURLs = map[string]string{
						"vocals":        vocalsURL,
						"accompaniment": accompanimentURL,
					}

					expectedStemFileContent = map[string][]byte{
						vocalsURL:        []byte(string(originalTrackData) + "-vocals"),
						accompanimentURL: []byte(string(originalTrackData) + "-accompaniment"),
					}
				})

				It("succeeds", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("publishes the next job", expectPublishNextJob)

				It("uploaded the stem files", expectUploadedStemFiles)
			})

			Describe("4stems", func() {
				BeforeEach(func() {
					trackType = entity.SplitFourStemsType

					vocalsURL := remoteURLBase + "/4stems/vocals.mp3"
					otherURL := remoteURLBase + "/4stems/other.mp3"
					bassURL := remoteURLBase + "/4stems/bass.mp3"
					drumsURL := remoteURLBase + "/4stems/drums.mp3"

					expectedPublishedStemURLs = map[string]string{
						"vocals": vocalsURL,
						"other":  otherURL,
						"bass":   bassURL,
						"drums":  drumsURL,
					}

					expectedStemFileContent = map[string][]byte{
						vocalsURL: []byte(string(originalTrackData) + "-vocals"),
						otherURL:  []byte(string(originalTrackData) + "-other"),
						bassURL:   []byte(string(originalTrackData) + "-bass"),
						drumsURL:  []byte(string(originalTrackData) + "-drums"),
					}
				})

				It("succeeds", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("publishes the next job", expectPublishNextJob)

				It("uploaded the stem files", expectUploadedStemFiles)
			})

			Describe("5stems", func() {
				BeforeEach(func() {
					trackType = entity.SplitFiveStemsType

					vocalsURL := remoteURLBase + "/5stems/vocals.mp3"
					otherURL := remoteURLBase + "/5stems/other.mp3"
					pianoURL := remoteURLBase + "/5stems/piano.mp3"
					bassURL := remoteURLBase + "/5stems/bass.mp3"
					drumsURL := remoteURLBase + "/5stems/drums.mp3"

					expectedPublishedStemURLs = map[string]string{
						"vocals": vocalsURL,
						"other":  otherURL,
						"piano":  pianoURL,
						"bass":   bassURL,
						"drums":  drumsURL,
					}

					expectedStemFileContent = map[string][]byte{
						vocalsURL: []byte(string(originalTrackData) + "-vocals"),
						otherURL:  []byte(string(originalTrackData) + "-other"),
						pianoURL:  []byte(string(originalTrackData) + "-piano"),
						bassURL:   []byte(string(originalTrackData) + "-bass"),
						drumsURL:  []byte(string(originalTrackData) + "-drums"),
					}
				})

				It("succeeds", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("publishes the next job", expectPublishNextJob)

				It("uploaded the stem files", expectUploadedStemFiles)
			})
		})

		Describe("When the file store is down", func() {
			BeforeEach(func() {
				dummyFileStore.Unavailable = true
			})

			It("returns an error", func() {
				err := handler.HandleMessage(message)
				Expect(err).To(HaveOccurred())
			})
		})

		Describe("When the track store is down", func() {
			BeforeEach(func() {
				dummyTrackStore.Unavailable = true
			})

			It("returns an error", func() {
				err := handler.HandleMessage(message)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Malformed message", func() {
		BeforeEach(func() {
			job := split.JobParams{
				TrackListID: tracklistID,
				TrackID:     trackID,
			}

			var err error
			message, err = json.Marshal(job)
			Expect(err).NotTo(HaveOccurred())

			trackType = entity.TwoStemsType
		})

		It("failaroo", func() {
			err := handler.HandleMessage(message)
			Expect(err).To(HaveOccurred())
		})
	})
})
