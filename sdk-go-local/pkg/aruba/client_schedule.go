package aruba

import "context"

type ScheduleClient interface {
	Jobs() JobsClient
}

type scheduleClientImpl struct {
	jobsClient JobsClient
}

var _ ScheduleClient = (*scheduleClientImpl)(nil)

func (c *scheduleClientImpl) Jobs() JobsClient {
	return c.jobsClient
}

type JobsClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*Job], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*Job, error)
	Create(ctx context.Context, j *Job, opts ...CallOption) (*Job, error)
	Update(ctx context.Context, j *Job, opts ...CallOption) (*Job, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}
