package aruba

import "github.com/Arubacloud/sdk-go/pkg/types"

// CallOption configures an optional per-call parameter. Use these instead of
// constructing a *types.RequestParameters by hand. Options are applied in order;
// the last writer for any given field wins.
type CallOption func(*callOptions)

type callOptions struct {
	filter     *string
	sort       *string
	projection *string
	offset     *int32
	limit      *int32
	apiVersion *string
}

// WithFilter sets the server-side filter expression.
func WithFilter(f string) CallOption {
	return func(o *callOptions) { o.filter = &f }
}

// WithSort sets the sort expression.
func WithSort(s string) CallOption {
	return func(o *callOptions) { o.sort = &s }
}

// WithProjection sets the field projection.
func WithProjection(p string) CallOption {
	return func(o *callOptions) { o.projection = &p }
}

// WithLimit sets the page size.
func WithLimit(n int) CallOption {
	v := int32(n)
	return func(o *callOptions) { o.limit = &v }
}

// WithOffset sets the pagination offset.
func WithOffset(n int) CallOption {
	v := int32(n)
	return func(o *callOptions) { o.offset = &v }
}

// WithAPIVersion overrides the default API version for the call.
func WithAPIVersion(v string) CallOption {
	return func(o *callOptions) { o.apiVersion = &v }
}

// WithRawParameters seeds the call options from p. Fields in p overwrite any
// previously set options; fields that are nil in p are not written, preserving
// earlier options for those fields. Subsequent CallOption values applied after
// WithRawParameters overwrite individual fields again.
func WithRawParameters(p *types.RequestParameters) CallOption {
	return func(o *callOptions) {
		if p == nil {
			return
		}
		if p.Filter != nil {
			o.filter = p.Filter
		}
		if p.Sort != nil {
			o.sort = p.Sort
		}
		if p.Projection != nil {
			o.projection = p.Projection
		}
		if p.Offset != nil {
			o.offset = p.Offset
		}
		if p.Limit != nil {
			o.limit = p.Limit
		}
		if p.APIVersion != nil {
			o.apiVersion = p.APIVersion
		}
	}
}

// applyCallOptions applies opts in order and returns the assembled callOptions.
func applyCallOptions(opts []CallOption) callOptions {
	var o callOptions
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// toRequestParameters converts the resolved callOptions into a *types.RequestParameters.
// Never returns nil.
func (o *callOptions) toRequestParameters() *types.RequestParameters {
	return &types.RequestParameters{
		Filter:     o.filter,
		Sort:       o.sort,
		Projection: o.projection,
		Offset:     o.offset,
		Limit:      o.limit,
		APIVersion: o.apiVersion,
	}
}
