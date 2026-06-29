package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &VPNTunnelResource{}
var _ resource.ResourceWithImportState = &VPNTunnelResource{}

func NewVPNTunnelResource() resource.Resource {
	return &VPNTunnelResource{}
}

type VPNTunnelResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Uri        types.String `tfsdk:"uri"`
	Name       types.String `tfsdk:"name"`
	Location   types.String `tfsdk:"location"`
	Tags       types.List   `tfsdk:"tags"`
	ProjectId  types.String `tfsdk:"project_id"`
	Properties types.Object `tfsdk:"properties"`
}

type VPNTunnelResource struct {
	client *ArubaCloudClient
}

func (r *VPNTunnelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpntunnel"
}

func (r *VPNTunnelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an ArubaCloud IPSec VPN Tunnel — a site-to-site connection between an ArubaCloud VPC and an external on-premises or cloud network. The tunnel is established using a pre-shared key. Use `arubacloud_vpnroute` resources to configure which CIDRs are routed over the tunnel.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources (e.g., as a `*_uri_ref` attribute).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the VPN tunnel.",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier for the resource (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"properties": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration properties for the VPN tunnel.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"vpn_type": schema.StringAttribute{
						MarkdownDescription: "Type of VPN tunnel. Accepted values: `Site-To-Site`.",
						Optional:            true,
					},
					"vpn_client_protocol": schema.StringAttribute{
						MarkdownDescription: "IKE protocol version. Accepted values: `ikev2`.",
						Optional:            true,
					},
					"billing_period": schema.StringAttribute{
						MarkdownDescription: "Billing cycle for the resource. Accepted values: `Hour`, `Month`, `Year`.",
						Optional:            true,
						Validators:          []validator.String{stringvalidator.OneOf("Hour", "Month", "Year")},
					},
					"ip_configurations": schema.SingleNestedAttribute{
						MarkdownDescription: "Network references for the VPN tunnel — VPC, subnet, and public IP.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"vpc": schema.SingleNestedAttribute{
								MarkdownDescription: "Reference to the VPC this tunnel is attached to.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										MarkdownDescription: "ID of the VPC.",
										Optional:            true,
									},
								},
							},
							"subnet": schema.SingleNestedAttribute{
								MarkdownDescription: "Reference to the subnet used by this tunnel.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										MarkdownDescription: "ID of the subnet.",
										Optional:            true,
									},
								},
							},
							"public_ip": schema.SingleNestedAttribute{
								MarkdownDescription: "Reference to the elastic public IP assigned to this tunnel.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										MarkdownDescription: "ID of the elastic public IP.",
										Optional:            true,
									},
								},
							},
						},
					},
					"vpn_client_settings": schema.SingleNestedAttribute{
						MarkdownDescription: "IKE/ESP/PSK client settings for the VPN tunnel.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"ike": schema.SingleNestedAttribute{
								MarkdownDescription: "IKE (Internet Key Exchange) phase-1 settings.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"lifetime": schema.Int64Attribute{
										MarkdownDescription: "IKE phase-1 lifetime in seconds.",
										Optional:            true,
									},
									"encryption": schema.StringAttribute{
										MarkdownDescription: "IKE encryption algorithm (e.g., `aes256`).",
										Optional:            true,
									},
									"hash": schema.StringAttribute{
										MarkdownDescription: "IKE integrity/hash algorithm (e.g., `sha256`).",
										Optional:            true,
									},
									"dh_group": schema.StringAttribute{
										MarkdownDescription: "IKE Diffie-Hellman group (e.g., `modp2048`).",
										Optional:            true,
									},
									"dpd_action": schema.StringAttribute{
										MarkdownDescription: "Dead Peer Detection action on failure (e.g., `restart`).",
										Optional:            true,
									},
									"dpd_interval": schema.Int64Attribute{
										MarkdownDescription: "DPD keep-alive interval in seconds.",
										Optional:            true,
									},
									"dpd_timeout": schema.Int64Attribute{
										MarkdownDescription: "DPD timeout before the peer is considered dead, in seconds.",
										Optional:            true,
									},
								},
							},
							"esp": schema.SingleNestedAttribute{
								MarkdownDescription: "ESP (Encapsulating Security Payload) phase-2 settings.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"lifetime": schema.Int64Attribute{
										MarkdownDescription: "ESP phase-2 lifetime in seconds.",
										Optional:            true,
									},
									"encryption": schema.StringAttribute{
										MarkdownDescription: "ESP encryption algorithm (e.g., `aes256`).",
										Optional:            true,
									},
									"hash": schema.StringAttribute{
										MarkdownDescription: "ESP integrity/hash algorithm (e.g., `sha256`).",
										Optional:            true,
									},
									"pfs": schema.StringAttribute{
										MarkdownDescription: "ESP Perfect Forward Secrecy group (e.g., `modp2048`).",
										Optional:            true,
									},
								},
							},
							"psk": schema.SingleNestedAttribute{
								MarkdownDescription: "Pre-Shared Key (PSK) authentication settings.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"cloud_site": schema.StringAttribute{
										MarkdownDescription: "Pre-shared key for the ArubaCloud side of the tunnel.",
										Optional:            true,
									},
									"on_prem_site": schema.StringAttribute{
										MarkdownDescription: "Pre-shared key for the on-premises side of the tunnel.",
										Optional:            true,
									},
									"secret": schema.StringAttribute{
										MarkdownDescription: "Shared secret used to authenticate the VPN tunnel. Write-only — this value is sent to the API but is not returned in subsequent read responses.",
										Optional:            true,
									},
								},
							},
							"peer_client_public_ip": schema.StringAttribute{
								MarkdownDescription: "Public IP address of the remote peer (on-premises gateway).",
								Optional:            true,
							},
						},
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
	client, ok := req.ProviderData.(*ArubaCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ArubaCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func vpnTunnelRef(data *VPNTunnelResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.VPNTunnelRef(data.ProjectId.ValueString(), data.Id.ValueString())
}

// buildVPNTunnel constructs a *aruba.VPNTunnel builder from model properties.
func buildVPNTunnel(_ context.Context, data *VPNTunnelResourceModel, tags []string) *aruba.VPNTunnel {
	projectID := data.ProjectId.ValueString()

	builder := aruba.NewVPNTunnel().
		Named(data.Name.ValueString()).
		InProject(aruba.URI("/projects/" + projectID)).
		InRegion(aruba.Region(data.Location.ValueString())).
		Tagged(tags...)

	if data.Properties.IsNull() || data.Properties.IsUnknown() {
		return builder
	}
	attrs := data.Properties.Attributes()

	if v, ok := attrs["vpn_type"]; ok {
		if s, ok := v.(types.String); ok && !s.IsNull() {
			builder = builder.OfType(aruba.VPNType(s.ValueString()))
		}
	}
	if v, ok := attrs["vpn_client_protocol"]; ok {
		if s, ok := v.(types.String); ok && !s.IsNull() {
			builder = builder.WithVPNClientProtocol(aruba.VPNClientProtocol(s.ValueString()))
		}
	}
	if v, ok := attrs["billing_period"]; ok {
		if s, ok := v.(types.String); ok && !s.IsNull() {
			builder = builder.BilledBy(aruba.BillingPeriod(s.ValueString()))
		}
	}

	// IP configurations
	if v, ok := attrs["ip_configurations"]; ok {
		if ipCfgObj, ok := v.(types.Object); ok && !ipCfgObj.IsNull() {
			ipCfgAttrs := ipCfgObj.Attributes()
			ipCfg := aruba.NewVPNIPConfig()

			if vpcAttr, ok := ipCfgAttrs["vpc"]; ok {
				if vpcObj, ok := vpcAttr.(types.Object); ok && !vpcObj.IsNull() {
					if idAttr, ok := vpcObj.Attributes()["id"]; ok {
						if s, ok := idAttr.(types.String); ok && !s.IsNull() {
							vpcID := s.ValueString()
							if !strings.HasPrefix(vpcID, "/") {
								vpcID = fmt.Sprintf("/projects/%s/providers/Aruba.Network/vpcs/%s", projectID, vpcID)
							}
							ipCfg.WithVPC(aruba.URI(vpcID))
						}
					}
				}
			}
			if subnetAttr, ok := ipCfgAttrs["subnet"]; ok {
				if subnetObj, ok := subnetAttr.(types.Object); ok && !subnetObj.IsNull() {
					if idAttr, ok := subnetObj.Attributes()["id"]; ok {
						if s, ok := idAttr.(types.String); ok && !s.IsNull() {
							// subnet name and CIDR — we use the id as name, CIDR left empty
							ipCfg.WithSubnet(s.ValueString(), "")
						}
					}
				}
			}
			if pubIPAttr, ok := ipCfgAttrs["public_ip"]; ok {
				if pubIPObj, ok := pubIPAttr.(types.Object); ok && !pubIPObj.IsNull() {
					if idAttr, ok := pubIPObj.Attributes()["id"]; ok {
						if s, ok := idAttr.(types.String); ok && !s.IsNull() {
							eipID := s.ValueString()
							if !strings.HasPrefix(eipID, "/") {
								eipID = fmt.Sprintf("/projects/%s/providers/Aruba.Network/elasticips/%s", projectID, eipID)
							}
							ipCfg.WithElasticIP(aruba.URI(eipID))
						}
					}
				}
			}
			builder = builder.WithIPConfig(ipCfg)
		}
	}

	// VPN client settings
	if v, ok := attrs["vpn_client_settings"]; ok {
		if clientObj, ok := v.(types.Object); ok && !clientObj.IsNull() {
			clientAttrs := clientObj.Attributes()

			if v, ok := clientAttrs["peer_client_public_ip"]; ok {
				if s, ok := v.(types.String); ok && !s.IsNull() {
					builder = builder.WithPeerClientPublicIP(s.ValueString())
				}
			}

			if ikeAttr, ok := clientAttrs["ike"]; ok {
				if ikeObj, ok := ikeAttr.(types.Object); ok && !ikeObj.IsNull() {
					ike := aruba.NewVPNIKE()
					ikeAttrs := ikeObj.Attributes()
					if v, ok := ikeAttrs["lifetime"]; ok {
						if n, ok := v.(types.Int64); ok && !n.IsNull() {
							ike.WithLifetimeSeconds(int(n.ValueInt64()))
						}
					}
					if v, ok := ikeAttrs["encryption"]; ok {
						if s, ok := v.(types.String); ok && !s.IsNull() {
							ike.WithEncryption(aruba.IKEEncryption(s.ValueString()))
						}
					}
					if v, ok := ikeAttrs["hash"]; ok {
						if s, ok := v.(types.String); ok && !s.IsNull() {
							ike.WithHash(aruba.IKEHash(s.ValueString()))
						}
					}
					if v, ok := ikeAttrs["dh_group"]; ok {
						if s, ok := v.(types.String); ok && !s.IsNull() {
							ike.WithDHGroup(aruba.IKEDHGroup(s.ValueString()))
						}
					}
					if v, ok := ikeAttrs["dpd_action"]; ok {
						if s, ok := v.(types.String); ok && !s.IsNull() {
							ike.WithDPDAction(aruba.IKEDPDAction(s.ValueString()))
						}
					}
					if v, ok := ikeAttrs["dpd_interval"]; ok {
						if n, ok := v.(types.Int64); ok && !n.IsNull() {
							ike.WithDPDIntervalSeconds(int(n.ValueInt64()))
						}
					}
					if v, ok := ikeAttrs["dpd_timeout"]; ok {
						if n, ok := v.(types.Int64); ok && !n.IsNull() {
							ike.WithDPDTimeoutSeconds(int(n.ValueInt64()))
						}
					}
					builder = builder.WithIKESettings(ike)
				}
			}

			if espAttr, ok := clientAttrs["esp"]; ok {
				if espObj, ok := espAttr.(types.Object); ok && !espObj.IsNull() {
					esp := aruba.NewVPNESP()
					espAttrs := espObj.Attributes()
					if v, ok := espAttrs["lifetime"]; ok {
						if n, ok := v.(types.Int64); ok && !n.IsNull() {
							esp.WithLifetimeSeconds(int(n.ValueInt64()))
						}
					}
					if v, ok := espAttrs["encryption"]; ok {
						if s, ok := v.(types.String); ok && !s.IsNull() {
							esp.WithEncryption(aruba.ESPEncryption(s.ValueString()))
						}
					}
					if v, ok := espAttrs["hash"]; ok {
						if s, ok := v.(types.String); ok && !s.IsNull() {
							esp.WithHash(aruba.ESPHash(s.ValueString()))
						}
					}
					if v, ok := espAttrs["pfs"]; ok {
						if s, ok := v.(types.String); ok && !s.IsNull() {
							esp.WithPFS(aruba.ESPPFSGroup(s.ValueString()))
						}
					}
					builder = builder.WithESPSettings(esp)
				}
			}

			if pskAttr, ok := clientAttrs["psk"]; ok {
				if pskObj, ok := pskAttr.(types.Object); ok && !pskObj.IsNull() {
					psk := aruba.NewVPNPSK()
					pskAttrs := pskObj.Attributes()
					if v, ok := pskAttrs["cloud_site"]; ok {
						if s, ok := v.(types.String); ok && !s.IsNull() {
							psk.WithCloudSite(s.ValueString())
						}
					}
					if v, ok := pskAttrs["on_prem_site"]; ok {
						if s, ok := v.(types.String); ok && !s.IsNull() {
							psk.WithOnPremSite(s.ValueString())
						}
					}
					if v, ok := pskAttrs["secret"]; ok {
						if s, ok := v.(types.String); ok && !s.IsNull() {
							psk.WithKey(s.ValueString())
						}
					}
					builder = builder.WithPSKSettings(psk)
				}
			}
		}
	}

	return builder
}

func (r *VPNTunnelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VPNTunnelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tunnel, err := r.client.Client.FromNetwork().VPNTunnels().Create(ctx, buildVPNTunnel(ctx, &data, tags))
	if provErr := CheckResponseErr("create", "VPNTunnel", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(tunnel.ID())
	data.Uri = strVal(tunnel.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if waitErr := tunnel.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "VPNTunnel", data.Id.ValueString())
		return
	}

	tflog.Trace(ctx, "created a VPN Tunnel resource", map[string]interface{}{
		"vpntunnel_id":   data.Id.ValueString(),
		"vpntunnel_name": data.Name.ValueString(),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPNTunnelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VPNTunnelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	tunnel, err := r.client.Client.FromNetwork().VPNTunnels().Get(ctx, vpnTunnelRef(&data))
	if provErr := CheckResponseErr("read", "VPNTunnel", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	st := string(tunnel.State())
	switch {
	case isFailedState(st):
		resp.Diagnostics.AddWarning("Resource in Failed State",
			fmt.Sprintf("VPNTunnel %q is in a terminal failure state (%s). "+
				"Run `terraform destroy` to clean it up, or `terraform apply -replace=<address>` to recreate it.", data.Id.ValueString(), st))
	case IsCreatingState(st):
		if waitErr := tunnel.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "VPNTunnel", data.Id.ValueString())
			return
		}
		tunnel, err = r.client.Client.FromNetwork().VPNTunnels().Get(ctx, vpnTunnelRef(&data))
		if provErr := CheckResponseErr("read", "VPNTunnel", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	data.Id = types.StringValue(tunnel.ID())
	data.Uri = strVal(tunnel.URI())
	data.Name = types.StringValue(tunnel.Name())
	data.Tags = TagsToListPreserveNull(tunnel.Tags(), data.Tags)
	if tunnel.Region() != "" {
		data.Location = types.StringValue(string(tunnel.Region()))
	}
	// Properties (PSK secrets etc.) are not returned by API — preserve existing state.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPNTunnelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VPNTunnelResourceModel
	var state VPNTunnelResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tunnel, err := r.client.Client.FromNetwork().VPNTunnels().Get(ctx, vpnTunnelRef(&state))
	if provErr := CheckResponseErr("read", "VPNTunnel", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	tunnel.Named(data.Name.ValueString())
	if tags != nil {
		tunnel.RetaggedAs(tags...)
	} else {
		tunnel.RetaggedAs(tunnel.Tags()...)
	}

	updated, err := r.client.Client.FromNetwork().VPNTunnels().Update(ctx, tunnel)
	if provErr := CheckResponseErr("update", "VPNTunnel", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = state.Id
	data.Uri = state.Uri
	data.ProjectId = state.ProjectId
	data.Location = state.Location
	data.Properties = state.Properties // Properties are immutable — preserve from state.
	data.Name = types.StringValue(updated.Name())
	data.Tags = TagsToListPreserveNull(updated.Tags(), data.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPNTunnelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VPNTunnelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := vpnTunnelRef(&data)
	tunnelID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromNetwork().VPNTunnels().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "VPNTunnel", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErrAsError("delete", "VPNTunnel",
			r.client.Client.FromNetwork().VPNTunnels().Delete(ctx, ref))
	}, "VPNTunnel", tunnelID, r.client.ResourceTimeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting VPNTunnel", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "VPNTunnel", tunnelID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for VPNTunnel deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a VPN Tunnel resource", map[string]interface{}{"vpntunnel_id": tunnelID})
}

func (r *VPNTunnelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseImportID(req.ID, "<project_id>/<tunnel_id>", "proj-abc/tun-xyz", 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
