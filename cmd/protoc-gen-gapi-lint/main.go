package main

import (
	"fmt"
	"io"
	"os"

	"github.com/felipeparaujo/protoc-gen-gapi-lint/internal/lint"
	"github.com/felipeparaujo/protoc-gen-gapi-lint/internal/lint/format"
	"github.com/jhump/protoreflect/desc"
	"github.com/spf13/pflag"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"google.golang.org/protobuf/compiler/protogen"
)

func NewFlagSet(config *lint.Config) *pflag.FlagSet {
	args := pflag.NewFlagSet("protoc-gen-gapi-lint", pflag.ExitOnError)
	args.StringVar(&config.Path, "config", "", "The linter config file.")
	args.StringVar(&config.OutputFormat, "output-format", "", "The format of the linting results.\nSupported formats include \"yaml\", \"json\", and \"summary\" table.\nYAML is the default.")
	args.StringVarP(&config.OutputPath, "output-path", "o", "", "The output file path.\nIf not given, the linting results will be printed out to STDERR.")
	args.StringArrayVar(&config.EnabledRules, "enable-rule", nil, "Enable a rule with the given name.\nMay be specified multiple times.")
	args.StringArrayVar(&config.DisabledRules, "disable-rule", nil, "Disable a rule with the given name.\nMay be specified multiple times.")
	args.BoolVar(&config.IgnoreCommentDisables, "ignore-comment-disables", false, "If set to true, disable comments will be ignored.\nThis is helpful when strict enforcement of AIPs are necessary and\nproto definitions should not be able to disable checks.")
	return args
}

func main() {
	config := &lint.Config{}

	args := NewFlagSet(config)
	handler := protogen.Options{
		ParamFunc: args.Set,
	}

	handler.Run(func(plugin *protogen.Plugin) error {
		var collection []lint.Response

		linter, err := lint.New(config)
		if err != nil {
			return err
		}

		for _, file := range plugin.Files {
			if !file.Generate {
				continue
			}

			fdesc, err := desc.WrapFile(file.Desc)
			if err != nil {
				return err
			}

			batch, err := linter.LintProtos(fdesc)
			if err != nil {
				return err
			}

			for _, item := range batch {
				if len(item.Problems) != 0 {
					collection = append(collection, item)
				}
			}
		}

		// have to read previous report if it exists otherwise we'd simply overwrite it.
		// buf splits the files across different runs, so in order to have a full report
		// we need to read the previous report and add the new errors to it.
		if len(collection) != 0 {
			previousReport, err := readReport(config)
			if err != nil {
				return fmt.Errorf("failed to read previous report: %v", err)
			}

			if err := writeReport(config, previousReport, collection); err != nil {
				return fmt.Errorf("failed to write report: %v", err)
			}

			return fmt.Errorf("found errors in %d files, see report in %s for more details", len(collection), config.OutputPath)
		}

		return nil
	})
}

func readReport(config *lint.Config) ([]format.Response, error) {
	// if no path provided, nothing to read
	if config.OutputPath == "" {
		return []format.Response{}, nil
	}

	// if file doesn't exist return empty collection
	data, err := os.ReadFile(config.OutputPath)
	if os.IsNotExist(err) {
		return []format.Response{}, nil
	}

	if err != nil {
		return nil, err
	}

	return format.Decode(data, config.OutputFormat)
}

func writeReport(config *lint.Config, previousReport []format.Response, collection []lint.Response) error {
	result, err := format.ConvertLintReponsesToLocalReponses(collection)
	if err != nil {
		return err
	}

	writer := io.WriteCloser(os.Stderr)
	if config.OutputPath != "" {
		w, err := os.Create(config.OutputPath)
		if err != nil {
			return err
		}
		writer = w
	}

	encoded, err := format.Encode(dedupeReport(previousReport, result), config.OutputFormat)
	if err != nil {
		return err
	}

	_, err = writer.Write(encoded)
	if err != nil {
		return err
	}

	return writer.Close()
}

func dedupeReport(old []format.Response, new []format.Response) []format.Response {
	responses := orderedmap.New[string, format.Response]()

	for _, r := range old {
		responses.Set(r.FilePath, r)
	}

	for _, r := range new {
		responses.Set(r.FilePath, r)
	}

	fullReport := []format.Response{}
	for resp := responses.Oldest(); resp != nil; resp = resp.Next() {
		fullReport = append(fullReport, resp.Value)
	}

	return fullReport
}
