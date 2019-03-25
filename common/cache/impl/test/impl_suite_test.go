package test_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestImpl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Impl Suite")
}
