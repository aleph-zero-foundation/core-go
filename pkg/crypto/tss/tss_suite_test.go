package tss_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTSS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TSS Suite")
}
