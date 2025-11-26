// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &VPNRouteResource{}
var _ resource.ResourceWithImportState = &VPNRouteResource{}

func NewVPNRouteResource() resource.Resource {
	return &VPNRouteResource{}
}

type VPNRouteProperties struct {
	CloudSubnet  types.String `tfsdk:"cloud_subnet"`
	OnPremSubnet types.String `tfsdk:"on_prem_subnet"`
}

type VPNRouteResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Location    types.String `tfsdk:"location"`
	Tags        types.List   `tfsdk:"tags"`
	ProjectId   types.String `tfsdk:"project_id"`
	VPNTunnelId types.String `tfsdk:"vpn_tunnel_id"`
	Properties  types.Object `tfsdk:"properties"`
}

type VPNRouteResource struct {
	client *ArubaCloudClient
}

func (r *VPNRouteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpnroute"
}

func (r *VPNRouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VPN Route resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VPN Route identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "VPN Route name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "VPN Route location",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the VPN Route",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this VPN Route belongs to",
				Required:            true,
			},
			"vpn_tunnel_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPN Tunnel this route belongs to",
				Required:            true,
			},
			"properties": schema.SingleNestedAttribute{
				MarkdownDescription: "Properties of the VPN Route",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"cloud_subnet": schema.StringAttribute{
						MarkdownDescription: "CIDR of the cloud subnet",
						Optional:            true,
						// Validators removed for v1.16.1 compatibility
					},
					"on_prem_subnet": schema.StringAttribute{
						MarkdownDescription: "CIDR of the on-prem subnet",
						Optional:            true,
						// Validators removed for v1.16.1 compatibility
					},
				},
			},
		},
	}
}

func (r *VPNRouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ArubaCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *VPNRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VPNRouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Simulate API response
	data.Id = types.StringValue("vpnroute-id")
	tflog.Trace(ctx, "created a VPN Route resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPNRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VPNRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPNRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VPNRouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPNRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VPNRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *VPNRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
