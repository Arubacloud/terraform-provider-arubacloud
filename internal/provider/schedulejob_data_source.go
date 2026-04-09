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
		MarkdownDescription: "Schedule Job data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Schedule Job identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Schedule Job name",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this Schedule Job belongs to",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Schedule Job description",
				Computed:            true,
			},
			"cron": schema.StringAttribute{
				MarkdownDescription: "Cron expression for the schedule job",
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
		resp.Diagnostics.AddError("Error reading schedule job", fmt.Sprintf("Unable to read schedule job: %s", err))
		return
	}
	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.Diagnostics.AddError("Schedule Job not found", fmt.Sprintf("No schedule job found with ID %q in project %q", jobID, projectID))
			return
		}
		resp.Diagnostics.AddError("API Error", FormatAPIError(ctx, response.Error, "Failed to read schedule job", map[string]interface{}{"project_id": projectID, "job_id": jobID}))
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
