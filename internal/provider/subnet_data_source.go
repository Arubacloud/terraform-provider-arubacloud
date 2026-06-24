package provider

import (
	"context"
	"fmt"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
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
		MarkdownDescription: "Retrieves read-only information about an existing `arubacloud_subnet`. Use this data source to reference a subnet's URI when attaching a CloudServer to a subnet managed in a separate configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the subnet to look up.",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources (e.g., as a `*_uri_ref` attribute).",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the subnet.",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier for the resource (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the parent VPC this subnet belongs to.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Subnet type. Accepted values: `Basic` (no custom CIDR), `Advanced` (requires the `network` block).",
				Computed:            true,
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "Subnet CIDR in RFC-1918 notation (e.g., `10.0.1.0/24`). Must fall within the parent VPC CIDR.",
				Computed:            true,
			},
			"dhcp_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether DHCP is enabled on this subnet.",
				Computed:            true,
			},
			"dhcp_range_start": schema.StringAttribute{
				MarkdownDescription: "First IP address in the DHCP allocation range.",
				Computed:            true,
			},
			"dhcp_range_count": schema.Int64Attribute{
				MarkdownDescription: "Number of consecutive IP addresses in the DHCP pool.",
				Computed:            true,
			},
			"dhcp_routes": schema.ListNestedAttribute{
				MarkdownDescription: "Static routes distributed to DHCP clients.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							MarkdownDescription: "Destination network in CIDR notation (e.g., `0.0.0.0/0` for a default route).",
							Computed:            true,
						},
						"gateway": schema.StringAttribute{
							MarkdownDescription: "Gateway IP address for this route.",
							Computed:            true,
						},
					},
				},
			},
			"dns": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of DNS server IP addresses distributed to DHCP clients.",
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
			fmt.Sprintf("Expected *ArubaCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	subnetID := data.Id.ValueString()
	if projectID == "" || vpcID == "" || subnetID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID, VPC ID, and Subnet ID are required to read the subnet")
		return
	}

	subnet, err := d.client.Client.FromNetwork().Subnets().Get(ctx,
		aruba.SubnetRef(projectID, vpcID, subnetID))
	if provErr := CheckResponseErr("read", "Subnet", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(subnet.ID())
	data.Uri = strVal(subnet.URI())
	data.Name = types.StringValue(subnet.Name())
	if subnet.Region() != "" {
		data.Location = types.StringValue(string(subnet.Region()))
	} else {
		data.Location = types.StringNull()
	}
	data.ProjectId = types.StringValue(projectID)
	data.VpcId = types.StringValue(vpcID)
	data.Type = types.StringValue(string(subnet.Type()))

	if cidr := subnet.CIDR(); cidr != "" {
		data.Address = types.StringValue(cidr)
	} else {
		data.Address = types.StringNull()
	}

	routeObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"address": types.StringType,
			"gateway": types.StringType,
		},
	}

	dhcp := subnet.DHCP()
	if dhcp != nil {
		data.DhcpEnabled = types.BoolValue(dhcp.IsEnabled())

		if dhcp.RangeStart() != "" || dhcp.RangeCount() > 0 {
			data.DhcpRangeStart = types.StringValue(dhcp.RangeStart())
			data.DhcpRangeCount = types.Int64Value(int64(dhcp.RangeCount()))
		} else {
			data.DhcpRangeStart = types.StringNull()
			data.DhcpRangeCount = types.Int64Null()
		}

		routes := dhcp.Routes()
		if len(routes) > 0 {
			routeObjs := make([]attr.Value, len(routes))
			for i, route := range routes {
				routeObj, d := types.ObjectValue(routeObjType.AttrTypes, map[string]attr.Value{
					"address": types.StringValue(route.Address),
					"gateway": types.StringValue(route.Gateway),
				})
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				routeObjs[i] = routeObj
			}
			data.DhcpRoutes = types.ListValueMust(routeObjType, routeObjs)
		} else {
			data.DhcpRoutes = types.ListValueMust(routeObjType, []attr.Value{})
		}

		dns := dhcp.DNS()
		if len(dns) > 0 {
			dnsValues := make([]attr.Value, len(dns))
			for i, s := range dns {
				dnsValues[i] = types.StringValue(s)
			}
			data.Dns = types.ListValueMust(types.StringType, dnsValues)
		} else {
			data.Dns = types.ListValueMust(types.StringType, []attr.Value{})
		}
	} else {
		data.DhcpEnabled = types.BoolNull()
		data.DhcpRangeStart = types.StringNull()
		data.DhcpRangeCount = types.Int64Null()
		data.DhcpRoutes = types.ListValueMust(routeObjType, []attr.Value{})
		data.Dns = types.ListValueMust(types.StringType, []attr.Value{})
	}

	data.Tags = TagsToListPreserveNull(subnet.Tags(), data.Tags)

	tflog.Trace(ctx, "read a Subnet data source", map[string]interface{}{"subnet_id": subnetID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
