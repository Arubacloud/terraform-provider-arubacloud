// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ArubaCloudProvider satisfies various provider interfaces.
var _ provider.Provider = &ArubaCloudProvider{}
var _ provider.ProviderWithFunctions = &ArubaCloudProvider{}
var _ provider.ProviderWithEphemeralResources = &ArubaCloudProvider{}

// ArubaCloudProvider defines the provider implementation.
type ArubaCloudProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ArubaCloudProviderModel describes the provider data model.
type ArubaCloudProviderModel struct {
	ApiKey          types.String `tfsdk:"api_key"`
	ApiSecret       types.String `tfsdk:"api_secret"`
	ResourceTimeout types.String `tfsdk:"resource_timeout"`
}

func (p *ArubaCloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "arubacloud"
	resp.Version = p.version
}

func (p *ArubaCloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key for ArubaCloud",
				Required:            true,
			},
			"api_secret": schema.StringAttribute{
				MarkdownDescription: "API secret for ArubaCloud",
				Required:            true,
			},
			"resource_timeout": schema.StringAttribute{
				MarkdownDescription: "Timeout for waiting for resources to become active after creation (e.g., \"5m\", \"10m\", \"15m\"). This timeout applies to all resources that need to wait for active state. Default: \"10m\"",
				Optional:            true,
			},
		},
	}
}

func (p *ArubaCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

	// Default values to environment variables, but override with Terraform configuration value if set.
	apiKey := os.Getenv("ARUBACLOUD_API_KEY")
	apiSecret := os.Getenv("ARUBACLOUD_API_SECRET")

	// Retrieve provider data from configuration
	var config ArubaCloudProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.ApiKey.IsNull() {
		apiKey = config.ApiKey.ValueString()
	}

	if !config.ApiSecret.IsNull() {
		apiSecret = config.ApiSecret.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown ArubaCloud API Key",
			"The provider cannot create the ArubaCloud API client as there is an unknown configuration value for the API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ARUBACLOUD_API_KEY environment variable.",
		)
	}

	if apiSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_secret"),
			"Unknown ArubaCloud API Secret",
			"The provider cannot create the ArubaCloud API client as there is an unknown configuration value for the API secret. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ARUBACLOUD_API_SECRET environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create SDK client with credentials using DefaultOptions
	options := aruba.DefaultOptions(apiKey, apiSecret)
	options = options.WithDefaultLogger()

	sdkClient, err := aruba.NewClient(options)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create ArubaCloud SDK client",
			fmt.Sprintf("Unable to create ArubaCloud SDK client: %s", err),
		)
		return
	}

	// Parse timeout configuration with default (10 minutes - enough for most resources including CloudServer)
	resourceTimeout := parseTimeout(config.ResourceTimeout, 10*time.Minute)

	// Create a new ArubaCloud client using the SDK client
	client := &ArubaCloudClient{
		ApiKey:          apiKey,
		ApiSecret:       apiSecret,
		Client:          sdkClient,
		ResourceTimeout: resourceTimeout,
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

// parseTimeout parses a timeout string (e.g., "5m", "10m") and returns the duration.
// If the string is empty or invalid, returns the default duration.
func parseTimeout(timeoutStr types.String, defaultDuration time.Duration) time.Duration {
	if timeoutStr.IsNull() || timeoutStr.IsUnknown() || timeoutStr.ValueString() == "" {
		return defaultDuration
	}

	duration, err := time.ParseDuration(timeoutStr.ValueString())
	if err != nil {
		// If parsing fails, return default
		return defaultDuration
	}

	return duration
}

// ArubaCloudClient wraps the SDK client with API credentials and timeout configuration.
type ArubaCloudClient struct {
	ApiKey          string
	ApiSecret       string
	Client          aruba.Client
	ResourceTimeout time.Duration
}

func (p *ArubaCloudProvider) Resources(ctx context.Context) []func() resource.Resource {

	return []func() resource.Resource{
		NewProjectResource,
		NewCloudServerResource,
		NewKeypairResource,
		NewElasticIPResource,
		NewBlockStorageResource,
		NewSnapshotResource,
		NewVPCResource,
		NewVPNTunnelResource,
		NewVPNRouteResource,
		NewSubnetResource,
		NewSecurityGroupResource,
		NewSecurityRuleResource,
		NewVpcPeeringResource,
		NewVpcPeeringRouteResource,
		NewKaaSResource,
		NewContainerRegistryResource,
		NewBackupResource,
		NewRestoreResource,
		NewDBaaSResource,
		NewDatabaseResource,
		NewDatabaseGrantResource,
		NewDatabaseBackupResource,
		NewDBaaSUserResource,
		NewScheduleJobResource,
		NewKMSResource,
		// NewKMIPResource, // TODO: KMIP not available in SDK yet
	}

}

func (p *ArubaCloudProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *ArubaCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
		NewBlockStorageDataSource,
		NewSnapshotDataSource,
		NewVPCDataSource,
		NewKeypairDataSource,
		NewCloudServerDataSource,
		NewSubnetDataSource,
		NewElasticIPDataSource,
		NewSecurityGroupDataSource,
		NewSecurityRuleDataSource,
		NewVPCPeeringDataSource,
		NewVPCPeeringRouteDataSource,
		NewKaaSDataSource,
		NewContainerRegistryDataSource,
		NewBackupDataSource,
		NewDatabaseDataSource,
		NewDatabaseBackupDataSource,
		NewDatabaseGrantDataSource,
		NewDBaaSDataSource,
		// NewKMIPDataSource, // TODO: KMIP not available in SDK yet
		NewKMSDataSource,
		NewRestoreDataSource,
		NewScheduleJobDataSource,
		NewVPNRouteDataSource,
		NewVPNTunnelDataSource,
	}
}

func (p *ArubaCloudProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ArubaCloudProvider{
			version: version,
		}
	}
}
