package cerr

import (
	"errors"

	"github.com/apex/log"
)

type Fielder interface {
	Fields() map[string]interface{}
}

func Log(err error) {
	fields := log.Fields{}

	wrappedErr := err
	for wrappedErr != nil {
		appendField(fields, wrappedErr)
		wrappedErr = errors.Unwrap(wrappedErr)
	}
	log.WithFields(fields).Error(err.Error())
}

func appendField(fields log.Fields, err error) {
	fieldErr, ok := err.(Fielder)
	if !ok {
		return
	}

	for k, v := range fieldErr.Fields() {
		fields[k] = v
	}
}
