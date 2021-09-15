package split_test

import (
	"chord-paper-be-workers/src/application/cloud_storage/store"
	"chord-paper-be-workers/src/application/integration_test/dummy"
	"chord-paper-be-workers/src/application/jobs/job_message"
	"chord-paper-be-workers/src/application/jobs/split"
	"chord-paper-be-workers/src/application/jobs/split/splitter"
	"chord-paper-be-workers/src/application/jobs/split/splitter/file_splitter"
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
			handler = split.NewJobHandler(trackSplitter)
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
				TrackIdentifier: job_message.TrackIdentifier{
					TrackListID: tracklistID,
					TrackID:     trackID,
				},
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
				err               error
				returnedJobParams split.JobParams
				returnedStemUrls  splitter.StemFilePaths

				expectedReturnedStemUrls splitter.StemFilePaths
				expectedStemFileContent  map[string][]byte

				expectUploadedStemFiles = func() {
					Expect(expectedStemFileContent).NotTo(BeEmpty())
					for stemURL, stemFileContent := range expectedStemFileContent {
						storedBytes, err := dummyFileStore.GetFile(context.Background(), stemURL)
						Expect(err).NotTo(HaveOccurred())
						Expect(storedBytes).To(Equal(stemFileContent))
					}
				}

				expectReturnValues = func() {
					Expect(returnedJobParams.TrackListID).To(Equal(tracklistID))
					Expect(returnedJobParams.TrackID).To(Equal(trackID))
					Expect(returnedStemUrls).To(Equal(expectedReturnedStemUrls))
				}
			)

			BeforeEach(func() {
				err = nil
				returnedJobParams = split.JobParams{}
				returnedStemUrls = nil
			})

			JustBeforeEach(func() {
				returnedJobParams, returnedStemUrls, err = handler.HandleSplitJob(message)
			})

			Describe("2stems", func() {
				BeforeEach(func() {
					trackType = entity.SplitTwoStemsType

					vocalsURL := remoteURLBase + "/2stems/vocals.mp3"
					accompanimentURL := remoteURLBase + "/2stems/accompaniment.mp3"

					expectedReturnedStemUrls = map[string]string{
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

				It("uploaded the stem files", expectUploadedStemFiles)

				It("returns the right values", expectReturnValues)
			})

			Describe("4stems", func() {
				BeforeEach(func() {
					trackType = entity.SplitFourStemsType

					vocalsURL := remoteURLBase + "/4stems/vocals.mp3"
					otherURL := remoteURLBase + "/4stems/other.mp3"
					bassURL := remoteURLBase + "/4stems/bass.mp3"
					drumsURL := remoteURLBase + "/4stems/drums.mp3"

					expectedReturnedStemUrls = map[string]string{
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

				It("uploaded the stem files", expectUploadedStemFiles)

				It("returns the right values", expectReturnValues)
			})

			Describe("5stems", func() {
				BeforeEach(func() {
					trackType = entity.SplitFiveStemsType

					vocalsURL := remoteURLBase + "/5stems/vocals.mp3"
					otherURL := remoteURLBase + "/5stems/other.mp3"
					pianoURL := remoteURLBase + "/5stems/piano.mp3"
					bassURL := remoteURLBase + "/5stems/bass.mp3"
					drumsURL := remoteURLBase + "/5stems/drums.mp3"

					expectedReturnedStemUrls = map[string]string{
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

				It("uploaded the stem files", expectUploadedStemFiles)

				It("returns the right values", expectReturnValues)
			})
		})

		Describe("When the file store is down", func() {
			BeforeEach(func() {
				dummyFileStore.Unavailable = true
			})

			It("returns an error", func() {
				_, _, err := handler.HandleSplitJob(message)
				Expect(err).To(HaveOccurred())
			})
		})

		Describe("When the track store is down", func() {
			BeforeEach(func() {
				dummyTrackStore.Unavailable = true
			})

			It("returns an error", func() {
				_, _, err := handler.HandleSplitJob(message)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Malformed message", func() {
		BeforeEach(func() {
			job := split.JobParams{
				TrackIdentifier: job_message.TrackIdentifier{
					TrackListID: tracklistID,
					TrackID:     trackID,
				},
			}

			var err error
			message, err = json.Marshal(job)
			Expect(err).NotTo(HaveOccurred())

			trackType = entity.TwoStemsType
		})

		It("failaroo", func() {
			_, _, err := handler.HandleSplitJob(message)
			Expect(err).To(HaveOccurred())
		})
	})
})
