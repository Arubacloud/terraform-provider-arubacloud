package provider

import (
	"context"
	"fmt"

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
		MarkdownDescription: "VPN Tunnel data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VPN Tunnel identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "VPN Tunnel name",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this VPN Tunnel belongs to",
				Required:            true,
			},
			"remote_peer": schema.StringAttribute{
				MarkdownDescription: "Remote peer address for the VPN tunnel",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the VPN tunnel",
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

	response, err := d.client.Client.FromNetwork().VPNTunnels().Get(ctx, projectID, tunnelID, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading VPN tunnel", NewTransportError("read", "Vpntunnel", err).Error())
		return
	}
	if apiErr := CheckResponse("read", "Vpntunnel", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError("No data returned", "VPN Tunnel Get returned no data")
		return
	}

	tunnel := response.Data
	if tunnel.Metadata.ID != nil {
		data.Id = types.StringValue(*tunnel.Metadata.ID)
	}
	if tunnel.Metadata.Name != nil {
		data.Name = types.StringValue(*tunnel.Metadata.Name)
	}
	data.ProjectId = types.StringValue(projectID)
	// remote_peer and status are not directly available in the metadata response
	data.RemotePeer = types.StringNull()
	data.Status = types.StringNull()

	tflog.Trace(ctx, "read a VPN Tunnel data source", map[string]interface{}{"tunnel_id": tunnelID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
