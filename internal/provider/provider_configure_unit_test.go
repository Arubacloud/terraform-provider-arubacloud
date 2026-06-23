package provider

import (
	"context"
	"testing"

	providerframe "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// buildProviderConfig constructs a tfsdk.Config for the ArubaCloud provider
// from a map of attribute name → tftypes.Value overrides.  Attributes not
// mentioned in overrides are set to null (all provider attributes are Optional).
func buildProviderConfig(t *testing.T, p *ArubaCloudProvider, overrides map[string]tftypes.Value) tfsdk.Config {
	t.Helper()
	ctx := context.Background()

	schemaResp := &providerframe.SchemaResponse{}
	p.Schema(ctx, providerframe.SchemaRequest{}, schemaResp)

	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("buildProviderConfig: provider schema root is not an object type")
	}

	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, ty := range objType.AttributeTypes {
		if v, hasOverride := overrides[name]; hasOverride {
			attrs[name] = v
		} else {
			attrs[name] = tftypes.NewValue(ty, nil) // null for unspecified attributes
		}
	}

	return tfsdk.Config{
		Raw:    tftypes.NewValue(objType, attrs),
		Schema: schemaResp.Schema,
	}
}

// TestProviderConfigure_MissingAPIKey verifies that Configure() adds an
// attribute-level error diagnostic when api_key is absent.
func TestProviderConfigure_MissingAPIKey(t *testing.T) {
	ctx := context.Background()
	p := New("test")().(*ArubaCloudProvider)

	// Provide api_secret but NOT api_key.
	config := buildProviderConfig(t, p, map[string]tftypes.Value{
		"api_secret": tftypes.NewValue(tftypes.String, "test-secret"),
	})
	req := providerframe.ConfigureRequest{Config: config}
	resp := &providerframe.ConfigureResponse{}
	p.Configure(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostic for missing api_key, got none")
	}
}

// TestProviderConfigure_MissingAPISecret verifies that Configure() adds an
// attribute-level error diagnostic when api_secret is absent.
func TestProviderConfigure_MissingAPISecret(t *testing.T) {
	ctx := context.Background()
	p := New("test")().(*ArubaCloudProvider)

	config := buildProviderConfig(t, p, map[string]tftypes.Value{
		"api_key": tftypes.NewValue(tftypes.String, "test-key"),
	})
	req := providerframe.ConfigureRequest{Config: config}
	resp := &providerframe.ConfigureResponse{}
	p.Configure(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostic for missing api_secret, got none")
	}
}

// TestProviderConfigure_MissingBothCredentials verifies that Configure() adds
// two error diagnostics when both api_key and api_secret are absent.
func TestProviderConfigure_MissingBothCredentials(t *testing.T) {
	ctx := context.Background()
	p := New("test")().(*ArubaCloudProvider)

	config := buildProviderConfig(t, p, nil) // all null
	req := providerframe.ConfigureRequest{Config: config}
	resp := &providerframe.ConfigureResponse{}
	p.Configure(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostics for missing credentials, got none")
	}
	// Expect exactly two errors: one for api_key, one for api_secret.
	var errCount int
	for _, d := range resp.Diagnostics {
		if d.Severity() == 1 { // Error severity
			errCount++
		}
	}
	if errCount != 2 {
		t.Errorf("expected 2 error diagnostics, got %d: %v", errCount, resp.Diagnostics)
	}
}

// TestProviderConfigure_Success verifies that Configure() succeeds and sets
// ResourceData / DataSourceData when valid credentials are provided.
// The SDK client creation is lazy (no token fetch during NewClient), so no
// live server is required.
func TestProviderConfigure_Success(t *testing.T) {
	ctx := context.Background()
	p := New("test")().(*ArubaCloudProvider)

	config := buildProviderConfig(t, p, map[string]tftypes.Value{
		"api_key":    tftypes.NewValue(tftypes.String, "test-key"),
		"api_secret": tftypes.NewValue(tftypes.String, "test-secret"),
	})
	req := providerframe.ConfigureRequest{Config: config}
	resp := &providerframe.ConfigureResponse{}
	p.Configure(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error from Configure(): %v", resp.Diagnostics)
	}
	if resp.ResourceData == nil {
		t.Fatal("Configure() did not set ResourceData")
	}
	if resp.DataSourceData == nil {
		t.Fatal("Configure() did not set DataSourceData")
	}
	client, ok := resp.ResourceData.(*ArubaCloudClient)
	if !ok {
		t.Fatalf("ResourceData is %T, want *ArubaCloudClient", resp.ResourceData)
	}
	if client.ApiKey != "test-key" {
		t.Errorf("client.ApiKey = %q, want %q", client.ApiKey, "test-key")
	}
	if client.ApiSecret != "test-secret" {
		t.Errorf("client.ApiSecret = %q, want %q", client.ApiSecret, "test-secret")
	}
}

// TestProviderConfigure_WithBaseURL verifies that Configure() accepts an
// explicit base_url and token_issuer_url and still succeeds.
func TestProviderConfigure_WithBaseURL(t *testing.T) {
	ctx := context.Background()
	p := New("test")().(*ArubaCloudProvider)

	config := buildProviderConfig(t, p, map[string]tftypes.Value{
		"api_key":          tftypes.NewValue(tftypes.String, "test-key"),
		"api_secret":       tftypes.NewValue(tftypes.String, "test-secret"),
		"base_url":         tftypes.NewValue(tftypes.String, "https://example.com/api"),
		"token_issuer_url": tftypes.NewValue(tftypes.String, "https://example.com/token"),
	})
	req := providerframe.ConfigureRequest{Config: config}
	resp := &providerframe.ConfigureResponse{}
	p.Configure(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error from Configure() with base_url: %v", resp.Diagnostics)
	}
}

// TestProviderConfigure_InvalidLogLevel verifies that Configure() adds a
// warning (not an error) and proceeds successfully when an invalid log_level
// is supplied.
func TestProviderConfigure_InvalidLogLevel(t *testing.T) {
	ctx := context.Background()
	p := New("test")().(*ArubaCloudProvider)

	config := buildProviderConfig(t, p, map[string]tftypes.Value{
		"api_key":    tftypes.NewValue(tftypes.String, "test-key"),
		"api_secret": tftypes.NewValue(tftypes.String, "test-secret"),
		"log_level":  tftypes.NewValue(tftypes.String, "INVALID_LEVEL"),
	})
	req := providerframe.ConfigureRequest{Config: config}
	resp := &providerframe.ConfigureResponse{}
	p.Configure(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error for invalid log_level (only warning), got: %v", resp.Diagnostics)
	}
	var hasWarning bool
	for _, d := range resp.Diagnostics {
		if d.Severity() == 2 { // Warning severity
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("expected a warning diagnostic for invalid log_level, got none")
	}
}

// TestProviderConfigure_ValidResourceTimeout verifies that Configure()
// accepts a valid resource_timeout string.
func TestProviderConfigure_ValidResourceTimeout(t *testing.T) {
	ctx := context.Background()
	p := New("test")().(*ArubaCloudProvider)

	config := buildProviderConfig(t, p, map[string]tftypes.Value{
		"api_key":          tftypes.NewValue(tftypes.String, "test-key"),
		"api_secret":       tftypes.NewValue(tftypes.String, "test-secret"),
		"resource_timeout": tftypes.NewValue(tftypes.String, "5m"),
	})
	req := providerframe.ConfigureRequest{Config: config}
	resp := &providerframe.ConfigureResponse{}
	p.Configure(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error from Configure() with valid timeout: %v", resp.Diagnostics)
	}
	client, ok := resp.ResourceData.(*ArubaCloudClient)
	if !ok {
		t.Fatalf("ResourceData is %T, want *ArubaCloudClient", resp.ResourceData)
	}
	if client.ResourceTimeout.Minutes() != 5 {
		t.Errorf("ResourceTimeout = %v, want 5m", client.ResourceTimeout)
	}
}
