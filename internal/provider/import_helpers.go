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
	if n <= 0 {
		return nil, fmt.Errorf("invalid segment count parameter n=%d", n)
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("expected format %q (e.g. %q) but got empty import ID", format, example)
	}
	parts := strings.Split(id, "/")
	if len(parts) != n {
		return nil, fmt.Errorf(
			"expected format %q (e.g. %q) but got %q — "+
				"incorrect number of segments (got %d, want %d)",
			format, example, id, len(parts), n,
		)
	}
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			return nil, fmt.Errorf(
				"expected format %q (e.g. %q) but got %q — "+
					"segment %d is empty",
				format, example, id, i+1,
			)
		}
		parts[i] = p
	}
	return parts, nil
}
