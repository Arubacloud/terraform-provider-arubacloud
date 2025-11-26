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

type VPCResource struct {
	client *ArubaCloudClient
}

type VPCResourceModel struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Location types.String `tfsdk:"location"`
	Tags     types.List   `tfsdk:"tags"`
}

var _ resource.Resource = &VPCResource{}
var _ resource.ResourceWithImportState = &VPCResource{}

func NewVPCResource() resource.Resource {
	return &VPCResource{}
}

func (r *VPCResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VPC resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VPC identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "VPC name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "VPC location",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the VPC",
				Optional:            true,
			},
		},
	}
}

func (r *VPCResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ArubaCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ArubaCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *VPCResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VPCResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Simulate API response
	data.Id = types.StringValue("vpc-id")
	tflog.Trace(ctx, "created a VPC resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPCResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VPCResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPCResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VPCResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPCResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VPCResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *VPCResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *VPCResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}
