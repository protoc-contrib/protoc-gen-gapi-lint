package lint

import (
	"github.com/googleapis/api-linter/lint"
	"github.com/googleapis/api-linter/rules"
)

// registry contains the rules registry
var registry = lint.NewRuleRegistry()

// registry should be added to the rules
func init() {
	if err := rules.Add(registry); err != nil {
		panic(err)
	}
}

// Config represents the lint config
type Config struct {
	Path                  string
	OutputFormat          string
	OutputPath            string
	EnabledRules          []string
	DisabledRules         []string
	IgnoreCommentDisables bool
}

// Response describes the result returned by a rule.
type Response = lint.Response

// New returns a new linter
func New(config *Config) (*lint.Linter, error) {
	var options lint.Configs

	// Read linter config and append it to the default.
	if config.Path != "" {
		fconfig, err := lint.ReadConfigsFromFile(config.Path)
		if err != nil {
			return nil, err
		}

		options = append(options, fconfig...)
	}

	// Add configs for the enabled rules.
	options = append(options, lint.Config{
		EnabledRules: config.EnabledRules,
	})
	// Add configs for the disabled rules.
	options = append(options, lint.Config{
		DisabledRules: config.DisabledRules,
	})

	linter := lint.New(registry, options, lint.IgnoreCommentDisables(config.IgnoreCommentDisables))
	// so far so good :)
	return linter, nil
}
