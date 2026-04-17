package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"

	"github.com/protoc-contrib/protoc-gen-aip-lint/internal/linter"
	"github.com/urfave/cli/v3"
	"google.golang.org/protobuf/compiler/protogen"
)

// version is set at build time via ldflags (e.g. -X main.version=0.1.0).
// When empty, the version falls back to the Go module build info.
var version string

func main() {
	app := &cli.Command{
		Name:      "protoc-gen-aip-lint",
		Usage:     "A protoc plugin for the Google API Linter",
		UsageText: "protoc-gen-aip-lint [global options]",
		ErrWriter: os.Stderr,
		Action: func(_ context.Context, _ *cli.Command) error {
			cfg := &linter.Config{}

			var (
				outputFormat  string
				outputPath    string
				setExitStatus bool
			)

			// protoc passes plugin options via the CodeGeneratorRequest parameter
			// field, not as CLI flags. ParamFunc applies each key=value pair from
			// protoc to the config.
			handler := protogen.Options{
				ParamFunc: func(name, value string) error {
					switch name {
					case "config":
						cfg.Path = value
					case "ignore-comment-disables":
						cfg.IgnoreCommentDisables = value == "" || value == "true"
					case "output-format":
						outputFormat = value
					case "output-path":
						outputPath = value
					case "set-exit-status":
						setExitStatus = value == "" || value == "true"
					default:
						if strings.HasPrefix(name, "paths") {
							return nil
						}
						return fmt.Errorf("unknown parameter: %q", name)
					}
					return nil
				},
			}

			// protogen.Options.Run reads from stdin, runs the handler, writes to
			// stdout, and calls os.Exit on failure. It does not return.
			handler.Run(func(plugin *protogen.Plugin) error {
				var collection []linter.Response

				verifier, err := linter.New(cfg)
				if err != nil {
					return err
				}

				for _, file := range plugin.Files {
					if !file.Generate {
						continue
					}

					batch, xerr := verifier.LintProtos(file.Desc)
					if xerr != nil {
						return xerr
					}

					for _, item := range batch {
						if len(item.Problems) > 0 {
							collection = append(collection, item)
						}
					}
				}

				if len(collection) == 0 {
					return nil
				}

				writer, err := linter.NewWriter(outputPath)
				if err != nil {
					return err
				}
				defer writer.Close()

				encoder, err := linter.NewEncoder(writer, outputFormat)
				if err != nil {
					return err
				}

				if err := encoder.Encode(collection); err != nil {
					return err
				}

				if setExitStatus {
					total := 0
					for _, r := range collection {
						total += len(r.Problems)
					}
					return fmt.Errorf("found %d lint problem(s) across %d file(s)", total, len(collection))
				}

				return nil
			})

			return nil
		},
	}

	if version != "" {
		app.Version = version
	} else if info, ok := debug.ReadBuildInfo(); ok {
		app.Version = info.Main.Version
	}

	ctx := context.Background()
	// start the application
	if err := app.Run(ctx, os.Args); err != nil {
		slog.ErrorContext(ctx, "The linter has encountered a fatal error", slog.Any("error", err))
		os.Exit(1)
	}
}
