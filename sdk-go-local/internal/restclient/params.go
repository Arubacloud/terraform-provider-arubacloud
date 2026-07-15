package restclient

// Common parameter types that can be reused across different API calls

// ListParams represents common pagination and filtering parameters used across list operations
type ListParams struct {
	Filter     *string
	Sort       *string
	Projection *string
	Offset     *int32
	Limit      *int32
}

// NewListParams creates a new ListParams instance
func NewListParams() *ListParams {
	return &ListParams{}
}

// WithFilter sets the filter parameter (fluent API)
func (p *ListParams) WithFilter(filter string) *ListParams {
	p.Filter = &filter
	return p
}

// WithSort sets the sort parameter (fluent API)
func (p *ListParams) WithSort(sort string) *ListParams {
	p.Sort = &sort
	return p
}

// WithProjection sets the projection parameter (fluent API)
func (p *ListParams) WithProjection(projection string) *ListParams {
	p.Projection = &projection
	return p
}

// WithLimit sets the limit parameter (fluent API)
func (p *ListParams) WithLimit(limit int32) *ListParams {
	p.Limit = &limit
	return p
}

// WithOffset sets the offset parameter (fluent API)
func (p *ListParams) WithOffset(offset int32) *ListParams {
	p.Offset = &offset
	return p
}

// WithPagination sets both offset and limit (fluent API)
func (p *ListParams) WithPagination(offset, limit int32) *ListParams {
	p.Offset = &offset
	p.Limit = &limit
	return p
}

// GetParams represents common parameters for get/retrieve operations
type GetParams struct {
	IgnoreDeletedStatus *bool
}

// NewGetParams creates a new GetParams instance
func NewGetParams() *GetParams {
	return &GetParams{}
}

// WithIgnoreDeletedStatus sets whether to ignore deleted resources (fluent API)
func (p *GetParams) WithIgnoreDeletedStatus(ignore bool) *GetParams {
	p.IgnoreDeletedStatus = &ignore
	return p
}

// DeleteParams represents common parameters for delete operations
type DeleteParams struct {
	NewDefaultResource *string
}

// NewDeleteParams creates a new DeleteParams instance
func NewDeleteParams() *DeleteParams {
	return &DeleteParams{}
}

// WithNewDefaultResource sets the replacement default resource (fluent API)
func (p *DeleteParams) WithNewDefaultResource(uri string) *DeleteParams {
	p.NewDefaultResource = &uri
	return p
}
