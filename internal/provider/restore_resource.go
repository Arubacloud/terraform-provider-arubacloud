package provider

import (
	"context"
	"fmt"
	"time"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
		MarkdownDescription: "Manages an ArubaCloud Block Storage Restore operation — restores a backup to a block storage volume.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource.",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the restore operation.",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"backup_id": schema.StringAttribute{
				MarkdownDescription: "ID of the backup to restore from.",
				Required:            true,
			},
			"volume_id": schema.StringAttribute{
				MarkdownDescription: "ID of the target block storage volume to restore the backup onto.",
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

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the backup details to get the full URI
	backupResponse, err := r.client.Client.FromStorage().Backups().Get(ctx, projectID, backupID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting backup details",
			NewTransportError("read", "Restore", err).Error(),
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
			NewTransportError("read", "Restore", err).Error(),
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
			NewTransportError("create", "Restore", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("create", "Restore", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
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
		ReportWaitResult(&resp.Diagnostics, err, "Restore", restoreID)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

	if data.Id.IsUnknown() || data.Id.IsNull() || restoreID == "" {
		tflog.Debug(ctx, "Restore ID is empty, removing resource from state", map[string]interface{}{"restore_id": restoreID})
		resp.State.RemoveResource(ctx)
		return
	}

	if projectID == "" || backupID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Backup ID are required to read the restore",
		)
		return
	}

	// Get restore details using the SDK
	response, err := r.client.Client.FromStorage().Restores().Get(ctx, projectID, backupID, restoreID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading restore",
			NewTransportError("read", "Restore", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("read", "Restore", response); apiErr != nil {
		if IsNotFound(apiErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}

	// If the resource is still provisioning (e.g. after a Create timeout that saved
	// partial state), resume the wait so the next terraform apply reconciles correctly.
	if response.Data.Status.State != nil {
		switch st := *response.Data.Status.State; {
		case isFailedState(st):
			resp.Diagnostics.AddError(
				"Resource in Failed State",
				fmt.Sprintf("Restore %q reached a terminal failure state (%s) and will not recover on its own. "+
					"Use `terraform apply -replace=<address>` to recreate it.", restoreID, st),
			)
			return
		case IsCreatingState(st):
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
			if err := WaitForResourceActive(ctx, checker, "Restore", restoreID, r.client.ResourceTimeout); err != nil {
				ReportWaitResult(&resp.Diagnostics, err, "Restore", restoreID)
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
			// Re-read to get the final active state.
			response, err = r.client.Client.FromStorage().Restores().Get(ctx, projectID, backupID, restoreID, nil)
			if err != nil {
				resp.Diagnostics.AddError("Error reading Restore after provisioning wait",
					NewTransportError("read", "Restore", err).Error())
				return
			}
			if apiErr := CheckResponse("read", "Restore", response); apiErr != nil {
				if IsNotFound(apiErr) {
					resp.State.RemoveResource(ctx)
					return
				}
				resp.Diagnostics.AddError("API Error", apiErr.Error())
				return
			}
		}
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

		data.Tags = TagsToList(restore.Metadata.Tags)
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
			NewTransportError("read", "Restore", err).Error(),
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

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if tags == nil {
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
			NewTransportError("update", "Restore", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("update", "Restore", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
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
	deletionChecker := func(ctx context.Context) (bool, error) {
		getResp, getErr := r.client.Client.FromStorage().Restores().Get(ctx, projectID, backupID, restoreID, nil)
		if getErr != nil {
			return false, NewTransportError("get", "Restore", getErr)
		}
		if provErr := CheckResponse("get", "Restore", getResp); provErr != nil {
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
			resp, err := r.client.Client.FromStorage().Restores().Delete(ctx, projectID, backupID, restoreID, nil)
			if err != nil {
				return NewTransportError("delete", "Restore", err)
			}
			return CheckResponse("delete", "Restore", resp)
		},
		"Restore",
		restoreID,
		r.client.ResourceTimeout,
		deletionChecker,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting restore",
			NewTransportError("delete", "Restore", err).Error(),
		)
		return
	}

	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "Restore", restoreID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for Restore deletion", waitErr.Error())
		return
	}

	tflog.Trace(ctx, "deleted a Restore resource", map[string]interface{}{
		"restore_id": restoreID,
	})
}

func (r *RestoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
