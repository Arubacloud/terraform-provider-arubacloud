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

type KMIPDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Endpoint    types.String `tfsdk:"endpoint"`
}

type KMIPDataSource struct {
	client *ArubaCloudClient
}

var _ datasource.DataSource = &KMIPDataSource{}

func NewKMIPDataSource() datasource.DataSource {
	return &KMIPDataSource{}
}

func (d *KMIPDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kmip"
}

func (d *KMIPDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "KMIP data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "KMIP identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the KMIP resource",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the KMIP resource",
				Computed:            true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "KMIP endpoint URL",
				Computed:            true,
			},
		},
	}
}

func (d *KMIPDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ArubaCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *KMIPDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data KMIPDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Simulate API response
	data.Id = types.StringValue("kmip-id")
	data.Description = types.StringValue("Simulated KMIP description")
	data.Endpoint = types.StringValue("https://kmip.example.com")
	tflog.Trace(ctx, "read a KMIP data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
