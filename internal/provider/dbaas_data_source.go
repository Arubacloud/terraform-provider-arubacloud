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

var _ datasource.DataSource = &DBaaSDataSource{}

func NewDBaaSDataSource() datasource.DataSource {
	return &DBaaSDataSource{}
}

type DBaaSDataSource struct {
	client *ArubaCloudClient
}

type DBaaSDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Zone          types.String `tfsdk:"zone"`
	Tags          types.List   `tfsdk:"tags"`
	ProjectID     types.String `tfsdk:"project_id"`
	EngineID      types.String `tfsdk:"engine_id"`
	Flavor        types.String `tfsdk:"flavor"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	// Storage fields (flattened)
	StorageSizeGB             types.Int64 `tfsdk:"storage_size_gb"`
	AutoscalingEnabled        types.Bool  `tfsdk:"autoscaling_enabled"`
	AutoscalingAvailableSpace types.Int64 `tfsdk:"autoscaling_available_space"`
	AutoscalingStepSize       types.Int64 `tfsdk:"autoscaling_step_size"`
	// Network fields (flattened)
	VpcUriRef           types.String `tfsdk:"vpc_uri_ref"`
	SubnetUriRef        types.String `tfsdk:"subnet_uri_ref"`
	SecurityGroupUriRef types.String `tfsdk:"security_group_uri_ref"`
	ElasticIpUriRef     types.String `tfsdk:"elastic_ip_uri_ref"`
}

func (d *DBaaSDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dbaas"
}

func (d *DBaaSDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "DBaaS data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "DBaaS identifier",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "DBaaS URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "DBaaS name",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "DBaaS location",
				Computed:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the DBaaS resource",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this DBaaS belongs to",
				Computed:            true,
			},
			"engine_id": schema.StringAttribute{
				MarkdownDescription: "Database engine ID (e.g., mysql-8.0, mssql-2022-web)",
				Computed:            true,
			},
			"flavor": schema.StringAttribute{
				MarkdownDescription: "DBaaS flavor name (e.g., DBO2A4)",
				Computed:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period (Hour, Month, Year)",
				Computed:            true,
			},
			"storage_size_gb": schema.Int64Attribute{
				MarkdownDescription: "Storage size in GB",
				Computed:            true,
			},
			"autoscaling_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable autoscaling",
				Computed:            true,
			},
			"autoscaling_available_space": schema.Int64Attribute{
				MarkdownDescription: "Minimum available space threshold in GB for autoscaling",
				Computed:            true,
			},
			"autoscaling_step_size": schema.Int64Attribute{
				MarkdownDescription: "Step size for autoscaling (in GB)",
				Computed:            true,
			},
			"vpc_uri_ref": schema.StringAttribute{
				MarkdownDescription: "VPC URI reference",
				Computed:            true,
			},
			"subnet_uri_ref": schema.StringAttribute{
				MarkdownDescription: "Subnet URI reference",
				Computed:            true,
			},
			"security_group_uri_ref": schema.StringAttribute{
				MarkdownDescription: "Security Group URI reference",
				Computed:            true,
			},
			"elastic_ip_uri_ref": schema.StringAttribute{
				MarkdownDescription: "Elastic IP URI reference",
				Computed:            true,
			},
		},
	}
}

func (d *DBaaSDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DBaaSDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DBaaSDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Populate all fields with example data
	data.Uri = types.StringValue("/v2/dbaas/dbaas-68398923fb2cb026400d4d31")
	data.Name = types.StringValue("example-dbaas")
	data.Location = types.StringValue("ITBG-Bergamo")
	data.Zone = types.StringValue("ITBG-1")
	data.Tags = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("database"),
		types.StringValue("production"),
	})
	data.ProjectID = types.StringValue("68398923fb2cb026400d4d31")
	data.EngineID = types.StringValue("mysql-8.0")
	data.Flavor = types.StringValue("DBO2A4")
	data.BillingPeriod = types.StringValue("Hour")
	// Storage fields
	data.StorageSizeGB = types.Int64Value(100)
	data.AutoscalingEnabled = types.BoolValue(true)
	data.AutoscalingAvailableSpace = types.Int64Value(10)
	data.AutoscalingStepSize = types.Int64Value(20)
	// Network fields
	data.VpcUriRef = types.StringValue("/v2/vpcs/vpc-68398923fb2cb026400d4d32")
	data.SubnetUriRef = types.StringValue("/v2/subnets/subnet-68398923fb2cb026400d4d33")
	data.SecurityGroupUriRef = types.StringValue("/v2/securitygroups/sg-68398923fb2cb026400d4d34")
	data.ElasticIpUriRef = types.StringValue("/v2/elasticips/eip-68398923fb2cb026400d4d35")
	tflog.Trace(ctx, "read a DBaaS data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
