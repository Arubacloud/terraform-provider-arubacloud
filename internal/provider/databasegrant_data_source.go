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

var _ datasource.DataSource = &DatabaseGrantDataSource{}

func NewDatabaseGrantDataSource() datasource.DataSource {
	return &DatabaseGrantDataSource{}
}

type DatabaseGrantDataSource struct {
	client *http.Client
}

type DatabaseGrantDataSourceModel struct {
	Id       types.String `tfsdk:"id"`
	Database types.String `tfsdk:"database"`
	UserID   types.String `tfsdk:"user_id"`
	Role     types.String `tfsdk:"role"`
}

func (d *DatabaseGrantDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databasegrant"
}

func (d *DatabaseGrantDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Database Grant data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Database Grant identifier",
				Required:            true,
			},
			"database": schema.StringAttribute{
				MarkdownDescription: "Database name or ID",
				Computed:            true,
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "User ID to grant access",
				Computed:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Role to grant (e.g., read, write, admin)",
				Computed:            true,
			},
		},
	}
}

func (d *DatabaseGrantDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DatabaseGrantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DatabaseGrantDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Simulate API response for all attributes
	data.Database = types.StringValue("example-database")
	data.UserID = types.StringValue("example-user-id")
	data.Role = types.StringValue("read")
	tflog.Trace(ctx, "read a Database Grant data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
