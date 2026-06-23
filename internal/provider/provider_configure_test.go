package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// TestAllResourcesConfigure exercises the three branches of Configure() on
// every resource: nil ProviderData (no-op), wrong type (error diagnostic),
// and correct *ArubaCloudClient (stored silently).
func TestAllResourcesConfigure(t *testing.T) {
	ctx := context.Background()

	resources := []struct {
		name string
		fn   func() resource.Resource
	}{
		{"project", NewProjectResource},
		{"cloudserver", NewCloudServerResource},
		{"keypair", NewKeypairResource},
		{"elasticip", NewElasticIPResource},
		{"blockstorage", NewBlockStorageResource},
		{"snapshot", NewSnapshotResource},
		{"vpc", NewVPCResource},
		{"vpntunnel", NewVPNTunnelResource},
		{"vpnroute", NewVPNRouteResource},
		{"subnet", NewSubnetResource},
		{"securitygroup", NewSecurityGroupResource},
		{"securityrule", NewSecurityRuleResource},
		{"vpcpeering", NewVpcPeeringResource},
		{"vpcpeeringroute", NewVpcPeeringRouteResource},
		{"kaas", NewKaaSResource},
		{"containerregistry", NewContainerRegistryResource},
		{"backup", NewBackupResource},
		{"restore", NewRestoreResource},
		{"dbaas", NewDBaaSResource},
		{"database", NewDatabaseResource},
		{"databasegrant", NewDatabaseGrantResource},
		{"databasebackup", NewDatabaseBackupResource},
		{"dbaasuser", NewDBaaSUserResource},
		{"schedulejob", NewScheduleJobResource},
		{"kms", NewKMSResource},
	}

	for _, tc := range resources {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.fn()
			cfg, ok := r.(resource.ResourceWithConfigure)
			if !ok {
				t.Fatalf("%s does not implement ResourceWithConfigure", tc.name)
			}

			// Branch 1: nil ProviderData — must be a no-op.
			resp1 := &resource.ConfigureResponse{}
			cfg.Configure(ctx, resource.ConfigureRequest{ProviderData: nil}, resp1)
			if resp1.Diagnostics.HasError() {
				t.Errorf("nil ProviderData: unexpected error: %v", resp1.Diagnostics)
			}

			// Branch 2: wrong type — must add exactly one error containing "Unexpected".
			resp2 := &resource.ConfigureResponse{}
			cfg.Configure(ctx, resource.ConfigureRequest{ProviderData: "wrong-type"}, resp2)
			if !resp2.Diagnostics.HasError() {
				t.Errorf("wrong type: expected error diagnostic, got none")
			}
			hasUnexpected := false
			for _, d := range resp2.Diagnostics {
				if strings.Contains(d.Summary(), "Unexpected") {
					hasUnexpected = true
					break
				}
			}
			if !hasUnexpected {
				t.Errorf("wrong type: expected 'Unexpected' in diagnostic summary, got: %v", resp2.Diagnostics)
			}

			// Branch 3: correct *ArubaCloudClient — must succeed with no diagnostics.
			resp3 := &resource.ConfigureResponse{}
			cfg.Configure(ctx, resource.ConfigureRequest{ProviderData: &ArubaCloudClient{}}, resp3)
			if resp3.Diagnostics.HasError() {
				t.Errorf("correct type: unexpected error: %v", resp3.Diagnostics)
			}
		})
	}
}

// TestAllDataSourcesConfigure exercises the three branches of Configure() on
// every datasource: nil ProviderData (no-op), wrong type (error diagnostic),
// and correct *ArubaCloudClient (stored silently).
func TestAllDataSourcesConfigure(t *testing.T) {
	ctx := context.Background()

	datasources := []struct {
		name string
		fn   func() datasource.DataSource
	}{
		{"project", NewProjectDataSource},
		{"blockstorage", NewBlockStorageDataSource},
		{"snapshot", NewSnapshotDataSource},
		{"vpc", NewVPCDataSource},
		{"keypair", NewKeypairDataSource},
		{"cloudserver", NewCloudServerDataSource},
		{"subnet", NewSubnetDataSource},
		{"elasticip", NewElasticIPDataSource},
		{"securitygroup", NewSecurityGroupDataSource},
		{"securityrule", NewSecurityRuleDataSource},
		{"vpcpeering", NewVPCPeeringDataSource},
		{"vpcpeeringroute", NewVPCPeeringRouteDataSource},
		{"kaas", NewKaaSDataSource},
		{"containerregistry", NewContainerRegistryDataSource},
		{"backup", NewBackupDataSource},
		{"database", NewDatabaseDataSource},
		{"databasebackup", NewDatabaseBackupDataSource},
		{"databasegrant", NewDatabaseGrantDataSource},
		{"dbaas", NewDBaaSDataSource},
		{"dbaasuser", NewDBaaSUserDataSource},
		{"kms", NewKMSDataSource},
		{"restore", NewRestoreDataSource},
		{"schedulejob", NewScheduleJobDataSource},
		{"vpnroute", NewVPNRouteDataSource},
		{"vpntunnel", NewVPNTunnelDataSource},
	}

	for _, tc := range datasources {
		t.Run(tc.name, func(t *testing.T) {
			ds := tc.fn()
			cfg, ok := ds.(datasource.DataSourceWithConfigure)
			if !ok {
				t.Fatalf("%s does not implement DataSourceWithConfigure", tc.name)
			}

			// Branch 1: nil ProviderData — must be a no-op.
			resp1 := &datasource.ConfigureResponse{}
			cfg.Configure(ctx, datasource.ConfigureRequest{ProviderData: nil}, resp1)
			if resp1.Diagnostics.HasError() {
				t.Errorf("nil ProviderData: unexpected error: %v", resp1.Diagnostics)
			}

			// Branch 2: wrong type — must add exactly one error containing "Unexpected".
			resp2 := &datasource.ConfigureResponse{}
			cfg.Configure(ctx, datasource.ConfigureRequest{ProviderData: "wrong-type"}, resp2)
			if !resp2.Diagnostics.HasError() {
				t.Errorf("wrong type: expected error diagnostic, got none")
			}
			hasUnexpected := false
			for _, d := range resp2.Diagnostics {
				if strings.Contains(d.Summary(), "Unexpected") {
					hasUnexpected = true
					break
				}
			}
			if !hasUnexpected {
				t.Errorf("wrong type: expected 'Unexpected' in diagnostic summary, got: %v", resp2.Diagnostics)
			}

			// Branch 3: correct *ArubaCloudClient — must succeed with no diagnostics.
			resp3 := &datasource.ConfigureResponse{}
			cfg.Configure(ctx, datasource.ConfigureRequest{ProviderData: &ArubaCloudClient{}}, resp3)
			if resp3.Diagnostics.HasError() {
				t.Errorf("correct type: unexpected error: %v", resp3.Diagnostics)
			}
		})
	}
}
