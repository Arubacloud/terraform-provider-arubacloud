package aruba

import (
	"context"
	"fmt"
)

// ProjectClient is the wrapper-based public surface for project CRUD.
type ProjectClient interface {
	Create(ctx context.Context, p *Project, opts ...CallOption) (*Project, error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*Project, error)
	Update(ctx context.Context, p *Project, opts ...CallOption) (*Project, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
	List(ctx context.Context, opts ...CallOption) (*List[*Project], error)
}

// projectIDFromRef extracts a project ID from a Ref, preferring the typed
// withProjectID assertion and falling back to URI path parsing.
func projectIDFromRef(ref Ref) (string, error) {
	id, ok := extractID(ref, func(r Ref) (string, bool) {
		if p, ok := r.(withProjectID); ok {
			return p.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || id == "" {
		return "", fmt.Errorf("cannot determine project ID from Ref %q", ref.URI())
	}
	return id, nil
}
