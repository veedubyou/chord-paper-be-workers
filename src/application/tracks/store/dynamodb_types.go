package store

import (
	"chord-paper-be-workers/src/lib/cerr"
	"strconv"

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

func getIntField(object map[string]*dynamodb.AttributeValue, fieldKey string) (int, error) {
	intVal, ok := object[fieldKey]
	if !ok {
		return 0, cerr.Error("Missing string key on object")
	}

	if intVal.N == nil {
		return 0, cerr.Error("Int value is empty")
	}

	value, err := strconv.Atoi(*intVal.N)
	if err != nil {
		return 0, cerr.Wrap(err).Error("Failed to convert dynamodb string to int")
	}

	return value, nil
}
