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

var _ datasource.DataSource = &SubnetDataSource{}

func NewSubnetDataSource() datasource.DataSource {
	return &SubnetDataSource{}
}

type SubnetDataSource struct {
	client *ArubaCloudClient
}

type SubnetDataSourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	Tags      types.List   `tfsdk:"tags"`
	ProjectId types.String `tfsdk:"project_id"`
	VpcId     types.String `tfsdk:"vpc_id"`
	Type      types.String `tfsdk:"type"`
	Network   types.Object `tfsdk:"network"`
	Dhcp      types.Object `tfsdk:"dhcp"`
	Routes    types.List   `tfsdk:"routes"`
	Dns       types.List   `tfsdk:"dns"`
}

func (d *SubnetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (d *SubnetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Subnet data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Subnet identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Subnet name",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Subnet location",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the subnet",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this subnet belongs to",
				Computed:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPC this subnet belongs to",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Subnet type (Basic or Advanced)",
				Computed:            true,
			},
			"network": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						MarkdownDescription: "Address of the network in CIDR notation (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)",
						Computed:            true,
					},
				},
				Computed: true,
			},
			"dhcp": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable DHCP",
						Computed:            true,
					},
					"range": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"start": schema.StringAttribute{
								MarkdownDescription: "Starting IP address",
								Computed:            true,
							},
							"count": schema.Int64Attribute{
								MarkdownDescription: "Number of available IP addresses",
								Computed:            true,
							},
						},
						Computed: true,
					},
				},
				Computed: true,
			},
			"routes": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							MarkdownDescription: "IP address of the route",
							Computed:            true,
						},
						"gateway": schema.StringAttribute{
							MarkdownDescription: "Gateway",
							Computed:            true,
						},
					},
				},
				Computed: true,
			},
			"dns": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of DNS IP addresses",
				Computed:            true,
			},
		},
	}
}

func (d *SubnetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SubnetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubnetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Name = types.StringValue("example-subnet")
	tflog.Trace(ctx, "read a Subnet data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
