// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &VPNTunnelResource{}
var _ resource.ResourceWithImportState = &VPNTunnelResource{}

func NewVPNTunnelResource() resource.Resource {
	return &VPNTunnelResource{}
}

type ReferenceResource struct {
	Id types.String `tfsdk:"id"`
}

type IPConfigurations struct {
	VPC      *ReferenceResource `tfsdk:"vpc"`
	Subnet   *ReferenceResource `tfsdk:"subnet"`
	PublicIP *ReferenceResource `tfsdk:"public_ip"`
}

type IKESettings struct {
	Lifetime    types.Int64  `tfsdk:"lifetime"`
	Encryption  types.String `tfsdk:"encryption"`
	Hash        types.String `tfsdk:"hash"`
	DHGroup     types.String `tfsdk:"dh_group"`
	DPDAction   types.String `tfsdk:"dpd_action"`
	DPDInterval types.Int64  `tfsdk:"dpd_interval"`
	DPDTimeout  types.Int64  `tfsdk:"dpd_timeout"`
}

type ESPSettings struct {
	Lifetime   types.Int64  `tfsdk:"lifetime"`
	Encryption types.String `tfsdk:"encryption"`
	Hash       types.String `tfsdk:"hash"`
	PFS        types.String `tfsdk:"pfs"`
}

type PSKSettings struct {
	CloudSite  types.String `tfsdk:"cloud_site"`
	OnPremSite types.String `tfsdk:"on_prem_site"`
	Secret     types.String `tfsdk:"secret"`
}

type VPNClientSettings struct {
	IKE *IKESettings `tfsdk:"ike"`
	ESP *ESPSettings `tfsdk:"esp"`
	PSK *PSKSettings `tfsdk:"psk"`
}

type VPNTunnelPropertiesRequest struct {
	VPNType            types.String       `tfsdk:"vpn_type"`
	VPNClientProtocol  types.String       `tfsdk:"vpn_client_protocol"`
	IPConfigurations   *IPConfigurations  `tfsdk:"ip_configurations"`
	VPNClientSettings  *VPNClientSettings `tfsdk:"vpn_client_settings"`
	PeerClientPublicIP types.String       `tfsdk:"peer_client_public_ip"`
}

type VPNTunnelResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Location   types.String `tfsdk:"location"`
	Tags       types.List   `tfsdk:"tags"`
	ProjectId  types.String `tfsdk:"project_id"`
	Properties types.Object `tfsdk:"properties"`
}

type VPNTunnelResource struct {
	client *http.Client
}

func (r *VPNTunnelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpntunnel"
}

func (r *VPNTunnelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VPN Tunnel resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VPN Tunnel identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "VPN Tunnel name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "VPN Tunnel location",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the VPN Tunnel",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this VPN Tunnel belongs to",
				Required:            true,
			},
			"properties": schema.SingleNestedAttribute{
				MarkdownDescription: "Properties of the VPN Tunnel",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"vpn_type": schema.StringAttribute{
						MarkdownDescription: "Type of VPN tunnel (Site-To-Site)",
						Optional:            true,
					},
					"vpn_client_protocol": schema.StringAttribute{
						MarkdownDescription: "Protocol of the VPN tunnel (ikev2)",
						Optional:            true,
					},
					"ip_configurations": schema.SingleNestedAttribute{
						MarkdownDescription: "Network configuration of the VPN tunnel",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"vpc": schema.SingleNestedAttribute{
								MarkdownDescription: "VPC reference",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										MarkdownDescription: "VPC id",
										Optional:            true,
									},
								},
							},
							"subnet": schema.SingleNestedAttribute{
								MarkdownDescription: "Subnet reference",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										MarkdownDescription: "Subnet id",
										Optional:            true,
									},
								},
							},
							"public_ip": schema.SingleNestedAttribute{
								MarkdownDescription: "Public IP reference",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										MarkdownDescription: "Public IP id",
										Optional:            true,
									},
								},
							},
						},
					},
					"vpn_client_settings": schema.SingleNestedAttribute{
						MarkdownDescription: "Client settings of the VPN tunnel",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"ike": schema.SingleNestedAttribute{
								MarkdownDescription: "IKE settings",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"lifetime": schema.Int64Attribute{
										MarkdownDescription: "IKE lifetime",
										Optional:            true,
									},
									"encryption": schema.StringAttribute{
										MarkdownDescription: "IKE encryption algorithm",
										Optional:            true,
									},
									"hash": schema.StringAttribute{
										MarkdownDescription: "IKE hash algorithm",
										Optional:            true,
									},
									"dh_group": schema.StringAttribute{
										MarkdownDescription: "IKE DH group",
										Optional:            true,
									},
									"dpd_action": schema.StringAttribute{
										MarkdownDescription: "IKE DPD action",
										Optional:            true,
									},
									"dpd_interval": schema.Int64Attribute{
										MarkdownDescription: "IKE DPD interval",
										Optional:            true,
									},
									"dpd_timeout": schema.Int64Attribute{
										MarkdownDescription: "IKE DPD timeout",
										Optional:            true,
									},
								},
							},
							"esp": schema.SingleNestedAttribute{
								MarkdownDescription: "ESP settings",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"lifetime": schema.Int64Attribute{
										MarkdownDescription: "ESP lifetime",
										Optional:            true,
									},
									"encryption": schema.StringAttribute{
										MarkdownDescription: "ESP encryption algorithm",
										Optional:            true,
									},
									"hash": schema.StringAttribute{
										MarkdownDescription: "ESP hash algorithm",
										Optional:            true,
									},
									"pfs": schema.StringAttribute{
										MarkdownDescription: "ESP PFS",
										Optional:            true,
									},
								},
							},
							"psk": schema.SingleNestedAttribute{
								MarkdownDescription: "PSK settings",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"cloud_site": schema.StringAttribute{
										MarkdownDescription: "PSK cloud site",
										Optional:            true,
									},
									"on_prem_site": schema.StringAttribute{
										MarkdownDescription: "PSK on-prem site",
										Optional:            true,
									},
									"secret": schema.StringAttribute{
										MarkdownDescription: "PSK secret",
										Optional:            true,
									},
								},
							},
						},
					},
					"peer_client_public_ip": schema.StringAttribute{
						MarkdownDescription: "Peer client public IP address",
						Optional:            true,
					},
				},
			},
		},
	}
}

func (r *VPNTunnelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*http.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *VPNTunnelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VPNTunnelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Simulate API response
	data.Id = types.StringValue("vpntunnel-id")
	tflog.Trace(ctx, "created a VPN Tunnel resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPNTunnelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VPNTunnelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPNTunnelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VPNTunnelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPNTunnelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VPNTunnelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *VPNTunnelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
