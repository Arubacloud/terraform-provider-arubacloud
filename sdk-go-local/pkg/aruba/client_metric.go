package aruba

import (
	"context"
)

type MetricClient interface {
	Alerts() AlertsClient
	Metrics() MetricsClient
}

type metricClientImpl struct {
	alertsClient  AlertsClient
	metricsClient MetricsClient
}

var _ MetricClient = (*metricClientImpl)(nil)

func (c *metricClientImpl) Alerts() AlertsClient {
	return c.alertsClient
}

func (c *metricClientImpl) Metrics() MetricsClient {
	return c.metricsClient
}

type AlertsClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*Alert], error)
}

type MetricsClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*Metric], error)
}
