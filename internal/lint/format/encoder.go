package format

import (
	"encoding/json"
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

// Encoder represents an encoder
type Encoder interface {
	Encode(interface{}) error
}

// NewEncoder creates a new encoder
func NewEncoder(writer io.Writer, format string) Encoder {
	switch format {
	case "yaml", "yml":
		return yaml.NewEncoder(writer)
	case "json":
		return json.NewEncoder(writer)
	default:
		return yaml.NewEncoder(writer)
	}
}

// NewWriter creates a new writer
func NewWriter(path string) (io.WriteCloser, error) {
	if path != "" {
		return os.Create(path)
	}
	return io.WriteCloser(os.Stderr), nil
}
