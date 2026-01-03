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

var _ resource.Resource = &SnapshotResource{}
var _ resource.ResourceWithImportState = &SnapshotResource{}

func NewSnapshotResource() resource.Resource {
	return &SnapshotResource{}
}

type SnapshotResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	ProjectId     types.String `tfsdk:"project_id"`
	Location      types.String `tfsdk:"location"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	VolumeId      types.String `tfsdk:"volume_id"`
	Tags          types.List   `tfsdk:"tags"`
}

type SnapshotResource struct {
	client *ArubaCloudClient
}

func (r *SnapshotResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshot"
}

func (r *SnapshotResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Snapshot resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Snapshot identifier",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Snapshot URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Snapshot name",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this Snapshot belongs to",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Snapshot location",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period (only 'Hour' allowed)",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"volume_id": schema.StringAttribute{
				MarkdownDescription: "URI or ID of the volume this snapshot is for. Can be the volume URI (e.g., `/projects/{project_id}/providers/Aruba.Storage/volumes/{volume_id}`) or just the volume ID. If an ID is provided, the URI will be constructed automatically.",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the snapshot",
				Optional:            true,
			},
		},
	}
}

func (r *SnapshotResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SnapshotResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"Project ID is required to create a snapshot",
		)
		return
	}

	// Extract tags from Terraform list
	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Build volume URI
	volumeURI := data.VolumeId.ValueString()
	if !strings.HasPrefix(volumeURI, "/") {
		volumeURI = fmt.Sprintf("/projects/%s/providers/Aruba.Storage/volumes/%s", projectID, data.VolumeId.ValueString())
	}

	// Build the create request
	createRequest := sdktypes.SnapshotRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.SnapshotPropertiesRequest{
			Volume: sdktypes.ReferenceResource{
				URI: volumeURI,
			},
		},
	}

	// Create the snapshot using the SDK
	response, err := r.client.Client.FromStorage().Snapshots().Create(ctx, projectID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating snapshot",
			fmt.Sprintf("Unable to create snapshot: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create snapshot"
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
		} else {
			resp.Diagnostics.AddError(
				"Invalid API Response",
				"Snapshot created but ID not returned from API",
			)
			return
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Snapshot created but no data returned from API",
		)
		return
	}

	// Wait for Snapshot to be active before returning (Snapshot references Volume)
	// This ensures Terraform doesn't proceed to create dependent resources until Snapshot is ready
	snapshotID := data.Id.ValueString()
	if snapshotID == "" {
		resp.Diagnostics.AddError(
			"Missing Snapshot ID",
			"Snapshot ID is required but was not set",
		)
		return
	}
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromStorage().Snapshots().Get(ctx, projectID, snapshotID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for Snapshot to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "Snapshot", snapshotID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Snapshot Not Active",
			fmt.Sprintf("Snapshot was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	// Read the Snapshot again to ensure ID and other fields are properly set from metadata
	getResp, err := r.client.Client.FromStorage().Snapshots().Get(ctx, projectID, snapshotID, nil)
	if err == nil && getResp != nil && getResp.Data != nil {
		// Ensure ID is set from metadata (should already be set, but double-check)
		if getResp.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*getResp.Data.Metadata.ID)
		}
		if getResp.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*getResp.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		// Also update other fields that might have changed
		if getResp.Data.Metadata.Name != nil {
			data.Name = types.StringValue(*getResp.Data.Metadata.Name)
		}
		if getResp.Data.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(getResp.Data.Metadata.LocationResponse.Value)
		}
		// Update tags from response
		if len(getResp.Data.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(getResp.Data.Metadata.Tags))
			for i, tag := range getResp.Data.Metadata.Tags {
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
	} else if err != nil {
		// If Get fails, log but don't fail - we already have the ID from create response
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh Snapshot after creation: %v", err))
	}

	tflog.Trace(ctx, "created a Snapshot resource", map[string]interface{}{
		"snapshot_id": data.Id.ValueString(),
		"snapshot_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SnapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	snapshotID := data.Id.ValueString()

	// If ID is unknown or null, check if this is a new resource (no state) or existing resource (state exists but ID missing)
	// For new resources (during plan), we can return early
	// For existing resources, we need the ID to read - if it's missing, that's an error
	if data.Id.IsUnknown() || data.Id.IsNull() || snapshotID == "" {
		// Check if we have any other state data that indicates this is an existing resource
		// If name is set in state, this is likely an existing resource with missing ID (error case)
		if !data.Name.IsUnknown() && !data.Name.IsNull() && data.Name.ValueString() != "" {
			tflog.Error(ctx, "Snapshot exists in state but ID is missing - this indicates a state corruption issue")
			resp.Diagnostics.AddError(
				"Missing Snapshot ID",
				"Snapshot ID is required to read the snapshot. The resource exists in state but the ID is missing. This may indicate a state corruption issue. Try running 'terraform refresh' or 'terraform import arubacloud_snapshot.test <snapshot_id>'.",
			)
			return
		}
		// Otherwise, this is likely a new resource during plan - return early
		tflog.Info(ctx, "Snapshot ID is unknown or null during read, skipping API call (likely new resource).")
		return // Do not error, as this is expected during plan for new resources
	}

	if projectID == "" {
		// Check if ProjectID is unknown (new resource) vs missing (error)
		if data.ProjectId.IsUnknown() || data.ProjectId.IsNull() {
			tflog.Info(ctx, "Snapshot Project ID is unknown or null during read, skipping API call (likely new resource).")
			return // Do not error, as this is expected during plan for new resources
		}
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID is required to read the snapshot",
		)
		return
	}

	// Get snapshot details using the SDK
	response, err := r.client.Client.FromStorage().Snapshots().Get(ctx, projectID, snapshotID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading snapshot",
			fmt.Sprintf("Unable to read snapshot: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read snapshot"
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
		snapshot := response.Data

		// Preserve immutable fields from state (they're not returned by API)
		projectIdFromState := data.ProjectId
		billingPeriodFromState := data.BillingPeriod
		volumeIdFromState := data.VolumeId
		tagsFromState := data.Tags
		locationFromState := data.Location

		if snapshot.Metadata.ID != nil {
			data.Id = types.StringValue(*snapshot.Metadata.ID)
		}
		if snapshot.Metadata.URI != nil {
			data.Uri = types.StringValue(*snapshot.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if snapshot.Metadata.Name != nil {
			data.Name = types.StringValue(*snapshot.Metadata.Name)
		}
		// Location is immutable - always preserve from state
		// This prevents false changes when the referenced block storage is updated
		if !locationFromState.IsUnknown() && !locationFromState.IsNull() {
			data.Location = locationFromState
		} else {
			// Only use API value if state doesn't have it (new resources during plan)
			if snapshot.Metadata.LocationResponse != nil {
				data.Location = types.StringValue(snapshot.Metadata.LocationResponse.Value)
			}
		}
		
		// Handle volume_id: Always preserve from state since it's immutable
		// The volume_id never changes after snapshot creation, so we should always use the value from state
		// This prevents false changes when the referenced block storage is updated
		if !volumeIdFromState.IsUnknown() && !volumeIdFromState.IsNull() {
			// Always preserve volume_id from state (it's immutable)
			data.VolumeId = volumeIdFromState
		} else {
			// Only use API value if state doesn't have it (shouldn't happen for existing resources)
			if snapshot.Properties.Volume.URI != nil && *snapshot.Properties.Volume.URI != "" {
				data.VolumeId = types.StringValue(*snapshot.Properties.Volume.URI)
			}
		}

		// Update tags from response
		// If tags are null/unknown in state, preserve null (user didn't specify tags)
		// If tags exist in state or API has tags, update from API
		if len(snapshot.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(snapshot.Metadata.Tags))
			for i, tag := range snapshot.Metadata.Tags {
				tagValues[i] = types.StringValue(tag)
			}
			tagsList, diags := types.ListValueFrom(ctx, types.StringType, tagValues)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = tagsList
			}
		} else {
			// API has no tags - preserve from state if null/unknown, otherwise set to empty list
			if tagsFromState.IsNull() || tagsFromState.IsUnknown() {
				data.Tags = tagsFromState // Preserve null/unknown from state
			} else {
				// State has tags (even if empty), update to empty list to match API
				emptyList, diags := types.ListValue(types.StringType, []attr.Value{})
				resp.Diagnostics.Append(diags...)
				if !resp.Diagnostics.HasError() {
					data.Tags = emptyList
				}
			}
		}

		// Restore immutable fields from state (they're not returned by API)
		data.ProjectId = projectIdFromState
		data.BillingPeriod = billingPeriodFromState
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SnapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SnapshotResourceModel
	var state SnapshotResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get IDs from state (not plan) - IDs are immutable and should always be in state
	projectID := state.ProjectId.ValueString()
	snapshotID := state.Id.ValueString()

	if projectID == "" || snapshotID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Snapshot ID are required to update the snapshot",
		)
		return
	}

	// Get current snapshot details
	getResponse, err := r.client.Client.FromStorage().Snapshots().Get(ctx, projectID, snapshotID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current snapshot",
			fmt.Sprintf("Unable to get current snapshot: %s", err),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"Snapshot Not Found",
			"Snapshot not found or no data returned",
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
			"Unable to determine region value for snapshot",
		)
		return
	}

	// Extract tags from Terraform list
	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		// Preserve existing tags if not provided
		tags = current.Metadata.Tags
	}

	// Build the update request
	updateRequest := sdktypes.SnapshotRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.SnapshotPropertiesRequest{
			// Volume cannot be updated - preserve from create request or state
			// Note: VolumeInfo may need to be converted to ReferenceResource
			// For now, Volume is read-only in updates
		},
	}

	// Update the snapshot using the SDK
	response, err := r.client.Client.FromStorage().Snapshots().Update(ctx, projectID, snapshotID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating snapshot",
			fmt.Sprintf("Unable to update snapshot: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to update snapshot"
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
	data.ProjectId = state.ProjectId
	data.VolumeId = state.VolumeId // Preserve volume_id from state (it's immutable)

	if response != nil && response.Data != nil {
		// Update from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = state.Uri
		}
		// Update tags from response
		if len(response.Data.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(response.Data.Metadata.Tags))
			for i, tag := range response.Data.Metadata.Tags {
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
		// If no response, preserve URI and tags from state
		data.Uri = state.Uri
		data.Tags = state.Tags
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SnapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	snapshotID := data.Id.ValueString()

	if projectID == "" || snapshotID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Snapshot ID are required to delete the snapshot",
		)
		return
	}

	// Delete the snapshot using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromStorage().Snapshots().Delete(ctx, projectID, snapshotID, nil)
		},
		ExtractSDKError,
		"Snapshot",
		snapshotID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting snapshot",
			fmt.Sprintf("Unable to delete snapshot: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a Snapshot resource", map[string]interface{}{
		"snapshot_id": snapshotID,
	})
}

func (r *SnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
