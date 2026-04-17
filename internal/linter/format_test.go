package linter_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/googleapis/api-linter/v2/lint"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/protoc-contrib/protoc-gen-aip-lint/internal/linter"
)

func testResponses() []lint.Response {
	return []lint.Response{
		{
			FilePath: "example.proto",
			Problems: []lint.Problem{
				{
					Message: "Field must be in lower_snake_case.",
					RuleID:  "core::0140::lower-snake",
					Location: &descriptorpb.SourceCodeInfo_Location{
						Span: []int32{9, 2, 30},
					},
				},
				{
					Message: "Resource must define a pattern.",
					RuleID:  "core::0123::resource-pattern",
					Location: &descriptorpb.SourceCodeInfo_Location{
						Span: []int32{3, 0, 8, 1},
					},
				},
			},
		},
	}
}

var _ = Describe("Encoder", func() {
	Describe("NewEncoder", func() {
		DescribeTable("supported formats",
			func(f string) {
				var buf bytes.Buffer
				enc, err := linter.NewEncoder(&buf, f)
				Expect(err).NotTo(HaveOccurred())
				Expect(enc).NotTo(BeNil())
			},
			Entry("yaml", "yaml"),
			Entry("yml", "yml"),
			Entry("json", "json"),
			Entry("github", "github"),
			Entry("summary", "summary"),
			Entry("empty (defaults to yaml)", ""),
		)

		It("should return error for unsupported format", func() {
			var buf bytes.Buffer
			_, err := linter.NewEncoder(&buf, "xml")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unsupported output format"))
		})
	})

	Describe("YAML encoder", func() {
		It("should encode responses as YAML", func() {
			var buf bytes.Buffer
			enc, err := linter.NewEncoder(&buf, "yaml")
			Expect(err).NotTo(HaveOccurred())

			Expect(enc.Encode(testResponses())).To(Succeed())

			output := buf.String()
			Expect(output).To(ContainSubstring("example.proto"))
			Expect(output).To(ContainSubstring("lower_snake_case"))
		})
	})

	Describe("JSON encoder", func() {
		It("should encode responses as valid JSON", func() {
			var buf bytes.Buffer
			enc, err := linter.NewEncoder(&buf, "json")
			Expect(err).NotTo(HaveOccurred())

			Expect(enc.Encode(testResponses())).To(Succeed())

			var result []any
			Expect(json.Unmarshal(buf.Bytes(), &result)).To(Succeed())
			Expect(result).To(HaveLen(1))
		})
	})

	Describe("GitHub encoder", func() {
		It("should emit GitHub Actions annotations", func() {
			var buf bytes.Buffer
			enc, err := linter.NewEncoder(&buf, "github")
			Expect(err).NotTo(HaveOccurred())

			Expect(enc.Encode(testResponses())).To(Succeed())

			lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
			Expect(lines).To(HaveLen(2))

			Expect(lines[0]).To(Equal(
				"::error file=example.proto,line=10,col=3::core::0140::lower-snake: Field must be in lower_snake_case.",
			))
			Expect(lines[1]).To(Equal(
				"::error file=example.proto,line=4,col=1::core::0123::resource-pattern: Resource must define a pattern.",
			))
		})
	})

	Describe("Summary encoder", func() {
		It("should produce a summary table", func() {
			var buf bytes.Buffer
			enc, err := linter.NewEncoder(&buf, "summary")
			Expect(err).NotTo(HaveOccurred())

			Expect(enc.Encode(testResponses())).To(Succeed())

			output := buf.String()
			Expect(output).To(ContainSubstring("core::0140::lower-snake"))
			Expect(output).To(ContainSubstring("core::0123::resource-pattern"))
			Expect(output).To(ContainSubstring("Total"))
		})
	})
})

var _ = Describe("Writer", func() {
	Describe("NewWriter", func() {
		It("should return stderr for empty path", func() {
			w, err := linter.NewWriter("")
			Expect(err).NotTo(HaveOccurred())
			Expect(w).To(Equal(os.Stderr))
		})

		It("should create a file for non-empty path", func() {
			path := filepath.Join(GinkgoT().TempDir(), "output.yaml")
			w, err := linter.NewWriter(path)
			Expect(err).NotTo(HaveOccurred())
			defer w.Close()

			_, err = w.Write([]byte("test"))
			Expect(err).NotTo(HaveOccurred())
			w.Close()

			data, err := os.ReadFile(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("test"))
		})
	})
})
