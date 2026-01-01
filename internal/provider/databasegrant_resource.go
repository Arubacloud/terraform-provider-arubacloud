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
	Id        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	DBaaSID   types.String `tfsdk:"dbaas_id"`
	Database  types.String `tfsdk:"database"`
	UserID    types.String `tfsdk:"user_id"`
	Role      types.String `tfsdk:"role"`
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
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this grant belongs to",
				Required:            true,
			},
			"dbaas_id": schema.StringAttribute{
				MarkdownDescription: "DBaaS ID this grant belongs to",
				Required:            true,
			},
			"database": schema.StringAttribute{
				MarkdownDescription: "Database name",
				Required:            true,
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "User ID (username) to grant access",
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
			fmt.Sprintf("Expected *ArubaCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()
	databaseName := data.Database.ValueString()
	userID := data.UserID.ValueString()

	if projectID == "" || dbaasID == "" || databaseName == "" || userID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, DBaaS ID, Database name, and User ID are required to create a database grant",
		)
		return
	}

	// Build the create request
	// NOTE: GrantRole is a custom type that doesn't accept string conversion
	// This requires checking the SDK for GrantRole enum constants or if Role field accepts string
	// TODO: Fix GrantRole - check if Role field in GrantRequest is actually string type
	// or if we need to use GrantRole enum constants like sdktypes.GrantRoleRead, etc.
	roleStr := data.Role.ValueString()
	
	resp.Diagnostics.AddError(
		"Unimplemented: GrantRole Type",
		fmt.Sprintf("DatabaseGrant resource requires proper GrantRole type handling. Role value: %s. Please check SDK for GrantRole enum values (e.g., sdktypes.GrantRoleRead) or if Role field accepts string directly.", roleStr),
	)
	return
}

func (r *DatabaseGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatabaseGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()
	databaseName := data.Database.ValueString()
	userID := data.UserID.ValueString()

	if projectID == "" || dbaasID == "" || databaseName == "" || userID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, DBaaS ID, Database name, and User ID are required to read the database grant",
		)
		return
	}

	// Get grant details using the SDK
	// Grants().Get signature: (ctx, projectID, dbaasID, databaseName, userID, nil)
	// NOTE: DatabaseGrant resource is temporarily disabled due to GrantRole type issue
	resp.Diagnostics.AddError(
		"Unimplemented: DatabaseGrant Resource",
		"DatabaseGrant resource is temporarily disabled. GrantRole type conversion needs to be resolved. Please check SDK for GrantRole enum values.",
	)
	return
}

func (r *DatabaseGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DatabaseGrantResourceModel
	var state DatabaseGrantResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()
	databaseName := data.Database.ValueString()
	userID := data.UserID.ValueString()

	if projectID == "" || dbaasID == "" || databaseName == "" || userID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, DBaaS ID, Database name, and User ID are required to update the database grant",
		)
		return
	}

	// Build update request
	// NOTE: Same GrantRole issue as Create
	roleStr := data.Role.ValueString()
	resp.Diagnostics.AddError(
		"Unimplemented: GrantRole Type",
		fmt.Sprintf("DatabaseGrant update requires proper GrantRole type handling. Role value: %s. Please check SDK for GrantRole enum values.", roleStr),
	)
	return
}

func (r *DatabaseGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatabaseGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()
	databaseName := data.Database.ValueString()
	userID := data.UserID.ValueString()

	if projectID == "" || dbaasID == "" || databaseName == "" || userID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, DBaaS ID, Database name, and User ID are required to delete the database grant",
		)
		return
	}

	// Delete the grant using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	grantID := data.Id.ValueString()
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromDatabase().Grants().Delete(ctx, projectID, dbaasID, databaseName, userID, nil)
		},
		ExtractSDKError,
		"DatabaseGrant",
		grantID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting database grant",
			fmt.Sprintf("Unable to delete database grant: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a Database Grant resource", map[string]interface{}{
		"grant_id": grantID,
	})
}

func (r *DatabaseGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
