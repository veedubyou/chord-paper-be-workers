package store

import (
	"chord-paper-be-workers/src/application/tracks/entity"
	"chord-paper-be-workers/src/lib/cerr"
	"chord-paper-be-workers/src/lib/env"
	"context"
	"fmt"

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
	newTrackTypeValueName = ":newTrackType"
	newStemURLsValueName  = ":newStemURLs"
	trackIDValueName      = ":trackID"
)

var _ entity.TrackStore = DynamoDBTrackStore{}

func NewDynamoDBTrackStore(environment env.Environment) DynamoDBTrackStore {
	dbSession := session.Must(session.NewSession())

	config := aws.NewConfig().WithRegion("us-east-2").WithCredentials(credentials.NewEnvCredentials())

	if environment == env.Development {
		config = config.WithEndpoint("http://localhost:8000")
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

	return entity.SplitStemTrack{
		BaseTrack: entity.BaseTrack{
			TrackType: trackType,
		},
		OriginalURL: originalURL,
	}, nil
}

func (d DynamoDBTrackStore) SetTrack(_ context.Context, trackListID string, trackID string, track entity.Track) error {
	stemTrack, ok := track.(entity.StemTrack)
	if !ok {
		return cerr.Error("Can only write stem tracks at this time")
	}

	var err error
	for i := 0; i < 10; i++ {
		err = d.updateTrack(i, trackListID, trackID, stemTrack)
		if err == nil {
			return nil
		}
	}

	return err
}

func (d DynamoDBTrackStore) updateTrack(index int, trackListID string, trackID string, stemTrack entity.StemTrack) error {
	key := makeKey(trackListID)

	updateExpression := func() string {
		trackTypeExpression := fmt.Sprintf("tracks[%d].track_type", index)
		stemURLsExpression := fmt.Sprintf("tracks[%d].stem_urls", index)

		val := fmt.Sprintf("SET %s = %s, %s = %s", trackTypeExpression, newTrackTypeValueName, stemURLsExpression, newStemURLsValueName)
		return val
	}()

	conditionExpression := fmt.Sprintf("tracks[%d].id = %s", index, trackIDValueName)

	expressionAttributeValues := func() map[string]*dynamodb.AttributeValue {
		trackIDValue := dynamodb.AttributeValue{}
		trackIDValue.SetS(trackID)

		newTrackType := dynamodb.AttributeValue{}
		newTrackType.SetS(string(stemTrack.TrackType))

		newStemURLs := dynamodb.AttributeValue{}
		newStemURLs.SetM(convertToAttributeValues(stemTrack.StemURLs))

		return map[string]*dynamodb.AttributeValue{
			newTrackTypeValueName: &newTrackType,
			newStemURLsValueName:  &newStemURLs,
			trackIDValueName:      &trackIDValue,
		}
	}()

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
