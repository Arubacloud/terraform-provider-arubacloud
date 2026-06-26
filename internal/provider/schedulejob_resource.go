package provider

import (
	"context"
	"fmt"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ScheduleJobResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Uri        types.String `tfsdk:"uri"`
	Name       types.String `tfsdk:"name"`
	ProjectID  types.String `tfsdk:"project_id"`
	Tags       types.List   `tfsdk:"tags"`
	Location   types.String `tfsdk:"location"`
	Properties types.Object `tfsdk:"properties"`
}

type ScheduleJobResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &ScheduleJobResource{}
var _ resource.ResourceWithImportState = &ScheduleJobResource{}

func NewScheduleJobResource() resource.Resource {
	return &ScheduleJobResource{}
}

func (r *ScheduleJobResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedulejob"
}

func (r *ScheduleJobResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an ArubaCloud Scheduled Job — a cron-triggered automation task.",
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
				MarkdownDescription: "Display name for the scheduled job.",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Optional:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
				Required:            true,
			},
			"properties": schema.SingleNestedAttribute{
				MarkdownDescription: "Job scheduling and execution configuration.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "When `true`, the job is active and will trigger on schedule.",
						Optional:            true,
					},
					"schedule_job_type": schema.StringAttribute{
						MarkdownDescription: "Execution mode of the job. Accepted values: `OneShot` (runs once at `schedule_at`), `Recurring` (repeats on a `cron` schedule).",
						Required:            true,
					},
					"schedule_at": schema.StringAttribute{
						MarkdownDescription: "ISO 8601 date-time at which the job executes once (required for `OneShot` type).",
						Optional:            true,
					},
					"execute_until": schema.StringAttribute{
						MarkdownDescription: "ISO 8601 date-time after which a `Recurring` job stops executing.",
						Optional:            true,
					},
					"cron": schema.StringAttribute{
						MarkdownDescription: "Cron expression defining the job schedule (e.g., `0 * * * *` for hourly). Standard 5-field cron format.",
						Optional:            true,
					},
					"steps": schema.ListNestedAttribute{
						MarkdownDescription: "Ordered list of API actions executed when the job triggers.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Optional human-readable label for the step.",
									Optional:            true,
								},
								"resource_uri": schema.StringAttribute{
									MarkdownDescription: "URI of the ArubaCloud resource that the step targets.",
									Required:            true,
								},
								"action_uri": schema.StringAttribute{
									MarkdownDescription: "URI of the API action to invoke on the target resource.",
									Required:            true,
								},
								"http_verb": schema.StringAttribute{
									MarkdownDescription: "HTTP method used to call the action URI (e.g., `GET`, `POST`, `PUT`, `DELETE`).",
									Required:            true,
								},
								"body": schema.StringAttribute{
									MarkdownDescription: "Optional JSON request body sent with the HTTP call.",
									Optional:            true,
								},
							},
						},
						Optional: true,
					},
				},
				Required: true,
			},
		},
	}
}

func (r *ScheduleJobResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func jobRef(data *ScheduleJobResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.URI("/projects/" + data.ProjectID.ValueString() +
		"/providers/Aruba.Schedule/jobs/" + data.Id.ValueString())
}

// stepObjectAttrTypes returns the attr.Type map for a step object.
func stepObjectAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":         types.StringType,
		"resource_uri": types.StringType,
		"action_uri":   types.StringType,
		"http_verb":    types.StringType,
		"body":         types.StringType,
	}
}

// propertiesAttrTypes returns the attr.Type map for the properties object.
func propertiesAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":           types.BoolType,
		"schedule_job_type": types.StringType,
		"schedule_at":       types.StringType,
		"execute_until":     types.StringType,
		"cron":              types.StringType,
		"steps":             types.ListType{ElemType: types.ObjectType{AttrTypes: stepObjectAttrTypes()}},
	}
}

// buildJobSteps builds *aruba.JobStep builders from the Terraform properties object.
func buildJobSteps(propertiesObj map[string]attr.Value, ctx context.Context, diagnostics *diag.Diagnostics) []*aruba.JobStep {
	stepsAttr, ok := propertiesObj["steps"]
	if !ok {
		return nil
	}
	stepsList, ok := stepsAttr.(types.List)
	if !ok || stepsList.IsNull() || stepsList.IsUnknown() {
		return nil
	}

	var stepsElements []types.Object
	d := stepsList.ElementsAs(ctx, &stepsElements, false)
	diagnostics.Append(d...)
	if diagnostics.HasError() {
		return nil
	}

	steps := make([]*aruba.JobStep, 0, len(stepsElements))
	for _, stepObj := range stepsElements {
		stepAttrs := stepObj.Attributes()
		builder := aruba.NewJobStep()

		if nameAttr, ok := stepAttrs["name"]; ok {
			if nameStr, ok := nameAttr.(types.String); ok && !nameStr.IsNull() {
				builder = builder.Named(nameStr.ValueString())
			}
		}
		if resURIAttr, ok := stepAttrs["resource_uri"]; ok {
			if resURIStr, ok := resURIAttr.(types.String); ok && !resURIStr.IsNull() {
				builder = builder.Targeting(aruba.URI(resURIStr.ValueString()))
			}
		}
		if actURIAttr, ok := stepAttrs["action_uri"]; ok {
			if actURIStr, ok := actURIAttr.(types.String); ok && !actURIStr.IsNull() {
				builder = builder.WithAction(actURIStr.ValueString())
			}
		}
		if verbAttr, ok := stepAttrs["http_verb"]; ok {
			if verbStr, ok := verbAttr.(types.String); ok && !verbStr.IsNull() {
				builder = builder.WithVerb(aruba.HTTPVerb(verbStr.ValueString()))
			}
		}
		if bodyAttr, ok := stepAttrs["body"]; ok {
			if bodyStr, ok := bodyAttr.(types.String); ok && !bodyStr.IsNull() {
				builder = builder.WithBody(bodyStr.ValueString())
			}
		}
		steps = append(steps, builder)
	}
	return steps
}

// applyJobToModel populates data from the job wrapper and raw response.
func applyJobToModel(job *aruba.Job, data *ScheduleJobResourceModel, diagnostics *diag.Diagnostics) {
	data.Id = types.StringValue(job.ID())
	if uri := job.URI(); uri != "" {
		data.Uri = types.StringValue(uri)
	} else {
		data.Uri = types.StringNull()
	}
	data.Name = types.StringValue(job.Name())
	if r := string(job.Region()); r != "" {
		data.Location = types.StringValue(r)
	}
	data.Tags = TagsToListPreserveNull(job.Tags(), data.Tags)

	raw := job.Raw()
	if raw == nil {
		return
	}

	stepObjType := types.ObjectType{AttrTypes: stepObjectAttrTypes()}
	var stepsListValue types.List
	if len(raw.Properties.Steps) > 0 {
		stepObjects := make([]attr.Value, 0, len(raw.Properties.Steps))
		for _, step := range raw.Properties.Steps {
			stepAttrs := map[string]attr.Value{
				"name":         types.StringNull(),
				"resource_uri": types.StringNull(),
				"action_uri":   types.StringNull(),
				"http_verb":    types.StringNull(),
				"body":         types.StringNull(),
			}
			if step.Name != nil {
				stepAttrs["name"] = types.StringValue(*step.Name)
			}
			if step.ResourceURI != nil {
				stepAttrs["resource_uri"] = types.StringValue(*step.ResourceURI)
			}
			if step.ActionURI != nil {
				stepAttrs["action_uri"] = types.StringValue(*step.ActionURI)
			}
			if step.HttpVerb != nil {
				stepAttrs["http_verb"] = types.StringValue(*step.HttpVerb)
			}
			if step.Body != nil {
				stepAttrs["body"] = types.StringValue(*step.Body)
			}
			obj, d := types.ObjectValue(stepObjectAttrTypes(), stepAttrs)
			diagnostics.Append(d...)
			if diagnostics.HasError() {
				return
			}
			stepObjects = append(stepObjects, obj)
		}
		var d diag.Diagnostics
		stepsListValue, d = types.ListValue(stepObjType, stepObjects)
		diagnostics.Append(d...)
		if diagnostics.HasError() {
			return
		}
	} else {
		stepsListValue = types.ListNull(stepObjType)
	}

	propertiesAttrs := map[string]attr.Value{
		"enabled":           types.BoolValue(raw.Properties.Enabled),
		"schedule_job_type": types.StringValue(string(raw.Properties.JobType)),
		"schedule_at":       types.StringNull(),
		"execute_until":     types.StringNull(),
		"cron":              types.StringNull(),
		"steps":             stepsListValue,
	}
	if raw.Properties.ScheduleAt != nil {
		propertiesAttrs["schedule_at"] = types.StringValue(*raw.Properties.ScheduleAt)
	}
	if raw.Properties.ExecuteUntil != nil {
		propertiesAttrs["execute_until"] = types.StringValue(*raw.Properties.ExecuteUntil)
	}
	if raw.Properties.Cron != nil {
		propertiesAttrs["cron"] = types.StringValue(*raw.Properties.Cron)
	}

	propertiesObj, d := types.ObjectValue(propertiesAttrTypes(), propertiesAttrs)
	diagnostics.Append(d...)
	if !diagnostics.HasError() {
		data.Properties = propertiesObj
	}
}

func (r *ScheduleJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ScheduleJobResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID is required to create a schedule job")
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	propertiesObj := data.Properties.Attributes()

	jobTypeAttr, ok := propertiesObj["schedule_job_type"].(types.String)
	if !ok {
		resp.Diagnostics.AddError("Invalid Type", "schedule_job_type must be a String")
		return
	}
	jobTypeStr := jobTypeAttr.ValueString()

	enabled := true
	if enabledAttr, ok := propertiesObj["enabled"]; ok {
		if enabledBool, ok := enabledAttr.(types.Bool); ok && !enabledBool.IsNull() {
			enabled = enabledBool.ValueBool()
		}
	}

	var scheduleAt *string
	if scheduleAtAttr, ok := propertiesObj["schedule_at"]; ok {
		if scheduleAtStr, ok := scheduleAtAttr.(types.String); ok && !scheduleAtStr.IsNull() {
			v := scheduleAtStr.ValueString()
			scheduleAt = &v
		}
	}

	var cron *string
	if cronAttr, ok := propertiesObj["cron"]; ok {
		if cronStr, ok := cronAttr.(types.String); ok && !cronStr.IsNull() {
			v := cronStr.ValueString()
			cron = &v
		}
	}

	var executeUntil *string
	if executeUntilAttr, ok := propertiesObj["execute_until"]; ok {
		if executeUntilStr, ok := executeUntilAttr.(types.String); ok && !executeUntilStr.IsNull() {
			v := executeUntilStr.ValueString()
			executeUntil = &v
		}
	}

	steps := buildJobSteps(propertiesObj, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	builder := aruba.NewJob().
		Named(data.Name.ValueString()).
		InProject(aruba.URI("/projects/" + projectID)).
		InRegion(aruba.Region(data.Location.ValueString())).
		Tagged(tags...).
		WithSteps(steps...)

	switch jobTypeStr {
	case string(aruba.JobTypeOneShot):
		if scheduleAt != nil {
			t, parseErr := time.Parse(time.RFC3339, *scheduleAt)
			if parseErr == nil {
				builder = builder.OneShotAt(t)
			}
		} else {
			builder = builder.OfType(aruba.JobTypeOneShot)
		}
	case string(aruba.JobTypeRecurring):
		if cron != nil {
			builder = builder.WithCron(*cron)
		}
		if scheduleAt != nil {
			t, parseErr := time.Parse(time.RFC3339, *scheduleAt)
			if parseErr == nil {
				builder = builder.StartingAt(t)
			}
		}
		if executeUntil != nil {
			t, parseErr := time.Parse(time.RFC3339, *executeUntil)
			if parseErr == nil {
				builder = builder.RecurringUntil(t)
			}
		}
	default:
		builder = builder.OfType(aruba.JobType(jobTypeStr))
	}

	if enabled {
		builder = builder.Enabled()
	} else {
		builder = builder.Disabled()
	}

	job, err := r.client.Client.FromSchedule().Jobs().Create(ctx, builder)
	if provErr := CheckResponseErr("create", "ScheduleJob", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(job.ID())
	if uri := job.URI(); uri != "" {
		data.Uri = types.StringValue(uri)
	} else {
		data.Uri = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if waitErr := job.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "ScheduleJob", data.Id.ValueString())
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	applyJobToModel(job, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a Schedule Job resource", map[string]interface{}{
		"job_id":   data.Id.ValueString(),
		"job_name": data.Name.ValueString(),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScheduleJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ScheduleJobResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id.IsUnknown() || data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	job, err := r.client.Client.FromSchedule().Jobs().Get(ctx, jobRef(&data))
	if provErr := CheckResponseErr("read", "ScheduleJob", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	st := string(job.State())
	switch {
	case isFailedState(st):
		resp.Diagnostics.AddWarning("Resource in Failed State",
			fmt.Sprintf("ScheduleJob %q is in a terminal failure state (%s). "+
				"Run `terraform destroy` to clean it up, or `terraform apply -replace=<address>` to recreate it.", data.Id.ValueString(), st))
	case IsCreatingState(st):
		if waitErr := job.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "ScheduleJob", data.Id.ValueString())
			return
		}
		job, err = r.client.Client.FromSchedule().Jobs().Get(ctx, jobRef(&data))
		if provErr := CheckResponseErr("read", "ScheduleJob", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	applyJobToModel(job, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ProjectID = types.StringValue(data.ProjectID.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScheduleJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ScheduleJobResourceModel
	var state ScheduleJobResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	job, err := r.client.Client.FromSchedule().Jobs().Get(ctx, jobRef(&state))
	if provErr := CheckResponseErr("read", "ScheduleJob", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	job.Named(data.Name.ValueString())
	if tags != nil {
		job.RetaggedAs(tags...)
	} else {
		job.RetaggedAs(job.Tags()...)
	}

	propertiesObj := data.Properties.Attributes()
	if enabledAttr, ok := propertiesObj["enabled"]; ok {
		if enabledBool, ok := enabledAttr.(types.Bool); ok && !enabledBool.IsNull() {
			if enabledBool.ValueBool() {
				job.Enabled()
			} else {
				job.Disabled()
			}
		}
	}

	updated, err := r.client.Client.FromSchedule().Jobs().Update(ctx, job)
	if provErr := CheckResponseErr("update", "ScheduleJob", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = state.Id
	data.ProjectID = state.ProjectID
	applyJobToModel(updated, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Id = state.Id
	data.ProjectID = state.ProjectID

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScheduleJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ScheduleJobResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id.IsUnknown() || data.Id.IsNull() || data.Id.ValueString() == "" {
		return
	}

	ref := jobRef(&data)
	jobID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromSchedule().Jobs().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "ScheduleJob", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErrAsError("delete", "ScheduleJob",
			r.client.Client.FromSchedule().Jobs().Delete(ctx, ref))
	}, "ScheduleJob", jobID, r.client.ResourceTimeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting schedule job", err.Error())
		return
	}

	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "ScheduleJob", jobID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		if IsWaitTimeout(waitErr) {
			resp.Diagnostics.AddWarning(
				"ScheduleJob Deletion Pending",
				fmt.Sprintf("ScheduleJob %q delete was accepted but the resource is still visible in the API. "+
					"It will be cleaned up asynchronously. (%s)", jobID, waitErr),
			)
		} else {
			resp.Diagnostics.AddError("Error waiting for ScheduleJob deletion", waitErr.Error())
			return
		}
	}

	tflog.Trace(ctx, "deleted a Schedule Job resource", map[string]interface{}{"job_id": jobID})
}

func (r *ScheduleJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
