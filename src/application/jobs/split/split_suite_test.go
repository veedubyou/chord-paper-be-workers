package split_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSplit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Split Suite")
}

var workingDir string

var _ = BeforeSuite(func() {
	workingDir = "./unit_test_wd"
	err := os.MkdirAll(workingDir, os.ModePerm)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	_ = os.RemoveAll(workingDir)
})
