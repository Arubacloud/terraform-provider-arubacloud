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

var _ datasource.DataSource = &ElasticIPDataSource{}

func NewElasticIPDataSource() datasource.DataSource {
	return &ElasticIPDataSource{}
}

type ElasticIPDataSource struct {
	client *ArubaCloudClient
}

type ElasticIPDataSourceModel struct {
	Id types.String `tfsdk:"id"`
	Ip types.String `tfsdk:"ip"`
}

func (d *ElasticIPDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticip"
}

func (d *ElasticIPDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Elastic IP data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Elastic IP identifier",
				Required:            true,
			},
			"ip": schema.StringAttribute{
				MarkdownDescription: "Elastic IP address",
				Computed:            true,
			},
		},
	}
}

func (d *ElasticIPDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ElasticIPDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ElasticIPDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Ip = types.StringValue("1.2.3.4")
	tflog.Trace(ctx, "read an Elastic IP data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
