// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"fmt"
	"net/http"

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
	client *http.Client
}

type ScheduleJobDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
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
	client, ok := req.ProviderData.(*http.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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
	data.Name = types.StringValue("example-schedulejob")
	data.Description = types.StringValue("Simulated schedule job description")
	data.Cron = types.StringValue("0 0 * * *")
	tflog.Trace(ctx, "read a Schedule Job data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
