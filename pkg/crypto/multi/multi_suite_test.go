package multi_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMulti(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multi Suite")
}
