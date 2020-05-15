package rmcbox_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRmc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rmc Suite")
}
