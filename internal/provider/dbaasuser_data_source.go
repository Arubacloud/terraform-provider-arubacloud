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

type DBaaSUserDataSourceModel struct {
	Id       types.String `tfsdk:"id"`
	DBaaSID  types.String `tfsdk:"dbaas_id"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type DBaaSUserDataSource struct {
	client *http.Client
}

var _ datasource.DataSource = &DBaaSUserDataSource{}

func NewDBaaSUserDataSource() datasource.DataSource {
	return &DBaaSUserDataSource{}
}

func (d *DBaaSUserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dbaasuser"
}

func (d *DBaaSUserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "DBaaS User data source",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for the DBaaS user (lookup key)",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "DBaaS User identifier",
				Computed:            true,
			},
			"dbaas_id": schema.StringAttribute{
				MarkdownDescription: "DBaaS ID this user belongs to",
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for the DBaaS user",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (d *DBaaSUserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*http.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *DBaaSUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DBaaSUserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Simulate API response
	data.Id = types.StringValue("dbaasuser-id")
	data.Password = types.StringValue("simulated-password")
	tflog.Trace(ctx, "read a DBaaS User data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
