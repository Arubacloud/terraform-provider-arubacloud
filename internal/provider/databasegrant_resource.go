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

type DatabaseGrantResourceModel struct {
	Id       types.String `tfsdk:"id"`
	Database types.String `tfsdk:"database"`
	UserID   types.String `tfsdk:"user_id"`
	Role     types.String `tfsdk:"role"`
}

type DatabaseGrantResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &DatabaseGrantResource{}
var _ resource.ResourceWithImportState = &DatabaseGrantResource{}

func NewDatabaseGrantResource() resource.Resource {
	return &DatabaseGrantResource{}
}

func (r *DatabaseGrantResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databasegrant"
}

func (r *DatabaseGrantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Database Grant resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Database Grant identifier",
				Computed:            true,
			},
			"database": schema.StringAttribute{
				MarkdownDescription: "Database name or ID",
				Required:            true,
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "User ID to grant access",
				Required:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Role to grant (e.g., read, write, admin)",
				Required:            true,
			},
		},
	}
}

func (r *DatabaseGrantResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DatabaseGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DatabaseGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Simulate API response
	data.Id = types.StringValue("databasegrant-id")
	tflog.Trace(ctx, "created a Database Grant resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatabaseGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DatabaseGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatabaseGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DatabaseGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
