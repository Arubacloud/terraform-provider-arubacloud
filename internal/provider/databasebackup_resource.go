package provider

import (
	"context"
	"fmt"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type DatabaseBackupResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	ProjectID     types.String `tfsdk:"project_id"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	Zone          types.String `tfsdk:"zone"`
	DBaaSID       types.String `tfsdk:"dbaas_id"`
	Database      types.String `tfsdk:"database"`
	BillingPeriod types.String `tfsdk:"billing_period"`
}

type DatabaseBackupResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &DatabaseBackupResource{}
var _ resource.ResourceWithImportState = &DatabaseBackupResource{}

func NewDatabaseBackupResource() resource.Resource {
	return &DatabaseBackupResource{}
}

func (r *DatabaseBackupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databasebackup"
}

func (r *DatabaseBackupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a backup of an ArubaCloud DBaaS database.",
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
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the database backup.",
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
			"zone": schema.StringAttribute{
				MarkdownDescription: "Availability zone within the region where the backup is stored.",
				Required:            true,
			},
			"dbaas_id": schema.StringAttribute{
				MarkdownDescription: "ID of the DBaaS cluster or database to back up.",
				Required:            true,
			},
			"database": schema.StringAttribute{
				MarkdownDescription: "Name of the logical database within the DBaaS cluster to back up.",
				Required:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle. Accepted values: `Hour`, `Month`, `Year`.",
				Required:            true,
			},
		},
	}
}

func (r *DatabaseBackupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func databaseBackupRef(data *DatabaseBackupResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.URI("/projects/" + data.ProjectID.ValueString() +
		"/providers/Aruba.Database/backups/" + data.Id.ValueString())
}

func (r *DatabaseBackupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DatabaseBackupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()
	databaseName := data.Database.ValueString()

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	dbaasURI := "/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID
	databaseURI := dbaasURI + "/databases/" + databaseName

	backup, err := r.client.Client.FromDatabase().Backups().Create(ctx,
		aruba.NewDBaaSBackup().
			Named(data.Name.ValueString()).
			InProject(aruba.URI("/projects/"+projectID)).
			InRegion(aruba.Region(data.Location.ValueString())).
			InZone(aruba.Zone(data.Zone.ValueString())).
			FromDBaaS(aruba.URI(dbaasURI)).
			FromDatabase(aruba.URI(databaseURI)).
			BilledBy(aruba.BillingPeriod(data.BillingPeriod.ValueString())).
			Tagged(tags...),
	)
	if provErr := CheckResponseErr("create", "DatabaseBackup", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(backup.ID())
	data.Uri = strVal(backup.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if waitErr := backup.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "DatabaseBackup", data.Id.ValueString())
		return
	}

	tflog.Trace(ctx, "created a Database Backup resource", map[string]interface{}{
		"backup_id":   data.Id.ValueString(),
		"backup_name": data.Name.ValueString(),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseBackupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatabaseBackupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	backup, err := r.client.Client.FromDatabase().Backups().Get(ctx, databaseBackupRef(&data))
	if provErr := CheckResponseErr("read", "DatabaseBackup", err); provErr != nil {
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
			fmt.Sprintf("DatabaseBackup %q is in a terminal failure state (%s). "+
				"Run `terraform destroy` to clean it up, or `terraform apply -replace=<address>` to recreate it.", data.Id.ValueString(), st))
	case IsCreatingState(st):
		if waitErr := backup.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "DatabaseBackup", data.Id.ValueString())
			return
		}
		backup, err = r.client.Client.FromDatabase().Backups().Get(ctx, databaseBackupRef(&data))
		if provErr := CheckResponseErr("read", "DatabaseBackup", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	data.Id = types.StringValue(backup.ID())
	data.Uri = strVal(backup.URI())
	data.Name = types.StringValue(backup.Name())
	data.Tags = TagsToListPreserveNull(backup.Tags(), data.Tags)
	if backup.Region() != "" {
		data.Location = types.StringValue(string(backup.Region()))
	}
	if z := string(backup.Zone()); z != "" {
		data.Zone = types.StringValue(z)
	}
	if bp := billingPeriodFromAPI(string(backup.BillingPeriod())); bp != "" {
		data.BillingPeriod = types.StringValue(bp)
	}
	// DBaaSID and Database are preserved from state (not in API response directly).

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseBackupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Database backups do not support updates.
	resp.Diagnostics.AddWarning(
		"Update Not Supported",
		"Database backups do not support updates. Changes will be ignored.",
	)
	var data DatabaseBackupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseBackupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatabaseBackupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := databaseBackupRef(&data)
	backupID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromDatabase().Backups().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "DatabaseBackup", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErrAsError("delete", "DatabaseBackup",
			r.client.Client.FromDatabase().Backups().Delete(ctx, ref))
	}, "DatabaseBackup", backupID, r.client.ResourceTimeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting DatabaseBackup", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "DatabaseBackup", backupID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for DatabaseBackup deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a Database Backup resource", map[string]interface{}{"backup_id": backupID})
}

func (r *DatabaseBackupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseImportID(req.ID, "<project_id>/<backup_id>", "proj-abc/dbbackup-xyz", 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
