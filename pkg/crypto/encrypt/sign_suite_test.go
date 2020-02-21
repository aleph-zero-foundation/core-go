package encrypt_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSigning(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Encryption Suite")
}
