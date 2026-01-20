// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// TestDataSourceSchemas tests that all datasources have valid schemas.
func TestDataSourceSchemas(t *testing.T) {
	ctx := context.Background()

	// Create provider
	p := New("test")()

	// Get provider schema
	providerReq := provider.SchemaRequest{}
	providerResp := &provider.SchemaResponse{}
	p.Schema(ctx, providerReq, providerResp)

	if providerResp.Diagnostics.HasError() {
		t.Fatalf("Provider schema has errors: %v", providerResp.Diagnostics)
	}

	// Get datasources directly
	datasources := p.DataSources(ctx)

	// Test each datasource schema
	for _, dsFunc := range datasources {
		ds := dsFunc()

		// Get metadata
		metadataReq := datasource.MetadataRequest{
			ProviderTypeName: "arubacloud",
		}
		metadataResp := &datasource.MetadataResponse{}
		ds.Metadata(ctx, metadataReq, metadataResp)

		// Get schema
		schemaReq := datasource.SchemaRequest{}
		schemaResp := &datasource.SchemaResponse{}
		ds.Schema(ctx, schemaReq, schemaResp)

		if schemaResp.Diagnostics.HasError() {
			t.Errorf("DataSource %s schema has errors: %v", metadataResp.TypeName, schemaResp.Diagnostics)
		}

		// Validate required attributes exist
		if schemaResp.Schema.Attributes == nil {
			t.Errorf("DataSource %s has no attributes", metadataResp.TypeName)
			continue
		}

		// Check for id attribute
		if _, ok := schemaResp.Schema.Attributes["id"]; !ok {
			t.Errorf("DataSource %s missing required 'id' attribute", metadataResp.TypeName)
		}
	}
}

// TestDataSourceMetadata tests that all datasources have proper metadata.
func TestDataSourceMetadata(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name           string
		dsFunc         func() datasource.DataSource
		expectedPrefix string
	}{
		{
			name:           "blockstorage",
			dsFunc:         NewBlockStorageDataSource,
			expectedPrefix: "arubacloud_blockstorage",
		},
		{
			name:           "cloudserver",
			dsFunc:         NewCloudServerDataSource,
			expectedPrefix: "arubacloud_cloudserver",
		},
		{
			name:           "project",
			dsFunc:         NewProjectDataSource,
			expectedPrefix: "arubacloud_project",
		},
		{
			name:           "vpc",
			dsFunc:         NewVPCDataSource,
			expectedPrefix: "arubacloud_vpc",
		},
		{
			name:           "subnet",
			dsFunc:         NewSubnetDataSource,
			expectedPrefix: "arubacloud_subnet",
		},
		{
			name:           "securitygroup",
			dsFunc:         NewSecurityGroupDataSource,
			expectedPrefix: "arubacloud_securitygroup",
		},
		{
			name:           "securityrule",
			dsFunc:         NewSecurityRuleDataSource,
			expectedPrefix: "arubacloud_securityrule",
		},
		{
			name:           "elasticip",
			dsFunc:         NewElasticIPDataSource,
			expectedPrefix: "arubacloud_elasticip",
		},
		{
			name:           "keypair",
			dsFunc:         NewKeypairDataSource,
			expectedPrefix: "arubacloud_keypair",
		},
		{
			name:           "containerregistry",
			dsFunc:         NewContainerRegistryDataSource,
			expectedPrefix: "arubacloud_containerregistry",
		},
		{
			name:           "kaas",
			dsFunc:         NewKaaSDataSource,
			expectedPrefix: "arubacloud_kaas",
		},
		{
			name:           "database",
			dsFunc:         NewDatabaseDataSource,
			expectedPrefix: "arubacloud_database",
		},
		{
			name:           "dbaas",
			dsFunc:         NewDBaaSDataSource,
			expectedPrefix: "arubacloud_dbaas",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := tc.dsFunc()

			metadataReq := datasource.MetadataRequest{
				ProviderTypeName: "arubacloud",
			}
			metadataResp := &datasource.MetadataResponse{}
			ds.Metadata(ctx, metadataReq, metadataResp)

			if metadataResp.TypeName != tc.expectedPrefix {
				t.Errorf("Expected type name %s, got %s", tc.expectedPrefix, metadataResp.TypeName)
			}
		})
	}
}

// TestBlockStorageDataSourceFlattenedSchema tests that blockstorage datasource has flattened fields.
func TestBlockStorageDataSourceFlattenedSchema(t *testing.T) {
	ctx := context.Background()
	ds := NewBlockStorageDataSource()

	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema has errors: %v", schemaResp.Diagnostics)
	}

	// Test that fields are flattened (not nested in properties)
	requiredFields := []string{"id", "name", "project_id", "location", "size_gb", "billing_period", "zone", "type", "tags", "snapshot_id", "bootable", "image"}

	for _, field := range requiredFields {
		if _, ok := schemaResp.Schema.Attributes[field]; !ok {
			t.Errorf("Missing flattened field: %s", field)
		}
	}

	// Ensure properties attribute doesn't exist (should be flattened)
	if _, ok := schemaResp.Schema.Attributes["properties"]; ok {
		t.Error("Schema should not have 'properties' attribute - fields should be flattened")
	}
}

// TestCloudServerDataSourceFlattenedSchema tests that cloudserver datasource has flattened fields.
func TestCloudServerDataSourceFlattenedSchema(t *testing.T) {
	ctx := context.Background()
	ds := NewCloudServerDataSource()

	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema has errors: %v", schemaResp.Diagnostics)
	}

	// Test that network, settings, storage fields are flattened
	requiredFields := []string{
		"id", "uri", "name", "location", "project_id", "zone", "tags",
		// Network fields (flattened)
		"vpc_uri_ref", "elastic_ip_uri_ref", "subnet_uri_refs", "securitygroup_uri_refs",
		// Settings fields (flattened)
		"flavor_name", "key_pair_uri_ref", "user_data",
		// Storage fields (flattened)
		"boot_volume_uri_ref",
	}

	for _, field := range requiredFields {
		if _, ok := schemaResp.Schema.Attributes[field]; !ok {
			t.Errorf("Missing flattened field: %s", field)
		}
	}

	// Ensure nested objects don't exist
	nestedObjects := []string{"network", "settings", "storage"}
	for _, obj := range nestedObjects {
		if _, ok := schemaResp.Schema.Attributes[obj]; ok {
			t.Errorf("Schema should not have '%s' attribute - fields should be flattened", obj)
		}
	}
}

// TestSubnetDataSourceFlattenedSchema tests that subnet datasource has flattened fields.
func TestSubnetDataSourceFlattenedSchema(t *testing.T) {
	ctx := context.Background()
	ds := NewSubnetDataSource()

	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema has errors: %v", schemaResp.Diagnostics)
	}

	// Test that network and dhcp fields are flattened
	requiredFields := []string{
		"id", "uri", "name", "location", "project_id", "vpc_id", "type", "tags",
		// Network field (flattened)
		"address",
		// DHCP fields (flattened)
		"dhcp_enabled", "dhcp_routes",
	}

	for _, field := range requiredFields {
		if _, ok := schemaResp.Schema.Attributes[field]; !ok {
			t.Errorf("Missing flattened field: %s", field)
		}
	}

	// Ensure nested objects don't exist
	nestedObjects := []string{"network", "dhcp"}
	for _, obj := range nestedObjects {
		if _, ok := schemaResp.Schema.Attributes[obj]; ok {
			t.Errorf("Schema should not have '%s' attribute - fields should be flattened", obj)
		}
	}
}

// TestProviderFactories tests that provider can be instantiated.
func TestProviderFactories(t *testing.T) {
	if _, err := providerserver.NewProtocol6WithError(New("test")())(); err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
}
