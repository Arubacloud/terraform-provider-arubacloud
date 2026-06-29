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

var _ datasource.DataSource = &VPNTunnelDataSource{}

func NewVPNTunnelDataSource() datasource.DataSource {
	return &VPNTunnelDataSource{}
}

type VPNTunnelDataSource struct {
	client *ArubaCloudClient
}

type VPNTunnelDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	Uri        types.String `tfsdk:"uri"`
	Name       types.String `tfsdk:"name"`
	ProjectId  types.String `tfsdk:"project_id"`
	RemotePeer types.String `tfsdk:"remote_peer"`
	Status     types.String `tfsdk:"status"`
}

func (d *VPNTunnelDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpntunnel"
}

func (d *VPNTunnelDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves read-only information about an existing `arubacloud_vpntunnel`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the VPN tunnel to look up.",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI. Use this value in `*_uri_ref` attributes of other resources.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the VPN tunnel.",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"remote_peer": schema.StringAttribute{
				MarkdownDescription: "Public IP address of the remote peer (on-premises gateway).",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current operational status of the VPN tunnel.",
				Computed:            true,
			},
		},
	}
}

func (d *VPNTunnelDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VPNTunnelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VPNTunnelDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	tunnelID := data.Id.ValueString()
	if projectID == "" || tunnelID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID and VPN Tunnel ID are required to read the VPN tunnel")
		return
	}

	tunnel, err := d.client.Client.FromNetwork().VPNTunnels().Get(ctx,
		aruba.VPNTunnelRef(projectID, tunnelID))
	if provErr := CheckResponseErr("read", "VPNTunnel", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(tunnel.ID())
	data.Uri = strVal(tunnel.URI())
	data.Name = types.StringValue(tunnel.Name())
	data.ProjectId = types.StringValue(projectID)
	// PeerClientPublicIP is the remote peer — exposed via wrapper accessor.
	if peer := tunnel.PeerClientPublicIP(); peer != "" {
		data.RemotePeer = types.StringValue(peer)
	} else {
		data.RemotePeer = types.StringNull()
	}
	if st := string(tunnel.State()); st != "" {
		data.Status = types.StringValue(st)
	} else {
		data.Status = types.StringNull()
	}

	tflog.Trace(ctx, "read a VPN Tunnel data source", map[string]interface{}{"tunnel_id": tunnelID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
