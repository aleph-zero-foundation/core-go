package bn256_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBn256(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bn256 Suite")
}
