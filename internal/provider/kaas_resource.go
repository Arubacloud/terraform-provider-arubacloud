// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type KaaSResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &KaaSResource{}
var _ resource.ResourceWithImportState = &KaaSResource{}

func NewKaaSResource() resource.Resource {
	return &KaaSResource{}
}

type KaaSResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	ProjectID     types.String `tfsdk:"project_id"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	ManagementIP  types.String `tfsdk:"management_ip"`
	Network       types.Object `tfsdk:"network"`
	Settings      types.Object `tfsdk:"settings"`
}

type KaaSNodeCIDRModel struct {
	Address types.String `tfsdk:"address"`
	Name    types.String `tfsdk:"name"`
}

type KaaSNodePoolModel struct {
	Name        types.String `tfsdk:"name"`
	Nodes       types.Int64  `tfsdk:"nodes"`
	Instance    types.String `tfsdk:"instance"`
	Zone        types.String `tfsdk:"zone"`
	Autoscaling types.Bool   `tfsdk:"autoscaling"`
	MinCount    types.Int64  `tfsdk:"min_count"`
	MaxCount    types.Int64  `tfsdk:"max_count"`
}

type KaaSNetworkModel struct {
	VpcUriRef         types.String `tfsdk:"vpc_uri_ref"`
	SubnetUriRef      types.String `tfsdk:"subnet_uri_ref"`
	NodeCIDR          types.Object `tfsdk:"node_cidr"`
	SecurityGroupName types.String `tfsdk:"security_group_name"`
	PodCIDR           types.String `tfsdk:"pod_cidr"`
}

type KaaSSettingsModel struct {
	KubernetesVersion types.String `tfsdk:"kubernetes_version"`
	NodePools         types.List   `tfsdk:"node_pools"`
	HA                types.Bool   `tfsdk:"ha"`
}

func (r *KaaSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kaas"
}

func (r *KaaSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "KaaS (Kubernetes as a Service) resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "KaaS identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "KaaS URI",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "KaaS name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "KaaS location",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the KaaS resource",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this KaaS resource belongs to",
				Required:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period (Hour, Month, Year)",
				Optional:            true,
			},
			"management_ip": schema.StringAttribute{
				MarkdownDescription: "Management IP address (available when KaaS is active)",
				Computed:            true,
			},
			"network": schema.SingleNestedAttribute{
				MarkdownDescription: "Network configuration for the KaaS cluster",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"vpc_uri_ref": schema.StringAttribute{
						MarkdownDescription: "VPC URI reference for the KaaS resource (e.g., arubacloud_vpc.example.uri)",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"subnet_uri_ref": schema.StringAttribute{
						MarkdownDescription: "Subnet URI reference for the KaaS resource (e.g., arubacloud_subnet.example.uri)",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"node_cidr": schema.SingleNestedAttribute{
						MarkdownDescription: "Node CIDR configuration",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"address": schema.StringAttribute{
								MarkdownDescription: "Node CIDR address in CIDR notation (e.g., 10.0.0.0/24)",
								Required:            true,
							},
							"name": schema.StringAttribute{
								MarkdownDescription: "Node CIDR name",
								Required:            true,
							},
						},
					},
					"security_group_name": schema.StringAttribute{
						MarkdownDescription: "Security group name",
						Required:            true,
					},
					"pod_cidr": schema.StringAttribute{
						MarkdownDescription: "Pod CIDR in CIDR notation (e.g., 10.0.3.0/24)",
						Optional:            true,
					},
				},
			},
			"settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Kubernetes cluster settings",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"kubernetes_version": schema.StringAttribute{
						MarkdownDescription: "Kubernetes version. Available versions are described in the [ArubaCloud API documentation](https://api.arubacloud.com/docs/metadata#kubernetes-version). For example, `1.33.2`.",
						Required:            true,
					},
					"node_pools": schema.ListNestedAttribute{
						MarkdownDescription: "Node pools configuration",
						Required:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Node pool name",
									Required:            true,
								},
								"nodes": schema.Int64Attribute{
									MarkdownDescription: "Number of nodes in the node pool",
									Required:            true,
								},
								"instance": schema.StringAttribute{
									MarkdownDescription: "KaaS flavor name for nodes. Available flavors are described in the [ArubaCloud API documentation](https://api.arubacloud.com/docs/metadata#kaas-flavors). For example, `K2A4` means 2 CPU, 4GB RAM, and 40GB storage.",
									Required:            true,
								},
								"zone": schema.StringAttribute{
									MarkdownDescription: "Datacenter/zone code for nodes",
									Required:            true,
								},
								"autoscaling": schema.BoolAttribute{
									MarkdownDescription: "Enable autoscaling for node pool",
									Optional:            true,
								},
								"min_count": schema.Int64Attribute{
									MarkdownDescription: "Minimum number of nodes for autoscaling",
									Optional:            true,
								},
								"max_count": schema.Int64Attribute{
									MarkdownDescription: "Maximum number of nodes for autoscaling",
									Optional:            true,
								},
							},
						},
					},
					"ha": schema.BoolAttribute{
						MarkdownDescription: "High availability",
						Optional:            true,
					},
				},
			},
		},
	}
}

func (r *KaaSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KaaSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KaaSResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"Project ID is required to create a KaaS cluster",
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

	// Extract Network configuration
	var networkModel KaaSNetworkModel
	diags := data.Network.As(ctx, &networkModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use VPC and Subnet URIs from network config
	vpcURI := networkModel.VpcUriRef.ValueString()
	subnetURI := networkModel.SubnetUriRef.ValueString()

	// Extract Node CIDR from network config
	var nodeCIDRModel KaaSNodeCIDRModel
	diags = networkModel.NodeCIDR.As(ctx, &nodeCIDRModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract Settings configuration
	var settingsModel KaaSSettingsModel
	diags = data.Settings.As(ctx, &settingsModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract Node Pools from settings
	var nodePoolModels []KaaSNodePoolModel
	if !settingsModel.NodePools.IsNull() && !settingsModel.NodePools.IsUnknown() {
		diags := settingsModel.NodePools.ElementsAs(ctx, &nodePoolModels, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if len(nodePoolModels) == 0 {
		resp.Diagnostics.AddError(
			"Missing Node Pools",
			"At least one node pool is required",
		)
		return
	}

	// Build node pools
	nodePools := make([]sdktypes.NodePoolProperties, len(nodePoolModels))
	for i, np := range nodePoolModels {
		nodePool := sdktypes.NodePoolProperties{
			Name:        np.Name.ValueString(),
			Nodes:       int32(np.Nodes.ValueInt64()),
			Instance:    np.Instance.ValueString(),
			Zone:        np.Zone.ValueString(),
			Autoscaling: np.Autoscaling.ValueBool(),
		}
		if !np.MinCount.IsNull() && np.MinCount.ValueInt64() > 0 {
			minCount := int32(np.MinCount.ValueInt64())
			nodePool.MinCount = &minCount
		}
		if !np.MaxCount.IsNull() && np.MaxCount.ValueInt64() > 0 {
			maxCount := int32(np.MaxCount.ValueInt64())
			nodePool.MaxCount = &maxCount
		}
		nodePools[i] = nodePool
	}

	// Build the create request
	createRequest := sdktypes.KaaSRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.KaaSPropertiesRequest{
			VPC: sdktypes.ReferenceResource{
				URI: vpcURI,
			},
			Subnet: sdktypes.ReferenceResource{
				URI: subnetURI,
			},
			NodeCIDR: sdktypes.NodeCIDRProperties{
				Address: nodeCIDRModel.Address.ValueString(),
				Name:    nodeCIDRModel.Name.ValueString(),
			},
			SecurityGroup: sdktypes.SecurityGroupProperties{
				Name: networkModel.SecurityGroupName.ValueString(),
			},
			KubernetesVersion: sdktypes.KubernetesVersionInfo{
				Value: settingsModel.KubernetesVersion.ValueString(),
			},
			NodePools: nodePools,
		},
	}

	// Add optional fields
	if !networkModel.PodCIDR.IsNull() && !networkModel.PodCIDR.IsUnknown() {
		podCIDR := networkModel.PodCIDR.ValueString()
		createRequest.Properties.PodCIDR = &podCIDR
	}

	if !settingsModel.HA.IsNull() && !settingsModel.HA.IsUnknown() {
		ha := settingsModel.HA.ValueBool()
		createRequest.Properties.HA = &ha
	}

	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		createRequest.Properties.BillingPlan = sdktypes.BillingPeriodResource{
			BillingPeriod: data.BillingPeriod.ValueString(),
		}
	}

	// Create the KaaS cluster using the SDK
	response, err := r.client.Client.FromContainer().KaaS().Create(ctx, projectID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating KaaS cluster",
			fmt.Sprintf("Unable to create KaaS cluster: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		logContext := map[string]interface{}{}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to create KaaS cluster", logContext)
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

		// Build Network object from response
		networkAttrs := map[string]attr.Value{
			"vpc_uri_ref":         types.StringValue(vpcURI),
			"subnet_uri_ref":      types.StringValue(subnetURI),
			"security_group_name": types.StringValue(networkModel.SecurityGroupName.ValueString()),
		}

		// Set node_cidr
		nodeCIDRAttrs := map[string]attr.Value{
			"address": types.StringValue(nodeCIDRModel.Address.ValueString()),
			"name":    types.StringValue(nodeCIDRModel.Name.ValueString()),
		}
		nodeCIDRObj, diags := types.ObjectValue(map[string]attr.Type{
			"address": types.StringType,
			"name":    types.StringType,
		}, nodeCIDRAttrs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		networkAttrs["node_cidr"] = nodeCIDRObj

		// Set pod_cidr
		if !networkModel.PodCIDR.IsNull() && !networkModel.PodCIDR.IsUnknown() {
			networkAttrs["pod_cidr"] = types.StringValue(networkModel.PodCIDR.ValueString())
		} else {
			networkAttrs["pod_cidr"] = types.StringNull()
		}

		// Create Network object
		networkObj, diags := types.ObjectValue(map[string]attr.Type{
			"vpc_uri_ref":         types.StringType,
			"subnet_uri_ref":      types.StringType,
			"node_cidr":           types.ObjectType{AttrTypes: map[string]attr.Type{"address": types.StringType, "name": types.StringType}},
			"security_group_name": types.StringType,
			"pod_cidr":            types.StringType,
		}, networkAttrs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Network = networkObj

		// Build Settings object from response
		settingsAttrs := map[string]attr.Value{
			"kubernetes_version": types.StringValue(settingsModel.KubernetesVersion.ValueString()),
			"node_pools":         settingsModel.NodePools,
		}

		// Set HA
		if !settingsModel.HA.IsNull() && !settingsModel.HA.IsUnknown() {
			settingsAttrs["ha"] = types.BoolValue(settingsModel.HA.ValueBool())
		} else {
			settingsAttrs["ha"] = types.BoolNull()
		}

		// Create Settings object
		settingsObj, diags := types.ObjectValue(map[string]attr.Type{
			"kubernetes_version": types.StringType,
			"node_pools": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":        types.StringType,
						"nodes":       types.Int64Type,
						"instance":    types.StringType,
						"zone":        types.StringType,
						"autoscaling": types.BoolType,
						"min_count":   types.Int64Type,
						"max_count":   types.Int64Type,
					},
				},
			},
			"ha": types.BoolType,
		}, settingsAttrs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Settings = settingsObj
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"KaaS cluster created but no data returned from API",
		)
		return
	}

	// Wait for KaaS to be active before returning
	// This ensures Terraform doesn't proceed until KaaS is ready
	kaasID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromContainer().KaaS().Get(ctx, projectID, kaasID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for KaaS to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "KaaS", kaasID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"KaaS Not Active",
			fmt.Sprintf("KaaS cluster was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	// Re-read to get management IP now that KaaS is active
	finalGetResp, err := r.client.Client.FromContainer().KaaS().Get(ctx, projectID, kaasID, nil)
	if err == nil && finalGetResp != nil && finalGetResp.Data != nil {
		if finalGetResp.Data.Properties.ManagementIP != nil && *finalGetResp.Data.Properties.ManagementIP != "" {
			data.ManagementIP = types.StringValue(*finalGetResp.Data.Properties.ManagementIP)
		} else {
			data.ManagementIP = types.StringNull()
		}
	}

	tflog.Trace(ctx, "created a KaaS resource", map[string]interface{}{
		"kaas_id":   data.Id.ValueString(),
		"kaas_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KaaSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KaaSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve the original state to fallback for fields not returned by API
	var originalState KaaSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &originalState)...)

	projectID := data.ProjectID.ValueString()
	kaasID := data.Id.ValueString()

	if projectID == "" || kaasID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and KaaS ID are required to read the KaaS cluster",
		)
		return
	}

	// Get KaaS cluster details using the SDK
	response, err := r.client.Client.FromContainer().KaaS().Get(ctx, projectID, kaasID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading KaaS cluster",
			fmt.Sprintf("Unable to read KaaS cluster: %s", err),
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
			"kaas_id":    kaasID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to read KaaS cluster", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		kaas := response.Data

		if kaas.Metadata.ID != nil {
			data.Id = types.StringValue(*kaas.Metadata.ID)
		}
		if kaas.Metadata.URI != nil {
			data.Uri = types.StringValue(*kaas.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if kaas.Metadata.Name != nil {
			data.Name = types.StringValue(*kaas.Metadata.Name)
		}
		if kaas.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(kaas.Metadata.LocationResponse.Value)
		}

		if kaas.Properties.BillingPlan != nil && kaas.Properties.BillingPlan.BillingPeriod != nil {
			data.BillingPeriod = types.StringValue(*kaas.Properties.BillingPlan.BillingPeriod)
		}

		// Set Management IP if available
		if kaas.Properties.ManagementIP != nil && *kaas.Properties.ManagementIP != "" {
			data.ManagementIP = types.StringValue(*kaas.Properties.ManagementIP)
		} else {
			data.ManagementIP = types.StringNull()
		}

		// Build Network object
		networkAttrs := map[string]attr.Value{
			"vpc_uri_ref":         types.StringNull(),
			"subnet_uri_ref":      types.StringNull(),
			"node_cidr":           types.ObjectNull(map[string]attr.Type{"address": types.StringType, "name": types.StringType}),
			"security_group_name": types.StringNull(),
			"pod_cidr":            types.StringNull(),
		}

		if kaas.Properties.VPC.URI != nil && *kaas.Properties.VPC.URI != "" {
			networkAttrs["vpc_uri_ref"] = types.StringValue(*kaas.Properties.VPC.URI)
		}
		if kaas.Properties.Subnet.URI != nil && *kaas.Properties.Subnet.URI != "" {
			networkAttrs["subnet_uri_ref"] = types.StringValue(*kaas.Properties.Subnet.URI)
		}
		if kaas.Properties.SecurityGroup.Name != nil && *kaas.Properties.SecurityGroup.Name != "" {
			networkAttrs["security_group_name"] = types.StringValue(*kaas.Properties.SecurityGroup.Name)
		} else if !originalState.Network.IsNull() && !originalState.Network.IsUnknown() {
			// Preserve security_group_name from state if API doesn't return it
			var originalNetwork KaaSNetworkModel
			diags := originalState.Network.As(ctx, &originalNetwork, basetypes.ObjectAsOptions{})
			if !diags.HasError() && !originalNetwork.SecurityGroupName.IsNull() {
				networkAttrs["security_group_name"] = originalNetwork.SecurityGroupName
			}
		}
		if kaas.Properties.NodeCIDR.Address != nil && *kaas.Properties.NodeCIDR.Address != "" {
			nodeCIDRName := ""
			if kaas.Properties.NodeCIDR.Name != nil && *kaas.Properties.NodeCIDR.Name != "" {
				nodeCIDRName = *kaas.Properties.NodeCIDR.Name
			} else if !originalState.Network.IsNull() && !originalState.Network.IsUnknown() {
				// Preserve node_cidr.name from state if API doesn't return it
				var originalNetwork KaaSNetworkModel
				diags := originalState.Network.As(ctx, &originalNetwork, basetypes.ObjectAsOptions{})
				if !diags.HasError() && !originalNetwork.NodeCIDR.IsNull() {
					var originalNodeCIDR struct {
						Address types.String `tfsdk:"address"`
						Name    types.String `tfsdk:"name"`
					}
					diagsNodeCIDR := originalNetwork.NodeCIDR.As(ctx, &originalNodeCIDR, basetypes.ObjectAsOptions{})
					if !diagsNodeCIDR.HasError() && !originalNodeCIDR.Name.IsNull() {
						nodeCIDRName = originalNodeCIDR.Name.ValueString()
					}
				}
			}

			nodeCIDRObj, diags := types.ObjectValue(map[string]attr.Type{
				"address": types.StringType,
				"name":    types.StringType,
			}, map[string]attr.Value{
				"address": types.StringValue(*kaas.Properties.NodeCIDR.Address),
				"name":    types.StringValue(nodeCIDRName),
			})
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				networkAttrs["node_cidr"] = nodeCIDRObj
			}
		}
		if kaas.Properties.PodCIDR != nil && kaas.Properties.PodCIDR.Address != nil {
			networkAttrs["pod_cidr"] = types.StringValue(*kaas.Properties.PodCIDR.Address)
		}

		// Create Network object
		networkObj, diags := types.ObjectValue(map[string]attr.Type{
			"vpc_uri_ref":         types.StringType,
			"subnet_uri_ref":      types.StringType,
			"node_cidr":           types.ObjectType{AttrTypes: map[string]attr.Type{"address": types.StringType, "name": types.StringType}},
			"security_group_name": types.StringType,
			"pod_cidr":            types.StringType,
		}, networkAttrs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Network = networkObj

		// Build Settings object
		settingsAttrs := map[string]attr.Value{
			"kubernetes_version": types.StringNull(),
			"node_pools":         types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{"name": types.StringType, "nodes": types.Int64Type, "instance": types.StringType, "zone": types.StringType, "autoscaling": types.BoolType, "min_count": types.Int64Type, "max_count": types.Int64Type}}),
			"ha":                 types.BoolNull(),
		}

		if kaas.Properties.KubernetesVersion.Value != nil {
			settingsAttrs["kubernetes_version"] = types.StringValue(*kaas.Properties.KubernetesVersion.Value)
		}
		if kaas.Properties.HA != nil {
			settingsAttrs["ha"] = types.BoolValue(*kaas.Properties.HA)
		}

		// Build node pools
		if kaas.Properties.NodePools != nil && len(*kaas.Properties.NodePools) > 0 {
			nodePoolValues := make([]attr.Value, 0)
			for _, np := range *kaas.Properties.NodePools {
				nodePoolMap := map[string]attr.Value{
					"name": types.StringValue(func() string {
						if np.Name != nil {
							return *np.Name
						}
						return ""
					}()),
					"nodes": types.Int64Value(func() int64 {
						if np.Nodes != nil {
							return int64(*np.Nodes)
						}
						return 0
					}()),
					"instance": types.StringValue(func() string {
						if np.Instance != nil && np.Instance.Name != nil {
							return *np.Instance.Name
						}
						return ""
					}()),
					"zone": types.StringValue(func() string {
						if np.DataCenter != nil && np.DataCenter.Code != nil {
							return *np.DataCenter.Code
						}
						return ""
					}()),
					"autoscaling": types.BoolValue(np.Autoscaling),
				}
				if np.MinCount != nil {
					nodePoolMap["min_count"] = types.Int64Value(int64(*np.MinCount))
				} else {
					nodePoolMap["min_count"] = types.Int64Null()
				}
				if np.MaxCount != nil {
					nodePoolMap["max_count"] = types.Int64Value(int64(*np.MaxCount))
				} else {
					nodePoolMap["max_count"] = types.Int64Null()
				}

				nodePoolObj, diags := types.ObjectValue(map[string]attr.Type{
					"name":        types.StringType,
					"nodes":       types.Int64Type,
					"instance":    types.StringType,
					"zone":        types.StringType,
					"autoscaling": types.BoolType,
					"min_count":   types.Int64Type,
					"max_count":   types.Int64Type,
				}, nodePoolMap)
				resp.Diagnostics.Append(diags...)
				if !resp.Diagnostics.HasError() {
					nodePoolValues = append(nodePoolValues, nodePoolObj)
				}
			}
			nodePoolsList, diags := types.ListValue(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":        types.StringType,
					"nodes":       types.Int64Type,
					"instance":    types.StringType,
					"zone":        types.StringType,
					"autoscaling": types.BoolType,
					"min_count":   types.Int64Type,
					"max_count":   types.Int64Type,
				},
			}, nodePoolValues)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				settingsAttrs["node_pools"] = nodePoolsList
			}
		} else if !originalState.Settings.IsNull() && !originalState.Settings.IsUnknown() {
			// If API doesn't return node_pools, preserve from original state
			var originalSettings KaaSSettingsModel
			diags := originalState.Settings.As(ctx, &originalSettings, basetypes.ObjectAsOptions{})
			if !diags.HasError() && !originalSettings.NodePools.IsNull() {
				settingsAttrs["node_pools"] = originalSettings.NodePools
			}
		}

		// Create Settings object
		settingsObj, diags := types.ObjectValue(map[string]attr.Type{
			"kubernetes_version": types.StringType,
			"node_pools": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":        types.StringType,
						"nodes":       types.Int64Type,
						"instance":    types.StringType,
						"zone":        types.StringType,
						"autoscaling": types.BoolType,
						"min_count":   types.Int64Type,
						"max_count":   types.Int64Type,
					},
				},
			},
			"ha": types.BoolType,
		}, settingsAttrs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Settings = settingsObj

		// Update tags
		if len(kaas.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(kaas.Metadata.Tags))
			for i, tag := range kaas.Metadata.Tags {
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

func (r *KaaSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KaaSResourceModel
	var state KaaSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get IDs from state (not plan) - IDs are immutable and should always be in state
	projectID := state.ProjectID.ValueString()
	kaasID := state.Id.ValueString()

	if projectID == "" || kaasID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and KaaS ID are required to update the KaaS cluster",
		)
		return
	}

	// Get current KaaS cluster details
	getResponse, err := r.client.Client.FromContainer().KaaS().Get(ctx, projectID, kaasID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current KaaS cluster",
			fmt.Sprintf("Unable to get current KaaS cluster: %s", err),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"KaaS Cluster Not Found",
			"KaaS cluster not found or no data returned",
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
			"Unable to determine region value for KaaS cluster",
		)
		return
	}

	// Extract tags
	var tags []string
	var diags diag.Diagnostics
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags = data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		tags = current.Metadata.Tags
	}

	// Extract Settings configuration
	var settingsModel KaaSSettingsModel
	diags = data.Settings.As(ctx, &settingsModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract Node Pools from settings
	var nodePoolModels []KaaSNodePoolModel
	if !settingsModel.NodePools.IsNull() && !settingsModel.NodePools.IsUnknown() {
		diags := settingsModel.NodePools.ElementsAs(ctx, &nodePoolModels, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Build node pools
	nodePools := make([]sdktypes.NodePoolProperties, len(nodePoolModels))
	for i, np := range nodePoolModels {
		nodePool := sdktypes.NodePoolProperties{
			Name:        np.Name.ValueString(),
			Nodes:       int32(np.Nodes.ValueInt64()),
			Instance:    np.Instance.ValueString(),
			Zone:        np.Zone.ValueString(),
			Autoscaling: np.Autoscaling.ValueBool(),
		}
		if !np.MinCount.IsNull() && np.MinCount.ValueInt64() > 0 {
			minCount := int32(np.MinCount.ValueInt64())
			nodePool.MinCount = &minCount
		}
		if !np.MaxCount.IsNull() && np.MaxCount.ValueInt64() > 0 {
			maxCount := int32(np.MaxCount.ValueInt64())
			nodePool.MaxCount = &maxCount
		}
		nodePools[i] = nodePool
	}

	// Build Kubernetes version update
	kubernetesVersionValue := settingsModel.KubernetesVersion.ValueString()
	if kubernetesVersionValue == "" && current.Properties.KubernetesVersion.Value != nil {
		kubernetesVersionValue = *current.Properties.KubernetesVersion.Value
	}

	kubernetesVersionUpdate := sdktypes.KubernetesVersionInfoUpdate{
		Value: kubernetesVersionValue,
	}

	// Build update request
	updateRequest := sdktypes.KaaSUpdateRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.KaaSPropertiesUpdateRequest{
			KubernetesVersion: kubernetesVersionUpdate,
			NodePools:         nodePools,
		},
	}

	// Add optional fields
	if !settingsModel.HA.IsNull() && !settingsModel.HA.IsUnknown() {
		ha := settingsModel.HA.ValueBool()
		updateRequest.Properties.HA = &ha
	} else if current.Properties.HA != nil {
		updateRequest.Properties.HA = current.Properties.HA
	}

	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		updateRequest.Properties.BillingPlan = &sdktypes.BillingPeriodResource{
			BillingPeriod: data.BillingPeriod.ValueString(),
		}
	} else if current.Properties.BillingPlan != nil && current.Properties.BillingPlan.BillingPeriod != nil {
		updateRequest.Properties.BillingPlan = &sdktypes.BillingPeriodResource{
			BillingPeriod: *current.Properties.BillingPlan.BillingPeriod,
		}
	}

	// Update the KaaS cluster using the SDK
	response, err := r.client.Client.FromContainer().KaaS().Update(ctx, projectID, kaasID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating KaaS cluster",
			fmt.Sprintf("Unable to update KaaS cluster: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		logContext := map[string]interface{}{}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to update KaaS cluster", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.Uri = state.Uri // Preserve URI from state
	data.ProjectID = state.ProjectID
	// Don't overwrite Network and Settings yet - preserve from plan until we read from API

	if response != nil && response.Data != nil {
		// Update from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		}
	} else {
		// If no response, re-read the KaaS cluster to get the latest state including URI
		getResp, err := r.client.Client.FromContainer().KaaS().Get(ctx, projectID, kaasID, nil)
		if err == nil && getResp != nil && getResp.Data != nil {
			if getResp.Data.Metadata.URI != nil {
				data.Uri = types.StringValue(*getResp.Data.Metadata.URI)
			} else {
				data.Uri = state.Uri // Fallback to state if not in response
			}
		} else {
			// If re-read fails, preserve from state
			data.Uri = state.Uri
		}
	}

	// Re-read to get the latest state and update all fields
	getResp, err := r.client.Client.FromContainer().KaaS().Get(ctx, projectID, kaasID, nil)
	if err == nil && getResp != nil && getResp.Data != nil {
		kaas := getResp.Data
		// Update URI if available
		if kaas.Metadata.URI != nil {
			data.Uri = types.StringValue(*kaas.Metadata.URI)
		} else {
			data.Uri = state.Uri // Fallback to state if not available
		}
		// Update other fields from re-read to ensure consistency
		if kaas.Metadata.Name != nil {
			data.Name = types.StringValue(*kaas.Metadata.Name)
		}
		if kaas.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(kaas.Metadata.LocationResponse.Value)
		}
		if kaas.Properties.BillingPlan != nil && kaas.Properties.BillingPlan.BillingPeriod != nil {
			data.BillingPeriod = types.StringValue(*kaas.Properties.BillingPlan.BillingPeriod)
		} else {
			data.BillingPeriod = types.StringNull()
		}

		// Set Management IP if available
		if kaas.Properties.ManagementIP != nil && *kaas.Properties.ManagementIP != "" {
			data.ManagementIP = types.StringValue(*kaas.Properties.ManagementIP)
		} else {
			data.ManagementIP = types.StringNull()
		}

		// Build Network object from re-read, preserving fields not returned by API from the plan
		var planNetwork KaaSNetworkModel
		diagsNet := data.Network.As(ctx, &planNetwork, basetypes.ObjectAsOptions{})

		networkAttrs := map[string]attr.Value{
			"vpc_uri_ref":         types.StringNull(),
			"subnet_uri_ref":      types.StringNull(),
			"node_cidr":           types.ObjectNull(map[string]attr.Type{"address": types.StringType, "name": types.StringType}),
			"security_group_name": types.StringNull(),
			"pod_cidr":            types.StringNull(),
		}

		if kaas.Properties.VPC.URI != nil && *kaas.Properties.VPC.URI != "" {
			networkAttrs["vpc_uri_ref"] = types.StringValue(*kaas.Properties.VPC.URI)
		}
		if kaas.Properties.Subnet.URI != nil && *kaas.Properties.Subnet.URI != "" {
			networkAttrs["subnet_uri_ref"] = types.StringValue(*kaas.Properties.Subnet.URI)
		}

		// Preserve security_group_name from plan if API doesn't return it
		if kaas.Properties.SecurityGroup.Name != nil && *kaas.Properties.SecurityGroup.Name != "" {
			networkAttrs["security_group_name"] = types.StringValue(*kaas.Properties.SecurityGroup.Name)
		} else if !diagsNet.HasError() && !planNetwork.SecurityGroupName.IsNull() {
			networkAttrs["security_group_name"] = planNetwork.SecurityGroupName
		}

		// Build node_cidr, preserving name from plan if API doesn't return it
		if kaas.Properties.NodeCIDR.Address != nil && *kaas.Properties.NodeCIDR.Address != "" {
			nodeCIDRName := ""
			if kaas.Properties.NodeCIDR.Name != nil && *kaas.Properties.NodeCIDR.Name != "" {
				nodeCIDRName = *kaas.Properties.NodeCIDR.Name
			} else if !diagsNet.HasError() && !planNetwork.NodeCIDR.IsNull() {
				var planNodeCIDR struct {
					Address types.String `tfsdk:"address"`
					Name    types.String `tfsdk:"name"`
				}
				diagsNodeCIDR := planNetwork.NodeCIDR.As(ctx, &planNodeCIDR, basetypes.ObjectAsOptions{})
				if !diagsNodeCIDR.HasError() && !planNodeCIDR.Name.IsNull() {
					nodeCIDRName = planNodeCIDR.Name.ValueString()
				}
			}

			nodeCIDRObj, diags := types.ObjectValue(map[string]attr.Type{
				"address": types.StringType,
				"name":    types.StringType,
			}, map[string]attr.Value{
				"address": types.StringValue(*kaas.Properties.NodeCIDR.Address),
				"name":    types.StringValue(nodeCIDRName),
			})
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				networkAttrs["node_cidr"] = nodeCIDRObj
			}
		}

		if kaas.Properties.PodCIDR != nil && kaas.Properties.PodCIDR.Address != nil {
			networkAttrs["pod_cidr"] = types.StringValue(*kaas.Properties.PodCIDR.Address)
		}

		// Create Network object
		networkObj, diags := types.ObjectValue(map[string]attr.Type{
			"vpc_uri_ref":         types.StringType,
			"subnet_uri_ref":      types.StringType,
			"node_cidr":           types.ObjectType{AttrTypes: map[string]attr.Type{"address": types.StringType, "name": types.StringType}},
			"security_group_name": types.StringType,
			"pod_cidr":            types.StringType,
		}, networkAttrs)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Network = networkObj
		}

		// Build Settings object from re-read, preserving node_pools from plan if API doesn't return them
		var planSettings KaaSSettingsModel
		diagsSettings := data.Settings.As(ctx, &planSettings, basetypes.ObjectAsOptions{})

		settingsAttrs := map[string]attr.Value{
			"kubernetes_version": types.StringNull(),
			"node_pools":         types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{"name": types.StringType, "nodes": types.Int64Type, "instance": types.StringType, "zone": types.StringType, "autoscaling": types.BoolType, "min_count": types.Int64Type, "max_count": types.Int64Type}}),
			"ha":                 types.BoolNull(),
		}

		if kaas.Properties.KubernetesVersion.Value != nil {
			settingsAttrs["kubernetes_version"] = types.StringValue(*kaas.Properties.KubernetesVersion.Value)
		}
		if kaas.Properties.HA != nil {
			settingsAttrs["ha"] = types.BoolValue(*kaas.Properties.HA)
		}

		// Build node pools from re-read, or preserve from plan if not returned
		if kaas.Properties.NodePools != nil && len(*kaas.Properties.NodePools) > 0 {
			nodePoolValues := make([]attr.Value, 0)
			for _, np := range *kaas.Properties.NodePools {
				nodePoolMap := map[string]attr.Value{
					"name": types.StringValue(func() string {
						if np.Name != nil {
							return *np.Name
						}
						return ""
					}()),
					"nodes": types.Int64Value(func() int64 {
						if np.Nodes != nil {
							return int64(*np.Nodes)
						}
						return 0
					}()),
					"instance": types.StringValue(func() string {
						if np.Instance != nil && np.Instance.Name != nil {
							return *np.Instance.Name
						}
						return ""
					}()),
					"zone": types.StringValue(func() string {
						if np.DataCenter != nil && np.DataCenter.Code != nil {
							return *np.DataCenter.Code
						}
						return ""
					}()),
					"autoscaling": types.BoolValue(np.Autoscaling),
				}
				if np.MinCount != nil {
					nodePoolMap["min_count"] = types.Int64Value(int64(*np.MinCount))
				} else {
					nodePoolMap["min_count"] = types.Int64Null()
				}
				if np.MaxCount != nil {
					nodePoolMap["max_count"] = types.Int64Value(int64(*np.MaxCount))
				} else {
					nodePoolMap["max_count"] = types.Int64Null()
				}

				nodePoolObj, diags := types.ObjectValue(map[string]attr.Type{
					"name":        types.StringType,
					"nodes":       types.Int64Type,
					"instance":    types.StringType,
					"zone":        types.StringType,
					"autoscaling": types.BoolType,
					"min_count":   types.Int64Type,
					"max_count":   types.Int64Type,
				}, nodePoolMap)
				resp.Diagnostics.Append(diags...)
				if !resp.Diagnostics.HasError() {
					nodePoolValues = append(nodePoolValues, nodePoolObj)
				}
			}
			nodePoolsList, diags := types.ListValue(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":        types.StringType,
					"nodes":       types.Int64Type,
					"instance":    types.StringType,
					"zone":        types.StringType,
					"autoscaling": types.BoolType,
					"min_count":   types.Int64Type,
					"max_count":   types.Int64Type,
				},
			}, nodePoolValues)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				settingsAttrs["node_pools"] = nodePoolsList
			}
		} else if !diagsSettings.HasError() && !planSettings.NodePools.IsNull() {
			// Preserve node_pools from plan if API doesn't return them
			settingsAttrs["node_pools"] = planSettings.NodePools
		}

		// Create Settings object
		settingsObj, diags := types.ObjectValue(map[string]attr.Type{
			"kubernetes_version": types.StringType,
			"node_pools": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":        types.StringType,
						"nodes":       types.Int64Type,
						"instance":    types.StringType,
						"zone":        types.StringType,
						"autoscaling": types.BoolType,
						"min_count":   types.Int64Type,
						"max_count":   types.Int64Type,
					},
				},
			},
			"ha": types.BoolType,
		}, settingsAttrs)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Settings = settingsObj
		} else {
			data.Settings = state.Settings // Fallback to state on error
		}

		// Update tags from re-read
		if len(kaas.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(kaas.Metadata.Tags))
			for i, tag := range kaas.Metadata.Tags {
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
		// If re-read fails, preserve fields from state
		data.Uri = state.Uri
		data.Network = state.Network
		data.Settings = state.Settings
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KaaSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KaaSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	kaasID := data.Id.ValueString()

	if projectID == "" || kaasID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and KaaS ID are required to delete the KaaS cluster",
		)
		return
	}

	// Delete the KaaS cluster using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromContainer().KaaS().Delete(ctx, projectID, kaasID, nil)
		},
		ExtractSDKError,
		"KaaS",
		kaasID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting KaaS cluster",
			fmt.Sprintf("Unable to delete KaaS cluster: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a KaaS resource", map[string]interface{}{
		"kaas_id": kaasID,
	})
}

func (r *KaaSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
