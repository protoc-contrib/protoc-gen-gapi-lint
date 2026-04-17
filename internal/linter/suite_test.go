package linter_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLinter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Linter Suite")
}
