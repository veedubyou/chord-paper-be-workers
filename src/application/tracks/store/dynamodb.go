package store

import (
	"chord-paper-be-workers/src/application/tracks/entity"
	"chord-paper-be-workers/src/lib/cerr"
	"chord-paper-be-workers/src/lib/env"
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var (
	tableName = "TrackLists"
	idField   = "song_id"
)

const (
	jobStatusAttr         = "job_status"
	jobStatusMessageAttr  = "job_status_message"
	jobStatusDebugLogAttr = "job_status_debug_log"
	jobProgressAttr       = "job_progress"

	newTrackTypeValueName      = ":newTrackType"
	newStemURLsValueName       = ":newStemURLs"
	newStatusValueName         = ":newStatus"
	newStatusMessageValueName  = ":newStatusMessage"
	newStatusDebugLogValueName = ":newStatusDebugLog"
	newStatusProgressValueName = ":newStatusProgress"
	trackIDValueName           = ":trackID"
	MaxTrackIndex              = 10
)

var _ entity.TrackStore = DynamoDBTrackStore{}

func NewDynamoDBTrackStore(environment env.Environment) DynamoDBTrackStore {
	dbSession := session.Must(session.NewSession())

	config := aws.NewConfig().WithCredentials(credentials.NewEnvCredentials())

	switch environment {
	case env.Production:
		config = config.WithRegion("us-east-2")

	case env.Development:
		config = config.WithEndpoint("http://localhost:8000").
			WithRegion("localhost")

	default:
		panic("Unrecognized environment")
	}

	client := dynamodb.New(dbSession, config)
	return DynamoDBTrackStore{
		dynamoDBClient: client,
	}
}

type DynamoDBTrackStore struct {
	dynamoDBClient *dynamodb.DynamoDB
}

func (d DynamoDBTrackStore) GetTrack(_ context.Context, tracklistID string, trackID string) (entity.Track, error) {
	consistentRead := true
	key := makeKey(tracklistID)

	output, err := d.dynamoDBClient.GetItem(&dynamodb.GetItemInput{
		ConsistentRead: &consistentRead,
		Key:            key,
		TableName:      &tableName,
	})

	if err != nil {
		return entity.BaseTrack{}, cerr.Wrap(err).Error("Failed to get TrackList from DynamoDB")
	}

	track, err := trackFromDynamoTrackList(trackID, output.Item)
	if err != nil {
		return entity.BaseTrack{}, cerr.Wrap(err).Error("Failed to extract track from output items")
	}

	return track, nil
}

func trackFromDynamoTrackList(targetTrackID string, tracklist map[string]*dynamodb.AttributeValue) (entity.Track, error) {
	tracks, ok := tracklist["tracks"]
	if !ok || tracks.L == nil {
		return entity.BaseTrack{}, cerr.Error("Missing tracks field")
	}

	for _, trackItem := range tracks.L {
		if trackItem.M == nil {
			return entity.BaseTrack{}, cerr.Error("Track is not an object")
		}

		trackID, err := getStringField(trackItem.M, "id")
		if err != nil {
			return entity.BaseTrack{}, cerr.Wrap(err).Error("Failed to get string field")
		}

		if trackID == targetTrackID {
			return trackFromDynamoTrack(trackItem.M)
		}
	}

	return entity.BaseTrack{}, cerr.Error("No matching track IDs found")
}

func trackFromDynamoTrack(track map[string]*dynamodb.AttributeValue) (entity.Track, error) {
	trackType, err := getStringField(track, "track_type")
	if err != nil {
		return entity.BaseTrack{}, cerr.Wrap(err).Error("Failed to get track type")
	}

	switch trackType {
	case
		string(entity.TwoStemsType),
		string(entity.FourStemsType),
		string(entity.FiveStemsType):
		{
			return entity.BaseTrack{}, cerr.Error("Not implemented at the moment")
		}
	case
		string(entity.SplitTwoStemsType),
		string(entity.SplitFourStemsType),
		string(entity.SplitFiveStemsType):
		{
			return splitStemTrackFromDynamoTrack(track)
		}
	default:
		{
			return entity.BaseTrack{}, cerr.Error("Unknown track type found")
		}
	}
}

func splitStemTrackFromDynamoTrack(track map[string]*dynamodb.AttributeValue) (entity.SplitStemTrack, error) {
	trackTypeVal, err := getStringField(track, "track_type")
	if err != nil {
		return entity.SplitStemTrack{}, cerr.Wrap(err).Error("Failed to get track type")
	}

	trackType, err := entity.ConvertToTrackType(trackTypeVal)
	if err != nil {
		return entity.SplitStemTrack{}, cerr.Wrap(err).Error("Failed to convert track type string value to enum")
	}

	originalURL, err := getStringField(track, "original_url")
	if err != nil {
		return entity.SplitStemTrack{}, cerr.Wrap(err).Error("Failed to get original URL")
	}

	statusVal, err := getStringField(track, jobStatusAttr)
	if err != nil {
		return entity.SplitStemTrack{}, cerr.Wrap(err).Error("Failed to get status")
	}

	status, err := entity.ConvertToStatus(statusVal)
	if err != nil {
		return entity.SplitStemTrack{}, cerr.Wrap(err).Error("Failed to convert status")
	}

	message, err := getStringField(track, jobStatusMessageAttr)
	if err != nil {
		return entity.SplitStemTrack{}, cerr.Wrap(err).Error("Failed to get status message")
	}

	debugLog, err := getStringField(track, jobStatusDebugLogAttr)
	if err != nil {
		return entity.SplitStemTrack{}, cerr.Wrap(err).Error("Failed to get status debug log")
	}

	progress, err := getIntField(track, jobProgressAttr)
	if err != nil {
		return entity.SplitStemTrack{}, cerr.Wrap(err).Error("Failed to get progress")
	}

	return entity.SplitStemTrack{
		BaseTrack: entity.BaseTrack{
			TrackType: trackType,
		},
		OriginalURL:       originalURL,
		JobStatus:         status,
		JobStatusMessage:  message,
		JobStatusDebugLog: debugLog,
		JobProgress:       progress,
	}, nil
}

func (d DynamoDBTrackStore) SetTrack(_ context.Context, trackListID string, trackID string, track entity.Track) error {
	switch typedTrack := track.(type) {
	case entity.StemTrack:
		return d.updateStemTrack(trackListID, trackID, typedTrack)

	case entity.SplitStemTrack:
		return d.updateSplitStemTrack(trackListID, trackID, typedTrack)

	default:
		return cerr.Error("Unrecognized track type, cannot write")
	}
}

func (d DynamoDBTrackStore) UpdateTrack(ctx context.Context, trackListID string, trackID string, updater entity.TrackUpdater) error {
	track, err := d.GetTrack(ctx, trackListID, trackID)
	if err != nil {
		return cerr.Wrap(err).Error("Failed to get track from DB")
	}

	updatedTrack, err := updater(track)
	if err != nil {
		return cerr.Wrap(err).Error("Track update function failed")
	}

	err = d.SetTrack(ctx, trackListID, trackID, updatedTrack)
	if err != nil {
		return cerr.Wrap(err).Error("Failed to set the updated track")
	}

	return nil
}

func (d DynamoDBTrackStore) updateSplitStemTrack(trackListID string, trackID string, splitStemTrack entity.SplitStemTrack) error {
	var err error
	for i := 0; i < MaxTrackIndex; i++ {
		// update every track conditionally, because we're not sure which index of the tracklist it is
		if err = d.updateSplitStemTrackForIndex(i, trackListID, trackID, splitStemTrack); err == nil {
			return nil
		}
	}

	return err
}

func (d DynamoDBTrackStore) updateSplitStemTrackForIndex(index int, trackListID string, trackID string, splitStemTrack entity.SplitStemTrack) error {
	updateExpression := func() string {
		statusExpression := fmt.Sprintf("tracks[%d].%s", index, jobStatusAttr)
		statusMessageExpression := fmt.Sprintf("tracks[%d].%s", index, jobStatusMessageAttr)
		statusDebugLogExpression := fmt.Sprintf("tracks[%d].%s", index, jobStatusDebugLogAttr)
		statusProgressExpression := fmt.Sprintf("tracks[%d].%s", index, jobProgressAttr)

		val := fmt.Sprintf(
			"SET %s = %s, %s = %s, %s = %s, %s = %s",
			statusExpression, newStatusValueName,
			statusMessageExpression, newStatusMessageValueName,
			statusDebugLogExpression, newStatusDebugLogValueName,
			statusProgressExpression, newStatusProgressValueName)
		return val
	}()

	expressionAttributeValues := func() map[string]*dynamodb.AttributeValue {
		newStatus := dynamodb.AttributeValue{}
		newStatus.SetS(string(splitStemTrack.JobStatus))

		newStatusMessage := dynamodb.AttributeValue{}
		newStatusMessage.SetS(splitStemTrack.JobStatusMessage)

		newStatusDebugLog := dynamodb.AttributeValue{}
		newStatusDebugLog.SetS(splitStemTrack.JobStatusDebugLog)

		newStatusProgress := dynamodb.AttributeValue{}
		newStatusProgress.SetN(strconv.Itoa(splitStemTrack.JobProgress))

		return map[string]*dynamodb.AttributeValue{
			newStatusValueName:         &newStatus,
			newStatusMessageValueName:  &newStatusMessage,
			newStatusDebugLogValueName: &newStatusDebugLog,
			newStatusProgressValueName: &newStatusProgress,
		}
	}()

	err := d.updateTrack(index, trackListID, trackID, updateExpression, expressionAttributeValues)

	if err != nil {
		return cerr.Wrap(err).Error("Failed to update track")
	}

	return nil
}

func (d DynamoDBTrackStore) updateStemTrack(trackListID string, trackID string, stemTrack entity.StemTrack) error {
	var err error
	for i := 0; i < MaxTrackIndex; i++ {
		// update every track conditionally, because we're not sure which index of the tracklist it is
		if err = d.updateStemTrackForIndex(i, trackListID, trackID, stemTrack); err == nil {
			return nil
		}
	}

	return err
}

func (d DynamoDBTrackStore) updateStemTrackForIndex(index int, trackListID string, trackID string, stemTrack entity.StemTrack) error {
	updateExpression := func() string {
		trackTypeExpression := fmt.Sprintf("tracks[%d].track_type", index)
		stemURLsExpression := fmt.Sprintf("tracks[%d].stem_urls", index)

		setNewValuesExpression := fmt.Sprintf("SET %s = %s, %s = %s",
			trackTypeExpression, newTrackTypeValueName,
			stemURLsExpression, newStemURLsValueName,
		)

		removeJobStatusExpression := makeRemoveJobStatusExpression(index)

		return fmt.Sprintf("%s %s", setNewValuesExpression, removeJobStatusExpression)
	}()

	expressionAttributeValues := func() map[string]*dynamodb.AttributeValue {
		newTrackType := dynamodb.AttributeValue{}
		newTrackType.SetS(string(stemTrack.TrackType))

		newStemURLs := dynamodb.AttributeValue{}
		newStemURLs.SetM(convertToAttributeValues(stemTrack.StemURLs))

		return map[string]*dynamodb.AttributeValue{
			newTrackTypeValueName: &newTrackType,
			newStemURLsValueName:  &newStemURLs,
		}
	}()

	err := d.updateTrack(index, trackListID, trackID, updateExpression, expressionAttributeValues)

	if err != nil {
		return cerr.Wrap(err).Error("Failed to update track")
	}

	return nil
}

func (d DynamoDBTrackStore) updateTrack(
	index int,
	trackListID string,
	trackID string,
	updateExpression string,
	expressionAttributeValues map[string]*dynamodb.AttributeValue,
) error {
	key := makeKey(trackListID)

	conditionExpression := fmt.Sprintf("tracks[%d].id = %s", index, trackIDValueName)

	trackIDValue := dynamodb.AttributeValue{}
	trackIDValue.SetS(trackID)
	expressionAttributeValues[trackIDValueName] = &trackIDValue

	_, err := d.dynamoDBClient.UpdateItem(&dynamodb.UpdateItemInput{
		ConditionExpression:       &conditionExpression,
		ExpressionAttributeValues: expressionAttributeValues,
		Key:                       key,
		TableName:                 &tableName,
		UpdateExpression:          &updateExpression,
	})

	if err != nil {
		return cerr.Wrap(err).Error("Failed to update dynamoDB item")
	}

	return nil
}

func convertToAttributeValues(m map[string]string) map[string]*dynamodb.AttributeValue {
	output := map[string]*dynamodb.AttributeValue{}

	for k, v := range m {
		attributeValue := dynamodb.AttributeValue{}
		attributeValue.SetS(v)
		output[k] = &attributeValue
	}

	return output
}

func makeKey(key string) map[string]*dynamodb.AttributeValue {
	attributeValue := dynamodb.AttributeValue{}
	attributeValue.SetS(key)
	return map[string]*dynamodb.AttributeValue{
		idField: &attributeValue,
	}
}

func makeRemoveJobStatusExpression(trackIndex int) string {
	trackElementPrefix := fmt.Sprintf("tracks[%d]", trackIndex)
	jobStatusExpression := fmt.Sprintf("%s.%s", trackElementPrefix, jobStatusAttr)
	jobStatusMessageExpression := fmt.Sprintf("%s.%s", trackElementPrefix, jobStatusMessageAttr)
	jobStatusDebugLogExpression := fmt.Sprintf("%s.%s", trackElementPrefix, jobStatusDebugLogAttr)
	jobProgressExpression := fmt.Sprintf("%s.%s", trackElementPrefix, jobProgressAttr)

	return fmt.Sprintf("REMOVE %s, %s, %s, %s", jobStatusExpression, jobStatusMessageExpression, jobStatusDebugLogExpression, jobProgressExpression)
}
