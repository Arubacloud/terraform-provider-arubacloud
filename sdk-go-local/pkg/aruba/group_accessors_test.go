package aruba

import "testing"

// TestClientImpl_Accessors verifies that every clientImpl getter returns the
// same pointer that was stored at construction time.
func TestClientImpl_Accessors(t *testing.T) {
	audit := &auditClientImpl{}
	compute := &computeClientImpl{}
	container := &containerClientImpl{}
	db := &databaseClientImpl{}
	metric := &metricClientImpl{}
	network := &networkClientImpl{}
	project := &projectClientAdapter{}
	schedule := &scheduleClientImpl{}
	security := &securityClientImpl{}
	storage := &storageClientImpl{}

	c := &clientImpl{
		auditClient:     audit,
		computeClient:   compute,
		containerClient: container,
		databaseClient:  db,
		metricsClient:   metric,
		networkClient:   network,
		projectClient:   project,
		scheduleClient:  schedule,
		securityClient:  security,
		storageClient:   storage,
	}

	if c.FromAudit() != audit {
		t.Error("FromAudit() returned unexpected value")
	}
	if c.FromCompute() != compute {
		t.Error("FromCompute() returned unexpected value")
	}
	if c.FromContainer() != container {
		t.Error("FromContainer() returned unexpected value")
	}
	if c.FromDatabase() != db {
		t.Error("FromDatabase() returned unexpected value")
	}
	if c.FromMetric() != metric {
		t.Error("FromMetric() returned unexpected value")
	}
	if c.FromNetwork() != network {
		t.Error("FromNetwork() returned unexpected value")
	}
	if c.FromProject() != project {
		t.Error("FromProject() returned unexpected value")
	}
	if c.FromSchedule() != schedule {
		t.Error("FromSchedule() returned unexpected value")
	}
	if c.FromSecurity() != security {
		t.Error("FromSecurity() returned unexpected value")
	}
	if c.FromStorage() != storage {
		t.Error("FromStorage() returned unexpected value")
	}
}

// TestAuditClientImpl_Accessors covers audit.go.
func TestAuditClientImpl_Accessors(t *testing.T) {
	events := &auditEventsClientAdapter{}
	c := &auditClientImpl{eventsClient: events}
	if c.Events() != events {
		t.Error("Events() returned unexpected value")
	}
}

// TestComputeClientImpl_Accessors covers compute.go.
func TestComputeClientImpl_Accessors(t *testing.T) {
	cs := &cloudServersClientAdapter{}
	kp := &keyPairsClientAdapter{}
	c := &computeClientImpl{cloudServerClient: cs, keyPairClient: kp}
	if c.CloudServers() != cs {
		t.Error("CloudServers() returned unexpected value")
	}
	if c.KeyPairs() != kp {
		t.Error("KeyPairs() returned unexpected value")
	}
}

// TestContainerClientImpl_Accessors covers container.go.
func TestContainerClientImpl_Accessors(t *testing.T) {
	kaas := &kaasClientAdapter{}
	cr := &containerRegistriesClientAdapter{}
	c := &containerClientImpl{kaasClient: kaas, containerRegistryClient: cr}
	if c.KaaS() != kaas {
		t.Error("KaaS() returned unexpected value")
	}
	if c.ContainerRegistry() != cr {
		t.Error("ContainerRegistry() returned unexpected value")
	}
}

// TestDatabaseClientImpl_Accessors covers database.go.
func TestDatabaseClientImpl_Accessors(t *testing.T) {
	dbaas := &dbaasClientAdapter{}
	dbs := &databasesClientAdapter{}
	backups := &dbaasBackupsClientAdapter{}
	users := &usersClientAdapter{}
	grants := &grantsClientAdapter{}
	c := databaseClientImpl{
		dbaasClient:     dbaas,
		databasesClient: dbs,
		backupsClient:   backups,
		usersClient:     users,
		grantsClient:    grants,
	}
	if c.DBaaS() != dbaas {
		t.Error("DBaaS() returned unexpected value")
	}
	if c.Databases() != dbs {
		t.Error("Databases() returned unexpected value")
	}
	if c.Backups() != backups {
		t.Error("Backups() returned unexpected value")
	}
	if c.Users() != users {
		t.Error("Users() returned unexpected value")
	}
	if c.Grants() != grants {
		t.Error("Grants() returned unexpected value")
	}
}

// TestMetricClientImpl_Accessors covers metric.go.
func TestMetricClientImpl_Accessors(t *testing.T) {
	alerts := &alertsClientAdapter{}
	metrics := &metricsClientAdapter{}
	c := &metricClientImpl{alertsClient: alerts, metricsClient: metrics}
	if c.Alerts() != alerts {
		t.Error("Alerts() returned unexpected value")
	}
	if c.Metrics() != metrics {
		t.Error("Metrics() returned unexpected value")
	}
}

// TestNetworkClientImpl_Accessors covers network.go.
func TestNetworkClientImpl_Accessors(t *testing.T) {
	eip := &elasticIPsClientAdapter{}
	lb := &loadBalancersClientAdapter{}
	sgr := &securityRulesClientAdapter{}
	sg := &securityGroupsClientAdapter{}
	sn := &subnetsClientAdapter{}
	vpcpr := &vpcPeeringRoutesClientAdapter{}
	vpcpe := &vpcPeeringsClientAdapter{}
	vpcs := &vpcsClientAdapter{}
	vpnr := &vpnRoutesClientAdapter{}
	vpnt := &vpnTunnelsClientAdapter{}
	c := &networkClientImpl{
		elasticIPsClient:         eip,
		loadBalancersClient:      lb,
		securityGroupRulesClient: sgr,
		securityGroupsClient:     sg,
		subnetsClient:            sn,
		vpcPeeringRoutesClient:   vpcpr,
		vpcPeeringsClient:        vpcpe,
		vpcsClient:               vpcs,
		vpnRoutesClient:          vpnr,
		vpnTunnelsClient:         vpnt,
	}
	if c.ElasticIPs() != eip {
		t.Error("ElasticIPs() returned unexpected value")
	}
	if c.LoadBalancers() != lb {
		t.Error("LoadBalancers() returned unexpected value")
	}
	if c.SecurityGroupRules() != sgr {
		t.Error("SecurityGroupRules() returned unexpected value")
	}
	if c.SecurityGroups() != sg {
		t.Error("SecurityGroups() returned unexpected value")
	}
	if c.Subnets() != sn {
		t.Error("Subnets() returned unexpected value")
	}
	if c.VPCPeeringRoutes() != vpcpr {
		t.Error("VPCPeeringRoutes() returned unexpected value")
	}
	if c.VPCPeerings() != vpcpe {
		t.Error("VPCPeerings() returned unexpected value")
	}
	if c.VPCs() != vpcs {
		t.Error("VPCs() returned unexpected value")
	}
	if c.VPNRoutes() != vpnr {
		t.Error("VPNRoutes() returned unexpected value")
	}
	if c.VPNTunnels() != vpnt {
		t.Error("VPNTunnels() returned unexpected value")
	}
}

// TestScheduleClientImpl_Accessors covers schedule.go.
func TestScheduleClientImpl_Accessors(t *testing.T) {
	jobs := &jobsClientAdapter{}
	c := &scheduleClientImpl{jobsClient: jobs}
	if c.Jobs() != jobs {
		t.Error("Jobs() returned unexpected value")
	}
}

// TestSecurityClientImpl_Accessors covers security.go.
func TestSecurityClientImpl_Accessors(t *testing.T) {
	kms := &kmsClientAdapter{}
	keys := &keysClientAdapter{}
	kmips := &kmipsClientAdapter{}
	c := &securityClientImpl{kmsClient: kms, keysClient: keys, kmipsClient: kmips}
	if c.KMS() != kms {
		t.Error("KMS() returned unexpected value")
	}
	if c.Keys() != keys {
		t.Error("Keys() returned unexpected value")
	}
	if c.Kmips() != kmips {
		t.Error("Kmips() returned unexpected value")
	}
}

// TestStorageClientImpl_Accessors covers storage.go.
func TestStorageClientImpl_Accessors(t *testing.T) {
	snaps := &snapshotsClientAdapter{}
	vols := &volumesClientAdapter{}
	backups := &storageBackupsClientAdapter{}
	restores := &storageRestoresClientAdapter{}
	c := &storageClientImpl{
		snapshotsClient: snaps,
		volumesClient:   vols,
		backupsClient:   backups,
		restoresClient:  restores,
	}
	if c.Snapshots() != snaps {
		t.Error("Snapshots() returned unexpected value")
	}
	if c.Volumes() != vols {
		t.Error("Volumes() returned unexpected value")
	}
	if c.Backups() != backups {
		t.Error("Backups() returned unexpected value")
	}
	if c.Restores() != restores {
		t.Error("Restores() returned unexpected value")
	}
}
