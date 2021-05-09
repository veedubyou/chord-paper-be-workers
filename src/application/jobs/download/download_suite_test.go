package download_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDownload(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Download Suite")
}

var workingDir string

var _ = BeforeSuite(func() {
	workingDir = "./unit_test_wd"
	err := os.MkdirAll(workingDir, os.ModePerm)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := os.RemoveAll(workingDir)
	Expect(err).NotTo(HaveOccurred())
})
