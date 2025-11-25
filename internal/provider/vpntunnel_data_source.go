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

var _ datasource.DataSource = &VPNTunnelDataSource{}

func NewVPNTunnelDataSource() datasource.DataSource {
	return &VPNTunnelDataSource{}
}

type VPNTunnelDataSource struct {
	client *http.Client
}

type VPNTunnelDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	RemotePeer types.String `tfsdk:"remote_peer"`
	Status     types.String `tfsdk:"status"`
}

func (d *VPNTunnelDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpntunnel"
}

func (d *VPNTunnelDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VPN Tunnel data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VPN Tunnel identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "VPN Tunnel name",
				Computed:            true,
			},
			"remote_peer": schema.StringAttribute{
				MarkdownDescription: "Remote peer address for the VPN tunnel",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the VPN tunnel",
				Computed:            true,
			},
		},
	}
}

func (d *VPNTunnelDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VPNTunnelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VPNTunnelDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Name = types.StringValue("example-vpntunnel")
	data.RemotePeer = types.StringValue("203.0.113.1")
	data.Status = types.StringValue("active")
	tflog.Trace(ctx, "read a VPN Tunnel data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
