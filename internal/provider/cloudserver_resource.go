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

type CloudServerResourceModel struct {
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

type CloudServerResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &CloudServerResource{}
var _ resource.ResourceWithImportState = &CloudServerResource{}

func NewCloudServerResource() resource.Resource {
	return &CloudServerResource{}
}

func (r *CloudServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudserver"
}

func (r *CloudServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CloudServer resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "CloudServer identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "CloudServer name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "CloudServer location",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID",
				Required:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "VPC ID",
				Required:            true,
			},
			"flavor_name": schema.StringAttribute{
				MarkdownDescription: "Flavor name",
				Required:            true,
			},
			"elastic_ip_id": schema.StringAttribute{
				MarkdownDescription: "Elastic IP ID",
				Optional:            true,
			},
			"boot_volume": schema.StringAttribute{
				MarkdownDescription: "Boot volume ID",
				Required:            true,
			},
			"key_pair_id": schema.StringAttribute{
				MarkdownDescription: "Key pair ID",
				Optional:            true,
			},
			"subnets": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of subnet IDs",
				Required:            true,
			},
			"securitygroups": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of security group reference IDs",
				Required:            true,
			},
		},
	}
}

func (r *CloudServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CloudServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CloudServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Simulate API response
	data.Id = types.StringValue("cloudserver-id")
	tflog.Trace(ctx, "created a CloudServer resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CloudServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CloudServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CloudServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CloudServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CloudServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CloudServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *CloudServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
