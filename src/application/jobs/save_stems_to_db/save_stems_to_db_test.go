package save_stems_to_db_test

import (
	"chord-paper-be-workers/src/application/integration_test/dummy"
	"chord-paper-be-workers/src/application/jobs/save_stems_to_db"
	"chord-paper-be-workers/src/application/tracks/entity"
	"context"
	"encoding/json"

	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("JobHandler", func() {
	var (
		tracklistID string
		trackID     string
		trackType   entity.TrackType

		dummyTrackStore *dummy.TrackStore
		handler         save_stems_to_db.JobHandler
	)

	BeforeEach(func() {
		tracklistID = "tracklist-ID"
		trackID = "track-ID"
		trackType = entity.InvalidType

		dummyTrackStore = dummy.NewDummyTrackStore()
		handler = save_stems_to_db.NewJobHandler(dummyTrackStore)
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

	Describe("Handle message", func() {
		var messageBytes []byte

		BeforeEach(func() {
			messageBytes = nil
		})

		Describe("Well formed job message", func() {
			var (
				stemURLs  map[string]string
				jobParams save_stems_to_db.JobParams
			)

			Describe("2stems", func() {
				BeforeEach(func() {
					stemURLs = map[string]string{
						"vocals":        "vocals.mp3",
						"accompaniment": "accompaniment.mp3",
					}
					jobParams = save_stems_to_db.JobParams{
						TrackListID: tracklistID,
						TrackID:     trackID,
						StemURLS:    stemURLs,
					}
					trackType = entity.SplitTwoStemsType

					var err error
					messageBytes, err = json.Marshal(jobParams)
					Expect(err).NotTo(HaveOccurred())
				})

				Describe("Store saves successfully", func() {
					It("does not error", func() {
						err := handler.HandleMessage(messageBytes)
						Expect(err).NotTo(HaveOccurred())
					})

					It("updates the track store", func() {
						_ = handler.HandleMessage(messageBytes)
						track, err := dummyTrackStore.GetTrack(context.Background(), tracklistID, trackID)
						Expect(err).NotTo(HaveOccurred())
						stemTrack, ok := track.(entity.StemTrack)
						Expect(ok).To(BeTrue())
						Expect(stemTrack.TrackType).To(Equal(entity.TwoStemsType))
						Expect(stemTrack.StemURLs).To(Equal(stemURLs))
					})
				})

				Describe("Store is unavailable", func() {
					BeforeEach(func() {
						dummyTrackStore.Unavailable = true
					})

					It("also returns a failure", func() {
						err := handler.HandleMessage(messageBytes)
						Expect(err).To(HaveOccurred())
					})
				})
			})

			Describe("4stems", func() {
				BeforeEach(func() {
					stemURLs = map[string]string{
						"vocals": "vocals.mp3",
						"other":  "other.mp3",
						"bass":   "bass.mp3",
						"drums":  "drums.mp3",
					}
					jobParams = save_stems_to_db.JobParams{
						TrackListID: tracklistID,
						TrackID:     trackID,
						StemURLS:    stemURLs,
					}
					trackType = entity.SplitFourStemsType

					var err error
					messageBytes, err = json.Marshal(jobParams)
					Expect(err).NotTo(HaveOccurred())
				})

				Describe("Store saves successfully", func() {
					It("does not error", func() {
						err := handler.HandleMessage(messageBytes)
						Expect(err).NotTo(HaveOccurred())
					})

					It("updates the track store", func() {
						_ = handler.HandleMessage(messageBytes)
						track, err := dummyTrackStore.GetTrack(context.Background(), tracklistID, trackID)
						Expect(err).NotTo(HaveOccurred())
						stemTrack, ok := track.(entity.StemTrack)
						Expect(ok).To(BeTrue())
						Expect(stemTrack.TrackType).To(Equal(entity.FourStemsType))
						Expect(stemTrack.StemURLs).To(Equal(stemURLs))
					})
				})

				Describe("Store is unavailable", func() {
					BeforeEach(func() {
						dummyTrackStore.Unavailable = true
					})

					It("also returns a failure", func() {
						err := handler.HandleMessage(messageBytes)
						Expect(err).To(HaveOccurred())
					})
				})
			})

			Describe("5stems", func() {
				BeforeEach(func() {
					stemURLs = map[string]string{
						"vocals": "vocals.mp3",
						"other":  "other.mp3",
						"bass":   "bass.mp3",
						"drums":  "drums.mp3",
						"piano":  "piano.mp3",
					}
					jobParams = save_stems_to_db.JobParams{
						TrackListID: tracklistID,
						TrackID:     trackID,
						StemURLS:    stemURLs,
					}
					trackType = entity.SplitFiveStemsType

					var err error
					messageBytes, err = json.Marshal(jobParams)
					Expect(err).NotTo(HaveOccurred())
				})

				Describe("Store saves successfully", func() {
					It("does not error", func() {
						err := handler.HandleMessage(messageBytes)
						Expect(err).NotTo(HaveOccurred())
					})

					It("updates the track store", func() {
						_ = handler.HandleMessage(messageBytes)
						track, err := dummyTrackStore.GetTrack(context.Background(), tracklistID, trackID)
						Expect(err).NotTo(HaveOccurred())
						stemTrack, ok := track.(entity.StemTrack)
						Expect(ok).To(BeTrue())
						Expect(stemTrack.TrackType).To(Equal(entity.FiveStemsType))
						Expect(stemTrack.StemURLs).To(Equal(stemURLs))
					})
				})

				Describe("Store is unavailable", func() {
					BeforeEach(func() {
						dummyTrackStore.Unavailable = true
					})

					It("also returns a failure", func() {
						err := handler.HandleMessage(messageBytes)
						Expect(err).To(HaveOccurred())
					})
				})
			})

		})

		Describe("Malformed job message", func() {
			BeforeEach(func() {
				jobParams := save_stems_to_db.JobParams{
					TrackListID: "tracklist-id",
					TrackID:     "track-id",
				}

				var err error
				messageBytes, err = json.Marshal(jobParams)
				Expect(err).NotTo(HaveOccurred())

				trackType = entity.TwoStemsType
			})

			It("returns error", func() {
				err := handler.HandleMessage(messageBytes)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
