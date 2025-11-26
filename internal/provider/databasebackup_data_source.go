// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

var _ datasource.DataSource = &DatabaseBackupDataSource{}

func NewDatabaseBackupDataSource() datasource.DataSource {
	return &DatabaseBackupDataSource{}
}

type DatabaseBackupDataSource struct {
	client *http.Client
}

type DatabaseBackupDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	Zone          types.String `tfsdk:"zone"`
	DBaaSID       types.String `tfsdk:"dbaas_id"`
	Database      types.String `tfsdk:"database"`
	BillingPeriod types.String `tfsdk:"billing_period"`
}

func (d *DatabaseBackupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databasebackup"
}

func (d *DatabaseBackupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Database Backup data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Database Backup identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Database Backup name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Database Backup location",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the Database Backup resource",
				Optional:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone for the Database Backup",
				Required:            true,
			},
			"dbaas_id": schema.StringAttribute{
				MarkdownDescription: "DBaaS ID this backup belongs to",
				Required:            true,
			},
			"database": schema.StringAttribute{
				MarkdownDescription: "Database to backup (ID or name)",
				Required:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period",
				Required:            true,
			},
		},
	}
}

func (d *DatabaseBackupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DatabaseBackupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DatabaseBackupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Name = types.StringValue("example-databasebackup")
	tflog.Trace(ctx, "read a Database Backup data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
