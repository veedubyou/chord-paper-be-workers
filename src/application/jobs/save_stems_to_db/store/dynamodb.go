package store

import (
	"chord-paper-be-workers/src/lib/werror"
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

func NewDynamoDBTrackStore() DynamoDBTrackStore {
	dbSession := session.Must(session.NewSession())
	config := aws.NewConfig().WithRegion("us-east-2").WithCredentials(credentials.NewEnvCredentials())
	client := dynamodb.New(dbSession, config)
	return DynamoDBTrackStore{
		dynamoDBClient: client,
	}
}

type DynamoDBTrackStore struct {
	dynamoDBClient *dynamodb.DynamoDB
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

func (d DynamoDBTrackStore) WriteTrackStems(ctx context.Context, trackListID string, trackID string, trackType string, stemURLs map[string]string) error {
	var err error
	for i := 0; i < 10; i++ {
		err = d.updateTrack(i, trackListID, trackID, trackType, stemURLs)
		if err == nil {
			return nil
		}
	}

	return err
}

func (d DynamoDBTrackStore) updateTrack(index int, trackListID string, trackID string, trackType string, stemURLs map[string]string) error {
	key := func() map[string]*dynamodb.AttributeValue {
		attributeValue := dynamodb.AttributeValue{}
		attributeValue.SetS(trackListID)
		return map[string]*dynamodb.AttributeValue{
			idField: &attributeValue,
		}
	}()

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
		newTrackType.SetS(trackType)

		newStemURLs := dynamodb.AttributeValue{}
		newStemURLs.SetM(convertToAttributeValues(stemURLs))

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
		return werror.WrapError("Failed to update dynamoDB item", err)
	}

	return nil
}
