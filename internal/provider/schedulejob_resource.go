// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ScheduleJobResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	ProjectID  types.String `tfsdk:"project_id"`
	Tags       types.List   `tfsdk:"tags"`
	Location   types.String `tfsdk:"location"`
	Properties types.Object `tfsdk:"properties"`
}

type ScheduleJobResource struct {
	client *http.Client
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
	client, ok := req.ProviderData.(*http.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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
	// Simulate API response
	data.Id = types.StringValue("schedulejob-id")
	tflog.Trace(ctx, "created a Schedule Job resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScheduleJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ScheduleJobResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScheduleJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ScheduleJobResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScheduleJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ScheduleJobResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ScheduleJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
