package aruba

import "context"

type EventsClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*AuditEvent], error)
}

type AuditClient interface {
	Events() EventsClient
}

type auditClientImpl struct {
	eventsClient EventsClient
}

var _ AuditClient = (*auditClientImpl)(nil)

func (c *auditClientImpl) Events() EventsClient {
	return c.eventsClient
}
