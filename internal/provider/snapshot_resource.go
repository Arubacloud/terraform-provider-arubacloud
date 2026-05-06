package provider

import (
	"context"
	"fmt"
	"time"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	VolumeUri     types.String `tfsdk:"volume_uri"`
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
		MarkdownDescription: "Manages an ArubaCloud Snapshot — a point-in-time copy of a block storage volume.",
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
				MarkdownDescription: "Display name for the snapshot.",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center). (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle. Accepted values: `Hour`, `Month`, `Year`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("Hour", "Month", "Year"),
				},
			},
			"volume_uri": schema.StringAttribute{
				MarkdownDescription: "URI of the block storage volume this snapshot is taken from. Reference the `uri` attribute of an `arubacloud_blockstorage` resource (e.g., `/projects/{project_id}/providers/Aruba.Storage/volumes/{volume_id}`). (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
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
			fmt.Sprintf("Expected *ArubaCloudClient, got: %T. Please report this issue to the provider developers. Please report this issue to the provider developers.", req.ProviderData),
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

	// Get volume URI from the plan
	volumeURI := data.VolumeUri.ValueString()
	if volumeURI == "" {
		resp.Diagnostics.AddError(
			"Missing Volume URI",
			"Volume URI is required to create a snapshot",
		)
		return
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
			NewTransportError("create", "Snapshot", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("create", "Snapshot", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
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
		ReportWaitResult(&resp.Diagnostics, err, "Snapshot", snapshotID)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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
		"snapshot_id":   data.Id.ValueString(),
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

	if data.Id.IsUnknown() || data.Id.IsNull() || snapshotID == "" {
		tflog.Debug(ctx, "Snapshot ID is empty, removing resource from state", map[string]interface{}{"snapshot_id": snapshotID})
		resp.State.RemoveResource(ctx)
		return
	}

	if projectID == "" {
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
			NewTransportError("read", "Snapshot", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("read", "Snapshot", response); apiErr != nil {
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
				fmt.Sprintf("Snapshot %q reached a terminal failure state (%s) and will not recover on its own. "+
					"Use `terraform apply -replace=<address>` to recreate it.", snapshotID, st),
			)
			return
		case IsCreatingState(st):
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
			if err := WaitForResourceActive(ctx, checker, "Snapshot", snapshotID, r.client.ResourceTimeout); err != nil {
				ReportWaitResult(&resp.Diagnostics, err, "Snapshot", snapshotID)
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
			// Re-read to get the final active state.
			response, err = r.client.Client.FromStorage().Snapshots().Get(ctx, projectID, snapshotID, nil)
			if err != nil {
				resp.Diagnostics.AddError("Error reading Snapshot after provisioning wait",
					NewTransportError("read", "Snapshot", err).Error())
				return
			}
			if apiErr := CheckResponse("read", "Snapshot", response); apiErr != nil {
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
		snapshot := response.Data

		// Preserve immutable fields from state (they're not returned by API)
		projectIdFromState := data.ProjectId
		billingPeriodFromState := data.BillingPeriod
		volumeUriFromState := data.VolumeUri
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

		// Handle volume_uri: Always preserve from state since it's immutable
		// The volume_uri never changes after snapshot creation, so we should always use the value from state
		// This prevents false changes when the referenced block storage is updated
		if !volumeUriFromState.IsUnknown() && !volumeUriFromState.IsNull() {
			// Always preserve volume_uri from state (it's immutable)
			data.VolumeUri = volumeUriFromState
		} else {
			// Only use API value if state doesn't have it (shouldn't happen for existing resources)
			if snapshot.Properties.Volume.URI != nil && *snapshot.Properties.Volume.URI != "" {
				data.VolumeUri = types.StringValue(*snapshot.Properties.Volume.URI)
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
			NewTransportError("read", "Snapshot", err).Error(),
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
			NewTransportError("update", "Snapshot", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("update", "Snapshot", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.ProjectId = state.ProjectId
	data.VolumeUri = state.VolumeUri // Preserve volume_uri from state (it's immutable)

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
	deletionChecker := func(ctx context.Context) (bool, error) {
		getResp, getErr := r.client.Client.FromStorage().Snapshots().Get(ctx, projectID, snapshotID, nil)
		if getErr != nil {
			return false, NewTransportError("get", "Snapshot", getErr)
		}
		if provErr := CheckResponse("get", "Snapshot", getResp); provErr != nil {
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
			resp, err := r.client.Client.FromStorage().Snapshots().Delete(ctx, projectID, snapshotID, nil)
			if err != nil {
				return NewTransportError("delete", "Snapshot", err)
			}
			return CheckResponse("delete", "Snapshot", resp)
		},
		"Snapshot",
		snapshotID,
		r.client.ResourceTimeout,
		deletionChecker,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting snapshot",
			NewTransportError("delete", "Snapshot", err).Error(),
		)
		return
	}

	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "Snapshot", snapshotID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for Snapshot deletion", waitErr.Error())
		return
	}

	tflog.Trace(ctx, "deleted a Snapshot resource", map[string]interface{}{
		"snapshot_id": snapshotID,
	})
}

func (r *SnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
