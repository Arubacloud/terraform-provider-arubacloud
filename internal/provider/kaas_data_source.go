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

var _ datasource.DataSource = &KaaSDataSource{}

func NewKaaSDataSource() datasource.DataSource {
	return &KaaSDataSource{}
}

type KaaSDataSource struct {
	client *ArubaCloudClient
}

type KaaSDataSourceModel struct {
	Id                types.String    `tfsdk:"id"`
	Name              types.String    `tfsdk:"name"`
	Location          types.String    `tfsdk:"location"`
	Tags              types.List      `tfsdk:"tags"`
	ProjectID         types.String    `tfsdk:"project_id"`
	Preset            types.Bool      `tfsdk:"preset"`
	VpcID             types.String    `tfsdk:"vpc_id"`
	SubnetID          types.String    `tfsdk:"subnet_id"`
	NodeCIDR          NodeCIDRModel   `tfsdk:"node_cidr"`
	SecurityGroupName types.String    `tfsdk:"security_group_name"`
	Version           types.String    `tfsdk:"version"`
	NodePools         []NodePoolModel `tfsdk:"node_pools"`
	HA                types.Bool      `tfsdk:"ha"`
	BillingPeriod     types.String    `tfsdk:"billing_period"`
}

type NodeCIDRModel struct {
	Address    types.String `tfsdk:"address"`
	SubnetName types.String `tfsdk:"subnet_name"`
}

type NodePoolModel struct {
	NodePoolName types.String `tfsdk:"node_pool_name"`
	Replicas     types.Int64  `tfsdk:"replicas"`
	Type         types.String `tfsdk:"type"`
	Zone         types.String `tfsdk:"zone"`
}

func (d *KaaSDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kaas"
}

func (d *KaaSDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "KaaS data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "KaaS identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "KaaS name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "KaaS location",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the KaaS resource",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this KaaS resource belongs to",
				Required:            true,
			},
			"preset": schema.BoolAttribute{
				MarkdownDescription: "Whether to use a preset configuration",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "VPC ID for the KaaS resource",
				Required:            true,
			},
			"subnet_id": schema.StringAttribute{
				MarkdownDescription: "Subnet ID for the KaaS resource",
				Required:            true,
			},
			"node_cidr": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						MarkdownDescription: "Node CIDR address",
						Required:            true,
					},
					"subnet_name": schema.StringAttribute{
						MarkdownDescription: "Node CIDR subnet name",
						Required:            true,
					},
				},
				Required: true,
			},
			"security_group_name": schema.StringAttribute{
				MarkdownDescription: "Security group name",
				Required:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Kubernetes version",
				Required:            true,
			},
			"node_pools": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"node_pool_name": schema.StringAttribute{Required: true},
						"replicas":       schema.Int64Attribute{Required: true},
						"type":           schema.StringAttribute{Required: true},
						"zone":           schema.StringAttribute{Required: true},
					},
				},
				Required: true,
			},
			"ha": schema.BoolAttribute{
				MarkdownDescription: "High availability",
				Required:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period",
				Required:            true,
			},
		},
	}
}

func (d *KaaSDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KaaSDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data KaaSDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Name = types.StringValue("example-kaas")
	tflog.Trace(ctx, "read a KaaS data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
