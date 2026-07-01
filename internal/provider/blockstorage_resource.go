package provider

import (
	"context"
	"fmt"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	Timeout       types.String `tfsdk:"timeout"`
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
				Validators:          []validator.String{stringvalidator.OneOf("Hour", "Month", "Year")},
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Availability zone within the region. If omitted the volume is regional (accessible across all zones).",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Storage type. Accepted values: `Standard`, `Performance`. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
			"timeout": schema.StringAttribute{
				MarkdownDescription: "Per-resource timeout override (e.g. `\"15m\"`, `\"1h\"`). Overrides the provider-level `resource_timeout` for this resource's Create and Delete operations. Uses Go duration syntax.",
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
			fmt.Sprintf("Expected *ArubaCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func blockStorageRef(data *BlockStorageResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.URI("/projects/" + data.ProjectID.ValueString() + "/providers/Aruba.Storage/blockStorages/" + data.Id.ValueString())
}

func applyBlockStorageToModel(vol *aruba.BlockStorage, data *BlockStorageResourceModel) {
	data.Id = types.StringValue(vol.ID())
	data.Uri = strVal(vol.URI())
	data.Name = types.StringValue(vol.Name())
	data.Tags = TagsToListPreserveNull(vol.Tags(), data.Tags)
	if vol.Region() != "" {
		data.Location = types.StringValue(string(vol.Region()))
	}
	data.SizeGB = types.Int64Value(int64(vol.SizeGB()))
	data.Type = types.StringValue(string(vol.Type()))
	data.BillingPeriod = strVal(string(vol.BillingPeriod()))

	if z := string(vol.Zone()); z != "" {
		data.Zone = types.StringValue(z)
	} else {
		data.Zone = types.StringNull()
	}
	if vol.IsBootable() {
		data.Bootable = types.BoolValue(true)
	} else {
		data.Bootable = types.BoolNull()
	}
	if img := vol.Image(); img != "" {
		data.Image = types.StringValue(img)
	} else {
		data.Image = types.StringNull()
	}
}

func (r *BlockStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BlockStorageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()

	if !data.Bootable.IsNull() && data.Bootable.ValueBool() {
		if data.Image.IsNull() || data.Image.ValueString() == "" {
			resp.Diagnostics.AddError("Missing Image", "Image is required when bootable is set to true")
			return
		}
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	builder := aruba.NewBlockStorage().
		Named(data.Name.ValueString()).
		InProject(aruba.URI("/projects/" + projectID)).
		InRegion(aruba.Region(data.Location.ValueString())).
		SizedGB(int(data.SizeGB.ValueInt64())).
		OfType(aruba.BlockStorageType(data.Type.ValueString())).
		BilledBy(aruba.BillingPeriod(data.BillingPeriod.ValueString())).
		Tagged(tags...)

	if !data.Zone.IsNull() && data.Zone.ValueString() != "" {
		builder = builder.InZone(aruba.Zone(data.Zone.ValueString()))
	}
	if !data.Bootable.IsNull() && data.Bootable.ValueBool() {
		builder = builder.AsBootable()
	}
	if !data.Image.IsNull() && data.Image.ValueString() != "" {
		builder = builder.FromImage(data.Image.ValueString())
	}

	vol, err := r.client.Client.FromStorage().Volumes().Create(ctx, builder)
	if provErr := CheckResponseErr("create", "BlockStorage", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(vol.ID())
	data.Uri = strVal(vol.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if waitErr := vol.WaitUntilReady(ctx, sdkWaitOptions(effectiveTimeout(data.Timeout, r.client.ResourceTimeout))...); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "BlockStorage", data.Id.ValueString())
		return
	}

	fresh, freshErr := r.client.Client.FromStorage().Volumes().Get(ctx, blockStorageRef(&data))
	if freshErr == nil {
		projectID := data.ProjectID
		applyBlockStorageToModel(fresh, &data)
		data.ProjectID = projectID
	} else {
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh BlockStorage after creation: %v", freshErr))
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
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	vol, err := r.client.Client.FromStorage().Volumes().Get(ctx, blockStorageRef(&data))
	if provErr := CheckResponseErr("read", "BlockStorage", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	st := string(vol.State())
	switch {
	case isFailedState(st):
		resp.Diagnostics.AddWarning("Resource in Failed State",
			fmt.Sprintf("BlockStorage %q is in a terminal failure state (%s). "+
				"Run `terraform destroy` to clean it up, or `terraform apply -replace=<address>` to recreate it.", data.Id.ValueString(), st))
	case IsCreatingState(st):
		if waitErr := vol.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "BlockStorage", data.Id.ValueString())
			return
		}
		vol, err = r.client.Client.FromStorage().Volumes().Get(ctx, blockStorageRef(&data))
		if provErr := CheckResponseErr("read", "BlockStorage", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	projectID := data.ProjectID
	applyBlockStorageToModel(vol, &data)
	data.ProjectID = projectID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BlockStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BlockStorageResourceModel
	var state BlockStorageResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	vol, err := r.client.Client.FromStorage().Volumes().Get(ctx, blockStorageRef(&state))
	if provErr := CheckResponseErr("read", "BlockStorage", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	// Validate status allows update.
	st := string(vol.State())
	if st != "Used" && st != "NotUsed" {
		resp.Diagnostics.AddError("Cannot Update",
			fmt.Sprintf("Cannot update block storage with status %q. Only 'Used' or 'NotUsed' is permitted.", st))
		return
	}

	vol.Named(data.Name.ValueString())
	if tags != nil {
		vol.RetaggedAs(tags...)
	} else {
		vol.RetaggedAs(vol.Tags()...)
	}
	if !data.SizeGB.IsNull() {
		vol.SizedGB(int(data.SizeGB.ValueInt64()))
	}
	if !data.BillingPeriod.IsNull() {
		vol.BilledBy(aruba.BillingPeriod(data.BillingPeriod.ValueString()))
	}

	updated, err := r.client.Client.FromStorage().Volumes().Update(ctx, vol)
	if provErr := CheckResponseErr("update", "BlockStorage", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	// Wait for update to settle.
	if waitErr := updated.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "BlockStorage", state.Id.ValueString())
		return
	}

	// Re-read from the API to ensure all computed fields reflect the server's actual state.
	fresh, freshErr := r.client.Client.FromStorage().Volumes().Get(ctx, blockStorageRef(&state))
	if freshErr != nil {
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh BlockStorage after update: %v", freshErr))
		// Fall back to combining state immutables with updated response fields.
		data.Id = state.Id
		data.ProjectID = state.ProjectID
		data.Uri = state.Uri
		data.Zone = state.Zone
		data.Location = state.Location
		data.Type = state.Type
		data.Bootable = state.Bootable
		data.Image = state.Image
		data.Name = types.StringValue(updated.Name())
		data.Tags = TagsToListPreserveNull(updated.Tags(), data.Tags)
		data.SizeGB = types.Int64Value(int64(updated.SizeGB()))
		data.BillingPeriod = strVal(string(updated.BillingPeriod()))
	} else {
		projectID := state.ProjectID
		applyBlockStorageToModel(fresh, &data)
		data.ProjectID = projectID
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BlockStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BlockStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := blockStorageRef(&data)
	volumeID := data.Id.ValueString()
	projectID := data.ProjectID.ValueString()
	timeout := effectiveTimeout(data.Timeout, r.client.ResourceTimeout)

	// ArubaCloud API bug: a snapshot whose source volume is deleted first
	// becomes permanently undeletable. To avoid this, list all snapshots in the
	// project and delete any that reference this volume before touching the
	// volume itself.
	if err := r.deleteAssociatedSnapshots(ctx, projectID, volumeID, data.Uri.ValueString(), timeout); err != nil {
		resp.Diagnostics.AddError("Error deleting Snapshots before BlockStorage", err.Error())
		return
	}

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromStorage().Volumes().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "BlockStorage", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErrAsError("delete", "BlockStorage",
			r.client.Client.FromStorage().Volumes().Delete(ctx, ref))
	}, "BlockStorage", volumeID, timeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting BlockStorage", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "BlockStorage", volumeID, remainingTimeout(deleteStart, timeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for BlockStorage deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a Block Storage resource", map[string]interface{}{"volume_id": volumeID})
}

// deleteAssociatedSnapshots lists all snapshots for the project and removes any
// that were taken from this volume. This works around an ArubaCloud API bug where
// a snapshot becomes permanently undeletable once its source volume is destroyed.
func (r *BlockStorageResource) deleteAssociatedSnapshots(ctx context.Context, projectID, volumeID, volumeURI string, timeout time.Duration) error {
	projectRef := aruba.URI("/projects/" + projectID)
	snapList, listErr := r.client.Client.FromStorage().Snapshots().List(ctx, projectRef)
	if listErr != nil {
		// Non-fatal: log and continue so the volume delete can still proceed.
		tflog.Warn(ctx, fmt.Sprintf(
			"could not list snapshots before deleting volume %s — proceeding without pre-deletion: %v",
			volumeID, listErr))
		return nil
	}

	// Canonical volume URI path used as the fallback match key.
	canonicalVolumeURI := "/projects/" + projectID + "/providers/Aruba.Storage/volumes/" + volumeID

	var iterErr error
	_ = snapList.All(ctx, func(snap *aruba.Snapshot) bool {
		snapVolumeURI := snap.VolumeURI()
		if snapVolumeURI != volumeURI && snapVolumeURI != canonicalVolumeURI {
			return true // continue iteration
		}

		snapID := snap.ID()
		snapURI := snap.URI()
		var snapRef aruba.Ref
		if snapURI != "" {
			snapRef = aruba.URI(snapURI)
		} else {
			snapRef = aruba.URI("/projects/" + projectID + "/providers/Aruba.Storage/snapshots/" + snapID)
		}

		tflog.Info(ctx, "deleting snapshot before source volume to avoid API bug",
			map[string]interface{}{"snapshot_id": snapID, "volume_id": volumeID})

		snapDeletionChecker := func(ctx context.Context) (bool, error) {
			_, getErr := r.client.Client.FromStorage().Snapshots().Get(ctx, snapRef)
			if provErr := CheckResponseErr("get", "Snapshot", getErr); provErr != nil {
				if IsNotFound(provErr) {
					return true, nil
				}
				return false, provErr
			}
			return false, nil
		}

		deleteStart := time.Now()
		if delErr := DeleteResourceWithRetry(ctx, func() error {
			return CheckResponseErrAsError("delete", "Snapshot",
				r.client.Client.FromStorage().Snapshots().Delete(ctx, snapRef))
		}, "Snapshot", snapID, timeout, snapDeletionChecker); delErr != nil {
			iterErr = fmt.Errorf("failed to delete snapshot %s before volume %s: %w", snapID, volumeID, delErr)
			return false // stop iteration
		}
		if waitErr := WaitForResourceDeleted(ctx, snapDeletionChecker, "Snapshot", snapID, remainingTimeout(deleteStart, timeout)); waitErr != nil {
			iterErr = fmt.Errorf("timed out waiting for snapshot %s deletion before volume %s: %w", snapID, volumeID, waitErr)
			return false // stop iteration
		}

		tflog.Info(ctx, "deleted snapshot before source volume",
			map[string]interface{}{"snapshot_id": snapID, "volume_id": volumeID})
		return true // continue iteration
	})
	return iterErr
}

func (r *BlockStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseImportID(req.ID, "<project_id>/<volume_id>", "proj-abc/vol-xyz", 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
