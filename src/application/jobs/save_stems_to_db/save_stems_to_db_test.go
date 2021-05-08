package save_stems_to_db_test

import (
	"chord-paper-be-workers/src/application/jobs/save_stems_to_db"
	entityfakes2 "chord-paper-be-workers/src/application/track_store/entity/entityfakes"
	"encoding/json"
	"errors"

	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("JobHandler", func() {
	var (
		fakeStore *entityfakes2.FakeTrackStore
		handler   save_stems_to_db.JobHandler
	)

	BeforeEach(func() {
		fakeStore = &entityfakes2.FakeTrackStore{}
		handler = save_stems_to_db.NewJobHandler(fakeStore)
	})

	Describe("Handle message", func() {
		var messageBytes []byte

		BeforeEach(func() {
			messageBytes = nil
		})

		Describe("Well formed job message", func() {
			var (
				jobParams save_stems_to_db.JobParams
			)

			BeforeEach(func() {
				jobParams = save_stems_to_db.JobParams{
					TrackListID:  "tracklist-id",
					TrackID:      "track-id",
					NewTrackType: "4stems",
					StemURLS: map[string]string{
						"bass":   "bass.mp3",
						"drums":  "drums.mp3",
						"vocals": "vocals.mp3",
						"other":  "other.mp3",
					},
				}

				var err error
				messageBytes, err = json.Marshal(jobParams)
				Expect(err).NotTo(HaveOccurred())
			})

			Describe("Store saves successfully", func() {
				BeforeEach(func() {
					fakeStore.WriteTrackStemsReturns(nil)
				})

				It("also returns a failure", func() {
					err := handler.HandleMessage(messageBytes)
					Expect(err).NotTo(HaveOccurred())

					_, receivedTrackListID, receivedTrackID, receivedTrackType, receivedStemURLs := fakeStore.WriteTrackStemsArgsForCall(0)
					Expect(receivedTrackListID).To(Equal(jobParams.TrackListID))
					Expect(receivedTrackID).To(Equal(jobParams.TrackID))
					Expect(receivedTrackType).To(Equal(jobParams.NewTrackType))
					Expect(receivedStemURLs).To(Equal(jobParams.StemURLS))
				})
			})

			Describe("Store fails to saves", func() {
				BeforeEach(func() {
					fakeStore.WriteTrackStemsReturns(errors.New("i fail"))
				})

				It("also returns a failure", func() {
					err := handler.HandleMessage(messageBytes)
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Describe("Malformed job message", func() {
			BeforeEach(func() {
				jobParams := save_stems_to_db.JobParams{
					TrackListID: "tracklist-id",
					TrackID:     "track-id",
					StemURLS:    map[string]string{"bass": "bass.mp3"},
				}

				var err error
				messageBytes, err = json.Marshal(jobParams)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns error", func() {
				err := handler.HandleMessage(messageBytes)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
