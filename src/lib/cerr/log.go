package cerr

import (
	"github.com/apex/log"
)

func Log(err error) {
	ctxErr, ok := err.(ContextualError)
	if !ok {
		log.Error(err.Error())
		return
	}

	log.WithFields(log.Fields(ctxErr.Context.ContextFields)).Error(err.Error())
}
