package aruba

type Client interface {
	FromAudit() AuditClient
	FromCompute() ComputeClient
	FromContainer() ContainerClient
	FromDatabase() DatabaseClient
	FromMetric() MetricClient
	FromNetwork() NetworkClient
	FromProject() ProjectClient
	FromSchedule() ScheduleClient
	FromSecurity() SecurityClient
	FromStorage() StorageClient
}

type clientImpl struct {
	auditClient     AuditClient
	computeClient   ComputeClient
	containerClient ContainerClient
	databaseClient  DatabaseClient
	metricsClient   MetricClient
	networkClient   NetworkClient
	projectClient   ProjectClient
	scheduleClient  ScheduleClient
	securityClient  SecurityClient
	storageClient   StorageClient
}

var _ Client = (*clientImpl)(nil)

func (c *clientImpl) FromAudit() AuditClient {
	return c.auditClient
}
func (c *clientImpl) FromCompute() ComputeClient {
	return c.computeClient
}
func (c *clientImpl) FromContainer() ContainerClient {
	return c.containerClient
}
func (c *clientImpl) FromDatabase() DatabaseClient {
	return c.databaseClient
}
func (c *clientImpl) FromMetric() MetricClient {
	return c.metricsClient
}
func (c *clientImpl) FromNetwork() NetworkClient {
	return c.networkClient
}
func (c *clientImpl) FromProject() ProjectClient {
	return c.projectClient
}
func (c *clientImpl) FromSchedule() ScheduleClient {
	return c.scheduleClient
}
func (c *clientImpl) FromSecurity() SecurityClient {
	return c.securityClient
}
func (c *clientImpl) FromStorage() StorageClient {
	return c.storageClient
}
