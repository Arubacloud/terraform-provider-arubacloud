package types

import "strconv"

// AcceptHeader defines model for acceptHeader.
type AcceptHeader string

type RequestParameters struct {
	Filter     *string       `json:"filter,omitempty"`
	Sort       *string       `json:"sort,omitempty"`
	Projection *string       `json:"projection,omitempty"`
	Accept     *AcceptHeader `json:"-"`
	Offset     *int32        `json:"offset,omitempty"`
	Limit      *int32        `json:"limit,omitempty"`
	APIVersion *string       `json:"api-version,omitempty"`
}

// ToQueryParams converts RequestParameters to a map of query parameters
func (r *RequestParameters) ToQueryParams() map[string]string {
	params := make(map[string]string)

	if r == nil {
		return params
	}

	if r.Filter != nil && *r.Filter != "" {
		params["filter"] = *r.Filter
	}

	if r.Sort != nil && *r.Sort != "" {
		params["sort"] = *r.Sort
	}

	if r.Projection != nil && *r.Projection != "" {
		params["projection"] = *r.Projection
	}

	if r.Offset != nil {
		params["offset"] = strconv.FormatInt(int64(*r.Offset), 10)
	}

	if r.Limit != nil {
		params["limit"] = strconv.FormatInt(int64(*r.Limit), 10)
	}

	if r.APIVersion != nil && *r.APIVersion != "" {
		params["api-version"] = *r.APIVersion
	}

	return params
}

// ToHeaders converts RequestParameters to a map of HTTP headers
func (r *RequestParameters) ToHeaders() map[string]string {
	headers := make(map[string]string)

	if r == nil {
		return headers
	}

	if r.Accept != nil && *r.Accept != "" {
		headers["Accept"] = string(*r.Accept)
	}

	return headers
}
