package main

import (
	"fmt"

	"github.com/felipeparaujo/protoc-gen-gapi-lint/internal/lint"
	"github.com/felipeparaujo/protoc-gen-gapi-lint/internal/lint/format"
	"github.com/jhump/protoreflect/desc"
	"github.com/spf13/pflag"
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

		if len(collection) != 0 {
			writer, err := format.NewWriter(config.OutputPath)
			if err != nil {
				return err
			}
			defer writer.Close()

			if err := format.NewEncoder(writer, config.OutputFormat).Encode(collection); err != nil {
				return err
			}

			return fmt.Errorf("found %d error(s), see report in %s for more details", len(collection), config.OutputPath)
		}

		return nil
	})
}
