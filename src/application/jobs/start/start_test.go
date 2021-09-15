package start_test

import (
	"chord-paper-be-workers/src/application/integration_test/dummy"
	"chord-paper-be-workers/src/application/jobs/job_message"
	"chord-paper-be-workers/src/application/jobs/transfer"
	"chord-paper-be-workers/src/application/tracks/entity"
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"chord-paper-be-workers/src/application/jobs/start"
)

var _ = Describe("Start", func() {
	var (
		dummyTrackStore *dummy.TrackStore

		handler start.JobHandler

		message []byte

		tracklistID string
		trackID     string
	)

	BeforeEach(func() {
		By("Initializing all variables", func() {
			message = nil

			tracklistID = "tracklist-id"
			trackID = "track-id"

			dummyTrackStore = dummy.NewDummyTrackStore()
		})

		By("Setting up the dummy track store data", func() {
			err := dummyTrackStore.SetTrack(context.Background(), tracklistID, trackID, entity.SplitStemTrack{
				BaseTrack: entity.BaseTrack{
					TrackType: entity.SplitFourStemsType,
				},
				OriginalURL: "",
				JobStatus:   entity.RequestedStatus,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		By("Instantiating the handler", func() {
			handler = start.NewJobHandler(dummyTrackStore)
		})
	})

	Describe("Well formed message", func() {
		var job start.JobParams
		BeforeEach(func() {
			job = start.JobParams{
				TrackIdentifier: job_message.TrackIdentifier{
					TrackListID: tracklistID,
					TrackID:     trackID,
				},
			}

			var err error
			message, err = json.Marshal(job)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("Happy path", func() {
			var err error
			var jobParams start.JobParams

			BeforeEach(func() {
				jobParams, err = handler.HandleStartJob(message)
			})

			It("doesn't return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("updates the track status", func() {
				track, err := dummyTrackStore.GetTrack(context.Background(), tracklistID, trackID)
				Expect(err).NotTo(HaveOccurred())

				splitStemTrack, ok := track.(entity.SplitStemTrack)
				Expect(ok).To(BeTrue())

				Expect(splitStemTrack.JobStatus).To(Equal(entity.ProcessingStatus))
			})

			It("returns the processed data", func() {
				Expect(jobParams.TrackListID).To(Equal(job.TrackListID))
				Expect(jobParams.TrackID).To(Equal(job.TrackID))
			})
		})

		Describe("Can't reach track store", func() {
			BeforeEach(func() {
				dummyTrackStore.Unavailable = true
			})

			It("returns an error", func() {
				_, err := handler.HandleStartJob(message)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Poorly formed message", func() {
		BeforeEach(func() {
			job := transfer.JobParams{
				TrackIdentifier: job_message.TrackIdentifier{
					TrackListID: tracklistID,
				},
			}

			var err error
			message, err = json.Marshal(job)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns error", func() {
			_, err := handler.HandleStartJob(message)
			Expect(err).To(HaveOccurred())
		})
	})
})
