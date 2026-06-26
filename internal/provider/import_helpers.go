package provider

import (
	"fmt"
	"strings"
)

// parseImportID splits a composite import ID into exactly n parts on "/".
// Returns the parts slice and nil on success.
// Returns nil and an error describing the expected format on failure.
// Each part must be non-empty.
func parseImportID(id, format, example string, n int) ([]string, error) {
	parts := strings.SplitN(id, "/", n)
	if len(parts) != n {
		return nil, fmt.Errorf(
			"expected format %q (e.g. %q) but got %q — "+
				"not enough segments (got %d, want %d)",
			format, example, id, len(parts), n,
		)
	}
	for i, p := range parts {
		if p == "" {
			return nil, fmt.Errorf(
				"expected format %q (e.g. %q) but got %q — "+
					"segment %d is empty",
				format, example, id, i+1,
			)
		}
	}
	return parts, nil
}
