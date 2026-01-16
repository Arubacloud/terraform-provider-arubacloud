// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type RestoreResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	Tags      types.List   `tfsdk:"tags"`
	ProjectID types.String `tfsdk:"project_id"`
	BackupID  types.String `tfsdk:"backup_id"`
	VolumeID  types.String `tfsdk:"volume_id"`
}

type RestoreResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &RestoreResource{}
var _ resource.ResourceWithImportState = &RestoreResource{}

func NewRestoreResource() resource.Resource {
	return &RestoreResource{}
}

func (r *RestoreResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_restore"
}

func (r *RestoreResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Restore resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Restore identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Restore URI",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Restore name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Restore location",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the restore resource",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this restore belongs to",
				Required:            true,
			},
			"backup_id": schema.StringAttribute{
				MarkdownDescription: "Backup ID to restore from",
				Required:            true,
			},
			"volume_id": schema.StringAttribute{
				MarkdownDescription: "Volume ID to restore to",
				Required:            true,
			},
		},
	}
}

func (r *RestoreResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RestoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RestoreResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	backupID := data.BackupID.ValueString()
	volumeID := data.VolumeID.ValueString()

	if projectID == "" || backupID == "" || volumeID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, Backup ID, and Volume ID are required to create a restore",
		)
		return
	}

	// Extract tags
	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Get the backup details to get the full URI
	backupResponse, err := r.client.Client.FromStorage().Backups().Get(ctx, projectID, backupID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting backup details",
			fmt.Sprintf("Unable to get backup details: %s", err),
		)
		return
	}

	if backupResponse == nil || backupResponse.Data == nil {
		resp.Diagnostics.AddError(
			"Backup Not Found",
			"Backup not found",
		)
		return
	}

	// backupURI is not used - removed unused variable

	// Get the volume details to get the full URI
	volumeResponse, err := r.client.Client.FromStorage().Volumes().Get(ctx, projectID, volumeID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting volume details",
			fmt.Sprintf("Unable to get volume details: %s", err),
		)
		return
	}

	if volumeResponse == nil || volumeResponse.Data == nil {
		resp.Diagnostics.AddError(
			"Volume Not Found",
			"Volume not found",
		)
		return
	}

	volumeURI := ""
	if volumeResponse.Data.Metadata.URI != nil {
		volumeURI = *volumeResponse.Data.Metadata.URI
	} else {
		resp.Diagnostics.AddError(
			"Invalid Volume Response",
			"Volume URI not found in response",
		)
		return
	}

	// Build the restore request
	createRequest := sdktypes.RestoreRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.RestorePropertiesRequest{
			Target: sdktypes.ReferenceResource{
				URI: volumeURI,
			},
		},
	}

	// Create the restore using the SDK
	response, err := r.client.Client.FromStorage().Restores().Create(ctx, projectID, backupID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating restore",
			fmt.Sprintf("Unable to create restore: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create restore"
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
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Restore created but no data returned from API",
		)
		return
	}

	// Wait for Restore to be active before returning
	// This ensures Terraform doesn't proceed until Restore is ready
	restoreID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromStorage().Restores().Get(ctx, projectID, backupID, restoreID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for Restore to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "Restore", restoreID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Restore Not Active",
			fmt.Sprintf("Restore was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a Restore resource", map[string]interface{}{
		"restore_id":   data.Id.ValueString(),
		"restore_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RestoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RestoreResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	backupID := data.BackupID.ValueString()
	restoreID := data.Id.ValueString()

	if projectID == "" || backupID == "" || restoreID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, Backup ID, and Restore ID are required to read the restore",
		)
		return
	}

	// Get restore details using the SDK
	response, err := r.client.Client.FromStorage().Restores().Get(ctx, projectID, backupID, restoreID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading restore",
			fmt.Sprintf("Unable to read restore: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read restore"
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
		restore := response.Data

		if restore.Metadata.ID != nil {
			data.Id = types.StringValue(*restore.Metadata.ID)
		}
		if restore.Metadata.URI != nil {
			data.Uri = types.StringValue(*restore.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if restore.Metadata.Name != nil {
			data.Name = types.StringValue(*restore.Metadata.Name)
		}
		if restore.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(restore.Metadata.LocationResponse.Value)
		}
		// Note: Target field may not be available in RestorePropertiesResult
		// VolumeID is stored from the create request, preserve from state if needed
		// If Target is needed, check SDK for correct field name

		// Update tags
		if len(restore.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(restore.Metadata.Tags))
			for i, tag := range restore.Metadata.Tags {
				tagValues[i] = types.StringValue(tag)
			}
			tagsList, diags := types.ListValueFrom(ctx, types.StringType, tagValues)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = tagsList
			}
		} else {
			emptyList, diags := types.ListValue(types.StringType, []attr.Value{})
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = emptyList
			}
		}
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RestoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RestoreResourceModel
	var state RestoreResourceModel

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
	backupID := state.BackupID.ValueString()
	restoreID := state.Id.ValueString()

	if projectID == "" || backupID == "" || restoreID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, Backup ID, and Restore ID are required to update the restore",
		)
		return
	}

	// Get current restore details
	getResponse, err := r.client.Client.FromStorage().Restores().Get(ctx, projectID, backupID, restoreID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current restore",
			fmt.Sprintf("Unable to get current restore: %s", err),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"Restore Not Found",
			"Restore not found or no data returned",
		)
		return
	}

	current := getResponse.Data

	// Get region value
	regionValue := ""
	if current.Metadata.LocationResponse != nil {
		regionValue = current.Metadata.LocationResponse.Value
	}
	if regionValue == "" {
		resp.Diagnostics.AddError(
			"Missing Region",
			"Unable to determine region value for restore",
		)
		return
	}

	// Extract tags
	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		tags = current.Metadata.Tags
	}

	// Build update request - only name and tags can be updated
	updateRequest := sdktypes.RestoreRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.RestorePropertiesRequest{
			// Properties cannot be updated
			// Note: Target field may not be available in RestorePropertiesResult
			// Preserve from state or use original create request values
		},
	}

	// Update the restore using the SDK
	response, err := r.client.Client.FromStorage().Restores().Update(ctx, projectID, backupID, restoreID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating restore",
			fmt.Sprintf("Unable to update restore: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to update restore"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.ProjectID = state.ProjectID
	data.BackupID = state.BackupID
	data.Uri = state.Uri           // Preserve URI from state
	data.VolumeID = state.VolumeID // Immutable

	if response != nil && response.Data != nil {
		// Update from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		}
	} else {
		// If no response, re-read the restore to get the latest state including URI
		getResp, err := r.client.Client.FromStorage().Restores().Get(ctx, projectID, backupID, restoreID, nil)
		if err == nil && getResp != nil && getResp.Data != nil {
			if getResp.Data.Metadata.URI != nil {
				data.Uri = types.StringValue(*getResp.Data.Metadata.URI)
			} else {
				data.Uri = state.Uri // Fallback to state if not in response
			}
		} else {
			// If re-read fails, preserve from state
			data.Uri = state.Uri
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RestoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RestoreResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	backupID := data.BackupID.ValueString()
	restoreID := data.Id.ValueString()

	if projectID == "" || backupID == "" || restoreID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, Backup ID, and Restore ID are required to delete the restore",
		)
		return
	}

	// Delete the restore using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromStorage().Restores().Delete(ctx, projectID, backupID, restoreID, nil)
		},
		ExtractSDKError,
		"Restore",
		restoreID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting restore",
			fmt.Sprintf("Unable to delete restore: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a Restore resource", map[string]interface{}{
		"restore_id": restoreID,
	})
}

func (r *RestoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
