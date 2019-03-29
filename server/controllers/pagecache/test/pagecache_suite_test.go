package test_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPagecache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pagecache Suite")
}
