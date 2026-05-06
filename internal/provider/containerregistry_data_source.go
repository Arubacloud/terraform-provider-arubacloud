package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &ContainerRegistryDataSource{}

func NewContainerRegistryDataSource() datasource.DataSource {
	return &ContainerRegistryDataSource{}
}

type ContainerRegistryDataSource struct {
	client *ArubaCloudClient
}

type ContainerRegistryDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	ProjectID     types.String `tfsdk:"project_id"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	// Network fields (flattened)
	PublicIpUriRef      types.String `tfsdk:"public_ip_uri_ref"`
	VpcUriRef           types.String `tfsdk:"vpc_uri_ref"`
	SubnetUriRef        types.String `tfsdk:"subnet_uri_ref"`
	SecurityGroupUriRef types.String `tfsdk:"security_group_uri_ref"`
	// Storage fields (flattened)
	BlockStorageUriRef types.String `tfsdk:"block_storage_uri_ref"`
	// Settings fields (flattened)
	AdminUser             types.String `tfsdk:"admin_user"`
	ConcurrentUsersFlavor types.String `tfsdk:"concurrent_users_flavor"`
}

func (d *ContainerRegistryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_containerregistry"
}

func (d *ContainerRegistryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about an existing ArubaCloud Container Registry.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the container registry to look up.",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the container registry.",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle. Accepted values: `Hour`, `Month`, `Year`.",
				Computed:            true,
			},
			"public_ip_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI of the Elastic IP that exposes the registry endpoint.",
				Computed:            true,
			},
			"vpc_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI of the VPC that hosts the registry.",
				Computed:            true,
			},
			"subnet_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI of the subnet within the VPC.",
				Computed:            true,
			},
			"security_group_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI of the security group controlling registry traffic.",
				Computed:            true,
			},
			"block_storage_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI of the block storage volume backing the registry image store.",
				Computed:            true,
			},
			"admin_user": schema.StringAttribute{
				MarkdownDescription: "Administrator username for the registry.",
				Computed:            true,
			},
			"concurrent_users_flavor": schema.StringAttribute{
				MarkdownDescription: "Concurrency tier for simultaneous push/pull sessions (`Small`, `Medium`, `HighPerf`).",
				Computed:            true,
			},
		},
	}
}

func (d *ContainerRegistryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ArubaCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ArubaCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *ContainerRegistryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ContainerRegistryDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	registryID := data.Id.ValueString()
	if projectID == "" || registryID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID and Container Registry ID are required to read the container registry")
		return
	}

	response, err := d.client.Client.FromContainer().ContainerRegistry().Get(ctx, projectID, registryID, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading container registry", NewTransportError("read", "Containerregistry", err).Error())
		return
	}
	if apiErr := CheckResponse("read", "Containerregistry", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError("No data returned", "Container Registry Get returned no data")
		return
	}

	registry := response.Data
	if registry.Metadata.ID != nil {
		data.Id = types.StringValue(*registry.Metadata.ID)
	}
	if registry.Metadata.URI != nil {
		data.Uri = types.StringValue(*registry.Metadata.URI)
	} else {
		data.Uri = types.StringNull()
	}
	if registry.Metadata.Name != nil {
		data.Name = types.StringValue(*registry.Metadata.Name)
	}
	if registry.Metadata.LocationResponse != nil {
		data.Location = types.StringValue(registry.Metadata.LocationResponse.Value)
	} else {
		data.Location = types.StringNull()
	}
	data.ProjectID = types.StringValue(projectID)

	if registry.Properties.BillingPlan.BillingPeriod != "" {
		data.BillingPeriod = types.StringValue(registry.Properties.BillingPlan.BillingPeriod)
	} else {
		data.BillingPeriod = types.StringNull()
	}
	data.PublicIpUriRef = types.StringValue(registry.Properties.PublicIp.URI)
	data.VpcUriRef = types.StringValue(registry.Properties.VPC.URI)
	data.SubnetUriRef = types.StringValue(registry.Properties.Subnet.URI)
	data.SecurityGroupUriRef = types.StringValue(registry.Properties.SecurityGroup.URI)
	data.BlockStorageUriRef = types.StringValue(registry.Properties.BlockStorage.URI)

	if registry.Properties.AdminUser.Username != "" {
		data.AdminUser = types.StringValue(registry.Properties.AdminUser.Username)
	} else {
		data.AdminUser = types.StringNull()
	}
	if registry.Properties.ConcurrentUsers != nil && *registry.Properties.ConcurrentUsers != "" {
		data.ConcurrentUsersFlavor = types.StringValue(*registry.Properties.ConcurrentUsers)
	} else {
		data.ConcurrentUsersFlavor = types.StringNull()
	}

	data.Tags = TagsToList(registry.Metadata.Tags)

	tflog.Trace(ctx, "read a Container Registry data source", map[string]interface{}{"registry_id": registryID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
