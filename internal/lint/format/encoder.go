package format

import (
	"encoding/json"

	"github.com/googleapis/api-linter/lint"
	"gopkg.in/yaml.v2"
)

func Encode(collection []lint.Response, format string) ([]byte, error) {
	switch format {
	case "yaml", "yml":
		return yaml.Marshal(collection)
	case "json":
		return json.Marshal(collection)
	default:
		return yaml.Marshal(collection)
	}
}
