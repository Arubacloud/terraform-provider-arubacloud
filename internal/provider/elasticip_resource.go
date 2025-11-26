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

var _ resource.Resource = &ElasticIPResource{}
var _ resource.ResourceWithImportState = &ElasticIPResource{}

func NewElasticIPResource() resource.Resource {
	return &ElasticIPResource{}
}

type ElasticIPResource struct {
	client *ArubaCloudClient
}

type ElasticIPResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	Address       types.String `tfsdk:"address"`
	ProjectId     types.String `tfsdk:"project_id"`
}

func (r *ElasticIPResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticip"
}

func (r *ElasticIPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Project resource",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Elastic IP name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Elastic IP location",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the Elastic IP",
				Optional:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period for the Elastic IP (only 'hourly' allowed)",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this Elastic IP belongs to",
				Required:            true,
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "Elastic IP address",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Project Identifier",
				Computed:            true,
			},
		},
	}
}

func (r *ElasticIPResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ElasticIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ElasticIPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Simulate API response
	data.Id = types.StringValue("elasticip-id")
	data.Address = types.StringValue("203.0.113.1")
	tflog.Trace(ctx, "created an elastic IP resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ElasticIPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ElasticIPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ElasticIPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ElasticIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
