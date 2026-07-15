package aruba

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// marshalRawJSON returns v marshaled as JSON, or nil if v is nil or marshaling
// fails. The error path is unreachable for the canonical *types.XxxResponse /
// *types.XxxList shapes used by wrappers; the swallow keeps the public
// signature ergonomic ([]byte instead of ([]byte, error)).
func marshalRawJSON[T any](v *T) []byte {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

// marshalRawYAML mirrors marshalRawJSON but emits YAML via gopkg.in/yaml.v3.
func marshalRawYAML[T any](v *T) []byte {
	if v == nil {
		return nil
	}
	b, err := yaml.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}
