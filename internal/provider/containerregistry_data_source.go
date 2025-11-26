// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Container Registry name",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Container Registry location",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the Container Registry resource",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this Container Registry belongs to",
				Computed:            true,
			},
			"elasticip_id": schema.StringAttribute{
				MarkdownDescription: "Elastic IP ID",
				Computed:            true,
			},
			"subnet_id": schema.StringAttribute{
				MarkdownDescription: "Subnet ID",
				Computed:            true,
			},
			"security_group_id": schema.StringAttribute{
				MarkdownDescription: "Security Group ID",
				Computed:            true,
			},
			"block_storage_id": schema.StringAttribute{
				MarkdownDescription: "Block Storage ID",
				Computed:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period",
				Computed:            true,
			},
			"admin_user": schema.StringAttribute{
				MarkdownDescription: "Admin user for the Container Registry",
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
