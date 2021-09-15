package job_router_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestJobRouter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "JobRouter Suite")
}
