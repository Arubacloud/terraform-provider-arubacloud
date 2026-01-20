package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// TestResourceSchemas tests that all resources have valid schemas.
func TestResourceSchemas(t *testing.T) {
	ctx := context.Background()

	// Create provider
	p := New("test")()

	// Get resources directly
	resources := p.Resources(ctx)

	// Test each resource schema
	for _, rFunc := range resources {
		r := rFunc()

		// Get metadata
		metadataReq := resource.MetadataRequest{
			ProviderTypeName: "arubacloud",
		}
		metadataResp := &resource.MetadataResponse{}
		r.Metadata(ctx, metadataReq, metadataResp)

		// Get schema
		schemaReq := resource.SchemaRequest{}
		schemaResp := &resource.SchemaResponse{}
		r.Schema(ctx, schemaReq, schemaResp)

		if schemaResp.Diagnostics.HasError() {
			t.Errorf("Resource %s schema has errors: %v", metadataResp.TypeName, schemaResp.Diagnostics)
		}

		// Validate required attributes exist
		if schemaResp.Schema.Attributes == nil {
			t.Errorf("Resource %s has no attributes", metadataResp.TypeName)
			continue
		}

		// Check for id attribute
		if _, ok := schemaResp.Schema.Attributes["id"]; !ok {
			t.Errorf("Resource %s missing required 'id' attribute", metadataResp.TypeName)
		}
	}
}

// TestResourceMetadata tests that all resources have correct type names.
func TestResourceMetadata(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name           string
		rFunc          func() resource.Resource
		expectedPrefix string
	}{
		{
			name:           "backup",
			rFunc:          NewBackupResource,
			expectedPrefix: "arubacloud_backup",
		},
		{
			name:           "blockstorage",
			rFunc:          NewBlockStorageResource,
			expectedPrefix: "arubacloud_blockstorage",
		},
		{
			name:           "cloudserver",
			rFunc:          NewCloudServerResource,
			expectedPrefix: "arubacloud_cloudserver",
		},
		{
			name:           "containerregistry",
			rFunc:          NewContainerRegistryResource,
			expectedPrefix: "arubacloud_containerregistry",
		},
		{
			name:           "database",
			rFunc:          NewDatabaseResource,
			expectedPrefix: "arubacloud_database",
		},
		{
			name:           "databasebackup",
			rFunc:          NewDatabaseBackupResource,
			expectedPrefix: "arubacloud_databasebackup",
		},
		{
			name:           "databasegrant",
			rFunc:          NewDatabaseGrantResource,
			expectedPrefix: "arubacloud_databasegrant",
		},
		{
			name:           "dbaas",
			rFunc:          NewDBaaSResource,
			expectedPrefix: "arubacloud_dbaas",
		},
		{
			name:           "dbaasuser",
			rFunc:          NewDBaaSUserResource,
			expectedPrefix: "arubacloud_dbaasuser",
		},
		{
			name:           "elasticip",
			rFunc:          NewElasticIPResource,
			expectedPrefix: "arubacloud_elasticip",
		},
		{
			name:           "kaas",
			rFunc:          NewKaaSResource,
			expectedPrefix: "arubacloud_kaas",
		},
		{
			name:           "keypair",
			rFunc:          NewKeypairResource,
			expectedPrefix: "arubacloud_keypair",
		},
		{
			name:           "kms",
			rFunc:          NewKMSResource,
			expectedPrefix: "arubacloud_kms",
		},
		{
			name:           "project",
			rFunc:          NewProjectResource,
			expectedPrefix: "arubacloud_project",
		},
		{
			name:           "restore",
			rFunc:          NewRestoreResource,
			expectedPrefix: "arubacloud_restore",
		},
		{
			name:           "schedulejob",
			rFunc:          NewScheduleJobResource,
			expectedPrefix: "arubacloud_schedulejob",
		},
		{
			name:           "securitygroup",
			rFunc:          NewSecurityGroupResource,
			expectedPrefix: "arubacloud_securitygroup",
		},
		{
			name:           "securityrule",
			rFunc:          NewSecurityRuleResource,
			expectedPrefix: "arubacloud_securityrule",
		},
		{
			name:           "snapshot",
			rFunc:          NewSnapshotResource,
			expectedPrefix: "arubacloud_snapshot",
		},
		{
			name:           "subnet",
			rFunc:          NewSubnetResource,
			expectedPrefix: "arubacloud_subnet",
		},
		{
			name:           "vpc",
			rFunc:          NewVPCResource,
			expectedPrefix: "arubacloud_vpc",
		},
		{
			name:           "vpcpeering",
			rFunc:          NewVpcPeeringResource,
			expectedPrefix: "arubacloud_vpcpeering",
		},
		{
			name:           "vpcpeeringroute",
			rFunc:          NewVpcPeeringRouteResource,
			expectedPrefix: "arubacloud_vpcpeeringroute",
		},
		{
			name:           "vpnroute",
			rFunc:          NewVPNRouteResource,
			expectedPrefix: "arubacloud_vpnroute",
		},
		{
			name:           "vpntunnel",
			rFunc:          NewVPNTunnelResource,
			expectedPrefix: "arubacloud_vpntunnel",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.rFunc()

			metadataReq := resource.MetadataRequest{
				ProviderTypeName: "arubacloud",
			}
			metadataResp := &resource.MetadataResponse{}
			r.Metadata(ctx, metadataReq, metadataResp)

			if metadataResp.TypeName != tc.expectedPrefix {
				t.Errorf("Expected type name %s, got %s", tc.expectedPrefix, metadataResp.TypeName)
			}
		})
	}
}

// TestResourceImportState tests that resources support ImportState.
func TestResourceImportState(t *testing.T) {
	// Test a subset of critical resources
	testCases := []struct {
		name  string
		rFunc func() resource.Resource
	}{
		{name: "blockstorage", rFunc: NewBlockStorageResource},
		{name: "cloudserver", rFunc: NewCloudServerResource},
		{name: "vpc", rFunc: NewVPCResource},
		{name: "subnet", rFunc: NewSubnetResource},
		{name: "securitygroup", rFunc: NewSecurityGroupResource},
		{name: "elasticip", rFunc: NewElasticIPResource},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.rFunc()

			// Check if resource implements ImportState
			if _, ok := r.(resource.ResourceWithImportState); !ok {
				t.Errorf("Resource %s does not implement ImportState", tc.name)
			}
		})
	}
}

// TestBlockStorageResourceSchema validates blockstorage resource schema structure.
func TestBlockStorageResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := NewBlockStorageResource()

	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema has errors: %v", schemaResp.Diagnostics)
	}

	// Check for required attributes
	requiredAttrs := []string{"id", "name", "size_gb", "zone"}
	for _, attr := range requiredAttrs {
		if _, ok := schemaResp.Schema.Attributes[attr]; !ok {
			t.Errorf("Missing required attribute: %s", attr)
		}
	}
}

// TestCloudServerResourceSchema validates cloudserver resource schema structure.
func TestCloudServerResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := NewCloudServerResource()

	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema has errors: %v", schemaResp.Diagnostics)
	}

	// Check for required attributes
	requiredAttrs := []string{"id", "name", "zone", "location", "project_id"}
	for _, attr := range requiredAttrs {
		if _, ok := schemaResp.Schema.Attributes[attr]; !ok {
			t.Errorf("Missing required attribute: %s", attr)
		}
	}
}

// TestVPCResourceSchema validates VPC resource schema structure.
func TestVPCResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := NewVPCResource()

	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema has errors: %v", schemaResp.Diagnostics)
	}

	// Check for required attributes
	requiredAttrs := []string{"id", "name", "location", "project_id"}
	for _, attr := range requiredAttrs {
		if _, ok := schemaResp.Schema.Attributes[attr]; !ok {
			t.Errorf("Missing required attribute: %s", attr)
		}
	}
}

// TestSubnetResourceSchema validates subnet resource schema structure.
// TestSubnetResourceSchema validates subnet resource schema structure.
func TestSubnetResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := NewSubnetResource()

	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema has errors: %v", schemaResp.Diagnostics)
	}

	// Check for required attributes
	requiredAttrs := []string{"id", "name", "vpc_uri_ref", "location", "project_id"}
	for _, attr := range requiredAttrs {
		if _, ok := schemaResp.Schema.Attributes[attr]; !ok {
			t.Errorf("Missing required attribute: %s", attr)
		}
	}
}
