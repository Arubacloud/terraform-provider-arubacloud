package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &ScheduleJobDataSource{}

func NewScheduleJobDataSource() datasource.DataSource {
	return &ScheduleJobDataSource{}
}

type ScheduleJobDataSource struct {
	client *ArubaCloudClient
}

type ScheduleJobDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ProjectID   types.String `tfsdk:"project_id"`
	Description types.String `tfsdk:"description"`
	Cron        types.String `tfsdk:"cron"`
}

func (d *ScheduleJobDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedulejob"
}

func (d *ScheduleJobDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about an existing ArubaCloud Scheduled Job.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the scheduled job to look up.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the scheduled job.",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional human-readable description of the scheduled job.",
				Computed:            true,
			},
			"cron": schema.StringAttribute{
				MarkdownDescription: "Cron expression defining the job schedule (e.g., `0 * * * *` for hourly). Standard 5-field cron format.",
				Computed:            true,
			},
		},
	}
}

func (d *ScheduleJobDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ArubaCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ArubaCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *ScheduleJobDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ScheduleJobDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	jobID := data.Id.ValueString()
	if projectID == "" || jobID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID and Schedule Job ID are required to read the schedule job")
		return
	}

	response, err := d.client.Client.FromSchedule().Jobs().Get(ctx, projectID, jobID, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading schedule job", NewTransportError("read", "Schedulejob", err).Error())
		return
	}
	if apiErr := CheckResponse("read", "Schedulejob", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError("No data returned", "Schedule Job Get returned no data")
		return
	}

	job := response.Data
	if job.Metadata.ID != nil {
		data.Id = types.StringValue(*job.Metadata.ID)
	}
	if job.Metadata.Name != nil {
		data.Name = types.StringValue(*job.Metadata.Name)
	}
	data.ProjectID = types.StringValue(projectID)
	// description and cron are not returned in the metadata response
	data.Description = types.StringNull()
	data.Cron = types.StringNull()

	tflog.Trace(ctx, "read a Schedule Job data source", map[string]interface{}{"job_id": jobID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
