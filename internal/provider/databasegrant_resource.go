// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type DatabaseGrantResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Database Grant URI",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this grant belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dbaas_id": schema.StringAttribute{
				MarkdownDescription: "DBaaS ID this grant belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database": schema.StringAttribute{
				MarkdownDescription: "Database name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "User ID (username) to grant access",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Role to grant: readonly, readwrite, or liteadmin",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("readonly", "readwrite", "liteadmin"),
				},
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
	roleStr := data.Role.ValueString()
	createRequest := sdktypes.GrantRequest{
		User: sdktypes.GrantUser{
			Username: userID,
		},
		Role: sdktypes.GrantRole{
			Name: roleStr,
		},
	}

	// Serialize request to JSON for debugging
	requestJSON, _ := json.Marshal(createRequest)

	// Debug logging for request
	tflog.Debug(ctx, "Creating database grant", map[string]interface{}{
		"project_id":    projectID,
		"dbaas_id":      dbaasID,
		"database_name": databaseName,
		"user_id":       userID,
		"role":          roleStr,
		"request":       fmt.Sprintf("%+v", createRequest),
		"request_json":  string(requestJSON),
	})

	// Create the grant using the SDK
	response, err := r.client.Client.FromDatabase().Grants().Create(ctx, projectID, dbaasID, databaseName, createRequest, nil)

	// Serialize response to JSON for debugging
	responseJSON, _ := json.Marshal(response)

	// Debug logging for response
	tflog.Debug(ctx, "Create database grant response", map[string]interface{}{
		"error":         fmt.Sprintf("%v", err),
		"response":      fmt.Sprintf("%+v", response),
		"response_json": string(responseJSON),
	})

	if err != nil {
		tflog.Error(ctx, "SDK returned error", map[string]interface{}{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError(
			"Error creating database grant",
			fmt.Sprintf("Unable to create database grant: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create database grant"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}

		// Enhanced error logging with full response details
		errorJSON, _ := json.Marshal(response.Error)
		tflog.Error(ctx, "API Error creating database grant", map[string]interface{}{
			"error_title":  response.Error.Title,
			"error_detail": response.Error.Detail,
			"error_status": response.Error.Status,
			"full_error":   fmt.Sprintf("%+v", response.Error),
			"error_json":   string(errorJSON),
		})

		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		// Set ID as composite: projectID/dbaasID/databaseName/userID
		grantID := fmt.Sprintf("%s/%s/%s/%s", projectID, dbaasID, databaseName, userID)
		data.Id = types.StringValue(grantID)
		data.Uri = types.StringNull() // Grants don't have URIs
		data.ProjectID = types.StringValue(projectID)
		data.DBaaSID = types.StringValue(dbaasID)
		data.Database = types.StringValue(databaseName)
		data.UserID = types.StringValue(userID)
		data.Role = types.StringValue(response.Data.Role.Name)
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"The API response did not contain expected data",
		)
		return
	}

	tflog.Trace(ctx, "created a Database Grant resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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
	response, err := r.client.Client.FromDatabase().Grants().Get(ctx, projectID, dbaasID, databaseName, userID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading database grant",
			fmt.Sprintf("Unable to read database grant: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		// Check if it's a 404 - resource not found
		if response.Error.Status != nil && *response.Error.Status == 404 {
			// Resource no longer exists, remove from state
			resp.State.RemoveResource(ctx)
			return
		}

		errorMsg := "Failed to read database grant"
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
		// Update state with current values
		data.Role = types.StringValue(response.Data.Role.Name)
		data.UserID = types.StringValue(response.Data.User.Username)
		data.Database = types.StringValue(response.Data.Database.Name)
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"The API response did not contain expected data",
		)
		return
	}

	tflog.Trace(ctx, "read a Database Grant resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Database grants cannot be updated - they can only be created or deleted
	// If you need to change a grant's role or other attributes, delete the existing grant and create a new one
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Database grants cannot be updated. To change a grant's role or other attributes, delete the existing grant and create a new one.",
	)
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
