// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type DatabaseResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	ProjectID types.String `tfsdk:"project_id"`
	DBaaSID   types.String `tfsdk:"dbaas_id"`
	Name      types.String `tfsdk:"name"`
}

type DatabaseResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &DatabaseResource{}
var _ resource.ResourceWithImportState = &DatabaseResource{}

func NewDatabaseResource() resource.Resource {
	return &DatabaseResource{}
}

func (r *DatabaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *DatabaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Database resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Database identifier (same as name)",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Database URI",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this database belongs to",
				Required:            true,
			},
			"dbaas_id": schema.StringAttribute{
				MarkdownDescription: "DBaaS ID this database belongs to",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Database name",
				Required:            true,
			},
		},
	}
}

func (r *DatabaseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DatabaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()

	if projectID == "" || dbaasID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and DBaaS ID are required to create a database",
		)
		return
	}

	// Build the create request
	createRequest := sdktypes.DatabaseRequest{
		Name: data.Name.ValueString(),
	}

	// Create the database using the SDK
	response, err := r.client.Client.FromDatabase().Databases().Create(ctx, projectID, dbaasID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating database",
			fmt.Sprintf("Unable to create database: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create database"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		// Database uses name as ID
		data.Id = types.StringValue(response.Data.Name)
		// Database response doesn't have Metadata.URI
		data.Uri = types.StringNull()
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Database created but no data returned from API",
		)
		return
	}

	// Wait for Database to be active before returning
	// This ensures Terraform doesn't proceed until Database is ready
	databaseName := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromDatabase().Databases().Get(ctx, projectID, dbaasID, databaseName, nil)
		if err != nil {
			return "", err
		}
		// Databases don't have a status field, so if we can get it, it's ready
		if getResp != nil && getResp.Data != nil {
			return "Active", nil
		}
		return "Unknown", nil
	}

	// Wait for Database to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "Database", databaseName, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Database Not Active",
			fmt.Sprintf("Database was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a Database resource", map[string]interface{}{
		"database_id": data.Id.ValueString(),
		"database_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()
	databaseName := data.Id.ValueString()

	if projectID == "" || dbaasID == "" || databaseName == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, DBaaS ID, and Database Name are required to read the database",
		)
		return
	}

	// Get database details using the SDK
	response, err := r.client.Client.FromDatabase().Databases().Get(ctx, projectID, dbaasID, databaseName, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading database",
			fmt.Sprintf("Unable to read database: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read database"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		db := response.Data
		data.Id = types.StringValue(db.Name)
		data.Name = types.StringValue(db.Name)
		// Database response doesn't have Metadata.URI
		data.Uri = types.StringNull()
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DatabaseResourceModel
	var state DatabaseResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get IDs from state (not plan) - IDs are immutable and should always be in state
	projectID := state.ProjectID.ValueString()
	dbaasID := state.DBaaSID.ValueString()
	oldDatabaseName := state.Id.ValueString()

	if projectID == "" || dbaasID == "" || oldDatabaseName == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, DBaaS ID, and Database Name are required to update the database",
		)
		return
	}

	// Build update request
	updateRequest := sdktypes.DatabaseRequest{
		Name: data.Name.ValueString(),
	}

	// Update the database using the SDK
	response, err := r.client.Client.FromDatabase().Databases().Update(ctx, projectID, dbaasID, oldDatabaseName, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating database",
			fmt.Sprintf("Unable to update database: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to update database"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		// Update ID to new name
		data.Id = types.StringValue(response.Data.Name)
		// Database response doesn't have Metadata.URI
		data.Uri = types.StringNull()
	}

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.ProjectID = state.ProjectID
	data.DBaaSID = state.DBaaSID

	if response != nil && response.Data != nil {
		// Update ID from response (database name can change, so use response)
		data.Id = types.StringValue(response.Data.Name)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()
	databaseName := data.Id.ValueString()

	if projectID == "" || dbaasID == "" || databaseName == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, DBaaS ID, and Database Name are required to delete the database",
		)
		return
	}

	// Delete the database using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromDatabase().Databases().Delete(ctx, projectID, dbaasID, databaseName, nil)
		},
		ExtractSDKError,
		"Database",
		databaseName,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting database",
			fmt.Sprintf("Unable to delete database: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a Database resource", map[string]interface{}{
		"database_id": databaseName,
	})
}

func (r *DatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
