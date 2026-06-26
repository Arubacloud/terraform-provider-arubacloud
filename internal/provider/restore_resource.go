package provider

import (
	"context"
	"fmt"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
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

func restoreRef(data *RestoreResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.URI("/projects/" + data.ProjectID.ValueString() +
		"/providers/Aruba.Storage/backups/" + data.BackupID.ValueString() +
		"/restores/" + data.Id.ValueString())
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

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get volume URI from the volume ID.
	vol, err := r.client.Client.FromStorage().Volumes().Get(ctx,
		aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/blockStorages/"+volumeID))
	if provErr := CheckResponseErr("read", "Volume", err); provErr != nil {
		resp.Diagnostics.AddError("Error getting volume details", provErr.Error())
		return
	}
	if vol.URI() == "" {
		resp.Diagnostics.AddError("Invalid Volume Response", "Volume URI not found in response")
		return
	}

	backupRef := aruba.URI("/projects/" + projectID + "/providers/Aruba.Storage/backups/" + backupID)
	restore, err := r.client.Client.FromStorage().Restores().Create(ctx,
		aruba.NewStorageRestore().
			Named(data.Name.ValueString()).
			InRegion(aruba.Region(data.Location.ValueString())).
			FromBackup(backupRef).
			ToVolume(aruba.URI(vol.URI())).
			Tagged(tags...),
	)
	if provErr := CheckResponseErr("create", "Restore", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(restore.ID())
	data.Uri = strVal(restore.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if waitErr := restore.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "Restore", data.Id.ValueString())
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
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	restore, err := r.client.Client.FromStorage().Restores().Get(ctx, restoreRef(&data))
	if provErr := CheckResponseErr("read", "Restore", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	st := string(restore.State())
	switch {
	case isFailedState(st):
		resp.Diagnostics.AddWarning("Resource in Failed State",
			fmt.Sprintf("Restore %q is in a terminal failure state (%s). "+
				"Run `terraform destroy` to clean it up, or `terraform apply -replace=<address>` to recreate it.", data.Id.ValueString(), st))
	case IsCreatingState(st):
		if waitErr := restore.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "Restore", data.Id.ValueString())
			return
		}
		restore, err = r.client.Client.FromStorage().Restores().Get(ctx, restoreRef(&data))
		if provErr := CheckResponseErr("read", "Restore", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	data.Id = types.StringValue(restore.ID())
	data.Uri = strVal(restore.URI())
	data.Name = types.StringValue(restore.Name())
	data.Tags = TagsToListPreserveNull(restore.Tags(), data.Tags)
	if restore.Region() != "" {
		data.Location = types.StringValue(string(restore.Region()))
	}
	// VolumeID is preserved from state (not returned by API).

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RestoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RestoreResourceModel
	var state RestoreResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	restore, err := r.client.Client.FromStorage().Restores().Get(ctx, restoreRef(&state))
	if provErr := CheckResponseErr("read", "Restore", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	restore.Named(data.Name.ValueString())
	if tags != nil {
		restore.RetaggedAs(tags...)
	} else {
		restore.RetaggedAs(restore.Tags()...)
	}

	updated, err := r.client.Client.FromStorage().Restores().Update(ctx, restore)
	if provErr := CheckResponseErr("update", "Restore", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = state.Id
	data.ProjectID = state.ProjectID
	data.BackupID = state.BackupID
	data.VolumeID = state.VolumeID
	data.Uri = state.Uri
	data.Location = state.Location
	data.Name = types.StringValue(updated.Name())
	data.Tags = TagsToListPreserveNull(updated.Tags(), data.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RestoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RestoreResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := restoreRef(&data)
	restoreID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromStorage().Restores().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Restore", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErrAsError("delete", "Restore",
			r.client.Client.FromStorage().Restores().Delete(ctx, ref))
	}, "Restore", restoreID, r.client.ResourceTimeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Restore", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "Restore", restoreID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for Restore deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a Restore resource", map[string]interface{}{"restore_id": restoreID})
}

func (r *RestoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
