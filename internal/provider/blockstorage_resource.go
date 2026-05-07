package provider

import (
	"context"
	"fmt"
	"time"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &BlockStorageResource{}
var _ resource.ResourceWithImportState = &BlockStorageResource{}

func NewBlockStorageResource() resource.Resource {
	return &BlockStorageResource{}
}

type BlockStorageResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	ProjectID     types.String `tfsdk:"project_id"`
	Location      types.String `tfsdk:"location"`
	SizeGB        types.Int64  `tfsdk:"size_gb"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	Zone          types.String `tfsdk:"zone"`
	Type          types.String `tfsdk:"type"`
	Bootable      types.Bool   `tfsdk:"bootable"`
	Image         types.String `tfsdk:"image"`
	Tags          types.List   `tfsdk:"tags"`
}

type BlockStorageResource struct {
	client *ArubaCloudClient
}

func (r *BlockStorageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blockstorage"
}

func (r *BlockStorageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an ArubaCloud Block Storage volume.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource.",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the block storage volume.",
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
			"size_gb": schema.Int64Attribute{
				MarkdownDescription: "Size of the block storage volume in GiB. Must be a positive integer.",
				Required:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle. Accepted values: `Hour`, `Month`, `Year`.",
				Required:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Availability zone within the region. If omitted the volume is regional (accessible across all zones).",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Storage type. Accepted values: `Standard`, `Performance`.",
				Required:            true,
			},
			"bootable": schema.BoolAttribute{
				MarkdownDescription: "Whether this volume can be used as a boot volume for an `arubacloud_cloudserver`. Must be `true` when `image` is set.",
				Optional:            true,
			},
			"image": schema.StringAttribute{
				MarkdownDescription: "Image ID to use when creating a bootable volume. Required when `bootable` is `true`. See the [available images](https://api.arubacloud.com/docs/metadata/#cloud-server-bootvolume).",
				Optional:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Optional:            true,
			},
		},
	}
}

func (r *BlockStorageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BlockStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BlockStorageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"Project ID is required to create block storage",
		)
		return
	}

	// Validate bootable and image
	if !data.Bootable.IsNull() && !data.Bootable.IsUnknown() && data.Bootable.ValueBool() {
		if data.Image.IsNull() || data.Image.IsUnknown() || data.Image.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Missing Image",
				"Image is required when bootable is set to true",
			)
			return
		}
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the create request
	createRequest := sdktypes.BlockStorageRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.BlockStoragePropertiesRequest{
			SizeGB:        int(data.SizeGB.ValueInt64()),
			BillingPeriod: data.BillingPeriod.ValueString(),
			Type:          sdktypes.BlockStorageType(data.Type.ValueString()),
		},
	}

	// Add zone if provided
	if !data.Zone.IsNull() && !data.Zone.IsUnknown() {
		zone := data.Zone.ValueString()
		createRequest.Properties.Zone = &zone
	}

	// Add bootable and image if provided
	if !data.Bootable.IsNull() && !data.Bootable.IsUnknown() {
		bootable := data.Bootable.ValueBool()
		createRequest.Properties.Bootable = &bootable
	}
	if !data.Image.IsNull() && !data.Image.IsUnknown() {
		image := data.Image.ValueString()
		createRequest.Properties.Image = &image
	}

	// Create the block storage using the SDK
	response, err := r.client.Client.FromStorage().Volumes().Create(ctx, projectID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating block storage",
			NewTransportError("create", "Blockstorage", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("create", "Blockstorage", response); apiErr != nil {
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
		if response.Data.Properties.Zone != "" {
			data.Zone = types.StringValue(response.Data.Properties.Zone)
		} else {
			data.Zone = types.StringNull()
		}
		// Populate bootable and image from API response
		// Only set if API provides a value, otherwise preserve null from plan
		if response.Data.Properties.Bootable != nil {
			data.Bootable = types.BoolValue(*response.Data.Properties.Bootable)
		} else {
			// Keep as null if API doesn't provide a value (preserves plan state)
			data.Bootable = types.BoolNull()
		}
		if response.Data.Properties.Image != nil {
			data.Image = types.StringValue(*response.Data.Properties.Image)
		} else {
			data.Image = types.StringNull()
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Block storage created but no data returned from API",
		)
		return
	}

	// Wait for Block Storage to be active before returning (Volume is referenced by Snapshot, CloudServer)
	// This ensures Terraform doesn't proceed to create dependent resources until BlockStorage is ready
	volumeID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromStorage().Volumes().Get(ctx, projectID, volumeID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for Block Storage to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "BlockStorage", volumeID, r.client.ResourceTimeout); err != nil {
		ReportWaitResult(&resp.Diagnostics, err, "BlockStorage", volumeID)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Read the Block Storage again to ensure URI and other fields are properly set from metadata
	getResp, err := r.client.Client.FromStorage().Volumes().Get(ctx, projectID, volumeID, nil)
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
		if getResp.Data.Properties.Zone != "" {
			data.Zone = types.StringValue(getResp.Data.Properties.Zone)
		} else {
			data.Zone = types.StringNull()
		}
		// Update bootable and image from re-read
		if getResp.Data.Properties.Bootable != nil {
			data.Bootable = types.BoolValue(*getResp.Data.Properties.Bootable)
		} else {
			data.Bootable = types.BoolNull()
		}
		if getResp.Data.Properties.Image != nil {
			data.Image = types.StringValue(*getResp.Data.Properties.Image)
		} else {
			data.Image = types.StringNull()
		}
	} else if err != nil {
		// If Get fails, log but don't fail - we already have the ID from create response
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh Block Storage after creation: %v", err))
	}

	tflog.Trace(ctx, "created a Block Storage resource", map[string]interface{}{
		"blockstorage_id":   data.Id.ValueString(),
		"blockstorage_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BlockStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BlockStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	volumeID := data.Id.ValueString()

	if data.Id.IsUnknown() || data.Id.IsNull() || volumeID == "" {
		tflog.Debug(ctx, "Block Storage ID is empty, removing resource from state", map[string]interface{}{"volume_id": volumeID})
		resp.State.RemoveResource(ctx)
		return
	}

	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID is required to read the block storage",
		)
		return
	}

	// Get block storage details using the SDK
	response, err := r.client.Client.FromStorage().Volumes().Get(ctx, projectID, volumeID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading block storage",
			NewTransportError("read", "Blockstorage", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("read", "Blockstorage", response); apiErr != nil {
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
				fmt.Sprintf("BlockStorage %q reached a terminal failure state (%s) and will not recover on its own. "+
					"Use `terraform apply -replace=<address>` to recreate it.", volumeID, st),
			)
			return
		case IsCreatingState(st):
			checker := func(ctx context.Context) (string, error) {
				getResp, err := r.client.Client.FromStorage().Volumes().Get(ctx, projectID, volumeID, nil)
				if err != nil {
					return "", err
				}
				if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
					return *getResp.Data.Status.State, nil
				}
				return "Unknown", nil
			}
			if err := WaitForResourceActive(ctx, checker, "BlockStorage", volumeID, r.client.ResourceTimeout); err != nil {
				ReportWaitResult(&resp.Diagnostics, err, "BlockStorage", volumeID)
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
			// Re-read to get the final active state.
			response, err = r.client.Client.FromStorage().Volumes().Get(ctx, projectID, volumeID, nil)
			if err != nil {
				resp.Diagnostics.AddError("Error reading BlockStorage after provisioning wait",
					NewTransportError("read", "Blockstorage", err).Error())
				return
			}
			if apiErr := CheckResponse("read", "Blockstorage", response); apiErr != nil {
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
		volume := response.Data

		if volume.Metadata.ID != nil {
			data.Id = types.StringValue(*volume.Metadata.ID)
		}
		if volume.Metadata.URI != nil {
			data.Uri = types.StringValue(*volume.Metadata.URI)
		} else {
			// If API doesn't return URI, try to preserve from state, or construct it from ID
			if !data.Uri.IsNull() && !data.Uri.IsUnknown() {
				// Preserve URI from state if available
				// (data.Uri already has the state value, so no change needed)
			} else if volume.Metadata.ID != nil {
				// Construct URI from ID if we have it
				uri := fmt.Sprintf("/projects/%s/providers/Aruba.Storage/volumes/%s", projectID, *volume.Metadata.ID)
				data.Uri = types.StringValue(uri)
			} else {
				data.Uri = types.StringNull()
			}
		}
		if volume.Metadata.Name != nil {
			data.Name = types.StringValue(*volume.Metadata.Name)
		}
		if volume.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(volume.Metadata.LocationResponse.Value)
		}
		data.SizeGB = types.Int64Value(int64(volume.Properties.SizeGB))
		data.Type = types.StringValue(string(volume.Properties.Type))
		// Zone: if empty, it's regional storage; if set, it's zonal storage
		if volume.Properties.Zone != "" {
			data.Zone = types.StringValue(volume.Properties.Zone)
		} else {
			// Regional storage - zone is null/empty
			data.Zone = types.StringNull()
		}

		// Populate bootable and image from API response
		// Only set if API provides a value, otherwise preserve null from plan
		if volume.Properties.Bootable != nil {
			data.Bootable = types.BoolValue(*volume.Properties.Bootable)
		} else {
			// Keep as null if API doesn't provide a value (preserves plan state)
			data.Bootable = types.BoolNull()
		}
		if volume.Properties.Image != nil {
			data.Image = types.StringValue(*volume.Properties.Image)
		} else {
			data.Image = types.StringNull()
		}

		data.Tags = TagsToListPreserveNull(volume.Metadata.Tags, data.Tags)
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BlockStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BlockStorageResourceModel
	var state BlockStorageResourceModel

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
	volumeID := state.Id.ValueString()

	if projectID == "" || volumeID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Volume ID are required to update the block storage",
		)
		return
	}

	// Get current block storage details
	getResponse, err := r.client.Client.FromStorage().Volumes().Get(ctx, projectID, volumeID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current block storage",
			NewTransportError("read", "Blockstorage", err).Error(),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"Block Storage Not Found",
			"Block storage not found or no data returned",
		)
		return
	}

	current := getResponse.Data

	// Check if the volume status allows updates
	if current.Status.State != nil {
		status := *current.Status.State
		if status != "Used" && status != "NotUsed" {
			resp.Diagnostics.AddError(
				"Cannot Update",
				fmt.Sprintf("Cannot update block storage with status '%s'. Block storage can only be updated when status is 'Used' or 'NotUsed'", status),
			)
			return
		}
	}

	// Get region value
	regionValue := ""
	if current.Metadata.LocationResponse != nil {
		regionValue = current.Metadata.LocationResponse.Value
	}
	if regionValue == "" {
		resp.Diagnostics.AddError(
			"Missing Region",
			"Unable to determine region value for block storage",
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

	// Handle zone
	var zone *string
	if current.Properties.Zone != "" {
		zone = &current.Properties.Zone
	}

	// Build the update request
	// Use size from plan (data) if provided, otherwise preserve current size
	sizeGB := current.Properties.SizeGB
	if !data.SizeGB.IsNull() && !data.SizeGB.IsUnknown() {
		sizeGB = int(data.SizeGB.ValueInt64())
	}

	// Use billing period from plan (data) if provided, otherwise preserve current
	billingPeriod := current.Properties.BillingPeriod
	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		billingPeriod = data.BillingPeriod.ValueString()
	}

	updateRequest := sdktypes.BlockStorageRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.BlockStoragePropertiesRequest{
			SizeGB:        sizeGB,
			BillingPeriod: billingPeriod,
			Zone:          zone,
			Type:          current.Properties.Type,
		},
	}

	// Update the block storage using the SDK
	response, err := r.client.Client.FromStorage().Volumes().Update(ctx, projectID, volumeID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating block storage",
			NewTransportError("update", "Blockstorage", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("update", "Blockstorage", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}

	// Wait for Block Storage update to complete before returning
	// This ensures Terraform doesn't proceed until the update is fully applied
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromStorage().Volumes().Get(ctx, projectID, volumeID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for Block Storage to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "BlockStorage", volumeID, r.client.ResourceTimeout); err != nil {
		ReportWaitResult(&resp.Diagnostics, err, "BlockStorage", volumeID)
		return
	}

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.ProjectID = state.ProjectID
	data.Uri = state.Uri   // Preserve URI from state
	data.Zone = state.Zone // Preserve zone from state (immutable)

	if response != nil && response.Data != nil {
		// Update from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = state.Uri // Fallback to state if not in response
		}
		// Update zone from response (empty = regional, set = zonal)
		if response.Data.Properties.Zone != "" {
			data.Zone = types.StringValue(response.Data.Properties.Zone)
		} else {
			data.Zone = types.StringNull() // Regional storage
		}
		// Update size from response
		data.SizeGB = types.Int64Value(int64(response.Data.Properties.SizeGB))
		// Update billing period from response if available
		if response.Data.Properties.BillingPeriod != "" {
			data.BillingPeriod = types.StringValue(response.Data.Properties.BillingPeriod)
		}
	} else {
		// If no response, re-read the resource to get the latest state after update
		getResp, err := r.client.Client.FromStorage().Volumes().Get(ctx, projectID, volumeID, nil)
		if err == nil && getResp != nil && getResp.Data != nil {
			// Update from the re-read response
			if getResp.Data.Metadata.URI != nil {
				data.Uri = types.StringValue(*getResp.Data.Metadata.URI)
			} else {
				data.Uri = state.Uri
			}
			if getResp.Data.Properties.Zone != "" {
				data.Zone = types.StringValue(getResp.Data.Properties.Zone)
			} else {
				data.Zone = types.StringNull()
			}
			data.SizeGB = types.Int64Value(int64(getResp.Data.Properties.SizeGB))
			if getResp.Data.Properties.BillingPeriod != "" {
				data.BillingPeriod = types.StringValue(getResp.Data.Properties.BillingPeriod)
			} else {
				data.BillingPeriod = state.BillingPeriod
			}
		} else {
			// If re-read fails, preserve from state
			data.Uri = state.Uri
			data.Zone = state.Zone
			data.SizeGB = state.SizeGB
			data.BillingPeriod = state.BillingPeriod
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BlockStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BlockStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	volumeID := data.Id.ValueString()

	if projectID == "" || volumeID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Volume ID are required to delete the block storage",
		)
		return
	}

	// Delete the block storage using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	deletionChecker := func(ctx context.Context) (bool, error) {
		getResp, getErr := r.client.Client.FromStorage().Volumes().Get(ctx, projectID, volumeID, nil)
		if getErr != nil {
			return false, NewTransportError("get", "BlockStorage", getErr)
		}
		if provErr := CheckResponse("get", "BlockStorage", getResp); provErr != nil {
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
			resp, err := r.client.Client.FromStorage().Volumes().Delete(ctx, projectID, volumeID, nil)
			if err != nil {
				return NewTransportError("delete", "BlockStorage", err)
			}
			return CheckResponse("delete", "BlockStorage", resp)
		},
		"BlockStorage",
		volumeID,
		r.client.ResourceTimeout,
		deletionChecker,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting block storage",
			NewTransportError("delete", "Blockstorage", err).Error(),
		)
		return
	}

	// Poll until the block storage volume is confirmed deleted (async deletion)
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "BlockStorage", volumeID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError(
			"Error waiting for BlockStorage deletion",
			waitErr.Error(),
		)
		return
	}

	tflog.Trace(ctx, "deleted a Block Storage resource", map[string]interface{}{
		"volume_id": volumeID,
	})
}

func (r *BlockStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
