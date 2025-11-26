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

type CloudServerDataSourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Location       types.String `tfsdk:"location"`
	ProjectID      types.String `tfsdk:"project_id"`
	Zone           types.String `tfsdk:"zone"`
	VpcID          types.String `tfsdk:"vpc_id"`
	FlavorName     types.String `tfsdk:"flavor_name"`
	ElasticIPID    types.String `tfsdk:"elastic_ip_id"`
	BootVolume     types.String `tfsdk:"boot_volume"`
	KeyPairID      types.String `tfsdk:"key_pair_id"`
	Subnets        types.List   `tfsdk:"subnets"`
	SecurityGroups types.List   `tfsdk:"securitygroups"`
}

type CloudServerDataSource struct {
	client *ArubaCloudClient
}

var _ datasource.DataSource = &CloudServerDataSource{}

func NewCloudServerDataSource() datasource.DataSource {
	return &CloudServerDataSource{}
}

func (d *CloudServerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudserver"
}

func (d *CloudServerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CloudServer data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "CloudServer identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "CloudServer name",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "CloudServer location",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID",
				Computed:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone",
				Computed:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "VPC ID",
				Computed:            true,
			},
			"flavor_name": schema.StringAttribute{
				MarkdownDescription: "Flavor name",
				Computed:            true,
			},
			"elastic_ip_id": schema.StringAttribute{
				MarkdownDescription: "Elastic IP ID",
				Computed:            true,
			},
			"boot_volume": schema.StringAttribute{
				MarkdownDescription: "Boot volume ID",
				Computed:            true,
			},
			"key_pair_id": schema.StringAttribute{
				MarkdownDescription: "Key pair ID",
				Computed:            true,
			},
			"subnets": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of subnet IDs",
				Computed:            true,
			},
			"securitygroups": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of security group reference IDs",
				Computed:            true,
			},
		},
	}
}

func (d *CloudServerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ArubaCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *ArubaCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *CloudServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CloudServerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Simulate API response
	data.Id = types.StringValue("cloudserver-id")
	tflog.Trace(ctx, "read a CloudServer data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
