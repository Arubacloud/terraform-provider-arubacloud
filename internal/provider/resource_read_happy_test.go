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
//   - status.state  — prevents WaitForResourceActive from looping (not in
//     a transitional state, so "Active" is treated as ready).
//
// All resource-specific properties are absent so nil-guard code is exercised
// rather than panicking.
const minimalActiveJSON = `{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"Active"}}`

// readHappyHandler returns a handler that serves minimalActiveJSON for every
// GET request and 500 for all other methods (write operations are not
// expected during a Read test).
func readHappyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		//nolint:errcheck
		w.Write([]byte(minimalActiveJSON))
		return
	}
	apiError(w, http.StatusInternalServerError)
}

// resourcesForReadHappy excludes resources whose Read() performs
// unsafe (non-nil-guarded) access on Properties fields that are absent from
// the minimal mock JSON.  Those resources require resource-specific response
// JSON to exercise their happy path safely.
var resourcesForReadHappy = []struct {
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

// TestResourceRead_Happy verifies that Read() succeeds (no error diagnostic)
// and keeps the resource in state when the API returns a minimal valid
// response.  Resources whose Read() has unsafe property access on nil
// pointer fields (containerregistry, cloudserver, securityrule, dbaas) are
// excluded because they require resource-specific JSON responses.
func TestResourceRead_Happy(t *testing.T) {
	// Speed up provider-side WaitForResourceActive (re-poll after Create timeout).
	// Read() can call WaitForResourceActive when it sees IsCreatingState, so 1ms
	// polling avoids a 5-second wait in the rare branch that our minimal JSON
	// does not exercise.
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	for _, tc := range resourcesForReadHappy {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, readHappyHandler)

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceReadReq(ctx, t, res)
			res.Read(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s: Read() reported error on 200 happy path: %v",
					tc.name, resp.Diagnostics)
			}
			if resp.State.Raw.IsNull() {
				t.Errorf("%s: Read() removed resource from state on 200 (expected state preserved)", tc.name)
			}
		})
	}
}

// snapshotJSON is the minimal JSON response for the snapshot datasource, which
// requires Properties.Volume (a *VolumeInfo pointer) to be non-nil to avoid
// a nil-pointer dereference at snapshot.Properties.Volume.URI.
const snapshotJSON = `{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"Active"},` +
	`"properties":{"volume":{"uri":"/projects/p/providers/Aruba.Storage/volumes/test-vol-id"}}}`

// containerregistryJSON is the minimal JSON for the containerregistry datasource.
// The datasource code dereferences BillingPlan.BillingPeriod and
// AdminUser.Username without nil guards, so both must be present.
const containerregistryJSON = `{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"Active"},` +
	`"properties":{"publicIp":{"uri":"test-pip-uri"},"vpc":{"uri":"test-vpc-uri"},` +
	`"subnet":{"uri":"test-subnet-uri"},"securityGroup":{"uri":"test-sg-uri"},` +
	`"blockStorage":{"uri":"test-bs-uri"},` +
	`"billingPlan":{"billingPeriod":"Month"},` +
	`"adminUser":{"username":"test-admin"}}}`

// TestDataSourceRead_HappyComplex covers datasources excluded from
// TestDataSourceRead_Happy because they require resource-specific response JSON.
func TestDataSourceRead_HappyComplex(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name    string
		newDS   func() datasource.DataSource
		handler http.HandlerFunc
	}{
		{
			"snapshot",
			NewSnapshotDataSource,
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(snapshotJSON)) //nolint:errcheck
			},
		},
		{
			"containerregistry",
			NewContainerRegistryDataSource,
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(containerregistryJSON)) //nolint:errcheck
			},
		},
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
			resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
			ds.Read(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s datasource: Read() reported error with targeted JSON: %v",
					tc.name, resp.Diagnostics)
			}
		})
	}
}

// TestDataSourceRead_Happy verifies that Read() for each datasource succeeds
// (no error diagnostic) when the API returns a minimal valid response.
// Snapshot and ContainerRegistry datasources are tested separately in
// TestDataSourceRead_HappyComplex using resource-specific JSON.
func TestDataSourceRead_Happy(t *testing.T) {
	ctx := context.Background()

	datasources := []struct {
		name  string
		newDS func() datasource.DataSource
	}{
		{"vpc", NewVPCDataSource},
		{"subnet", NewSubnetDataSource},
		{"securitygroup", NewSecurityGroupDataSource},
		{"securityrule", NewSecurityRuleDataSource},
		{"elasticip", NewElasticIPDataSource},
		{"keypair", NewKeypairDataSource},
		{"blockstorage", NewBlockStorageDataSource},
		{"backup", NewBackupDataSource},
		{"restore", NewRestoreDataSource},
		{"kms", NewKMSDataSource},
		{"project", NewProjectDataSource},
		{"cloudserver", NewCloudServerDataSource},
		{"vpcpeering", NewVPCPeeringDataSource},
		{"vpcpeeringroute", NewVPCPeeringRouteDataSource},
		{"vpntunnel", NewVPNTunnelDataSource},
		{"vpnroute", NewVPNRouteDataSource},
		{"dbaas", NewDBaaSDataSource},
		{"dbaasuser", NewDBaaSUserDataSource},
		{"database", NewDatabaseDataSource},
		{"databasebackup", NewDatabaseBackupDataSource},
		{"databasegrant", NewDatabaseGrantDataSource},
		{"kaas", NewKaaSDataSource},
		{"schedulejob", NewScheduleJobDataSource},
	}

	for _, tc := range datasources {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, readHappyHandler)

			ds := tc.newDS()
			configureDatasource(ctx, t, ds, mockClient)

			req := dsReadReq(ctx, t, ds, nil)
			// Initialize State with the schema so that ds.Read() can call
			// resp.State.Set(ctx, &data) without a nil-schema panic.
			schemaResp := &datasource.SchemaResponse{}
			ds.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
			resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
			ds.Read(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s datasource: Read() reported error on 200 happy path: %v",
					tc.name, resp.Diagnostics)
			}
		})
	}
}
