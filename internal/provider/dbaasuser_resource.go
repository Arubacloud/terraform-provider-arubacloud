package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "DBaaS User URI",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this user belongs to",
				Required:            true,
			},
			"dbaas_id": schema.StringAttribute{
				MarkdownDescription: "DBaaS ID this user belongs to",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for the DBaaS user",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for the DBaaS user",
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

	// Build the create request
	// Password must be base64 encoded for the API
	passwordBase64 := base64.StdEncoding.EncodeToString([]byte(data.Password.ValueString()))
	createRequest := sdktypes.UserRequest{
		Username: data.Username.ValueString(),
		Password: passwordBase64,
	}

	// Create the user using the SDK, retrying on transient errors (e.g. when the
	// parent DBaaS is still in InCreation and returns 400 category:transient).
	var response *sdktypes.Response[sdktypes.UserResponse]
	if createErr := CreateWithTransientRetry(
		ctx,
		func() error {
			var err error
			response, err = r.client.Client.FromDatabase().Users().Create(ctx, projectID, dbaasID, createRequest, nil)
			if err != nil {
				return NewTransportError("create", "Dbaasuser", err)
			}
			return CheckResponse("create", "Dbaasuser", response)
		},
		"DBaaSUser",
		data.Username.ValueString(),
		r.client.ResourceTimeout,
	); createErr != nil {
		resp.Diagnostics.AddError("Error creating DBaaS user", createErr.Error())
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
		ReportWaitResult(&resp.Diagnostics, err, "DBaaSUser", username)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

	if data.Id.IsUnknown() || data.Id.IsNull() || username == "" {
		tflog.Debug(ctx, "DBaaS User ID is empty, removing resource from state", map[string]interface{}{"username": username})
		resp.State.RemoveResource(ctx)
		return
	}

	if projectID == "" || dbaasID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and DBaaS ID are required to read the DBaaS user",
		)
		return
	}

	// Get user details using the SDK
	response, err := r.client.Client.FromDatabase().Users().Get(ctx, projectID, dbaasID, username, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading DBaaS user",
			NewTransportError("read", "Dbaasuser", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("read", "Dbaasuser", response); apiErr != nil {
		if IsNotFound(apiErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}

	if response != nil && response.Data != nil {
		user := response.Data
		data.Id = types.StringValue(user.Username)
		// UserResponse doesn't have Metadata.URI
		data.Uri = types.StringNull()
		data.Username = types.StringValue(user.Username)
		// Password is not returned from API, so we keep the existing value
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DBaaSUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DBaaSUserResourceModel
	var state DBaaSUserResourceModel

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
	username := state.Id.ValueString()

	if projectID == "" || dbaasID == "" || username == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, DBaaS ID, and Username are required to update the DBaaS user",
		)
		return
	}

	// Only password can be updated
	// Password must be base64 encoded for the API
	passwordBase64 := base64.StdEncoding.EncodeToString([]byte(data.Password.ValueString()))
	updateRequest := sdktypes.UserRequest{
		Username: username,
		Password: passwordBase64,
	}

	// Update the user using the SDK
	response, err := r.client.Client.FromDatabase().Users().Update(ctx, projectID, dbaasID, username, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating DBaaS user",
			NewTransportError("update", "Dbaasuser", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("update", "Dbaasuser", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}

	if response != nil && response.Data != nil {
		data.Id = types.StringValue(response.Data.Username)
		// UserResponse doesn't have Metadata.URI
		data.Uri = types.StringNull()
		data.Username = types.StringValue(response.Data.Username)
	}

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.ProjectID = state.ProjectID
	data.DBaaSID = state.DBaaSID

	if response != nil && response.Data != nil {
		// Update ID from response (username can change, so use response)
		data.Id = types.StringValue(response.Data.Username)
		data.Username = types.StringValue(response.Data.Username)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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
	deletionChecker := func(ctx context.Context) (bool, error) {
		getResp, getErr := r.client.Client.FromDatabase().Users().Get(ctx, projectID, dbaasID, username, nil)
		if getErr != nil {
			return false, NewTransportError("get", "DBaaSUser", getErr)
		}
		if provErr := CheckResponse("get", "DBaaSUser", getResp); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(
		ctx,
		func() error {
			resp, err := r.client.Client.FromDatabase().Users().Delete(ctx, projectID, dbaasID, username, nil)
			if err != nil {
				return NewTransportError("delete", "DBaaSUser", err)
			}
			return CheckResponse("delete", "DBaaSUser", resp)
		},
		"DBaaSUser",
		username,
		r.client.ResourceTimeout,
		deletionChecker,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting DBaaS user",
			NewTransportError("delete", "Dbaasuser", err).Error(),
		)
		return
	}

	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "DBaaSUser", username, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for DBaaSUser deletion", waitErr.Error())
		return
	}

	tflog.Trace(ctx, "deleted a DBaaS User resource", map[string]interface{}{
		"user_id": username,
	})
}

func (r *DBaaSUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
