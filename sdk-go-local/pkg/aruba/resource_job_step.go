package aruba

import (
	"fmt"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Sub-builder ----

// JobStep is a fluent builder for a single step within a Job.
// Construct with NewJobStep() and attach via Job.WithSteps.
type JobStep struct {
	errMixin
	name        *string
	resourceURI *string
	actionURI   *string
	httpVerb    *HTTPVerb
	body        *string
}

// NewJobStep returns a fresh *JobStep sub-builder for configuring a Job step.
func NewJobStep() *JobStep { return &JobStep{} }

// Named sets the step name.
func (s *JobStep) Named(name string) *JobStep { s.name = &name; return s }

// WithAction sets the action URI to invoke for this step.
func (s *JobStep) WithAction(action string) *JobStep { s.actionURI = &action; return s }

// WithVerb sets the HTTP verb for this step (e.g. POST, PUT).
func (s *JobStep) WithVerb(verb HTTPVerb) *JobStep { s.httpVerb = &verb; return s }

// WithBody sets the JSON request body for this step.
func (s *JobStep) WithBody(body string) *JobStep { s.body = &body; return s }

// Targeting sets the resource URI for this step. Errors if the ref's URI is empty.
func (s *JobStep) Targeting(res Ref) *JobStep {
	uri := res.URI()
	if uri == "" {
		s.addErr(fmt.Errorf("Targeting: empty URI"))
		return s
	}
	s.resourceURI = &uri
	return s
}

func (s *JobStep) build() types.JobStepRequest {
	out := types.JobStepRequest{}
	if s.name != nil {
		out.Name = s.name
	}
	if s.resourceURI != nil {
		out.ResourceURI = *s.resourceURI
	}
	if s.actionURI != nil {
		out.ActionURI = *s.actionURI
	}
	if s.httpVerb != nil {
		out.HttpVerb = *s.httpVerb
	}
	if s.body != nil {
		out.Body = s.body
	}
	return out
}
