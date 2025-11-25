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

var _ datasource.DataSource = &DBaaSDataSource{}

func NewDBaaSDataSource() datasource.DataSource {
	return &DBaaSDataSource{}
}

type DBaaSDataSource struct {
	client *http.Client
}

type DBaaSDataSourceModel struct {
	Id            types.String           `tfsdk:"id"`
	Name          types.String           `tfsdk:"name"`
	Location      types.String           `tfsdk:"location"`
	Tags          types.List             `tfsdk:"tags"`
	ProjectID     types.String           `tfsdk:"project_id"`
	Engine        types.String           `tfsdk:"engine"`
	Zone          types.String           `tfsdk:"zone"`
	Flavor        types.String           `tfsdk:"flavor"`
	StorageSize   types.Int64            `tfsdk:"storage_size"`
	BillingPeriod types.String           `tfsdk:"billing_period"`
	Network       *DBaaSNetworkModel     `tfsdk:"network"`
	Autoscaling   *DBaaSAutoscalingModel `tfsdk:"autoscaling"`
}

type DBaaSNetworkModel struct {
	VpcID           types.String `tfsdk:"vpc_id"`
	SubnetID        types.String `tfsdk:"subnet_id"`
	SecurityGroupID types.String `tfsdk:"security_group_id"`
	ElasticIpID     types.String `tfsdk:"elastic_ip_id"`
}

type DBaaSAutoscalingModel struct {
	Enabled        types.Bool  `tfsdk:"enabled"`
	AvailableSpace types.Int64 `tfsdk:"available_space"`
	StepSize       types.Int64 `tfsdk:"step_size"`
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
			"name": schema.StringAttribute{
				MarkdownDescription: "DBaaS name",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "DBaaS location",
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
			"engine": schema.StringAttribute{
				MarkdownDescription: "Database engine (mysql-8.0, mssql-2022-web, mssql-2022-standard, mssql-2022-enterprise)",
				Computed:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone (ITBG-1, ITBG-2, ITBG-3)",
				Computed:            true,
			},
			"flavor": schema.StringAttribute{
				MarkdownDescription: "Flavor type",
				Computed:            true,
			},
			"storage_size": schema.Int64Attribute{
				MarkdownDescription: "Storage size",
				Computed:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period",
				Computed:            true,
			},
			"network": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"vpc_id": schema.StringAttribute{
						MarkdownDescription: "VPC ID",
						Computed:            true,
					},
					"subnet_id": schema.StringAttribute{
						MarkdownDescription: "Subnet ID",
						Computed:            true,
					},
					"security_group_id": schema.StringAttribute{
						MarkdownDescription: "Security Group ID",
						Computed:            true,
					},
					"elastic_ip_id": schema.StringAttribute{
						MarkdownDescription: "Elastic IP ID",
						Computed:            true,
					},
				},
				Computed: true,
			},
			"autoscaling": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Autoscaling enabled",
						Computed:            true,
					},
					"available_space": schema.Int64Attribute{
						MarkdownDescription: "Available space for autoscaling",
						Computed:            true,
					},
					"step_size": schema.Int64Attribute{
						MarkdownDescription: "Step size for autoscaling",
						Computed:            true,
					},
				},
				Computed: true,
			},
		},
	}
}

func (d *DBaaSDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DBaaSDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DBaaSDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Name = types.StringValue("example-dbaas")
	tflog.Trace(ctx, "read a DBaaS data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
