package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClayServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ClayServer Suite")
}
