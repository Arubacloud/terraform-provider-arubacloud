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

type KMSDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Endpoint    types.String `tfsdk:"endpoint"`
}

type KMSDataSource struct {
	client *ArubaCloudClient
}

var _ datasource.DataSource = &KMSDataSource{}

func NewKMSDataSource() datasource.DataSource {
	return &KMSDataSource{}
}

func (d *KMSDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kms"
}

func (d *KMSDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "KMS data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "KMS identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the KMS resource",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the KMS resource",
				Computed:            true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "KMS endpoint URL",
				Computed:            true,
			},
		},
	}
}

func (d *KMSDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KMSDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data KMSDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Simulate API response
	data.Id = types.StringValue("kms-id")
	data.Description = types.StringValue("Simulated KMS description")
	data.Endpoint = types.StringValue("https://kms.example.com")
	tflog.Trace(ctx, "read a KMS data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
