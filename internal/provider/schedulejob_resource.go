// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ScheduleJobResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
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
		MarkdownDescription: "Schedule Job resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Schedule Job identifier",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Schedule Job URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Schedule Job name",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this job belongs to",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the job",
				Optional:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Location for the job",
				Required:            true,
			},
			"properties": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Whether the job is enabled.",
						Optional:            true,
					},
					"schedule_job_type": schema.StringAttribute{
						MarkdownDescription: "Type of job (OneShot, Recurring)",
						Required:            true,
					},
					"schedule_at": schema.StringAttribute{
						MarkdownDescription: "Date and time when the job should run (for OneShot)",
						Optional:            true,
					},
					"execute_until": schema.StringAttribute{
						MarkdownDescription: "End date until which the job can run (for Recurring)",
						Optional:            true,
					},
					"cron": schema.StringAttribute{
						MarkdownDescription: "CRON expression for recurrence (for Recurring)",
						Optional:            true,
					},
					"steps": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Descriptive name of the step.",
									Optional:            true,
								},
								"resource_uri": schema.StringAttribute{
									MarkdownDescription: "URI of the resource.",
									Required:            true,
								},
								"action_uri": schema.StringAttribute{
									MarkdownDescription: "URI of the action to execute.",
									Required:            true,
								},
								"http_verb": schema.StringAttribute{
									MarkdownDescription: "HTTP verb to use (GET, POST, etc.)",
									Required:            true,
								},
								"body": schema.StringAttribute{
									MarkdownDescription: "Optional HTTP request body.",
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

	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Extract properties
	var propertiesObj map[string]attr.Value
	diags := data.Properties.As(ctx, &propertiesObj, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobType := propertiesObj["schedule_job_type"].(types.String).ValueString()
	enabled := true
	if enabledAttr, ok := propertiesObj["enabled"]; ok && !enabledAttr.(types.Bool).IsNull() {
		enabled = enabledAttr.(types.Bool).ValueBool()
	}

	var scheduleAt *string
	if scheduleAtAttr, ok := propertiesObj["schedule_at"]; ok && !scheduleAtAttr.(types.String).IsNull() {
		scheduleAtStr := scheduleAtAttr.(types.String).ValueString()
		scheduleAt = &scheduleAtStr
	}

	var cron *string
	if cronAttr, ok := propertiesObj["cron"]; ok && !cronAttr.(types.String).IsNull() {
		cronStr := cronAttr.(types.String).ValueString()
		cron = &cronStr
	}

	var executeUntil *string
	if executeUntilAttr, ok := propertiesObj["execute_until"]; ok && !executeUntilAttr.(types.String).IsNull() {
		executeUntilStr := executeUntilAttr.(types.String).ValueString()
		executeUntil = &executeUntilStr
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
		},
	}

	// Create the job using the SDK
	response, err := r.client.Client.FromSchedule().Jobs().Create(ctx, projectID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating schedule job",
			fmt.Sprintf("Unable to create schedule job: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create schedule job"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
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
		resp.Diagnostics.AddError(
			"Schedule Job Not Active",
			fmt.Sprintf("Schedule job was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a Schedule Job resource", map[string]interface{}{
		"job_id": data.Id.ValueString(),
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

	if projectID == "" || jobID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Job ID are required to read the schedule job",
		)
		return
	}

	// Get job details using the SDK
	response, err := r.client.Client.FromSchedule().Jobs().Get(ctx, projectID, jobID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading schedule job",
			fmt.Sprintf("Unable to read schedule job: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read schedule job"
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

		// Reconstruct properties object
		propertiesAttrs := map[string]attr.Value{
			"enabled":            types.BoolValue(job.Properties.Enabled),
			"schedule_job_type":  types.StringValue(string(job.Properties.JobType)),
			"schedule_at":        types.StringNull(),
			"execute_until":      types.StringNull(),
			"cron":               types.StringNull(),
			"steps":              types.ListNull(types.ObjectType{}),
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
				"steps":             types.ListType{ElemType: types.ObjectType{}},
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
			fmt.Sprintf("Unable to get schedule job: %s", err),
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

	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		tags = current.Metadata.Tags
	}

	// Extract properties
	var propertiesObj map[string]attr.Value
	diags := data.Properties.As(ctx, &propertiesObj, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	enabled := current.Properties.Enabled
	if enabledAttr, ok := propertiesObj["enabled"]; ok && !enabledAttr.(types.Bool).IsNull() {
		enabled = enabledAttr.(types.Bool).ValueBool()
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
			ScheduleAt:     current.Properties.ScheduleAt,
			ExecuteUntil: current.Properties.ExecuteUntil,
			Cron:         current.Properties.Cron,
		},
	}

	// Update the job using the SDK
	response, err := r.client.Client.FromSchedule().Jobs().Update(ctx, projectID, jobID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating schedule job",
			fmt.Sprintf("Unable to update schedule job: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to update schedule job"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
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
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromSchedule().Jobs().Delete(ctx, projectID, jobID, nil)
		},
		ExtractSDKError,
		"ScheduleJob",
		jobID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting schedule job",
			fmt.Sprintf("Unable to delete schedule job: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a Schedule Job resource", map[string]interface{}{
		"job_id": jobID,
	})
}

func (r *ScheduleJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
