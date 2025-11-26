// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &SnapshotDataSource{}

func NewSnapshotDataSource() datasource.DataSource {
	return &SnapshotDataSource{}
}

type SnapshotDataSource struct {
	client *ArubaCloudClient
}

type SnapshotDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func (d *SnapshotDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshot"
}

func (d *SnapshotDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Snapshot data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Snapshot identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Snapshot name",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this Snapshot belongs to",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Snapshot location",
				Computed:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period (only 'Hour' allowed)",
				Computed:            true,
			},
			"volume_id": schema.StringAttribute{
				MarkdownDescription: "ID of the volume this snapshot is for",
				Computed:            true,
			},
		},
	}
}

func (d *SnapshotDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ArubaCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *SnapshotDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SnapshotDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Name = types.StringValue("example-snapshot")
	data.Description = types.StringValue("Simulated snapshot description")
	data.CreatedAt = types.StringValue("2025-11-25T00:00:00Z")
	tflog.Trace(ctx, "read a Snapshot data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
