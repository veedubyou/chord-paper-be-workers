package store

import (
	"chord-paper-be-workers/src/lib/werror"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func getStringField(object map[string]*dynamodb.AttributeValue, fieldKey string) (string, error) {
	stringVal, ok := object[fieldKey]
	if !ok {
		return "", werror.WrapError("Missing string key on object", nil)
	}

	if stringVal.S == nil {
		return "", werror.WrapError("String value is empty", nil)
	}

	return *stringVal.S, nil
}
