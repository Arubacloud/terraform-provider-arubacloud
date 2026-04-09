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
		MarkdownDescription: "VPN Route data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VPN Route identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "VPN Route name",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this VPN Route belongs to",
				Required:            true,
			},
			"vpn_tunnel_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPN Tunnel this route belongs to",
				Required:            true,
			},
			"destination": schema.StringAttribute{
				MarkdownDescription: "Destination network for the VPN route (CloudSubnet)",
				Computed:            true,
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "Gateway for the VPN route (OnPremSubnet)",
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
		resp.Diagnostics.AddError("Error reading VPN route", fmt.Sprintf("Unable to read VPN route: %s", err))
		return
	}
	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.Diagnostics.AddError("VPN Route not found", fmt.Sprintf("No VPN route found with ID %q in tunnel %q", routeID, vpnTunnelID))
			return
		}
		resp.Diagnostics.AddError("API Error", FormatAPIError(ctx, response.Error, "Failed to read VPN route", map[string]interface{}{"project_id": projectID, "vpn_tunnel_id": vpnTunnelID, "route_id": routeID}))
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
