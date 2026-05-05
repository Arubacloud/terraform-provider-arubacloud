package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &VPNRouteDataSource{}

func NewVPNRouteDataSource() datasource.DataSource {
	return &VPNRouteDataSource{}
}

type VPNRouteDataSource struct {
	client *ArubaCloudClient
}

type VPNRouteDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ProjectId   types.String `tfsdk:"project_id"`
	VpnTunnelId types.String `tfsdk:"vpn_tunnel_id"`
	Destination types.String `tfsdk:"destination"`
	Gateway     types.String `tfsdk:"gateway"`
}

func (d *VPNRouteDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpnroute"
}

func (d *VPNRouteDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves read-only information about an existing `arubacloud_vpnroute`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the VPN route to look up.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the VPN route.",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"vpn_tunnel_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPN tunnel this route is associated with.",
				Required:            true,
			},
			"destination": schema.StringAttribute{
				MarkdownDescription: "CIDR of the ArubaCloud-side subnet routed over this tunnel (maps to `cloud_subnet`).",
				Computed:            true,
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "CIDR of the on-premises subnet reachable through this tunnel (maps to `on_prem_subnet`).",
				Computed:            true,
			},
		},
	}
}

func (d *VPNRouteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VPNRouteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VPNRouteDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpnTunnelID := data.VpnTunnelId.ValueString()
	routeID := data.Id.ValueString()
	if projectID == "" || vpnTunnelID == "" || routeID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID, VPN Tunnel ID, and Route ID are required to read the VPN route")
		return
	}

	response, err := d.client.Client.FromNetwork().VPNRoutes().Get(ctx, projectID, vpnTunnelID, routeID, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading VPN route", NewTransportError("read", "Vpnroute", err).Error())
		return
	}
	if apiErr := CheckResponse("read", "Vpnroute", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError("No data returned", "VPN Route Get returned no data")
		return
	}

	route := response.Data
	if route.Metadata.ID != nil {
		data.Id = types.StringValue(*route.Metadata.ID)
	}
	if route.Metadata.Name != nil {
		data.Name = types.StringValue(*route.Metadata.Name)
	}
	data.ProjectId = types.StringValue(projectID)
	data.VpnTunnelId = types.StringValue(vpnTunnelID)
	data.Destination = types.StringValue(route.Properties.CloudSubnet)
	data.Gateway = types.StringValue(route.Properties.OnPremSubnet)

	tflog.Trace(ctx, "read a VPN Route data source", map[string]interface{}{"route_id": routeID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
