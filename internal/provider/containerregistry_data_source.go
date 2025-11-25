// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"fmt"
	"net/http"

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
	client *http.Client
}

type ContainerRegistryDataSourceModel struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Location        types.String `tfsdk:"location"`
	Tags            types.List   `tfsdk:"tags"`
	ProjectID       types.String `tfsdk:"project_id"`
	ElasticIPID     types.String `tfsdk:"elasticip_id"`
	SubnetID        types.String `tfsdk:"subnet_id"`
	SecurityGroupID types.String `tfsdk:"security_group_id"`
	BlockStorageID  types.String `tfsdk:"block_storage_id"`
	BillingPeriod   types.String `tfsdk:"billing_period"`
	AdminUser       types.String `tfsdk:"admin_user"`
}

func (d *ContainerRegistryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_containerregistry"
}

func (d *ContainerRegistryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Container Registry data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Container Registry identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Container Registry name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Container Registry location",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the Container Registry resource",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this Container Registry belongs to",
				Required:            true,
			},
			"elasticip_id": schema.StringAttribute{
				MarkdownDescription: "Elastic IP ID",
				Required:            true,
			},
			"subnet_id": schema.StringAttribute{
				MarkdownDescription: "Subnet ID",
				Required:            true,
			},
			"security_group_id": schema.StringAttribute{
				MarkdownDescription: "Security Group ID",
				Required:            true,
			},
			"block_storage_id": schema.StringAttribute{
				MarkdownDescription: "Block Storage ID",
				Required:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period",
				Required:            true,
			},
			"admin_user": schema.StringAttribute{
				MarkdownDescription: "Admin user for the Container Registry",
				Required:            true,
			},
		},
	}
}

func (d *ContainerRegistryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*http.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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
	data.Name = types.StringValue("example-containerregistry")
	tflog.Trace(ctx, "read a Container Registry data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
