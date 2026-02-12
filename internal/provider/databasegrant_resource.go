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
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Database Grant URI",
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
	// GrantRequest requires both User and Role fields
	createRequest := sdktypes.GrantRequest{
		User: sdktypes.GrantUser{Username: userID},
		Role: sdktypes.GrantRole{Name: data.Role.ValueString()},
	}

	// Create the grant using the SDK
	response, err := r.client.Client.FromDatabase().Grants().Create(ctx, projectID, dbaasID, databaseName, createRequest, nil)
	if err != nil {
		tflog.Error(ctx, "Database grant create error", map[string]interface{}{
			"error":      err.Error(),
			"project_id": projectID,
			"dbaas_id":   dbaasID,
			"database":   databaseName,
			"user_id":    userID,
		})
		resp.Diagnostics.AddError(
			"Error creating database grant",
			fmt.Sprintf("Unable to create database grant: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		logContext := map[string]interface{}{
			"project_id": projectID,
			"dbaas_id":   dbaasID,
			"database":   databaseName,
			"user_id":    userID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to create database grant", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		// Set the grant ID - using a composite key
		data.Id = types.StringValue(fmt.Sprintf("%s/%s/%s/%s", projectID, dbaasID, databaseName, userID))
		data.Uri = types.StringNull() // Grants don't have URIs
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Database grant created but no data returned from API",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Trace(ctx, "created a Database Grant resource", map[string]interface{}{
		"grant_id": data.Id.ValueString(),
	})
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
		tflog.Error(ctx, "Database grant read error", map[string]interface{}{
			"error":      err.Error(),
			"project_id": projectID,
			"dbaas_id":   dbaasID,
			"database":   databaseName,
			"user_id":    userID,
		})
		resp.Diagnostics.AddError(
			"Error reading database grant",
			fmt.Sprintf("Unable to read database grant: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		// Handle 404 - Resource no longer exists
		if response.Error.Status != nil && *response.Error.Status == 404 {
			tflog.Info(ctx, "Database grant not found, removing from state", map[string]interface{}{
				"project_id": projectID,
				"dbaas_id":   dbaasID,
				"database":   databaseName,
				"user_id":    userID,
			})
			resp.State.RemoveResource(ctx)
			return
		}

		logContext := map[string]interface{}{
			"project_id": projectID,
			"dbaas_id":   dbaasID,
			"database":   databaseName,
			"user_id":    userID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to read database grant", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		// Update state with current grant info
		data.Role = types.StringValue(response.Data.Role.Name)
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Database grant not found or no data returned from API",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

	// Get IDs from state (not plan) - IDs are immutable and should always be in state
	projectID := state.ProjectID.ValueString()
	dbaasID := state.DBaaSID.ValueString()
	databaseName := state.Database.ValueString()
	userID := state.UserID.ValueString()

	if projectID == "" || dbaasID == "" || databaseName == "" || userID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, DBaaS ID, Database name, and User ID are required to update the database grant",
		)
		return
	}

	// Build update request
	updateRequest := sdktypes.GrantRequest{
		User: sdktypes.GrantUser{Username: userID},
		Role: sdktypes.GrantRole{Name: data.Role.ValueString()},
	}

	// Update the grant using the SDK
	response, err := r.client.Client.FromDatabase().Grants().Update(ctx, projectID, dbaasID, databaseName, userID, updateRequest, nil)
	if err != nil {
		tflog.Error(ctx, "Database grant update error", map[string]interface{}{
			"error":      err.Error(),
			"project_id": projectID,
			"dbaas_id":   dbaasID,
			"database":   databaseName,
			"user_id":    userID,
		})
		resp.Diagnostics.AddError(
			"Error updating database grant",
			fmt.Sprintf("Unable to update database grant: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		logContext := map[string]interface{}{
			"project_id": projectID,
			"dbaas_id":   dbaasID,
			"database":   databaseName,
			"user_id":    userID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to update database grant", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		// Update state with the response data
		data.Role = types.StringValue(response.Data.Role.Name)
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Database grant updated but no data returned from API",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Trace(ctx, "updated a Database Grant resource", map[string]interface{}{
		"grant_id": data.Id.ValueString(),
	})
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
