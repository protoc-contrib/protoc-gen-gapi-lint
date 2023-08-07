package format

import (
	"encoding/json"

	"github.com/felipeparaujo/protoc-gen-gapi-lint/internal/lint"
	"gopkg.in/yaml.v2"
)

// This is a hack to get Problems to unmarshal correctly because
// the Problem type contains a protobuf field which can't be
// unmarshaled using yaml/json libs
type Response struct {
	FilePath string        `json:"file_path" yaml:"file_path"`
	Problems []interface{} `json:"problems" yaml:"problems"`
}

func Decode(data []byte, format string) ([]Response, error) {
	var collection []Response
	switch format {
	case "yaml", "yml":
		return collection, yaml.Unmarshal(data, &collection)
	case "json":
		return collection, json.Unmarshal(data, &collection)
	default:
		return collection, yaml.Unmarshal(data, &collection)
	}
}

func Encode(collection []Response, format string) ([]byte, error) {
	switch format {
	case "yaml", "yml":
		return yaml.Marshal(collection)
	case "json":
		return json.Marshal(collection)
	default:
		return yaml.Marshal(collection)
	}
}

func ConvertLintReponsesToLocalReponses(collection []lint.Response) ([]Response, error) {
	result := []Response{}

	for _, response := range collection {
		r := Response{
			FilePath: response.FilePath,
		}

		for _, problem := range response.Problems {
			result, err := problem.MarshalJSON()
			if err != nil {
				return nil, err
			}
			var remarshaled interface{}
			if err := json.Unmarshal(result, &remarshaled); err != nil {
				return nil, err
			}

			r.Problems = append(r.Problems, remarshaled)
		}

		result = append(result, r)
	}

	return result, nil
}
