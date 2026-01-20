// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	ProjectId     types.String `tfsdk:"project_id"`
	Address       types.String `tfsdk:"address"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	Tags          types.List   `tfsdk:"tags"`
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
			"name": schema.StringAttribute{
				MarkdownDescription: "Elastic IP name",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Elastic IP location",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this Elastic IP belongs to",
				Computed:            true,
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "Elastic IP address",
				Computed:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period for the Elastic IP",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the Elastic IP",
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

	// Populate all fields with example data
	data.Name = types.StringValue("example-elasticip")
	data.Location = types.StringValue("ITBG-Bergamo")
	data.ProjectId = types.StringValue("68398923fb2cb026400d4d31")
	data.Address = types.StringValue("203.0.113.10")
	data.BillingPeriod = types.StringValue("Hour")
	data.Tags = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("public-ip"),
		types.StringValue("production"),
	})

	tflog.Trace(ctx, "read an Elastic IP data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
