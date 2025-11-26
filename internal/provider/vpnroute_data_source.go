// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

var _ datasource.DataSource = &VPNRouteDataSource{}

func NewVPNRouteDataSource() datasource.DataSource {
	return &VPNRouteDataSource{}
}

type VPNRouteDataSource struct {
	client *http.Client
}

type VPNRouteDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Destination types.String `tfsdk:"destination"`
	Gateway     types.String `tfsdk:"gateway"`
}

func (d *VPNRouteDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpnroute"
}

func (d *VPNRouteDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VPN Route data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VPN Route identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "VPN Route name",
				Computed:            true,
			},
			"destination": schema.StringAttribute{
				MarkdownDescription: "Destination network for the VPN route",
				Computed:            true,
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "Gateway for the VPN route",
				Computed:            true,
			},
		},
	}
}

func (d *VPNRouteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VPNRouteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VPNRouteDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Name = types.StringValue("example-vpnroute")
	data.Destination = types.StringValue("10.0.0.0/24")
	data.Gateway = types.StringValue("10.0.0.1")
	tflog.Trace(ctx, "read a VPN Route data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
