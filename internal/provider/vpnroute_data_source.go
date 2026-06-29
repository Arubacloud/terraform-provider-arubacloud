package provider

import (
	"context"
	"fmt"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
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
	Uri         types.String `tfsdk:"uri"`
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
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI. Use this value in `*_uri_ref` attributes of other resources.",
				Computed:            true,
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

	route, err := d.client.Client.FromNetwork().VPNRoutes().Get(ctx,
		aruba.VPNRouteRef(projectID, vpnTunnelID, routeID))
	if provErr := CheckResponseErr("read", "VPNRoute", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(route.ID())
	data.Uri = strVal(route.URI())
	data.Name = types.StringValue(route.Name())
	data.ProjectId = types.StringValue(projectID)
	data.VpnTunnelId = types.StringValue(vpnTunnelID)
	data.Destination = types.StringValue(route.CloudSubnet())
	data.Gateway = types.StringValue(route.OnPremSubnet())

	tflog.Trace(ctx, "read a VPN Route data source", map[string]interface{}{"route_id": routeID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
