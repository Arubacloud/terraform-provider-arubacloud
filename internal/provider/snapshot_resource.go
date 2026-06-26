package provider

import (
	"context"
	"fmt"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
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
			fmt.Sprintf("Expected *ArubaCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func snapshotRef(data *SnapshotResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.URI("/projects/" + data.ProjectId.ValueString() + "/providers/Aruba.Storage/snapshots/" + data.Id.ValueString())
}

func (r *SnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SnapshotResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	volumeURI := data.VolumeUri.ValueString()
	if volumeURI == "" {
		resp.Diagnostics.AddError("Missing Volume URI", "Volume URI is required to create a snapshot")
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	snap, err := r.client.Client.FromStorage().Snapshots().Create(ctx,
		aruba.NewSnapshot().
			Named(data.Name.ValueString()).
			InProject(aruba.URI("/projects/"+projectID)).
			InRegion(aruba.Region(data.Location.ValueString())).
			BilledBy(aruba.BillingPeriod(data.BillingPeriod.ValueString())).
			FromVolume(aruba.URI(volumeURI)).
			Tagged(tags...),
	)
	if provErr := CheckResponseErr("create", "Snapshot", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(snap.ID())
	data.Uri = strVal(snap.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if waitErr := snap.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "Snapshot", data.Id.ValueString())
		return
	}

	fresh, freshErr := r.client.Client.FromStorage().Snapshots().Get(ctx, snapshotRef(&data))
	if freshErr == nil {
		data.Id = types.StringValue(fresh.ID())
		data.Uri = strVal(fresh.URI())
		data.Name = types.StringValue(fresh.Name())
		data.Tags = TagsToListPreserveNull(fresh.Tags(), data.Tags)
		if fresh.Region() != "" {
			data.Location = types.StringValue(string(fresh.Region()))
		}
		// volume_uri and billing_period are immutable — preserved from plan.
	} else {
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh Snapshot after creation: %v", freshErr))
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
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	snap, err := r.client.Client.FromStorage().Snapshots().Get(ctx, snapshotRef(&data))
	if provErr := CheckResponseErr("read", "Snapshot", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	st := string(snap.State())
	switch {
	case isFailedState(st):
		resp.Diagnostics.AddWarning("Resource in Failed State",
			fmt.Sprintf("Snapshot %q is in a terminal failure state (%s). "+
				"Run `terraform destroy` to clean it up, or `terraform apply -replace=<address>` to recreate it.", data.Id.ValueString(), st))
	case IsCreatingState(st):
		if waitErr := snap.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "Snapshot", data.Id.ValueString())
			return
		}
		snap, err = r.client.Client.FromStorage().Snapshots().Get(ctx, snapshotRef(&data))
		if provErr := CheckResponseErr("read", "Snapshot", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	// Preserve immutable fields from state.
	projectId := data.ProjectId
	billingPeriod := data.BillingPeriod
	volumeUri := data.VolumeUri

	data.Id = types.StringValue(snap.ID())
	data.Uri = strVal(snap.URI())
	data.Name = types.StringValue(snap.Name())

	// Location is immutable — preserve state unless blank.
	if !data.Location.IsNull() && !data.Location.IsUnknown() {
		// keep from state
	} else if snap.Region() != "" {
		data.Location = types.StringValue(string(snap.Region()))
	}

	// Tags: prefer API values when present, otherwise preserve state.
	if len(snap.Tags()) > 0 {
		tagVals := make([]attr.Value, len(snap.Tags()))
		for i, t := range snap.Tags() {
			tagVals[i] = types.StringValue(t)
		}
		tagsList, diags := types.ListValue(types.StringType, tagVals)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Tags = tagsList
		}
	}

	data.ProjectId = projectId
	data.BillingPeriod = billingPeriod
	data.VolumeUri = volumeUri

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SnapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SnapshotResourceModel
	var state SnapshotResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	snap, err := r.client.Client.FromStorage().Snapshots().Get(ctx, snapshotRef(&state))
	if provErr := CheckResponseErr("read", "Snapshot", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	snap.Named(data.Name.ValueString())
	if tags != nil {
		snap.RetaggedAs(tags...)
	} else {
		snap.RetaggedAs(snap.Tags()...)
	}

	updated, err := r.client.Client.FromStorage().Snapshots().Update(ctx, snap)
	if provErr := CheckResponseErr("update", "Snapshot", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = state.Id
	data.ProjectId = state.ProjectId
	data.VolumeUri = state.VolumeUri
	data.BillingPeriod = state.BillingPeriod
	data.Location = state.Location
	data.Uri = state.Uri
	data.Name = types.StringValue(updated.Name())
	data.Tags = TagsToListPreserveNull(updated.Tags(), data.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SnapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := snapshotRef(&data)
	snapshotID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromStorage().Snapshots().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Snapshot", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErrAsError("delete", "Snapshot",
			r.client.Client.FromStorage().Snapshots().Delete(ctx, ref))
	}, "Snapshot", snapshotID, r.client.ResourceTimeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Snapshot", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "Snapshot", snapshotID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for Snapshot deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a Snapshot resource", map[string]interface{}{"snapshot_id": snapshotID})
}

func (r *SnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
