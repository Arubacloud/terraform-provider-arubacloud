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

var _ datasource.DataSource = &BackupDataSource{}

func NewBackupDataSource() datasource.DataSource {
	return &BackupDataSource{}
}

type BackupDataSource struct {
	client *http.Client
}

type BackupDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	ProjectID     types.String `tfsdk:"project_id"`
	Type          types.String `tfsdk:"type"`
	VolumeID      types.String `tfsdk:"volume_id"`
	RetentionDays types.Int64  `tfsdk:"retention_days"`
	BillingPeriod types.String `tfsdk:"billing_period"`
}

func (d *BackupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup"
}

func (d *BackupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Backup data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Backup identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Backup name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Backup location",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the backup resource",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this backup belongs to",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of backup (Full, Incremental)",
				Required:            true,
			},
			"volume_id": schema.StringAttribute{
				MarkdownDescription: "Volume ID for the backup",
				Required:            true,
			},
			"retention_days": schema.Int64Attribute{
				MarkdownDescription: "Retention days for the backup",
				Optional:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period",
				Required:            true,
			},
		},
	}
}

func (d *BackupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BackupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BackupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Name = types.StringValue("example-backup")
	tflog.Trace(ctx, "read a Backup data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
