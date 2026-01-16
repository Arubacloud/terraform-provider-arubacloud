// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type DBaaSUserResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	ProjectID types.String `tfsdk:"project_id"`
	DBaaSID   types.String `tfsdk:"dbaas_id"`
	Username  types.String `tfsdk:"username"`
	Password  types.String `tfsdk:"password"`
}

type DBaaSUserResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &DBaaSUserResource{}
var _ resource.ResourceWithImportState = &DBaaSUserResource{}

func NewDBaaSUserResource() resource.Resource {
	return &DBaaSUserResource{}
}

func (r *DBaaSUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dbaasuser"
}

func (r *DBaaSUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "DBaaS User resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "DBaaS User identifier (same as username)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "DBaaS User URI",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this user belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dbaas_id": schema.StringAttribute{
				MarkdownDescription: "DBaaS ID this user belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for the DBaaS user",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for the DBaaS user (must be base64 encoded). The plain password must be 8-20 characters, using at least one number, one uppercase letter, one lowercase letter, and one special character. Spaces are not allowed. Use the `base64encode()` function to encode your plain password.",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *DBaaSUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DBaaSUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DBaaSUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()

	if projectID == "" || dbaasID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and DBaaS ID are required to create a DBaaS user",
		)
		return
	}

	// Password should already be base64 encoded from Terraform (using base64encode() function)
	// Pass it through to the API as-is
	createRequest := sdktypes.UserRequest{
		Username: data.Username.ValueString(),
		Password: data.Password.ValueString(),
	}

	// Create the user using the SDK
	response, err := r.client.Client.FromDatabase().Users().Create(ctx, projectID, dbaasID, createRequest, nil)
	if err != nil {
		tflog.Error(ctx, "DBaaS user create error", map[string]interface{}{
			"error":      err.Error(),
			"project_id": projectID,
			"dbaas_id":   dbaasID,
		})
		resp.Diagnostics.AddError(
			"Error creating DBaaS user",
			fmt.Sprintf("Unable to create DBaaS user: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create DBaaS user"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		
		// Log detailed error information for debugging
		errorDetails := map[string]interface{}{
			"project_id": projectID,
			"dbaas_id":   dbaasID,
			"username":   data.Username.ValueString(),
		}
		if response.Error.Title != nil {
			errorDetails["error_title"] = *response.Error.Title
		}
		if response.Error.Detail != nil {
			errorDetails["error_detail"] = *response.Error.Detail
		}
		if response.Error.Status != nil {
			errorDetails["error_status"] = *response.Error.Status
		}
		if response.Error.Type != nil {
			errorDetails["error_type"] = *response.Error.Type
		}
		
		// Log full request and error response JSON only on errors for debugging
		if requestJSON, jsonErr := json.MarshalIndent(createRequest, "", "  "); jsonErr == nil {
			tflog.Debug(ctx, "Full DBaaS user create request JSON (error case)", map[string]interface{}{
				"request_json": string(requestJSON),
			})
		}
		if errorJSON, jsonErr := json.MarshalIndent(response.Error, "", "  "); jsonErr == nil {
			tflog.Debug(ctx, "Full API error response JSON", map[string]interface{}{
				"error_json": string(errorJSON),
			})
		}
		
		tflog.Error(ctx, "DBaaS user create request failed", errorDetails)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		// User uses username as ID
		data.Id = types.StringValue(response.Data.Username)
		// UserResponse doesn't have Metadata.URI
		data.Uri = types.StringNull()
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"DBaaS user created but no data returned from API",
		)
		return
	}

	// Wait for DBaaS User to be active before returning
	// This ensures Terraform doesn't proceed until User is ready
	username := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromDatabase().Users().Get(ctx, projectID, dbaasID, username, nil)
		if err != nil {
			return "", err
		}
		// Users don't have a status field, so if we can get it, it's ready
		if getResp != nil && getResp.Data != nil {
			return "Active", nil
		}
		return "Unknown", nil
	}

	// Wait for DBaaS User to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "DBaaSUser", username, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"DBaaS User Not Active",
			fmt.Sprintf("DBaaS user was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a DBaaS User resource", map[string]interface{}{
		"user_id":  data.Id.ValueString(),
		"username": data.Username.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DBaaSUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DBaaSUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()
	username := data.Id.ValueString()

	if projectID == "" || dbaasID == "" || username == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, DBaaS ID, and Username are required to read the DBaaS user",
		)
		return
	}

	// Get user details using the SDK
	response, err := r.client.Client.FromDatabase().Users().Get(ctx, projectID, dbaasID, username, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading DBaaS user",
			fmt.Sprintf("Unable to read DBaaS user: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read DBaaS user"
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
		user := response.Data
		data.Id = types.StringValue(user.Username)
		// UserResponse doesn't have Metadata.URI
		data.Uri = types.StringNull()
		data.Username = types.StringValue(user.Username)
		// Password is not returned from API, so we keep the existing value
		// Preserve immutable fields from state (dbaas_id and project_id are not in API response)
		// These are already set from req.State.Get above, but we ensure they're preserved
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DBaaSUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// DBaaS users cannot be updated - they can only be created or deleted
	// If you need to change a user's password or other attributes, delete and recreate the user
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"DBaaS users cannot be updated. To change a user's password or other attributes, delete the existing user and create a new one.",
	)
}

func (r *DBaaSUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DBaaSUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()
	username := data.Id.ValueString()

	if projectID == "" || dbaasID == "" || username == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, DBaaS ID, and Username are required to delete the DBaaS user",
		)
		return
	}

	// Delete the user using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromDatabase().Users().Delete(ctx, projectID, dbaasID, username, nil)
		},
		ExtractSDKError,
		"DBaaSUser",
		username,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting DBaaS user",
			fmt.Sprintf("Unable to delete DBaaS user: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a DBaaS User resource", map[string]interface{}{
		"user_id": username,
	})
}

func (r *DBaaSUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
