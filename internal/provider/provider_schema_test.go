package provider

import (
	"context"
	"testing"
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
