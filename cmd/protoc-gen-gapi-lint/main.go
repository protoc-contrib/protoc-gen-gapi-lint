package main

import (
	"fmt"
	"os"

	"github.com/jhump/protoreflect/desc"
	"github.com/sensat/protoc-gen-gapi-lint/internal/lint"
	"github.com/sensat/protoc-gen-gapi-lint/internal/lint/format"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/compiler/protogen"
)

func NewFlagSet(config *lint.Config) *pflag.FlagSet {
	args := pflag.NewFlagSet("protoc-gen-gapi-lint", pflag.ExitOnError)
	args.StringVar(&config.Path, "config", "", "The linter config file.")
	args.StringVar(&config.OutputFormat, "output-format", "", "The format of the linting results.\nSupported formats include \"yaml\", \"json\", and \"summary\" table.\nYAML is the default.")
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
			res, err := format.Encode(collection, config.OutputFormat)
			if err != nil {
				return fmt.Errorf("failed to encode: %+v", err)
			}

			os.Stderr.Write(res)
			return fmt.Errorf("found errors in %d files, see report for more details", len(collection))
		}

		return nil
	})
}
