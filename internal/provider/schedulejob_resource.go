package provider

import (
	"context"
	"fmt"
	"time"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
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

func (r *ScheduleJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ScheduleJobResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()

	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID is required to create a schedule job",
		)
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract properties
	propertiesObj := data.Properties.Attributes()

	jobTypeAttr, ok := propertiesObj["schedule_job_type"].(types.String)
	if !ok {
		resp.Diagnostics.AddError("Invalid Type", "schedule_job_type must be a String")
		return
	}
	jobType := jobTypeAttr.ValueString()

	enabled := true
	if enabledAttr, ok := propertiesObj["enabled"]; ok {
		if enabledBool, ok := enabledAttr.(types.Bool); ok && !enabledBool.IsNull() {
			enabled = enabledBool.ValueBool()
		}
	}

	var scheduleAt *string
	if scheduleAtAttr, ok := propertiesObj["schedule_at"]; ok {
		if scheduleAtStr, ok := scheduleAtAttr.(types.String); ok && !scheduleAtStr.IsNull() {
			scheduleAtVal := scheduleAtStr.ValueString()
			scheduleAt = &scheduleAtVal
		}
	}

	var cron *string
	if cronAttr, ok := propertiesObj["cron"]; ok {
		if cronStr, ok := cronAttr.(types.String); ok && !cronStr.IsNull() {
			cronVal := cronStr.ValueString()
			cron = &cronVal
		}
	}

	var executeUntil *string
	if executeUntilAttr, ok := propertiesObj["execute_until"]; ok {
		if executeUntilStr, ok := executeUntilAttr.(types.String); ok && !executeUntilStr.IsNull() {
			executeUntilVal := executeUntilStr.ValueString()
			executeUntil = &executeUntilVal
		}
	}

	// Extract steps
	var steps []sdktypes.JobStep
	if stepsAttr, ok := propertiesObj["steps"]; ok {
		if stepsList, ok := stepsAttr.(types.List); ok && !stepsList.IsNull() {
			var stepsElements []types.Object
			diags := stepsList.ElementsAs(ctx, &stepsElements, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			for _, stepObj := range stepsElements {
				stepAttrs := stepObj.Attributes()

				var name *string
				if nameAttr, ok := stepAttrs["name"]; ok {
					if nameStr, ok := nameAttr.(types.String); ok && !nameStr.IsNull() {
						nameVal := nameStr.ValueString()
						name = &nameVal
					}
				}

				var resourceURI string
				if resURIAttr, ok := stepAttrs["resource_uri"]; ok {
					if resURIStr, ok := resURIAttr.(types.String); ok {
						resourceURI = resURIStr.ValueString()
					}
				}

				var actionURI string
				if actURIAttr, ok := stepAttrs["action_uri"]; ok {
					if actURIStr, ok := actURIAttr.(types.String); ok {
						actionURI = actURIStr.ValueString()
					}
				}

				var httpVerb string
				if verbAttr, ok := stepAttrs["http_verb"]; ok {
					if verbStr, ok := verbAttr.(types.String); ok {
						httpVerb = verbStr.ValueString()
					}
				}

				var body *string
				if bodyAttr, ok := stepAttrs["body"]; ok {
					if bodyStr, ok := bodyAttr.(types.String); ok && !bodyStr.IsNull() {
						bodyVal := bodyStr.ValueString()
						body = &bodyVal
					}
				}

				steps = append(steps, sdktypes.JobStep{
					Name:        name,
					ResourceURI: resourceURI,
					ActionURI:   actionURI,
					HttpVerb:    httpVerb,
					Body:        body,
				})
			}
		}
	}

	// Build the create request
	createRequest := sdktypes.JobRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.JobPropertiesRequest{
			Enabled:      enabled,
			JobType:      sdktypes.TypeJob(jobType),
			ScheduleAt:   scheduleAt,
			Cron:         cron,
			ExecuteUntil: executeUntil,
			Steps:        steps,
		},
	}

	// Create the job using the SDK
	response, err := r.client.Client.FromSchedule().Jobs().Create(ctx, projectID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating schedule job",
			NewTransportError("create", "Schedulejob", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("create", "Schedulejob", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}

	if response != nil && response.Data != nil && response.Data.Metadata.ID != nil {
		data.Id = types.StringValue(*response.Data.Metadata.ID)
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Schedule job created but no ID returned from API",
		)
		return
	}

	// Wait for Schedule Job to be active before returning
	// This ensures Terraform doesn't proceed until Job is ready
	jobID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromSchedule().Jobs().Get(ctx, projectID, jobID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for Schedule Job to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "ScheduleJob", jobID, r.client.ResourceTimeout); err != nil {
		ReportWaitResult(&resp.Diagnostics, err, "ScheduleJob", jobID)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

	projectID := data.ProjectID.ValueString()
	jobID := data.Id.ValueString()

	if data.Id.IsUnknown() || data.Id.IsNull() || jobID == "" {
		tflog.Debug(ctx, "Schedule Job ID is empty, removing resource from state", map[string]interface{}{"job_id": jobID})
		resp.State.RemoveResource(ctx)
		return
	}

	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID is required to read the schedule job",
		)
		return
	}

	// Get job details using the SDK
	response, err := r.client.Client.FromSchedule().Jobs().Get(ctx, projectID, jobID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading schedule job",
			NewTransportError("read", "Schedulejob", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("read", "Schedulejob", response); apiErr != nil {
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
				fmt.Sprintf("ScheduleJob %q reached a terminal failure state (%s) and will not recover on its own. "+
					"Use `terraform apply -replace=<address>` to recreate it.", jobID, st),
			)
			return
		case IsCreatingState(st):
			checker := func(ctx context.Context) (string, error) {
				getResp, err := r.client.Client.FromSchedule().Jobs().Get(ctx, projectID, jobID, nil)
				if err != nil {
					return "", err
				}
				if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
					return *getResp.Data.Status.State, nil
				}
				return "Unknown", nil
			}
			if err := WaitForResourceActive(ctx, checker, "ScheduleJob", jobID, r.client.ResourceTimeout); err != nil {
				ReportWaitResult(&resp.Diagnostics, err, "ScheduleJob", jobID)
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
			// Re-read to get the final active state.
			response, err = r.client.Client.FromSchedule().Jobs().Get(ctx, projectID, jobID, nil)
			if err != nil {
				resp.Diagnostics.AddError("Error reading ScheduleJob after provisioning wait",
					NewTransportError("read", "Schedulejob", err).Error())
				return
			}
			if apiErr := CheckResponse("read", "Schedulejob", response); apiErr != nil {
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
		job := response.Data
		if job.Metadata.ID != nil {
			data.Id = types.StringValue(*job.Metadata.ID)
			if response.Data.Metadata.URI != nil {
				data.Uri = types.StringValue(*response.Data.Metadata.URI)
			} else {
				data.Uri = types.StringNull()
			}
			if response.Data.Metadata.URI != nil {
				data.Uri = types.StringValue(*response.Data.Metadata.URI)
			} else {
				data.Uri = types.StringNull()
			}
			if response.Data.Metadata.URI != nil {
				data.Uri = types.StringValue(*response.Data.Metadata.URI)
			} else {
				data.Uri = types.StringNull()
			}
		}
		if job.Metadata.Name != nil {
			data.Name = types.StringValue(*job.Metadata.Name)
		}
		if job.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(job.Metadata.LocationResponse.Value)
		}
		if job.Metadata.Tags != nil {
			tagValues := make([]attr.Value, len(job.Metadata.Tags))
			for i, tag := range job.Metadata.Tags {
				tagValues[i] = types.StringValue(tag)
			}
			tagsList, diags := types.ListValue(types.StringType, tagValues)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = tagsList
			}
		} else {
			data.Tags = types.ListNull(types.StringType)
		}

		// Define step object type
		stepObjectType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":         types.StringType,
				"resource_uri": types.StringType,
				"action_uri":   types.StringType,
				"http_verb":    types.StringType,
				"body":         types.StringType,
			},
		}

		// Convert steps from API response
		var stepsListValue types.List
		if len(job.Properties.Steps) > 0 {
			stepObjects := make([]attr.Value, len(job.Properties.Steps))
			for i, step := range job.Properties.Steps {
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

				stepObj, diags := types.ObjectValue(stepObjectType.AttrTypes, stepAttrs)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				stepObjects[i] = stepObj
			}
			var diags diag.Diagnostics
			stepsListValue, diags = types.ListValue(stepObjectType, stepObjects)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		} else {
			stepsListValue = types.ListNull(stepObjectType)
		}

		// Reconstruct properties object
		propertiesAttrs := map[string]attr.Value{
			"enabled":           types.BoolValue(job.Properties.Enabled),
			"schedule_job_type": types.StringValue(string(job.Properties.JobType)),
			"schedule_at":       types.StringNull(),
			"execute_until":     types.StringNull(),
			"cron":              types.StringNull(),
			"steps":             stepsListValue,
		}

		if job.Properties.ScheduleAt != nil {
			propertiesAttrs["schedule_at"] = types.StringValue(*job.Properties.ScheduleAt)
		}
		if job.Properties.ExecuteUntil != nil {
			propertiesAttrs["execute_until"] = types.StringValue(*job.Properties.ExecuteUntil)
		}
		if job.Properties.Cron != nil {
			propertiesAttrs["cron"] = types.StringValue(*job.Properties.Cron)
		}

		propertiesObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"enabled":           types.BoolType,
				"schedule_job_type": types.StringType,
				"schedule_at":       types.StringType,
				"execute_until":     types.StringType,
				"cron":              types.StringType,
				"steps":             types.ListType{ElemType: stepObjectType},
			},
			propertiesAttrs,
		)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Properties = propertiesObj
		}
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScheduleJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ScheduleJobResourceModel
	var state ScheduleJobResourceModel

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
	jobID := state.Id.ValueString()

	if projectID == "" || jobID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Job ID are required to update the schedule job",
		)
		return
	}

	// Get current job to preserve fields
	getResp, err := r.client.Client.FromSchedule().Jobs().Get(ctx, projectID, jobID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting schedule job",
			NewTransportError("read", "Schedulejob", err).Error(),
		)
		return
	}

	if getResp == nil || getResp.Data == nil {
		resp.Diagnostics.AddError(
			"Job Not Found",
			"Schedule job not found",
		)
		return
	}

	current := getResp.Data
	regionValue := ""
	if current.Metadata.LocationResponse != nil {
		regionValue = current.Metadata.LocationResponse.Value
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if tags == nil {
		tags = current.Metadata.Tags
	}

	// Extract properties
	propertiesObj := data.Properties.Attributes()

	enabled := current.Properties.Enabled
	if enabledAttr, ok := propertiesObj["enabled"]; ok {
		if enabledBool, ok := enabledAttr.(types.Bool); ok && !enabledBool.IsNull() {
			enabled = enabledBool.ValueBool()
		}
	}

	// Extract steps from properties
	var steps []sdktypes.JobStep
	if stepsAttr, ok := propertiesObj["steps"]; ok {
		if stepsList, ok := stepsAttr.(types.List); ok && !stepsList.IsNull() {
			var stepsElements []types.Object
			diags := stepsList.ElementsAs(ctx, &stepsElements, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			for _, stepObj := range stepsElements {
				stepAttrs := stepObj.Attributes()

				var name *string
				if nameAttr, ok := stepAttrs["name"]; ok {
					if nameStr, ok := nameAttr.(types.String); ok && !nameStr.IsNull() {
						nameVal := nameStr.ValueString()
						name = &nameVal
					}
				}

				var resourceURI string
				if resURIAttr, ok := stepAttrs["resource_uri"]; ok {
					if resURIStr, ok := resURIAttr.(types.String); ok {
						resourceURI = resURIStr.ValueString()
					}
				}

				var actionURI string
				if actURIAttr, ok := stepAttrs["action_uri"]; ok {
					if actURIStr, ok := actURIAttr.(types.String); ok {
						actionURI = actURIStr.ValueString()
					}
				}

				var httpVerb string
				if verbAttr, ok := stepAttrs["http_verb"]; ok {
					if verbStr, ok := verbAttr.(types.String); ok {
						httpVerb = verbStr.ValueString()
					}
				}

				var body *string
				if bodyAttr, ok := stepAttrs["body"]; ok {
					if bodyStr, ok := bodyAttr.(types.String); ok && !bodyStr.IsNull() {
						bodyVal := bodyStr.ValueString()
						body = &bodyVal
					}
				}

				steps = append(steps, sdktypes.JobStep{
					Name:        name,
					ResourceURI: resourceURI,
					ActionURI:   actionURI,
					HttpVerb:    httpVerb,
					Body:        body,
				})
			}
		}
	} else {
		// If steps not provided in update, use current steps
		if current.Properties.Steps != nil {
			steps = make([]sdktypes.JobStep, len(current.Properties.Steps))
			for i, currentStep := range current.Properties.Steps {
				step := sdktypes.JobStep{
					Name: currentStep.Name,
					Body: currentStep.Body,
				}
				if currentStep.ResourceURI != nil {
					step.ResourceURI = *currentStep.ResourceURI
				}
				if currentStep.ActionURI != nil {
					step.ActionURI = *currentStep.ActionURI
				}
				if currentStep.HttpVerb != nil {
					step.HttpVerb = *currentStep.HttpVerb
				}
				steps[i] = step
			}
		}
	}

	// Build update request
	updateRequest := sdktypes.JobRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.JobPropertiesRequest{
			Enabled:      enabled,
			JobType:      current.Properties.JobType,
			ScheduleAt:   current.Properties.ScheduleAt,
			ExecuteUntil: current.Properties.ExecuteUntil,
			Cron:         current.Properties.Cron,
			Steps:        steps,
		},
	}

	// Update the job using the SDK
	response, err := r.client.Client.FromSchedule().Jobs().Update(ctx, projectID, jobID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating schedule job",
			NewTransportError("update", "Schedulejob", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("update", "Schedulejob", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}

	if response != nil && response.Data != nil && response.Data.Metadata.ID != nil {
		data.Id = types.StringValue(*response.Data.Metadata.ID)
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
	}

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.ProjectID = state.ProjectID

	if response != nil && response.Data != nil {
		// Update from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScheduleJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ScheduleJobResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	jobID := data.Id.ValueString()

	if projectID == "" || jobID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Job ID are required to delete the schedule job",
		)
		return
	}

	// Delete the job using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	deletionChecker := func(ctx context.Context) (bool, error) {
		getResp, getErr := r.client.Client.FromSchedule().Jobs().Get(ctx, projectID, jobID, nil)
		if getErr != nil {
			return false, NewTransportError("get", "ScheduleJob", getErr)
		}
		if provErr := CheckResponse("get", "ScheduleJob", getResp); provErr != nil {
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
			resp, err := r.client.Client.FromSchedule().Jobs().Delete(ctx, projectID, jobID, nil)
			if err != nil {
				return NewTransportError("delete", "ScheduleJob", err)
			}
			return CheckResponse("delete", "ScheduleJob", resp)
		},
		"ScheduleJob",
		jobID,
		r.client.ResourceTimeout,
		deletionChecker,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting schedule job",
			NewTransportError("delete", "Schedulejob", err).Error(),
		)
		return
	}

	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "ScheduleJob", jobID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for ScheduleJob deletion", waitErr.Error())
		return
	}

	tflog.Trace(ctx, "deleted a Schedule Job resource", map[string]interface{}{
		"job_id": jobID,
	})
}

func (r *ScheduleJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
