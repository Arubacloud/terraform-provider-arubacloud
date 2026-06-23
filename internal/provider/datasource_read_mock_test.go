package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// dsReadReq builds a datasource.ReadRequest whose Config satisfies the
// datasource schema: Required string attributes are set to
// "test-<attr-name>"; everything else is null.  extras overrides specific
// attribute values (use to force empty strings for missing-ID tests).
func dsReadReq(ctx context.Context, t *testing.T, ds datasource.DataSource, extras map[string]string) datasource.ReadRequest {
	t.Helper()

	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(ctx, datasource.SchemaRequest{}, schemaResp)

	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("dsReadReq: datasource schema root is not an object type")
	}

	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, ty := range objType.AttributeTypes {
		if override, found := extras[name]; found {
			attrs[name] = tftypes.NewValue(tftypes.String, override)
			continue
		}
		schAttr, exists := schemaResp.Schema.Attributes[name]
		if exists && schAttr.IsRequired() && ty.Is(tftypes.String) {
			attrs[name] = tftypes.NewValue(tftypes.String, "test-"+name)
		} else {
			attrs[name] = tftypes.NewValue(ty, nil)
		}
	}

	return datasource.ReadRequest{
		Config: tfsdk.Config{
			Raw:    tftypes.NewValue(objType, attrs),
			Schema: schemaResp.Schema,
		},
	}
}

// configureDatasource injects client into ds via its Configure() method.
func configureDatasource(ctx context.Context, t *testing.T, ds datasource.DataSource, client *ArubaCloudClient) {
	t.Helper()
	cfg, ok := ds.(datasource.DataSourceWithConfigure)
	if !ok {
		t.Fatalf("configureDatasource: %T does not implement DataSourceWithConfigure", ds)
	}
	resp := &datasource.ConfigureResponse{}
	cfg.Configure(ctx, datasource.ConfigureRequest{ProviderData: client}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("configureDatasource: %v", resp.Diagnostics)
	}
}

// TestDataSourceRead_APIErrors exercises the CheckResponse error branch of
// Read() for several representative datasources.  A single mock HTTP server
// returns 404 or 500 and we assert that Read() surfaces an error diagnostic.
func TestDataSourceRead_APIErrors(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name   string
		newDS  func() datasource.DataSource
		status int
	}{
		{"vpc/404", NewVPCDataSource, http.StatusNotFound},
		{"vpc/500", NewVPCDataSource, http.StatusInternalServerError},
		{"subnet/404", NewSubnetDataSource, http.StatusNotFound},
		{"subnet/500", NewSubnetDataSource, http.StatusInternalServerError},
		{"securitygroup/404", NewSecurityGroupDataSource, http.StatusNotFound},
		{"securitygroup/500", NewSecurityGroupDataSource, http.StatusInternalServerError},
		{"securityrule/404", NewSecurityRuleDataSource, http.StatusNotFound},
		{"securityrule/500", NewSecurityRuleDataSource, http.StatusInternalServerError},
		{"elasticip/404", NewElasticIPDataSource, http.StatusNotFound},
		{"elasticip/500", NewElasticIPDataSource, http.StatusInternalServerError},
		{"vpcpeering/404", NewVPCPeeringDataSource, http.StatusNotFound},
		{"vpcpeering/500", NewVPCPeeringDataSource, http.StatusInternalServerError},
		{"vpcpeeringroute/404", NewVPCPeeringRouteDataSource, http.StatusNotFound},
		{"vpcpeeringroute/500", NewVPCPeeringRouteDataSource, http.StatusInternalServerError},
		{"vpntunnel/404", NewVPNTunnelDataSource, http.StatusNotFound},
		{"vpntunnel/500", NewVPNTunnelDataSource, http.StatusInternalServerError},
		{"vpnroute/404", NewVPNRouteDataSource, http.StatusNotFound},
		{"vpnroute/500", NewVPNRouteDataSource, http.StatusInternalServerError},
		{"backup/404", NewBackupDataSource, http.StatusNotFound},
		{"backup/500", NewBackupDataSource, http.StatusInternalServerError},
		{"blockstorage/404", NewBlockStorageDataSource, http.StatusNotFound},
		{"blockstorage/500", NewBlockStorageDataSource, http.StatusInternalServerError},
		{"snapshot/404", NewSnapshotDataSource, http.StatusNotFound},
		{"snapshot/500", NewSnapshotDataSource, http.StatusInternalServerError},
		{"restore/404", NewRestoreDataSource, http.StatusNotFound},
		{"restore/500", NewRestoreDataSource, http.StatusInternalServerError},
		{"keypair/404", NewKeypairDataSource, http.StatusNotFound},
		{"keypair/500", NewKeypairDataSource, http.StatusInternalServerError},
		{"cloudserver/404", NewCloudServerDataSource, http.StatusNotFound},
		{"cloudserver/500", NewCloudServerDataSource, http.StatusInternalServerError},
		{"kaas/404", NewKaaSDataSource, http.StatusNotFound},
		{"kaas/500", NewKaaSDataSource, http.StatusInternalServerError},
		{"containerregistry/404", NewContainerRegistryDataSource, http.StatusNotFound},
		{"containerregistry/500", NewContainerRegistryDataSource, http.StatusInternalServerError},
		{"dbaas/404", NewDBaaSDataSource, http.StatusNotFound},
		{"dbaas/500", NewDBaaSDataSource, http.StatusInternalServerError},
		{"database/404", NewDatabaseDataSource, http.StatusNotFound},
		{"database/500", NewDatabaseDataSource, http.StatusInternalServerError},
		{"databasebackup/404", NewDatabaseBackupDataSource, http.StatusNotFound},
		{"databasebackup/500", NewDatabaseBackupDataSource, http.StatusInternalServerError},
		{"databasegrant/404", NewDatabaseGrantDataSource, http.StatusNotFound},
		{"databasegrant/500", NewDatabaseGrantDataSource, http.StatusInternalServerError},
		{"dbaasuser/404", NewDBaaSUserDataSource, http.StatusNotFound},
		{"dbaasuser/500", NewDBaaSUserDataSource, http.StatusInternalServerError},
		{"kms/404", NewKMSDataSource, http.StatusNotFound},
		{"kms/500", NewKMSDataSource, http.StatusInternalServerError},
		{"project/404", NewProjectDataSource, http.StatusNotFound},
		{"project/500", NewProjectDataSource, http.StatusInternalServerError},
		{"schedulejob/404", NewScheduleJobDataSource, http.StatusNotFound},
		{"schedulejob/500", NewScheduleJobDataSource, http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status := tc.status
			_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
				apiError(w, status)
			})

			ds := tc.newDS()
			configureDatasource(ctx, t, ds, mockClient)

			req := dsReadReq(ctx, t, ds, nil)
			resp := &datasource.ReadResponse{}

			ds.Read(ctx, req, resp)

			if !resp.Diagnostics.HasError() {
				t.Errorf("%s: Read() returned no error diagnostic for HTTP %d response", tc.name, tc.status)
			}
		})
	}
}

// TestDataSourceRead_MissingIDs confirms that Read() returns an error
// diagnostic when the required IDs are empty strings (no API call is made).
func TestDataSourceRead_MissingIDs(t *testing.T) {
	ctx := context.Background()

	// Use any non-nil client — the ID check fires before the SDK call.
	_, dummyClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Should never be reached.
		t.Errorf("unexpected API call to %s", r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	})

	datasources := []struct {
		name  string
		newDS func() datasource.DataSource
	}{
		{"vpc", NewVPCDataSource},
		{"backup", NewBackupDataSource},
		{"kms", NewKMSDataSource},
		{"elasticip", NewElasticIPDataSource},
		{"blockstorage", NewBlockStorageDataSource},
	}

	for _, tc := range datasources {
		t.Run(tc.name, func(t *testing.T) {
			ds := tc.newDS()
			configureDatasource(ctx, t, ds, dummyClient)

			// Force all string attributes to empty string to trigger missing-ID error.
			schemaResp := &datasource.SchemaResponse{}
			ds.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
			extras := make(map[string]string)
			for name, attr := range schemaResp.Schema.Attributes {
				if attr.IsRequired() {
					extras[name] = ""
				}
			}

			req := dsReadReq(ctx, t, ds, extras)
			resp := &datasource.ReadResponse{}

			ds.Read(ctx, req, resp)

			if !resp.Diagnostics.HasError() {
				t.Errorf("%s: Read() with empty IDs returned no error diagnostic", tc.name)
			}
		})
	}
}
