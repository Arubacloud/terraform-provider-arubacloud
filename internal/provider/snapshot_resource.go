// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
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
	Name          types.String `tfsdk:"name"`
	ProjectId     types.String `tfsdk:"project_id"`
	Location      types.String `tfsdk:"location"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	VolumeId      types.String `tfsdk:"volume_id"`
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
				MarkdownDescription: "ID of the volume this snapshot is for",
				Required:            true,
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

	if projectID == "" || snapshotID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Snapshot ID are required to read the snapshot",
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

		if snapshot.Metadata.ID != nil {
			data.Id = types.StringValue(*snapshot.Metadata.ID)
		}
		if snapshot.Metadata.Name != nil {
			data.Name = types.StringValue(*snapshot.Metadata.Name)
		}
		if snapshot.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(snapshot.Metadata.LocationResponse.Value)
		}
		if snapshot.Properties.Volume.URI != nil && *snapshot.Properties.Volume.URI != "" {
			// Extract volume ID from URI
			parts := strings.Split(*snapshot.Properties.Volume.URI, "/")
			if len(parts) > 0 {
				data.VolumeId = types.StringValue(parts[len(parts)-1])
			}
		}
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

	projectID := data.ProjectId.ValueString()
	snapshotID := data.Id.ValueString()

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

	// Build the update request
	updateRequest := sdktypes.SnapshotRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
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
