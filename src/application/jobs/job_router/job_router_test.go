package job_router_test

import (
	"chord-paper-be-workers/src/application/integration_test/dummy"
	"chord-paper-be-workers/src/application/jobs/job_message"
	"chord-paper-be-workers/src/application/jobs/job_router"
	"chord-paper-be-workers/src/application/jobs/save_stems_to_db"
	"chord-paper-be-workers/src/application/jobs/save_stems_to_db/save_stems_to_dbfakes"
	"chord-paper-be-workers/src/application/jobs/split"
	"chord-paper-be-workers/src/application/jobs/split/splitfakes"
	"chord-paper-be-workers/src/application/jobs/split/splitter"
	"chord-paper-be-workers/src/application/jobs/start"
	"chord-paper-be-workers/src/application/jobs/start/startfakes"
	"chord-paper-be-workers/src/application/jobs/transfer"
	"chord-paper-be-workers/src/application/jobs/transfer/transferfakes"
	"chord-paper-be-workers/src/application/tracks/entity"
	"chord-paper-be-workers/src/lib/cerr"
	"context"
	"encoding/json"

	"github.com/streadway/amqp"

	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("JobRouter", func() {
	var (
		tracklistID string
		trackID     string

		startHandler     *startfakes.FakeStartJobHandler
		transferHandler  *transferfakes.FakeTransferJobHandler
		splitHandler     *splitfakes.FakeSplitJobHandler
		saveStemsHandler *save_stems_to_dbfakes.FakeSaveStemsJobHandler

		trackStore *dummy.TrackStore
		rabbitMQ   *dummy.RabbitMQ

		jobRouter job_router.JobRouter

		message     amqp.Delivery
		messageJson []byte

		WhenJobFails = func(failureSetup func()) {
			Describe("When job fails", func() {
				BeforeEach(failureSetup)

				It("updates the track to error status", func() {
					Expect(message).NotTo(BeZero())

					_ = jobRouter.HandleMessage(message)

					track, err := trackStore.GetTrack(context.Background(), tracklistID, trackID)
					Expect(err).NotTo(HaveOccurred())

					stemTrack, ok := track.(entity.SplitStemTrack)
					Expect(ok).To(BeTrue())

					Expect(stemTrack.JobStatus).To(Equal(entity.ErrorStatus))
				})

				It("returns an error", func() {
					err := jobRouter.HandleMessage(message)
					Expect(err).To(HaveOccurred())
				})

				It("doesn't publish any new jobs", func() {
					Expect(rabbitMQ.MessageChannel).To(BeEmpty())
				})
			})
		}
	)

	BeforeEach(func() {
		tracklistID = "tracklist-id"
		trackID = "track-id"
		message = amqp.Delivery{}

		var err error
		messageJson, err = json.Marshal(job_message.TrackIdentifier{
			TrackListID: tracklistID,
			TrackID:     trackID,
		})
		Expect(err).NotTo(HaveOccurred())

		By("Initializing the router", func() {
			startHandler = &startfakes.FakeStartJobHandler{}
			transferHandler = &transferfakes.FakeTransferJobHandler{}
			splitHandler = &splitfakes.FakeSplitJobHandler{}
			saveStemsHandler = &save_stems_to_dbfakes.FakeSaveStemsJobHandler{}

			trackStore = dummy.NewDummyTrackStore()
			rabbitMQ = dummy.NewRabbitMQ()

			jobRouter = job_router.NewJobRouter(trackStore, rabbitMQ, startHandler, transferHandler, splitHandler, saveStemsHandler)
		})

		By("Setting up the track store", func() {
			track := entity.SplitStemTrack{
				BaseTrack: entity.BaseTrack{
					TrackType: entity.SplitFourStemsType,
				},
				OriginalURL: "",
				JobStatus:   entity.RequestedStatus,
			}
			err := trackStore.SetTrack(context.Background(), tracklistID, trackID, track)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Start job", func() {
		BeforeEach(func() {
			message = amqp.Delivery{
				Type: start.JobType,
				Body: messageJson,
			}
		})

		Describe("When job succeeds", func() {
			BeforeEach(func() {
				startHandler.HandleStartJobReturns(start.JobParams{
					TrackIdentifier: job_message.TrackIdentifier{
						TrackListID: tracklistID,
						TrackID:     trackID,
					},
				}, nil)
			})

			It("doesn't return an error", func() {
				err := jobRouter.HandleMessage(message)
				Expect(err).NotTo(HaveOccurred())
			})

			It("publishes the next job", func() {
				_ = jobRouter.HandleMessage(message)
				Expect(rabbitMQ.MessageChannel).To(HaveLen(1))

				nextJob := <-rabbitMQ.MessageChannel
				Expect(nextJob.Type).To(Equal(transfer.JobType))

				var transferJob transfer.JobParams
				err := json.Unmarshal(nextJob.Body, &transferJob)
				Expect(err).NotTo(HaveOccurred())
				Expect(transferJob.TrackListID).To(Equal(tracklistID))
				Expect(transferJob.TrackID).To(Equal(trackID))
			})
		})

		WhenJobFails(func() {
			startHandler.HandleStartJobReturns(start.JobParams{}, cerr.Error("i failed"))
		})
	})

	Describe("Transfer job", func() {
		BeforeEach(func() {
			message = amqp.Delivery{
				Type: transfer.JobType,
				Body: messageJson,
			}
		})

		Describe("When job succeeds", func() {
			var savedOriginalURL string

			BeforeEach(func() {
				savedOriginalURL = "saved-original-url"
				transferHandler.HandleTransferJobReturns(transfer.JobParams{
					TrackIdentifier: job_message.TrackIdentifier{
						TrackListID: tracklistID,
						TrackID:     trackID,
					},
				}, savedOriginalURL, nil)
			})

			It("doesn't return an error", func() {
				err := jobRouter.HandleMessage(message)
				Expect(err).NotTo(HaveOccurred())
			})

			It("publishes the next job", func() {
				_ = jobRouter.HandleMessage(message)
				Expect(rabbitMQ.MessageChannel).To(HaveLen(1))

				nextJob := <-rabbitMQ.MessageChannel
				Expect(nextJob.Type).To(Equal(split.JobType))

				var splitJob split.JobParams
				err := json.Unmarshal(nextJob.Body, &splitJob)
				Expect(err).NotTo(HaveOccurred())
				Expect(splitJob.TrackListID).To(Equal(tracklistID))
				Expect(splitJob.TrackID).To(Equal(trackID))
				Expect(splitJob.SavedOriginalURL).To(Equal(savedOriginalURL))
			})
		})

		WhenJobFails(func() {
			transferHandler.HandleTransferJobReturns(transfer.JobParams{}, "", cerr.Error("i failed"))
		})
	})

	Describe("Split job", func() {
		BeforeEach(func() {
			message = amqp.Delivery{
				Type: split.JobType,
				Body: messageJson,
			}
		})

		Describe("When job succeeds", func() {
			var stemURLs splitter.StemFilePaths

			BeforeEach(func() {
				stemURLs = map[string]string{
					"vocals": "vocals.mp3",
					"other":  "other.mp3",
					"bass":   "bass.mp3",
					"drums":  "drums.mp3",
				}

				splitHandler.HandleSplitJobReturns(split.JobParams{
					TrackIdentifier: job_message.TrackIdentifier{
						TrackListID: tracklistID,
						TrackID:     trackID,
					},
					SavedOriginalURL: "saved-original-url",
				}, stemURLs, nil)
			})

			It("doesn't return an error", func() {
				err := jobRouter.HandleMessage(message)
				Expect(err).NotTo(HaveOccurred())
			})

			It("publishes the next job", func() {
				_ = jobRouter.HandleMessage(message)
				Expect(rabbitMQ.MessageChannel).To(HaveLen(1))

				nextJob := <-rabbitMQ.MessageChannel
				Expect(nextJob.Type).To(Equal(save_stems_to_db.JobType))

				var saveStemsJob save_stems_to_db.JobParams
				err := json.Unmarshal(nextJob.Body, &saveStemsJob)
				Expect(err).NotTo(HaveOccurred())
				Expect(saveStemsJob.TrackListID).To(Equal(tracklistID))
				Expect(saveStemsJob.TrackID).To(Equal(trackID))
				Expect(saveStemsJob.StemURLS).To(Equal(stemURLs))
			})
		})

		WhenJobFails(func() {
			splitHandler.HandleSplitJobReturns(split.JobParams{}, nil, cerr.Error("i failed"))
		})
	})

	Describe("Save stem tracks job", func() {
		BeforeEach(func() {
			message = amqp.Delivery{
				Type: save_stems_to_db.JobType,
				Body: messageJson,
			}
		})

		Describe("When job succeeds", func() {
			BeforeEach(func() {
				saveStemsHandler.HandleSaveStemsToDBJobReturns(nil)
			})

			It("doesn't return an error", func() {
				err := jobRouter.HandleMessage(message)
				Expect(err).NotTo(HaveOccurred())
			})

			It("doesn't publishes the next job", func() {
				_ = jobRouter.HandleMessage(message)
				Expect(rabbitMQ.MessageChannel).To(BeEmpty())
			})
		})

		WhenJobFails(func() {
			saveStemsHandler.HandleSaveStemsToDBJobReturns(cerr.Error("i failed"))
		})
	})
})
