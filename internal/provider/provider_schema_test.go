package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
)

// TestProviderSchema validates that the provider can be instantiated successfully.
func TestProviderSchema(t *testing.T) {
	t.Parallel()

	p := New("test")()
	if p == nil {
		t.Fatal("provider is nil")
	}

	provider, ok := p.(*ArubaCloudProvider)
	if !ok {
		t.Fatal("provider is not of type *ArubaCloudProvider")
	}

	if provider.version != "test" {
		t.Errorf("expected version 'test', got %s", provider.version)
	}
}

// TestAllResourceSchemas validates that all resource schemas can be retrieved.

func TestAllResourceSchemas(t *testing.T) {
	t.Parallel()

	p := New("test")()
	provider, ok := p.(*ArubaCloudProvider)
	if !ok {
		t.Fatal("provider is not of type *ArubaCloudProvider")
	}

	// Get all resources
	resources := provider.Resources(context.TODO())

	if len(resources) == 0 {
		t.Fatal("no resources registered")
	}

	// Expected number of resources (excluding disabled Key and KMIP)
	expectedCount := 25 // Total active resources
	if len(resources) != expectedCount {
		t.Errorf("expected %d resources, got %d", expectedCount, len(resources))
	}

	// Validate each resource can be instantiated
	for _, resourceFunc := range resources {
		resource := resourceFunc()
		if resource == nil {
			t.Error("resource function returned nil")
		}
	}
}

// TestAllDataSourceSchemas validates that all data source schemas can be retrieved.

func TestAllDataSourceSchemas(t *testing.T) {
	t.Parallel()

	p := New("test")()
	provider, ok := p.(*ArubaCloudProvider)
	if !ok {
		t.Fatal("provider is not of type *ArubaCloudProvider")
	}

	// Get all data sources
	dataSources := provider.DataSources(context.TODO())

	if len(dataSources) == 0 {
		t.Fatal("no data sources registered")
	}

	// Expected number of data sources (excluding disabled Key and KMIP)
	expectedCount := 25 // Total active data sources
	if len(dataSources) != expectedCount {
		t.Errorf("expected %d data sources, got %d", expectedCount, len(dataSources))
	}

	// Validate each data source can be instantiated
	for _, dataSourceFunc := range dataSources {
		dataSource := dataSourceFunc()
		if dataSource == nil {
			t.Error("data source function returned nil")
		}
	}
}

// TestProviderSchemaAttributes validates the auth attribute rename from v0.1.x to v0.2.0.
func TestProviderSchemaAttributes(t *testing.T) {
	t.Parallel()

	p := New("test")()
	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.TODO(), provider.SchemaRequest{}, schemaResp)
	attrs := schemaResp.Schema.Attributes

	// v0.2.0 attributes must exist
	for _, name := range []string{"client_id", "client_secret"} {
		if _, ok := attrs[name]; !ok {
			t.Errorf("expected provider schema attribute %q to exist", name)
		}
	}
	// client_secret must be sensitive
	if cs, ok := attrs["client_secret"]; ok {
		if sa, ok := cs.(schema.StringAttribute); !ok || !sa.Sensitive {
			t.Error("client_secret must be Sensitive: true")
		}
	}
	// v0.1.x attributes must NOT exist
	for _, name := range []string{"api_key", "api_secret"} {
		if _, ok := attrs[name]; ok {
			t.Errorf("deprecated attribute %q must not be present in v0.2.0 schema", name)
		}
	}
}

// TestProviderMetadata validates provider metadata.
func TestProviderMetadata(t *testing.T) {
	t.Parallel()

	p := New("0.0.1")()
	provider, ok := p.(*ArubaCloudProvider)
	if !ok {
		t.Fatal("provider is not of type *ArubaCloudProvider")
	}

	if provider.version != "0.0.1" {
		t.Errorf("expected version 0.0.1, got %s", provider.version)
	}
}
