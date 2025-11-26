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

type KMIPResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	ProjectID types.String `tfsdk:"project_id"`
	KMSID     types.String `tfsdk:"kms_id"`
}

type KMIPResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &KMIPResource{}
var _ resource.ResourceWithImportState = &KMIPResource{}

func NewKMIPResource() resource.Resource {
	return &KMIPResource{}
}

func (r *KMIPResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kmip"
}

func (r *KMIPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "KMIP resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "KMIP identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "KMIP name",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this KMIP belongs to",
				Required:            true,
			},
			"kms_id": schema.StringAttribute{
				MarkdownDescription: "ID of the associated KMS",
				Required:            true,
			},
		},
	}
}

func (r *KMIPResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KMIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KMIPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Simulate API response
	data.Id = types.StringValue("kmip-id")
	tflog.Trace(ctx, "created a KMIP resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KMIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KMIPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KMIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KMIPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KMIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KMIPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *KMIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
