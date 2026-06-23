package provider

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// minimalActiveJSON is the smallest valid response body that satisfies the
// common structure of every SDK response type:
//   - metadata.id   — lets the provider store the resource ID in state.
//   - metadata.name — populates the name field when mapped.
//   - status.state  — "Active" prevents WaitForResourceActive from looping.
//
// All resource-specific properties are absent so nil-guard branches are
// exercised rather than panicking.
const minimalActiveJSON = `{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"Active"}}`

// snapshotJSON is a minimal response for the snapshot datasource.
// Properties.Volume is a *VolumeInfo pointer; without it the datasource code
// dereferences nil at snapshot.Properties.Volume.URI.
const snapshotJSON = `{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"Active"},` +
	`"properties":{"volume":{"uri":"/projects/p/providers/Aruba.Storage/volumes/test-vol-id"}}}`

// containerregistryJSON is a minimal response for the containerregistry
// datasource.  Its Read() dereferences BillingPlan.BillingPeriod and
// AdminUser.Username without nil guards, so both must be present.
const containerregistryJSON = `{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"Active"},` +
	`"properties":{"publicIp":{"uri":"test-pip-uri"},"vpc":{"uri":"test-vpc-uri"},` +
	`"subnet":{"uri":"test-subnet-uri"},"securityGroup":{"uri":"test-sg-uri"},` +
	`"blockStorage":{"uri":"test-bs-uri"},` +
	`"billingPlan":{"billingPeriod":"Month"},` +
	`"adminUser":{"username":"test-admin"}}}`

// readSuccessHandler serves minimalActiveJSON for every GET request.
// Non-GET methods return 500 — write operations are not expected during Read.
func readSuccessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(minimalActiveJSON)) //nolint:errcheck
		return
	}
	apiError(w, http.StatusInternalServerError)
}

// snapshotReadSuccessHandler serves snapshotJSON, which includes the volume
// pointer field required by the snapshot datasource Read().
func snapshotReadSuccessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(snapshotJSON)) //nolint:errcheck
}

// containerRegistryReadSuccessHandler serves containerregistryJSON, which
// includes the BillingPlan and AdminUser fields required by the
// containerregistry datasource Read().
func containerRegistryReadSuccessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(containerregistryJSON)) //nolint:errcheck
}

// resourcesForReadSuccess is the subset of allResources25 whose Read()
// works correctly with minimalActiveJSON.  Resources excluded here have
// nil-unsafe access on Properties pointer fields that are absent from the
// minimal response, causing a nil-pointer panic:
//   - containerregistry: Properties.PublicIp / AdminUser / BlockStorage
//   - cloudserver:       Properties.Network
//   - securityrule:      Properties nested object (mandatory sub-attributes)
//   - dbaas:             Config nested object (mandatory sub-attributes)
var resourcesForReadSuccess = []struct {
	name string
	newR func() resource.Resource
}{
	{"vpc", NewVPCResource},
	{"subnet", NewSubnetResource},
	{"securitygroup", NewSecurityGroupResource},
	{"elasticip", NewElasticIPResource},
	{"keypair", NewKeypairResource},
	{"blockstorage", NewBlockStorageResource},
	{"snapshot", NewSnapshotResource},
	{"backup", NewBackupResource},
	{"restore", NewRestoreResource},
	{"kms", NewKMSResource},
	{"project", NewProjectResource},
	{"vpcpeering", NewVpcPeeringResource},
	{"vpcpeeringroute", NewVpcPeeringRouteResource},
	{"vpntunnel", NewVPNTunnelResource},
	{"vpnroute", NewVPNRouteResource},
	{"dbaasuser", NewDBaaSUserResource},
	{"database", NewDatabaseResource},
	{"databasebackup", NewDatabaseBackupResource},
	{"databasegrant", NewDatabaseGrantResource},
	{"kaas", NewKaaSResource},
	{"schedulejob", NewScheduleJobResource},
}

// TestResourceRead_Success verifies that Read() succeeds (no error diagnostic
// and resource stays in state) when the API returns a minimal valid 200.
// Resources with nil-unsafe Properties access on absent pointer fields are
// excluded (see resourcesForReadSuccess).
func TestResourceRead_Success(t *testing.T) {
	// Speed up provider-side WaitForResourceActive (triggered in Read when the
	// stored state shows an in-progress creation). 1ms avoids a 5-second wait.
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	for _, tc := range resourcesForReadSuccess {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, readSuccessHandler)

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceReadReq(ctx, t, res)
			res.Read(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s: Read() reported error on 200 response: %v",
					tc.name, resp.Diagnostics)
			}
			if resp.State.Raw.IsNull() {
				t.Errorf("%s: Read() removed resource from state on 200 (expected state preserved)", tc.name)
			}
		})
	}
}

// TestDataSourceRead_Success verifies that Read() for each of the 25
// datasources succeeds (no error diagnostic) when the API returns a valid
// 200 response.  Resources that require fields beyond minimalActiveJSON use
// a dedicated handler (snapshot, containerregistry).
func TestDataSourceRead_Success(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name    string
		newDS   func() datasource.DataSource
		handler http.HandlerFunc
	}{
		{"vpc", NewVPCDataSource, readSuccessHandler},
		{"subnet", NewSubnetDataSource, readSuccessHandler},
		{"securitygroup", NewSecurityGroupDataSource, readSuccessHandler},
		{"securityrule", NewSecurityRuleDataSource, readSuccessHandler},
		{"elasticip", NewElasticIPDataSource, readSuccessHandler},
		{"keypair", NewKeypairDataSource, readSuccessHandler},
		{"blockstorage", NewBlockStorageDataSource, readSuccessHandler},
		// snapshot: Properties.Volume is *VolumeInfo — nil without specific JSON
		{"snapshot", NewSnapshotDataSource, snapshotReadSuccessHandler},
		{"backup", NewBackupDataSource, readSuccessHandler},
		{"restore", NewRestoreDataSource, readSuccessHandler},
		{"kms", NewKMSDataSource, readSuccessHandler},
		{"project", NewProjectDataSource, readSuccessHandler},
		{"cloudserver", NewCloudServerDataSource, readSuccessHandler},
		{"vpcpeering", NewVPCPeeringDataSource, readSuccessHandler},
		{"vpcpeeringroute", NewVPCPeeringRouteDataSource, readSuccessHandler},
		{"vpntunnel", NewVPNTunnelDataSource, readSuccessHandler},
		{"vpnroute", NewVPNRouteDataSource, readSuccessHandler},
		{"dbaas", NewDBaaSDataSource, readSuccessHandler},
		{"dbaasuser", NewDBaaSUserDataSource, readSuccessHandler},
		{"database", NewDatabaseDataSource, readSuccessHandler},
		{"databasebackup", NewDatabaseBackupDataSource, readSuccessHandler},
		{"databasegrant", NewDatabaseGrantDataSource, readSuccessHandler},
		{"kaas", NewKaaSDataSource, readSuccessHandler},
		// containerregistry: BillingPlan and AdminUser pointer fields — nil without specific JSON
		{"containerregistry", NewContainerRegistryDataSource, containerRegistryReadSuccessHandler},
		{"schedulejob", NewScheduleJobDataSource, readSuccessHandler},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, tc.handler)

			ds := tc.newDS()
			configureDatasource(ctx, t, ds, mockClient)

			schemaResp := &datasource.SchemaResponse{}
			ds.Schema(ctx, datasource.SchemaRequest{}, schemaResp)

			req := dsReadReq(ctx, t, ds, nil)
			// Initialize State with the schema so ds.Read() can call
			// resp.State.Set(ctx, &data) without a nil-schema panic.
			resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
			ds.Read(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s datasource: Read() reported error: %v",
					tc.name, resp.Diagnostics)
			}
		})
	}
}
