package linter_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/protoc-contrib/protoc-gen-aip-lint/internal/linter"
)

var _ = Describe("Linter", func() {
	Describe("New", func() {
		It("should create a linter with empty config", func() {
			l, err := linter.New(&linter.Config{})
			Expect(err).NotTo(HaveOccurred())
			Expect(l).NotTo(BeNil())
		})

		It("should return error for nonexistent config file", func() {
			_, err := linter.New(&linter.Config{
				Path: "/nonexistent/path/config.yaml",
			})
			Expect(err).To(HaveOccurred())
		})

		It("should load a valid config file", func() {
			dir := GinkgoT().TempDir()
			configPath := filepath.Join(dir, "config.yaml")

			Expect(os.WriteFile(configPath, []byte(`---
- disabled_rules:
    - core::0140::lower-snake
`), 0o644)).To(Succeed())

			l, err := linter.New(&linter.Config{Path: configPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(l).NotTo(BeNil())
		})

		It("should respect ignore-comment-disables", func() {
			l, err := linter.New(&linter.Config{
				IgnoreCommentDisables: true,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(l).NotTo(BeNil())
		})
	})
})
