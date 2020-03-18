package solidity

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/xlab/suplog"
)

func TestLoran(t *testing.T) {
	// avoid errors from suites that would try to break things
	log.DefaultLogger.SetLevel(log.FatalLevel)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Loran Test Suite")
}