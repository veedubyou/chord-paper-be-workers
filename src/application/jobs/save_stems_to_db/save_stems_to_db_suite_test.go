package save_stems_to_db_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSaveStemsToDb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SaveStemsToDb Suite")
}
