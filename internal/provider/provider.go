// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

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
	ApiKey    types.String `tfsdk:"api_key"`
	ApiSecret types.String `tfsdk:"api_secret"`
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

	// Create a new ArubaCloud client using the configuration values
	client := &ArubaCloudClient{
		ApiKey:    apiKey,
		ApiSecret: apiSecret,
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

// TODO: put here ArubaCloud API client.
type ArubaCloudClient struct {
	ApiKey    string
	ApiSecret string
}

func (p *ArubaCloudProvider) Resources(ctx context.Context) []func() resource.Resource {

	return []func() resource.Resource{
		NewProjectResource,
		NewElasticIPResource,
		NewVPCResource,
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
		NewKMIPResource,
	}

}

func (p *ArubaCloudProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *ArubaCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
		NewVPCDataSource,
		NewSubnetDataSource,
		NewElasticIPDataSource,
		NewSecurityGroupDataSource,
		NewSecurityRuleDataSource,
		NewVPCPeeringDataSource,
		NewVPCPeeringRouteDataSource,
		NewKaaSDataSource,
		NewContainerRegistryDataSource,
		NewBackupDataSource,
		NewBlockStorageDataSource,
		NewDatabaseDataSource,
		NewDatabaseBackupDataSource,
		NewDatabaseGrantDataSource,
		NewDBaaSDataSource,
		NewKMIPDataSource,
		NewKMSDataSource,
		NewRestoreDataSource,
		NewScheduleJobDataSource,
		NewSnapshotDataSource,
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
