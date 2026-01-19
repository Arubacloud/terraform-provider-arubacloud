// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
		MarkdownDescription: "VPN Tunnel resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VPN Tunnel identifier",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "VPN Tunnel URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "VPN Tunnel name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "VPN Tunnel location",
				Required:            true,
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
					"billing_period": schema.StringAttribute{
						MarkdownDescription: "Billing period (Hour, Month, Year)",
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
							"peer_client_public_ip": schema.StringAttribute{
								MarkdownDescription: "Peer client public IP address",
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

func (r *VPNTunnelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VPNTunnelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"Project ID is required to create a VPN tunnel",
		)
		return
	}

	// Extract tags
	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Extract properties from Terraform object
	propertiesObj, diags := data.Properties.ToObjectValue(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	propertiesAttrs := propertiesObj.Attributes()

	// Extract VPN type and protocol
	vpnType := "Site-To-Site"
	if vpnTypeAttr, ok := propertiesAttrs["vpn_type"]; ok && vpnTypeAttr != nil {
		if vpnTypeStr, ok := vpnTypeAttr.(types.String); ok && !vpnTypeStr.IsNull() {
			vpnType = vpnTypeStr.ValueString()
		}
	}

	protocol := "ikev2"
	if protocolAttr, ok := propertiesAttrs["vpn_client_protocol"]; ok && protocolAttr != nil {
		if protocolStr, ok := protocolAttr.(types.String); ok && !protocolStr.IsNull() {
			protocol = protocolStr.ValueString()
		}
	}

	billingPeriod := "Hour"
	if billingAttr, ok := propertiesAttrs["billing_period"]; ok && billingAttr != nil {
		if billingStr, ok := billingAttr.(types.String); ok && !billingStr.IsNull() {
			billingPeriod = billingStr.ValueString()
		}
	}

	// Extract IP configurations
	var ipConfig *sdktypes.IPConfigurations
	if ipConfigAttr, ok := propertiesAttrs["ip_configurations"]; ok && ipConfigAttr != nil {
		if ipConfigObj, ok := ipConfigAttr.(types.Object); ok && !ipConfigObj.IsNull() {
			ipConfigObjValue, diags := ipConfigObj.ToObjectValue(ctx)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				ipConfigAttrs := ipConfigObjValue.Attributes()
				ipConfig = &sdktypes.IPConfigurations{}

				// Extract VPC
				if vpcAttr, ok := ipConfigAttrs["vpc"]; ok && vpcAttr != nil {
					if vpcObj, ok := vpcAttr.(types.Object); ok && !vpcObj.IsNull() {
						vpcObjValue, diags := vpcObj.ToObjectValue(ctx)
						resp.Diagnostics.Append(diags...)
						if !resp.Diagnostics.HasError() {
							vpcAttrs := vpcObjValue.Attributes()
							if vpcIDAttr, ok := vpcAttrs["id"]; ok && vpcIDAttr != nil {
								if vpcIDStr, ok := vpcIDAttr.(types.String); ok && !vpcIDStr.IsNull() {
									vpcID := vpcIDStr.ValueString()
									if !strings.HasPrefix(vpcID, "/") {
										vpcID = fmt.Sprintf("/projects/%s/providers/Aruba.Network/vpcs/%s", projectID, vpcID)
									}
									ipConfig.VPC = &sdktypes.ReferenceResource{URI: vpcID}
								}
							}
						}
					}
				}

				// Extract Subnet
				if subnetAttr, ok := ipConfigAttrs["subnet"]; ok && subnetAttr != nil {
					if subnetObj, ok := subnetAttr.(types.Object); ok && !subnetObj.IsNull() {
						subnetObjValue, diags := subnetObj.ToObjectValue(ctx)
						resp.Diagnostics.Append(diags...)
						if !resp.Diagnostics.HasError() {
							subnetAttrs := subnetObjValue.Attributes()
							if subnetIDAttr, ok := subnetAttrs["id"]; ok && subnetIDAttr != nil {
								if subnetIDStr, ok := subnetIDAttr.(types.String); ok && !subnetIDStr.IsNull() {
									subnetID := subnetIDStr.ValueString()
									if !strings.HasPrefix(subnetID, "/") {
										// Need VPC ID for subnet URI - try to get from VPC if available
										if ipConfig.VPC != nil {
											// Extract VPC ID from URI
											vpcURI := ipConfig.VPC.URI
											parts := strings.Split(vpcURI, "/")
											if len(parts) > 0 {
												vpcID := parts[len(parts)-1]
												subnetID = fmt.Sprintf("/projects/%s/providers/Aruba.Network/vpcs/%s/subnets/%s", projectID, vpcID, subnetID)
											}
										}
									}
									// Subnet field expects SubnetInfo, which has CIDR and Name fields
									// Extract subnet name from URI or use the ID as name
									subnetName := subnetID
									if strings.Contains(subnetID, "/") {
										parts := strings.Split(subnetID, "/")
										if len(parts) > 0 {
											subnetName = parts[len(parts)-1]
										}
									}
									ipConfig.Subnet = &sdktypes.SubnetInfo{
										Name: subnetName,
										// CIDR can be empty if not provided
									}
								}
							}
						}
					}
				}

				// Extract Public IP
				if publicIPAttr, ok := ipConfigAttrs["public_ip"]; ok && publicIPAttr != nil {
					if publicIPObj, ok := publicIPAttr.(types.Object); ok && !publicIPObj.IsNull() {
						publicIPObjValue, diags := publicIPObj.ToObjectValue(ctx)
						resp.Diagnostics.Append(diags...)
						if !resp.Diagnostics.HasError() {
							publicIPAttrs := publicIPObjValue.Attributes()
							if publicIPIDAttr, ok := publicIPAttrs["id"]; ok && publicIPIDAttr != nil {
								if publicIPIDStr, ok := publicIPIDAttr.(types.String); ok && !publicIPIDStr.IsNull() {
									publicIPID := publicIPIDStr.ValueString()
									if !strings.HasPrefix(publicIPID, "/") {
										publicIPID = fmt.Sprintf("/projects/%s/providers/Aruba.Network/elasticips/%s", projectID, publicIPID)
									}
									ipConfig.PublicIP = &sdktypes.ReferenceResource{URI: publicIPID}
								}
							}
						}
					}
				}
			}
		}
	}

	// Extract VPN client settings
	var vpnClientSettings *sdktypes.VPNClientSettings
	if vpnClientAttr, ok := propertiesAttrs["vpn_client_settings"]; ok && vpnClientAttr != nil {
		if vpnClientObj, ok := vpnClientAttr.(types.Object); ok && !vpnClientObj.IsNull() {
			vpnClientObjValue, diags := vpnClientObj.ToObjectValue(ctx)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				vpnClientAttrs := vpnClientObjValue.Attributes()
				vpnClientSettings = &sdktypes.VPNClientSettings{}

				// Extract IKE settings
				if ikeAttr, ok := vpnClientAttrs["ike"]; ok && ikeAttr != nil {
					if ikeObj, ok := ikeAttr.(types.Object); ok && !ikeObj.IsNull() {
						ikeObjValue, diags := ikeObj.ToObjectValue(ctx)
						resp.Diagnostics.Append(diags...)
						if !resp.Diagnostics.HasError() {
							ikeAttrs := ikeObjValue.Attributes()
							ikeSettings := &sdktypes.IKESettings{}

							if lifetimeAttr, ok := ikeAttrs["lifetime"]; ok && lifetimeAttr != nil {
								if lifetimeInt, ok := lifetimeAttr.(types.Int64); ok && !lifetimeInt.IsNull() {
									lifetime := int32(lifetimeInt.ValueInt64())
									// Lifetime field expects int32, not *int32
									ikeSettings.Lifetime = lifetime
								}
							}
							if encryptionAttr, ok := ikeAttrs["encryption"]; ok && encryptionAttr != nil {
								if encryptionStr, ok := encryptionAttr.(types.String); ok && !encryptionStr.IsNull() {
									encryption := encryptionStr.ValueString()
									ikeSettings.Encryption = &encryption
								}
							}
							if hashAttr, ok := ikeAttrs["hash"]; ok && hashAttr != nil {
								if hashStr, ok := hashAttr.(types.String); ok && !hashStr.IsNull() {
									hash := hashStr.ValueString()
									ikeSettings.Hash = &hash
								}
							}
							if dhGroupAttr, ok := ikeAttrs["dh_group"]; ok && dhGroupAttr != nil {
								if dhGroupStr, ok := dhGroupAttr.(types.String); ok && !dhGroupStr.IsNull() {
									dhGroup := dhGroupStr.ValueString()
									ikeSettings.DHGroup = &dhGroup
								}
							}
							if dpdActionAttr, ok := ikeAttrs["dpd_action"]; ok && dpdActionAttr != nil {
								if dpdActionStr, ok := dpdActionAttr.(types.String); ok && !dpdActionStr.IsNull() {
									dpdAction := dpdActionStr.ValueString()
									ikeSettings.DPDAction = &dpdAction
								}
							}
							if dpdIntervalAttr, ok := ikeAttrs["dpd_interval"]; ok && dpdIntervalAttr != nil {
								if dpdIntervalInt, ok := dpdIntervalAttr.(types.Int64); ok && !dpdIntervalInt.IsNull() {
									dpdInterval := int32(dpdIntervalInt.ValueInt64())
									// DPDInterval field expects int32, not *int32
									ikeSettings.DPDInterval = dpdInterval
								}
							}
							if dpdTimeoutAttr, ok := ikeAttrs["dpd_timeout"]; ok && dpdTimeoutAttr != nil {
								if dpdTimeoutInt, ok := dpdTimeoutAttr.(types.Int64); ok && !dpdTimeoutInt.IsNull() {
									dpdTimeout := int32(dpdTimeoutInt.ValueInt64())
									// DPDTimeout field expects int32, not *int32
									ikeSettings.DPDTimeout = dpdTimeout
								}
							}

							vpnClientSettings.IKE = ikeSettings
						}
					}
				}

				// Extract ESP settings
				if espAttr, ok := vpnClientAttrs["esp"]; ok && espAttr != nil {
					if espObj, ok := espAttr.(types.Object); ok && !espObj.IsNull() {
						espObjValue, diags := espObj.ToObjectValue(ctx)
						resp.Diagnostics.Append(diags...)
						if !resp.Diagnostics.HasError() {
							espAttrs := espObjValue.Attributes()
							espSettings := &sdktypes.ESPSettings{}

							if lifetimeAttr, ok := espAttrs["lifetime"]; ok && lifetimeAttr != nil {
								if lifetimeInt, ok := lifetimeAttr.(types.Int64); ok && !lifetimeInt.IsNull() {
									lifetime := int32(lifetimeInt.ValueInt64())
									// Lifetime field expects int32, not *int32
									espSettings.Lifetime = lifetime
								}
							}
							if encryptionAttr, ok := espAttrs["encryption"]; ok && encryptionAttr != nil {
								if encryptionStr, ok := encryptionAttr.(types.String); ok && !encryptionStr.IsNull() {
									encryption := encryptionStr.ValueString()
									espSettings.Encryption = &encryption
								}
							}
							if hashAttr, ok := espAttrs["hash"]; ok && hashAttr != nil {
								if hashStr, ok := hashAttr.(types.String); ok && !hashStr.IsNull() {
									hash := hashStr.ValueString()
									espSettings.Hash = &hash
								}
							}
							if pfsAttr, ok := espAttrs["pfs"]; ok && pfsAttr != nil {
								if pfsStr, ok := pfsAttr.(types.String); ok && !pfsStr.IsNull() {
									pfs := pfsStr.ValueString()
									espSettings.PFS = &pfs
								}
							}

							vpnClientSettings.ESP = espSettings
						}
					}
				}

				// Extract PSK settings
				if pskAttr, ok := vpnClientAttrs["psk"]; ok && pskAttr != nil {
					if pskObj, ok := pskAttr.(types.Object); ok && !pskObj.IsNull() {
						pskObjValue, diags := pskObj.ToObjectValue(ctx)
						resp.Diagnostics.Append(diags...)
						if !resp.Diagnostics.HasError() {
							pskAttrs := pskObjValue.Attributes()
							pskSettings := &sdktypes.PSKSettings{}

							if cloudSiteAttr, ok := pskAttrs["cloud_site"]; ok && cloudSiteAttr != nil {
								if cloudSiteStr, ok := cloudSiteAttr.(types.String); ok && !cloudSiteStr.IsNull() {
									cloudSite := cloudSiteStr.ValueString()
									pskSettings.CloudSite = &cloudSite
								}
							}
							if onPremSiteAttr, ok := pskAttrs["on_prem_site"]; ok && onPremSiteAttr != nil {
								if onPremSiteStr, ok := onPremSiteAttr.(types.String); ok && !onPremSiteStr.IsNull() {
									onPremSite := onPremSiteStr.ValueString()
									pskSettings.OnPremSite = &onPremSite
								}
							}
							if secretAttr, ok := pskAttrs["secret"]; ok && secretAttr != nil {
								if secretStr, ok := secretAttr.(types.String); ok && !secretStr.IsNull() {
									secret := secretStr.ValueString()
									pskSettings.Secret = &secret
								}
							}

							vpnClientSettings.PSK = pskSettings
						}
					}
				}

				// Extract peer client public IP
				if peerIPAttr, ok := vpnClientAttrs["peer_client_public_ip"]; ok && peerIPAttr != nil {
					if peerIPStr, ok := peerIPAttr.(types.String); ok && !peerIPStr.IsNull() {
						peerIP := peerIPStr.ValueString()
						vpnClientSettings.PeerClientPublicIP = &peerIP
					}
				}
			}
		}
	}

	// Build the create request
	createRequest := sdktypes.VPNTunnelRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.VPNTunnelPropertiesRequest{
			VPNType:           &vpnType,
			VPNClientProtocol: &protocol,
			IPConfigurations:  ipConfig,
			VPNClientSettings: vpnClientSettings,
			BillingPlan: &sdktypes.BillingPeriodResource{
				BillingPeriod: billingPeriod,
			},
		},
	}

	// Create the VPN tunnel using the SDK
	response, err := r.client.Client.FromNetwork().VPNTunnels().Create(ctx, projectID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating VPN tunnel",
			fmt.Sprintf("Unable to create VPN tunnel: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		logContext := map[string]interface{}{}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to create VPN tunnel", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"VPN tunnel created but no data returned from API",
		)
		return
	}

	// Wait for VPN Tunnel to be active before returning
	// This ensures Terraform doesn't proceed to create dependent resources until VPN Tunnel is ready
	tunnelID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromNetwork().VPNTunnels().Get(ctx, projectID, tunnelID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for VPN Tunnel to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "VPNTunnel", tunnelID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"VPN Tunnel Not Active",
			fmt.Sprintf("VPN tunnel was created but did not become active within the timeout period: %s", err),
		)
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

	projectID := data.ProjectId.ValueString()
	tunnelID := data.Id.ValueString()

	if projectID == "" || tunnelID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Tunnel ID are required to read the VPN tunnel",
		)
		return
	}

	// Get VPN tunnel details using the SDK
	response, err := r.client.Client.FromNetwork().VPNTunnels().Get(ctx, projectID, tunnelID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading VPN tunnel",
			fmt.Sprintf("Unable to read VPN tunnel: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		logContext := map[string]interface{}{
			"project_id": projectID,
			"tunnel_id":  tunnelID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to read VPN tunnel", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		tunnel := response.Data

		if tunnel.Metadata.ID != nil {
			data.Id = types.StringValue(*tunnel.Metadata.ID)
		}
		if tunnel.Metadata.URI != nil {
			data.Uri = types.StringValue(*tunnel.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if tunnel.Metadata.Name != nil {
			data.Name = types.StringValue(*tunnel.Metadata.Name)
		}
		if tunnel.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(tunnel.Metadata.LocationResponse.Value)
		}

		// Update tags
		if len(tunnel.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(tunnel.Metadata.Tags))
			for i, tag := range tunnel.Metadata.Tags {
				tagValues[i] = types.StringValue(tag)
			}
			tagsList, diags := types.ListValueFrom(ctx, types.StringType, tagValues)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = tagsList
			}
		} else {
			emptyList, diags := types.ListValue(types.StringType, []attr.Value{})
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = emptyList
			}
		}

		// Note: Properties are complex nested structures - for now, we preserve the existing state
		// A full implementation would reconstruct the properties object from the API response
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPNTunnelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VPNTunnelResourceModel
	var state VPNTunnelResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get IDs from state (not plan) - IDs are immutable and should always be in state
	projectID := state.ProjectId.ValueString()
	tunnelID := state.Id.ValueString()

	if projectID == "" || tunnelID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Tunnel ID are required to update the VPN tunnel",
		)
		return
	}

	// Get current VPN tunnel details
	getResponse, err := r.client.Client.FromNetwork().VPNTunnels().Get(ctx, projectID, tunnelID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current VPN tunnel",
			fmt.Sprintf("Unable to get current VPN tunnel: %s", err),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"VPN Tunnel Not Found",
			"VPN tunnel not found or no data returned",
		)
		return
	}

	current := getResponse.Data

	// Check if VPN tunnel is in InCreation state
	if current.Status.State != nil && *current.Status.State == "InCreation" {
		resp.Diagnostics.AddError(
			"Cannot Update VPN Tunnel",
			"Cannot update VPN tunnel while it is in 'InCreation' state. Please wait until the VPN tunnel is fully created.",
		)
		return
	}

	// Get region value
	regionValue := ""
	if current.Metadata.LocationResponse != nil {
		regionValue = current.Metadata.LocationResponse.Value
	}
	if regionValue == "" {
		resp.Diagnostics.AddError(
			"Missing Region",
			"Unable to determine region value for VPN tunnel",
		)
		return
	}

	// Extract tags
	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		tags = current.Metadata.Tags
	}

	// Build update request - only name and tags can be updated, properties must remain unchanged
	updateRequest := sdktypes.VPNTunnelRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.VPNTunnelPropertiesRequest{
			// Properties cannot be updated - use current values
			VPNType:           current.Properties.VPNType,
			VPNClientProtocol: current.Properties.VPNClientProtocol,
			IPConfigurations:  current.Properties.IPConfigurations,
			VPNClientSettings: current.Properties.VPNClientSettings,
			BillingPlan:       current.Properties.BillingPlan,
		},
	}

	// Update the VPN tunnel using the SDK
	response, err := r.client.Client.FromNetwork().VPNTunnels().Update(ctx, projectID, tunnelID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating VPN tunnel",
			fmt.Sprintf("Unable to update VPN tunnel: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		logContext := map[string]interface{}{}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to update VPN tunnel", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.ProjectId = state.ProjectId

	if response != nil && response.Data != nil {
		// Update from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPNTunnelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VPNTunnelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	tunnelID := data.Id.ValueString()

	if projectID == "" || tunnelID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Tunnel ID are required to delete the VPN tunnel",
		)
		return
	}

	// Delete the VPN tunnel using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromNetwork().VPNTunnels().Delete(ctx, projectID, tunnelID, nil)
		},
		ExtractSDKError,
		"VPNTunnel",
		tunnelID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting VPN tunnel",
			fmt.Sprintf("Unable to delete VPN tunnel: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a VPN Tunnel resource", map[string]interface{}{
		"vpntunnel_id": tunnelID,
	})
}

func (r *VPNTunnelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
