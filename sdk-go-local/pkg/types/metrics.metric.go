package types

// MetricMetadataResponse represents metadata for a metric
type MetricMetadataResponse struct {
	Field string `json:"field,omitempty"`
	Value string `json:"value,omitempty"`
}

// MetricDataResponse represents data points for a metric
type MetricDataResponse struct {
	Time    string `json:"time,omitempty"`
	Measure string `json:"measure,omitempty"`
}

// Metric represents a metric response
type MetricResponse struct {
	ReferenceID   string                   `json:"referenceId,omitempty"`
	Name          string                   `json:"name,omitempty"`
	ReferenceName string                   `json:"referenceName,omitempty"`
	Metadata      []MetricMetadataResponse `json:"metadata,omitempty"`
	Data          []MetricDataResponse     `json:"data,omitempty"`
}

// MetricList represents a list of metrics
type MetricListResponse struct {
	ListResponse
	Values []MetricResponse `json:"values"`
}
