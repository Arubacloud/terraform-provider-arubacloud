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

var _ resource.Resource = &SubnetResource{}
var _ resource.ResourceWithImportState = &SubnetResource{}

func NewSubnetResource() resource.Resource {
	return &SubnetResource{}
}

type SubnetResource struct {
	client *ArubaCloudClient
}

type SubnetResourceModel struct {
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

type NetworkModel struct {
	Address types.String `tfsdk:"address"`
}

type DhcpModel struct {
	Enabled types.Bool   `tfsdk:"enabled"`
	Range   types.Object `tfsdk:"range"`
}

type DhcpRangeModel struct {
	Start types.String `tfsdk:"start"`
	Count types.Int64  `tfsdk:"count"`
}

type RouteModel struct {
	Address types.String `tfsdk:"address"`
	Gateway types.String `tfsdk:"gateway"`
}

func (r *SubnetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (r *SubnetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Subnet resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Subnet identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Subnet name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Subnet location",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the subnet",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this subnet belongs to",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPC this subnet belongs to",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Subnet type (Basic or Advanced)",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"network": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						MarkdownDescription: "Address of the network in CIDR notation (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"dhcp": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable DHCP",
						Optional:            true,
					},
					"range": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"start": schema.StringAttribute{
								MarkdownDescription: "Starting IP address",
								Optional:            true,
							},
							"count": schema.Int64Attribute{
								MarkdownDescription: "Number of available IP addresses",
								Optional:            true,
							},
						},
						Optional: true,
					},
				},
				Optional: true,
			},
			"routes": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							MarkdownDescription: "IP address of the route",
							Optional:            true,
						},
						"gateway": schema.StringAttribute{
							MarkdownDescription: "Gateway",
							Optional:            true,
						},
					},
				},
				Optional: true,
			},
			"dns": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of DNS IP addresses",
				Optional:            true,
			},
		},
	}
}

func (r *SubnetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Simulate API response
	data.Id = types.StringValue("subnet-id")
	tflog.Trace(ctx, "created a Subnet resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SubnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
