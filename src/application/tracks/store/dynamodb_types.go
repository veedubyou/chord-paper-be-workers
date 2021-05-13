package store

import (
	"chord-paper-be-workers/src/lib/cerr"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func getStringField(object map[string]*dynamodb.AttributeValue, fieldKey string) (string, error) {
	stringVal, ok := object[fieldKey]
	if !ok {
		return "", cerr.Error("Missing string key on object")
	}

	if stringVal.S == nil {
		return "", cerr.Error("String value is empty")
	}

	return *stringVal.S, nil
}
