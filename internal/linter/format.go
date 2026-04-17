package linter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/googleapis/api-linter/v2/lint"
	"gopkg.in/yaml.v3"
)

// Encoder encodes lint responses to a writer.
type Encoder interface {
	// Encode writes the given value to the underlying writer
	// in the encoder's format.
	Encode(any) error
}

// NewEncoder returns an [Encoder] for the given format.
//
// Supported formats:
//   - "yaml" or "yml": YAML output (default when format is empty)
//   - "json": JSON output
//   - "github": GitHub Actions workflow command annotations
//   - "summary": human-readable table with problem counts per rule
//
// Returns an error for unrecognized formats.
func NewEncoder(writer io.Writer, format string) (Encoder, error) {
	switch format {
	case "yaml", "yml", "":
		return yaml.NewEncoder(writer), nil
	case "json":
		return json.NewEncoder(writer), nil
	case "github":
		return &GithubEncoder{w: writer}, nil
	case "summary":
		return &SummaryEncoder{w: writer}, nil
	default:
		return nil, fmt.Errorf("unsupported output format: %q (supported: yaml, json, github, summary)", format)
	}
}

// NewWriter returns an [io.WriteCloser] for the given path.
// If path is empty, [os.Stderr] is returned (stdout is reserved
// for the protoc wire protocol).
func NewWriter(path string) (io.WriteCloser, error) {
	if path == "" {
		return os.Stderr, nil
	}

	writer, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return writer, nil
}

// GithubEncoder emits lint problems as GitHub Actions workflow commands.
// Each problem produces an "::error" annotation that appears inline on
// pull request diffs.
type GithubEncoder struct {
	w io.Writer
}

// Encode writes each problem as a GitHub Actions error annotation.
func (e *GithubEncoder) Encode(v any) error {
	responses, ok := v.([]lint.Response)
	if !ok {
		return fmt.Errorf("github encoder: expected []lint.Response, got %T", v)
	}

	for _, resp := range responses {
		for _, p := range resp.Problems {
			line, col := 1, 1

			if p.Location != nil && len(p.Location.Span) >= 3 {
				line = int(p.Location.Span[0]) + 1
				col = int(p.Location.Span[1]) + 1
			} else if p.Descriptor != nil {
				loc := p.Descriptor.ParentFile().SourceLocations().ByDescriptor(p.Descriptor)
				line = loc.StartLine + 1
				col = loc.StartColumn + 1
			}

			fmt.Fprintf(e.w, "::error file=%s,line=%d,col=%d::%s: %s\n",
				resp.FilePath, line, col, p.RuleID, p.Message)
		}
	}

	return nil
}

// SummaryEncoder produces a human-readable table that aggregates
// problem counts per rule, with a total at the bottom.
type SummaryEncoder struct {
	w io.Writer
}

// Encode writes a tabular summary of lint problems grouped by rule.
func (e *SummaryEncoder) Encode(v any) error {
	responses, ok := v.([]lint.Response)
	if !ok {
		return fmt.Errorf("summary encoder: expected []lint.Response, got %T", v)
	}

	ruleCounts := make(map[string]int)
	fileCounts := make(map[string]int)
	total := 0

	for _, resp := range responses {
		fileCounts[resp.FilePath] += len(resp.Problems)
		for _, p := range resp.Problems {
			ruleCounts[string(p.RuleID)]++
			total++
		}
	}

	tw := tabwriter.NewWriter(e.w, 0, 4, 2, ' ', 0)

	fmt.Fprintf(tw, "RULE\tCOUNT\n")
	fmt.Fprintf(tw, "%s\t%s\n", strings.Repeat("-", 40), strings.Repeat("-", 5))

	ruleIDs := make([]string, 0, len(ruleCounts))
	for id := range ruleCounts {
		ruleIDs = append(ruleIDs, id)
	}
	sort.Strings(ruleIDs)

	for _, id := range ruleIDs {
		fmt.Fprintf(tw, "%s\t%d\n", id, ruleCounts[id])
	}

	fmt.Fprintf(tw, "%s\t%s\n", strings.Repeat("-", 40), strings.Repeat("-", 5))
	fmt.Fprintf(tw, "Total (%d files)\t%d\n", len(fileCounts), total)

	return tw.Flush()
}
