package provider

import (
	"context"
	"fmt"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &SubnetResource{}
var _ resource.ResourceWithImportState = &SubnetResource{}

func NewSubnetResource() resource.Resource {
	return &SubnetResource{}
}

type SubnetResource struct {
	client *ArubaCloudClient
}

type SubnetResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	Tags      types.List   `tfsdk:"tags"`
	ProjectId types.String `tfsdk:"project_id"`
	VpcId     types.String `tfsdk:"vpc_id"`
	Type      types.String `tfsdk:"type"`
	Network   types.Object `tfsdk:"network"`
}

type NetworkModel struct {
	Address types.String `tfsdk:"address"`
	Dhcp    types.Object `tfsdk:"dhcp"`
}

type DhcpModel struct {
	Enabled types.Bool   `tfsdk:"enabled"`
	Range   types.Object `tfsdk:"range"`
	Routes  types.List   `tfsdk:"routes"`
	Dns     types.List   `tfsdk:"dns"`
}

type DhcpRangeModel struct {
	Start types.String `tfsdk:"start"`
	Count types.Int64  `tfsdk:"count"`
}

type RouteModel struct {
	Address types.String `tfsdk:"address"`
	Gateway types.String `tfsdk:"gateway"`
}

func (r *SubnetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (r *SubnetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Subnet resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Subnet identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Subnet URI",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Subnet name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Subnet location",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the subnet",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this subnet belongs to",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPC this subnet belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Subnet type (Basic or Advanced)",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"network": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						MarkdownDescription: "Address of the network in CIDR notation (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)",
						Optional:            true,
					},
					"dhcp": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								MarkdownDescription: "Enable DHCP",
								Optional:            true,
							},
							"range": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"start": schema.StringAttribute{
										MarkdownDescription: "Starting IP address",
										Optional:            true,
									},
									"count": schema.Int64Attribute{
										MarkdownDescription: "Number of available IP addresses",
										Optional:            true,
									},
								},
								Optional: true,
							},
							"routes": schema.ListNestedAttribute{
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"address": schema.StringAttribute{
											MarkdownDescription: "Destination network address in CIDR notation (e.g., 0.0.0.0/0)",
											Optional:            true,
										},
										"gateway": schema.StringAttribute{
											MarkdownDescription: "Gateway IP address for the route",
											Optional:            true,
										},
									},
								},
								MarkdownDescription: "DHCP routes configuration",
								Optional:            true,
							},
							"dns": schema.ListAttribute{
								ElementType:         types.StringType,
								MarkdownDescription: "DNS server addresses",
								Optional:            true,
							},
						},
						Optional: true,
					},
				},
				Optional: true,
			},
		},
	}
}

func (r *SubnetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ArubaCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *SubnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	if projectID == "" || vpcID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and VPC ID are required to create a subnet",
		)
		return
	}

	// Extract tags from Terraform list
	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	subnetType := data.Type.ValueString()

	// Validation: If type is Advanced, network block is mandatory
	if subnetType == "Advanced" {
		if data.Network.IsNull() || data.Network.IsUnknown() {
			resp.Diagnostics.AddError(
				"Missing Required Field",
				"The 'network' block is required when subnet type is 'Advanced'",
			)
			return
		}
	}

	// Extract network CIDR and DHCP if provided
	var network *sdktypes.SubnetNetwork
	var dhcp *sdktypes.SubnetDHCP
	if !data.Network.IsNull() && !data.Network.IsUnknown() {
		networkObj, diags := data.Network.ToObjectValue(ctx)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			attrs := networkObj.Attributes()

			// Extract network address
			var addressValue string
			if addressAttr, ok := attrs["address"]; ok && addressAttr != nil {
				if addressStr, ok := addressAttr.(types.String); ok && !addressStr.IsNull() {
					addressValue = addressStr.ValueString()
					if addressValue != "" {
						network = &sdktypes.SubnetNetwork{
							Address: addressValue,
						}
					}
				}
			}

			// Validation: If type is Advanced, address is mandatory
			if subnetType == "Advanced" && addressValue == "" {
				resp.Diagnostics.AddError(
					"Missing Required Field",
					"The 'network.address' field is required when subnet type is 'Advanced'",
				)
				return
			}

			// Extract DHCP configuration from network.dhcp
			var dhcpEnabledSet bool
			var dhcpRangeStart string
			var dhcpRangeCount int

			if dhcpAttr, ok := attrs["dhcp"]; ok && dhcpAttr != nil {
				if dhcpObj, ok := dhcpAttr.(types.Object); ok && !dhcpObj.IsNull() {
					dhcpAttrs := dhcpObj.Attributes()
					dhcp = &sdktypes.SubnetDHCP{}

					// Extract enabled
					if enabledAttr, ok := dhcpAttrs["enabled"]; ok && enabledAttr != nil {
						if enabledBool, ok := enabledAttr.(types.Bool); ok && !enabledBool.IsNull() {
							dhcp.Enabled = enabledBool.ValueBool()
							dhcpEnabledSet = true
						}
					}

					// Extract range
					if rangeAttr, ok := dhcpAttrs["range"]; ok && rangeAttr != nil {
						if rangeObj, ok := rangeAttr.(types.Object); ok && !rangeObj.IsNull() {
							rangeAttrs := rangeObj.Attributes()
							dhcpRange := &sdktypes.SubnetDHCPRange{}
							if startAttr, ok := rangeAttrs["start"]; ok && startAttr != nil {
								if startStr, ok := startAttr.(types.String); ok && !startStr.IsNull() {
									dhcpRange.Start = startStr.ValueString()
									dhcpRangeStart = dhcpRange.Start
								}
							}
							if countAttr, ok := rangeAttrs["count"]; ok && countAttr != nil {
								if countInt, ok := countAttr.(types.Int64); ok && !countInt.IsNull() {
									dhcpRange.Count = int(countInt.ValueInt64())
									dhcpRangeCount = dhcpRange.Count
								}
							}
							if dhcpRange.Start != "" || dhcpRange.Count > 0 {
								dhcp.Range = dhcpRange
							}
						}
					}

					// Extract routes
					if routesAttr, ok := dhcpAttrs["routes"]; ok && routesAttr != nil {
						if routesList, ok := routesAttr.(types.List); ok && !routesList.IsNull() {
							var routesData []types.Object
							diags := routesList.ElementsAs(ctx, &routesData, false)
							resp.Diagnostics.Append(diags...)
							if !resp.Diagnostics.HasError() {
								dhcpRoutes := make([]sdktypes.SubnetDHCPRoute, 0, len(routesData))
								for _, routeObj := range routesData {
									routeAttrs := routeObj.Attributes()
									route := sdktypes.SubnetDHCPRoute{}
									if addrAttr, ok := routeAttrs["address"]; ok && addrAttr != nil {
										if addrStr, ok := addrAttr.(types.String); ok && !addrStr.IsNull() {
											route.Address = addrStr.ValueString()
										}
									}
									if gwAttr, ok := routeAttrs["gateway"]; ok && gwAttr != nil {
										if gwStr, ok := gwAttr.(types.String); ok && !gwStr.IsNull() {
											route.Gateway = gwStr.ValueString()
										}
									}
									if route.Address != "" || route.Gateway != "" {
										dhcpRoutes = append(dhcpRoutes, route)
									}
								}
								if len(dhcpRoutes) > 0 {
									dhcp.Routes = dhcpRoutes
								}
							}
						}
					}

					// Extract DNS
					if dnsAttr, ok := dhcpAttrs["dns"]; ok && dnsAttr != nil {
						if dnsList, ok := dnsAttr.(types.List); ok && !dnsList.IsNull() {
							var dnsServers []string
							diags := dnsList.ElementsAs(ctx, &dnsServers, false)
							resp.Diagnostics.Append(diags...)
							if !resp.Diagnostics.HasError() && len(dnsServers) > 0 {
								dhcp.DNS = dnsServers
							}
						}
					}
				}
			}

			// Validation: If type is Advanced, dhcp block with enabled, range.start and range.count are mandatory
			if subnetType == "Advanced" {
				if dhcp == nil {
					resp.Diagnostics.AddError(
						"Missing Required Field",
						"The 'network.dhcp' block is required when subnet type is 'Advanced'",
					)
					return
				}
				if !dhcpEnabledSet {
					resp.Diagnostics.AddError(
						"Missing Required Field",
						"The 'network.dhcp.enabled' field is required when subnet type is 'Advanced'",
					)
					return
				}
				if dhcp.Range == nil || dhcpRangeStart == "" || dhcpRangeCount == 0 {
					resp.Diagnostics.AddError(
						"Missing Required Fields",
						"The 'network.dhcp.range' block with 'start' and 'count' fields is required when subnet type is 'Advanced'",
					)
					return
				}
			}
		}
	}

	// Determine SubnetType for SDK: Advanced if CIDR is provided, Basic otherwise
	sdkSubnetType := sdktypes.SubnetTypeBasic
	if network != nil && network.Address != "" {
		sdkSubnetType = sdktypes.SubnetTypeAdvanced
	} else if data.Type.ValueString() == "Advanced" {
		sdkSubnetType = sdktypes.SubnetTypeAdvanced
	}

	// Build the create request
	createRequest := sdktypes.SubnetRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.SubnetPropertiesRequest{
			Type:    sdkSubnetType,
			Network: network,
			DHCP:    dhcp,
		},
	}

	// Create the subnet using the SDK
	response, err := r.client.Client.FromNetwork().Subnets().Create(ctx, projectID, vpcID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating subnet",
			fmt.Sprintf("Unable to create subnet: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		logContext := map[string]interface{}{}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to create subnet", logContext)
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
			"Subnet created but no data returned from API",
		)
		return
	}

	// Wait for Subnet to be active before returning (Subnet is referenced by CloudServer)
	// This ensures Terraform doesn't proceed to create dependent resources until Subnet is ready
	subnetID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromNetwork().Subnets().Get(ctx, projectID, vpcID, subnetID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for Subnet to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "Subnet", subnetID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Subnet Not Active",
			fmt.Sprintf("Subnet was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	// Re-read the Subnet to get the URI and ensure all fields are properly set
	getResp, err := r.client.Client.FromNetwork().Subnets().Get(ctx, projectID, vpcID, subnetID, nil)
	if err == nil && getResp != nil && getResp.Data != nil {
		// Ensure ID is set from metadata (should already be set, but double-check)
		if getResp.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*getResp.Data.Metadata.ID)
		}
		if getResp.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*getResp.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		// Also update other fields that might have changed
		if getResp.Data.Metadata.Name != nil {
			data.Name = types.StringValue(*getResp.Data.Metadata.Name)
		}
		if getResp.Data.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(getResp.Data.Metadata.LocationResponse.Value)
		}
		// Update tags from response
		if len(getResp.Data.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(getResp.Data.Metadata.Tags))
			for i, tag := range getResp.Data.Metadata.Tags {
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
	} else if err != nil {
		// If Get fails, log but don't fail - we already have the ID from create response
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh Subnet after creation: %v", err))
	}

	tflog.Trace(ctx, "created a Subnet resource", map[string]interface{}{
		"subnet_id":   data.Id.ValueString(),
		"subnet_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	subnetID := data.Id.ValueString()

	if projectID == "" || vpcID == "" || subnetID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, and Subnet ID are required to read the subnet",
		)
		return
	}

	// Get subnet details using the SDK
	response, err := r.client.Client.FromNetwork().Subnets().Get(ctx, projectID, vpcID, subnetID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading subnet",
			fmt.Sprintf("Unable to read subnet: %s", err),
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
			"subnet_id":  subnetID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to read subnet", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		subnet := response.Data

		if subnet.Metadata.ID != nil {
			data.Id = types.StringValue(*subnet.Metadata.ID)
		}
		if subnet.Metadata.URI != nil {
			data.Uri = types.StringValue(*subnet.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if subnet.Metadata.Name != nil {
			data.Name = types.StringValue(*subnet.Metadata.Name)
		}
		if subnet.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(subnet.Metadata.LocationResponse.Value)
		}
		data.Type = types.StringValue(string(subnet.Properties.Type))
		subnetType := string(subnet.Properties.Type)

		// Update network and DHCP if available
		networkAttrs := make(map[string]attr.Value)
		networkAttrTypes := map[string]attr.Type{
			"address": types.StringType,
			"dhcp": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"enabled": types.BoolType,
					"range": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"start": types.StringType,
							"count": types.Int64Type,
						},
					},
					"routes": types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"address": types.StringType,
								"gateway": types.StringType,
							},
						},
					},
					"dns": types.ListType{ElemType: types.StringType},
				},
			},
		}

		// Set network address
		// For Basic subnets, only set network if it was in the original state (to avoid drift)
		// For Advanced subnets, always set network if available from API
		networkWasInState := !data.Network.IsNull() && !data.Network.IsUnknown()
		shouldSetNetwork := subnetType == "Advanced" || networkWasInState

		if shouldSetNetwork {
			// Set network address
			if subnet.Properties.Network != nil && subnet.Properties.Network.Address != "" {
				networkAttrs["address"] = types.StringValue(subnet.Properties.Network.Address)
			} else {
				networkAttrs["address"] = types.StringNull()
			}

			// Set DHCP if available
			if subnet.Properties.DHCP != nil {
				dhcpAttrs := make(map[string]attr.Value)
				dhcpAttrs["enabled"] = types.BoolValue(subnet.Properties.DHCP.Enabled)

				// Handle DHCP range
				if subnet.Properties.DHCP.Range != nil {
					rangeObj, diags := types.ObjectValue(map[string]attr.Type{
						"start": types.StringType,
						"count": types.Int64Type,
					}, map[string]attr.Value{
						"start": types.StringValue(subnet.Properties.DHCP.Range.Start),
						"count": types.Int64Value(int64(subnet.Properties.DHCP.Range.Count)),
					})
					resp.Diagnostics.Append(diags...)
					if !resp.Diagnostics.HasError() {
						dhcpAttrs["range"] = rangeObj
					}
				} else {
					dhcpAttrs["range"] = types.ObjectNull(map[string]attr.Type{
						"start": types.StringType,
						"count": types.Int64Type,
					})
				}

				// Handle DHCP routes
				if len(subnet.Properties.DHCP.Routes) > 0 {
					routeObjs := make([]attr.Value, 0, len(subnet.Properties.DHCP.Routes))
					for _, route := range subnet.Properties.DHCP.Routes {
						routeObj, diags := types.ObjectValue(map[string]attr.Type{
							"address": types.StringType,
							"gateway": types.StringType,
						}, map[string]attr.Value{
							"address": types.StringValue(route.Address),
							"gateway": types.StringValue(route.Gateway),
						})
						resp.Diagnostics.Append(diags...)
						if !resp.Diagnostics.HasError() {
							routeObjs = append(routeObjs, routeObj)
						}
					}
					routesList, diags := types.ListValue(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"address": types.StringType,
							"gateway": types.StringType,
						},
					}, routeObjs)
					resp.Diagnostics.Append(diags...)
					if !resp.Diagnostics.HasError() {
						dhcpAttrs["routes"] = routesList
					}
				} else {
					dhcpAttrs["routes"] = types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"address": types.StringType,
							"gateway": types.StringType,
						},
					})
				}

				// Handle DNS
				if len(subnet.Properties.DHCP.DNS) > 0 {
					dnsValues := make([]attr.Value, 0, len(subnet.Properties.DHCP.DNS))
					for _, dns := range subnet.Properties.DHCP.DNS {
						dnsValues = append(dnsValues, types.StringValue(dns))
					}
					dnsList, diags := types.ListValue(types.StringType, dnsValues)
					resp.Diagnostics.Append(diags...)
					if !resp.Diagnostics.HasError() {
						dhcpAttrs["dns"] = dnsList
					}
				} else {
					dhcpAttrs["dns"] = types.ListNull(types.StringType)
				}

				dhcpObj, diags := types.ObjectValue(map[string]attr.Type{
					"enabled": types.BoolType,
					"range": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"start": types.StringType,
							"count": types.Int64Type,
						},
					},
					"routes": types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"address": types.StringType,
								"gateway": types.StringType,
							},
						},
					},
					"dns": types.ListType{ElemType: types.StringType},
				}, dhcpAttrs)
				resp.Diagnostics.Append(diags...)
				if !resp.Diagnostics.HasError() {
					networkAttrs["dhcp"] = dhcpObj
				}
			} else {
				networkAttrs["dhcp"] = types.ObjectNull(map[string]attr.Type{
					"enabled": types.BoolType,
					"range": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"start": types.StringType,
							"count": types.Int64Type,
						},
					},
					"routes": types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"address": types.StringType,
								"gateway": types.StringType,
							},
						},
					},
					"dns": types.ListType{ElemType: types.StringType},
				})
			}

			// Build network object with nested dhcp
			networkObj, diags := types.ObjectValue(networkAttrTypes, networkAttrs)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Network = networkObj
			}
		} else {
			// Basic subnet without network in state - set entire network block to null
			// This prevents drift when API returns network info but it wasn't configured
			data.Network = types.ObjectNull(networkAttrTypes)
		}

		// Update tags from response
		if len(subnet.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(subnet.Metadata.Tags))
			for i, tag := range subnet.Metadata.Tags {
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
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SubnetResourceModel
	var state SubnetResourceModel

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
	vpcID := state.VpcId.ValueString()
	subnetID := state.Id.ValueString()

	if projectID == "" || vpcID == "" || subnetID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, and Subnet ID are required to update the subnet",
		)
		return
	}

	// Get current subnet details
	getResponse, err := r.client.Client.FromNetwork().Subnets().Get(ctx, projectID, vpcID, subnetID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current subnet",
			fmt.Sprintf("Unable to get current subnet: %s", err),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"Subnet Not Found",
			"Subnet not found or no data returned",
		)
		return
	}

	current := getResponse.Data

	// Get region value
	regionValue := ""
	if current.Metadata.LocationResponse != nil {
		regionValue = current.Metadata.LocationResponse.Value
	}
	if regionValue == "" {
		resp.Diagnostics.AddError(
			"Missing Region",
			"Unable to determine region value for subnet",
		)
		return
	}

	// Extract tags from Terraform list
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

	// Preserve network from current state
	network := current.Properties.Network

	// Extract DHCP configuration from network.dhcp if provided
	var dhcp *sdktypes.SubnetDHCP
	if !data.Network.IsNull() && !data.Network.IsUnknown() {
		networkObj, diags := data.Network.ToObjectValue(ctx)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			attrs := networkObj.Attributes()

			// Extract DHCP configuration from network.dhcp
			if dhcpAttr, ok := attrs["dhcp"]; ok && dhcpAttr != nil {
				if dhcpObj, ok := dhcpAttr.(types.Object); ok && !dhcpObj.IsNull() {
					dhcpAttrs := dhcpObj.Attributes()
					dhcp = &sdktypes.SubnetDHCP{}

					// Extract enabled
					if enabledAttr, ok := dhcpAttrs["enabled"]; ok && enabledAttr != nil {
						if enabledBool, ok := enabledAttr.(types.Bool); ok && !enabledBool.IsNull() {
							dhcp.Enabled = enabledBool.ValueBool()
						}
					}

					// Extract range
					if rangeAttr, ok := dhcpAttrs["range"]; ok && rangeAttr != nil {
						if rangeObj, ok := rangeAttr.(types.Object); ok && !rangeObj.IsNull() {
							rangeAttrs := rangeObj.Attributes()
							dhcpRange := &sdktypes.SubnetDHCPRange{}
							if startAttr, ok := rangeAttrs["start"]; ok && startAttr != nil {
								if startStr, ok := startAttr.(types.String); ok && !startStr.IsNull() {
									dhcpRange.Start = startStr.ValueString()
								}
							}
							if countAttr, ok := rangeAttrs["count"]; ok && countAttr != nil {
								if countInt, ok := countAttr.(types.Int64); ok && !countInt.IsNull() {
									dhcpRange.Count = int(countInt.ValueInt64())
								}
							}
							if dhcpRange.Start != "" || dhcpRange.Count > 0 {
								dhcp.Range = dhcpRange
							}
						}
					}

					// Extract routes
					if routesAttr, ok := dhcpAttrs["routes"]; ok && routesAttr != nil {
						if routesList, ok := routesAttr.(types.List); ok && !routesList.IsNull() {
							var routesData []types.Object
							diags := routesList.ElementsAs(ctx, &routesData, false)
							resp.Diagnostics.Append(diags...)
							if !resp.Diagnostics.HasError() {
								dhcpRoutes := make([]sdktypes.SubnetDHCPRoute, 0, len(routesData))
								for _, routeObj := range routesData {
									routeAttrs := routeObj.Attributes()
									route := sdktypes.SubnetDHCPRoute{}
									if addrAttr, ok := routeAttrs["address"]; ok && addrAttr != nil {
										if addrStr, ok := addrAttr.(types.String); ok && !addrStr.IsNull() {
											route.Address = addrStr.ValueString()
										}
									}
									if gwAttr, ok := routeAttrs["gateway"]; ok && gwAttr != nil {
										if gwStr, ok := gwAttr.(types.String); ok && !gwStr.IsNull() {
											route.Gateway = gwStr.ValueString()
										}
									}
									if route.Address != "" || route.Gateway != "" {
										dhcpRoutes = append(dhcpRoutes, route)
									}
								}
								if len(dhcpRoutes) > 0 {
									dhcp.Routes = dhcpRoutes
								}
							}
						}
					}

					// Extract DNS
					if dnsAttr, ok := dhcpAttrs["dns"]; ok && dnsAttr != nil {
						if dnsList, ok := dnsAttr.(types.List); ok && !dnsList.IsNull() {
							var dnsServers []string
							diags := dnsList.ElementsAs(ctx, &dnsServers, false)
							resp.Diagnostics.Append(diags...)
							if !resp.Diagnostics.HasError() && len(dnsServers) > 0 {
								dhcp.DNS = dnsServers
							}
						}
					}
				}
			}
		}
	}

	// If dhcp wasn't provided in the plan, preserve from current state
	if dhcp == nil {
		dhcp = current.Properties.DHCP
	}

	// Build the update request
	updateRequest := sdktypes.SubnetRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.SubnetPropertiesRequest{
			Type:    current.Properties.Type,
			Network: network,
			DHCP:    dhcp,
		},
	}

	// Update the subnet using the SDK
	response, err := r.client.Client.FromNetwork().Subnets().Update(ctx, projectID, vpcID, subnetID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating subnet",
			fmt.Sprintf("Unable to update subnet: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		logContext := map[string]interface{}{}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to update subnet", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.ProjectId = state.ProjectId
	data.VpcId = state.VpcId
	data.Uri = state.Uri // Preserve URI from state

	if response != nil && response.Data != nil {
		// Update from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			// If no URI in response, re-read the subnet to get the latest state
			getResp, err := r.client.Client.FromNetwork().Subnets().Get(ctx, projectID, vpcID, subnetID, nil)
			if err == nil && getResp != nil && getResp.Data != nil {
				if getResp.Data.Metadata.URI != nil {
					data.Uri = types.StringValue(*getResp.Data.Metadata.URI)
				} else {
					data.Uri = state.Uri // Fallback to state if not available
				}
			} else {
				data.Uri = state.Uri // Fallback to state if re-read fails
			}
		}
	} else {
		// If no response, preserve URI from state
		data.Uri = state.Uri
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	subnetID := data.Id.ValueString()

	if projectID == "" || vpcID == "" || subnetID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, and Subnet ID are required to delete the subnet",
		)
		return
	}

	// Delete the subnet using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromNetwork().Subnets().Delete(ctx, projectID, vpcID, subnetID, nil)
		},
		ExtractSDKError,
		"Subnet",
		subnetID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting subnet",
			fmt.Sprintf("Unable to delete subnet: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a Subnet resource", map[string]interface{}{
		"subnet_id": subnetID,
	})
}

func (r *SubnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
