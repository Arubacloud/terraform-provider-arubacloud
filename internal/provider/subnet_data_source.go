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

var _ datasource.DataSource = &SubnetDataSource{}

func NewSubnetDataSource() datasource.DataSource {
	return &SubnetDataSource{}
}

type SubnetDataSource struct {
	client *ArubaCloudClient
}

type SubnetDataSourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	Tags      types.List   `tfsdk:"tags"`
	ProjectId types.String `tfsdk:"project_id"`
	VpcId     types.String `tfsdk:"vpc_id"`
	Type      types.String `tfsdk:"type"`
	// Network fields (flattened)
	Address types.String `tfsdk:"address"`
	// DHCP fields (flattened)
	DhcpEnabled    types.Bool   `tfsdk:"dhcp_enabled"`
	DhcpRangeStart types.String `tfsdk:"dhcp_range_start"`
	DhcpRangeCount types.Int64  `tfsdk:"dhcp_range_count"`
	DhcpRoutes     types.List   `tfsdk:"dhcp_routes"`
	Dns            types.List   `tfsdk:"dns"`
}

type RouteDataSourceModel struct {
	Address types.String `tfsdk:"address"`
	Gateway types.String `tfsdk:"gateway"`
}

func (d *SubnetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (d *SubnetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Subnet data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Subnet identifier",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Subnet URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Subnet name",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Subnet location",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the subnet",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this subnet belongs to",
				Computed:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPC this subnet belongs to",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Subnet type (Basic or Advanced)",
				Computed:            true,
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "Address of the network in CIDR notation (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)",
				Computed:            true,
			},
			"dhcp_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable DHCP",
				Computed:            true,
			},
			"dhcp_range_start": schema.StringAttribute{
				MarkdownDescription: "Starting IP address for DHCP range",
				Computed:            true,
			},
			"dhcp_range_count": schema.Int64Attribute{
				MarkdownDescription: "Number of available IP addresses in DHCP range",
				Computed:            true,
			},
			"dhcp_routes": schema.ListNestedAttribute{
				MarkdownDescription: "DHCP routes configuration",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							MarkdownDescription: "Destination network address in CIDR notation",
							Computed:            true,
						},
						"gateway": schema.StringAttribute{
							MarkdownDescription: "Gateway IP address",
							Computed:            true,
						},
					},
				},
			},
			"dns": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of DNS IP addresses",
				Computed:            true,
			},
		},
	}
}

func (d *SubnetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SubnetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubnetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate all fields with example data
	data.Uri = types.StringValue("/v2/subnets/subnet-68398923fb2cb026400d4d32")
	data.Name = types.StringValue("example-subnet")
	data.Location = types.StringValue("ITBG-Bergamo")
	data.ProjectId = types.StringValue("68398923fb2cb026400d4d31")
	data.VpcId = types.StringValue("vpc-68398923fb2cb026400d4d32")
	data.Type = types.StringValue("Advanced")
	data.Address = types.StringValue("10.0.1.0/24")
	data.DhcpEnabled = types.BoolValue(true)
	data.DhcpRangeStart = types.StringValue("10.0.1.10")
	data.DhcpRangeCount = types.Int64Value(200)

	// Create routes list
	routeType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"address": types.StringType,
			"gateway": types.StringType,
		},
	}
	route1, _ := types.ObjectValue(routeType.AttrTypes, map[string]attr.Value{
		"address": types.StringValue("0.0.0.0/0"),
		"gateway": types.StringValue("10.0.1.1"),
	})
	route2, _ := types.ObjectValue(routeType.AttrTypes, map[string]attr.Value{
		"address": types.StringValue("192.168.0.0/16"),
		"gateway": types.StringValue("10.0.1.254"),
	})
	data.DhcpRoutes = types.ListValueMust(routeType, []attr.Value{route1, route2})

	data.Dns = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("8.8.8.8"),
		types.StringValue("8.8.4.4"),
	})
	data.Tags = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("network"),
		types.StringValue("private"),
	})

	tflog.Trace(ctx, "read a Subnet data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
