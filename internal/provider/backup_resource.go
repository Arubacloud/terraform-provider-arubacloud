// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type BackupResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	ProjectID     types.String `tfsdk:"project_id"`
	Type          types.String `tfsdk:"type"`
	VolumeID      types.String `tfsdk:"volume_id"`
	RetentionDays types.Int64  `tfsdk:"retention_days"`
	BillingPeriod types.String `tfsdk:"billing_period"`
}

type BackupResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &BackupResource{}
var _ resource.ResourceWithImportState = &BackupResource{}

func NewBackupResource() resource.Resource {
	return &BackupResource{}
}

func (r *BackupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup"
}

func (r *BackupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Backup resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Backup identifier",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Backup URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Backup name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Backup location",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the backup resource",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this backup belongs to",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of backup (Full, Incremental)",
				Required:            true,
			},
			"volume_id": schema.StringAttribute{
				MarkdownDescription: "Volume ID for the backup",
				Required:            true,
			},
			"retention_days": schema.Int64Attribute{
				MarkdownDescription: "Retention days for the backup",
				Optional:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period (Hour, Month, Year)",
				Optional:            true,
			},
		},
	}
}

func (r *BackupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BackupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BackupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	volumeID := data.VolumeID.ValueString()

	if projectID == "" || volumeID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Volume ID are required to create a backup",
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

	// Build the backup create request
	createRequest := sdktypes.StorageBackupRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.StorageBackupPropertiesRequest{
			StorageBackupType: sdktypes.StorageBackupType(data.Type.ValueString()),
			Origin: sdktypes.ReferenceResource{
				URI: volumeURI,
			},
		},
	}

	// Add optional fields
	if !data.RetentionDays.IsNull() && !data.RetentionDays.IsUnknown() {
		retentionDays := int(data.RetentionDays.ValueInt64())
		createRequest.Properties.RetentionDays = &retentionDays
	}

	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		billingPeriod := data.BillingPeriod.ValueString()
		createRequest.Properties.BillingPeriod = &billingPeriod
	}

	// Create the backup using the SDK
	response, err := r.client.Client.FromStorage().Backups().Create(ctx, projectID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating backup",
			fmt.Sprintf("Unable to create backup: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create backup"
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
			"Backup created but no data returned from API",
		)
		return
	}

	// Wait for Backup to be active before returning
	// This ensures Terraform doesn't proceed until Backup is ready
	backupID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromStorage().Backups().Get(ctx, projectID, backupID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for Backup to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "Backup", backupID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Backup Not Active",
			fmt.Sprintf("Backup was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a Backup resource", map[string]interface{}{
		"backup_id": data.Id.ValueString(),
		"backup_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BackupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BackupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	backupID := data.Id.ValueString()

	if projectID == "" || backupID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Backup ID are required to read the backup",
		)
		return
	}

	// Get backup details using the SDK
	response, err := r.client.Client.FromStorage().Backups().Get(ctx, projectID, backupID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading backup",
			fmt.Sprintf("Unable to read backup: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read backup"
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
		backup := response.Data

		if backup.Metadata.ID != nil {
			data.Id = types.StringValue(*backup.Metadata.ID)
		}
		if backup.Metadata.URI != nil {
			data.Uri = types.StringValue(*backup.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if backup.Metadata.Name != nil {
			data.Name = types.StringValue(*backup.Metadata.Name)
		}
		if backup.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(backup.Metadata.LocationResponse.Value)
		}
		// Note: StorageBackupType field may not be available in the response
		// If type is needed, it should be stored from the create request
		if backup.Properties.Origin.URI != "" {
			// Extract Volume ID from URI
			parts := strings.Split(backup.Properties.Origin.URI, "/")
			if len(parts) > 0 {
				data.VolumeID = types.StringValue(parts[len(parts)-1])
			}
		}
		if backup.Properties.RetentionDays != nil {
			data.RetentionDays = types.Int64Value(int64(*backup.Properties.RetentionDays))
		}
		if backup.Properties.BillingPeriod != nil {
			data.BillingPeriod = types.StringValue(*backup.Properties.BillingPeriod)
		}

		// Update tags
		if len(backup.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(backup.Metadata.Tags))
			for i, tag := range backup.Metadata.Tags {
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

func (r *BackupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BackupResourceModel
	var state BackupResourceModel

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
	backupID := state.Id.ValueString()

	if projectID == "" || backupID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Backup ID are required to update the backup",
		)
		return
	}

	// Get current backup details
	getResponse, err := r.client.Client.FromStorage().Backups().Get(ctx, projectID, backupID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current backup",
			fmt.Sprintf("Unable to get current backup: %s", err),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"Backup Not Found",
			"Backup not found or no data returned",
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
			"Unable to determine region value for backup",
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
	updateRequest := sdktypes.StorageBackupRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.StorageBackupPropertiesRequest{
			// Properties cannot be updated - use current values
			// Note: StorageBackupType may not be available in result type
			// If needed, preserve from state or use default
			Origin:        current.Properties.Origin,
			RetentionDays: current.Properties.RetentionDays,
			BillingPeriod: current.Properties.BillingPeriod,
		},
	}

	// Update the backup using the SDK
	response, err := r.client.Client.FromStorage().Backups().Update(ctx, projectID, backupID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating backup",
			fmt.Sprintf("Unable to update backup: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to update backup"
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

	if response != nil && response.Data != nil {
		// Update from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BackupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BackupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	backupID := data.Id.ValueString()

	if projectID == "" || backupID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Backup ID are required to delete the backup",
		)
		return
	}

	// Delete the backup using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromStorage().Backups().Delete(ctx, projectID, backupID, nil)
		},
		ExtractSDKError,
		"Backup",
		backupID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting backup",
			fmt.Sprintf("Unable to delete backup: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a Backup resource", map[string]interface{}{
		"backup_id": backupID,
	})
}

func (r *BackupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
