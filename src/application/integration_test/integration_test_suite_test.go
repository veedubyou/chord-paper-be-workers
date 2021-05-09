package integration_test_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegrationTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IntegrationTest Suite")
}

var workingDir string

var _ = BeforeSuite(func() {
	workingDir = "./integration_test_wd"
	err := os.MkdirAll(workingDir, os.ModePerm)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	_ = os.RemoveAll(workingDir)
})
