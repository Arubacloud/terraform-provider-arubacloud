package provider

import (
	"context"
	"fmt"
	"strings"
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
	Timeout       types.String `tfsdk:"timeout"`
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
		MarkdownDescription: "Manages an ArubaCloud Block Storage Backup.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the backup.",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center). (Immutable — changing this value forces the resource to be destroyed and re-created.)",
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
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Backup type. Accepted values: `Full`, `Incremental`. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("Full", "Incremental"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"volume_id": schema.StringAttribute{
				MarkdownDescription: "ID of the block storage volume to back up. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"retention_days": schema.Int64Attribute{
				MarkdownDescription: "Number of days to retain the backup before automatic deletion. Optional — if omitted, the backup is retained indefinitely.",
				Optional:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle. Accepted values: `Hour`, `Month`, `Year`.",
				Optional:            true,
				Validators:          []validator.String{stringvalidator.OneOf("Hour", "Month", "Year")},
			},
			"timeout": schema.StringAttribute{
				MarkdownDescription: "Per-resource timeout override (e.g. `\"15m\"`, `\"1h\"`). Overrides the provider-level `resource_timeout` for this resource's Create and Delete operations. Uses Go duration syntax.",
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

func backupRef(data *BackupResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.URI("/projects/" + data.ProjectID.ValueString() + "/providers/Aruba.Storage/backups/" + data.Id.ValueString())
}

func (r *BackupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BackupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	volumeID := data.VolumeID.ValueString()

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the volume to obtain its URI.
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

	builder := aruba.NewStorageBackup().
		Named(data.Name.ValueString()).
		InProject(aruba.URI("/projects/" + projectID)).
		InRegion(aruba.Region(data.Location.ValueString())).
		OfType(aruba.StorageBackupType(data.Type.ValueString())).
		FromVolume(aruba.URI(vol.URI())).
		Tagged(tags...)

	if !data.RetentionDays.IsNull() && !data.RetentionDays.IsUnknown() {
		builder = builder.RetainedForDays(int(data.RetentionDays.ValueInt64()))
	}
	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		builder = builder.BilledBy(aruba.BillingPeriod(data.BillingPeriod.ValueString()))
	}

	backup, err := r.client.Client.FromStorage().Backups().Create(ctx, builder)
	if provErr := CheckResponseErr("create", "Backup", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(backup.ID())
	data.Uri = strVal(backup.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if waitErr := backup.WaitUntilReady(ctx, sdkWaitOptions(effectiveTimeout(data.Timeout, r.client.ResourceTimeout))...); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "Backup", data.Id.ValueString())
		return
	}

	// Re-read to capture server-assigned fields (normalised URI, billing period, etc.)
	// that may differ between the Create response and the fully-provisioned state.
	fresh, freshErr := r.client.Client.FromStorage().Backups().Get(ctx, backupRef(&data))
	if freshErr == nil {
		applyBackupToState(fresh, &data)
	} else {
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh Backup after creation: %v", freshErr))
	}

	tflog.Trace(ctx, "created a Backup resource", map[string]interface{}{
		"backup_id":   data.Id.ValueString(),
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
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	backup, err := r.client.Client.FromStorage().Backups().Get(ctx, backupRef(&data))
	if provErr := CheckResponseErr("read", "Backup", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	st := string(backup.State())
	switch {
	case isFailedState(st):
		resp.Diagnostics.AddWarning("Resource in Failed State",
			fmt.Sprintf("Backup %q is in a terminal failure state (%s). "+
				"Run `terraform destroy` to clean it up, or `terraform apply -replace=<address>` to recreate it.", data.Id.ValueString(), st))
	case IsCreatingState(st):
		if waitErr := backup.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "Backup", data.Id.ValueString())
			return
		}
		backup, err = r.client.Client.FromStorage().Backups().Get(ctx, backupRef(&data))
		if provErr := CheckResponseErr("read", "Backup", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	applyBackupToState(backup, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// applyBackupToState maps fields from the SDK StorageBackup wrapper into data.
// Called from both Create (post-wait re-read) and Read.
func applyBackupToState(backup *aruba.StorageBackup, data *BackupResourceModel) {
	data.Id = types.StringValue(backup.ID())
	data.Uri = strVal(backup.URI())
	data.Name = types.StringValue(backup.Name())
	data.Tags = TagsToListPreserveNull(backup.Tags(), data.Tags)
	if backup.Region() != "" {
		data.Location = types.StringValue(string(backup.Region()))
	}
	if t := string(backup.Type()); t != "" {
		data.Type = types.StringValue(t)
	}
	// Extract volume ID from origin URI.
	if originURI := backup.OriginURI(); originURI != "" {
		parts := strings.Split(originURI, "/")
		if last := parts[len(parts)-1]; last != "" {
			data.VolumeID = types.StringValue(last)
		}
	}
	if days := backup.RetentionDays(); days > 0 {
		data.RetentionDays = types.Int64Value(int64(days))
	}
	if bp := billingPeriodFromAPI(string(backup.BillingPeriod())); bp != "" {
		data.BillingPeriod = types.StringValue(bp)
	}
}

func (r *BackupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BackupResourceModel
	var state BackupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	backup, err := r.client.Client.FromStorage().Backups().Get(ctx, backupRef(&state))
	if provErr := CheckResponseErr("read", "Backup", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	backup.Named(data.Name.ValueString())
	if tags != nil {
		backup.RetaggedAs(tags...)
	} else {
		backup.RetaggedAs(backup.Tags()...)
	}

	updated, err := r.client.Client.FromStorage().Backups().Update(ctx, backup)
	if provErr := CheckResponseErr("update", "Backup", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = state.Id
	data.ProjectID = state.ProjectID
	data.Uri = state.Uri
	data.VolumeID = state.VolumeID
	data.Type = state.Type
	data.RetentionDays = state.RetentionDays
	data.BillingPeriod = state.BillingPeriod
	data.Location = state.Location
	data.Name = types.StringValue(updated.Name())
	data.Tags = TagsToListPreserveNull(updated.Tags(), data.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BackupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BackupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := backupRef(&data)
	backupID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromStorage().Backups().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Backup", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErrAsError("delete", "Backup",
			r.client.Client.FromStorage().Backups().Delete(ctx, ref))
	}, "Backup", backupID, effectiveTimeout(data.Timeout, r.client.ResourceTimeout), deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Backup", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "Backup", backupID, remainingTimeout(deleteStart, effectiveTimeout(data.Timeout, r.client.ResourceTimeout))); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for Backup deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a Backup resource", map[string]interface{}{"backup_id": backupID})
}

func (r *BackupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseImportID(req.ID, "<project_id>/<backup_id>", "proj-abc/bkp-xyz", 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
