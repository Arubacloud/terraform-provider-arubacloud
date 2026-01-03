// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	Id                types.String `tfsdk:"id"`
	Uri               types.String `tfsdk:"uri"`
	Name              types.String `tfsdk:"name"`
	Location          types.String `tfsdk:"location"`
	Tags              types.List   `tfsdk:"tags"`
	ProjectID         types.String `tfsdk:"project_id"`
	VpcUriRef         types.String `tfsdk:"vpc_uri_ref"`
	SubnetUriRef      types.String `tfsdk:"subnet_uri_ref"`
	NodeCIDR          types.Object `tfsdk:"node_cidr"`
	SecurityGroupName types.String `tfsdk:"security_group_name"`
	KubernetesVersion types.String `tfsdk:"kubernetes_version"`
	NodePools         types.List   `tfsdk:"node_pools"`
	HA                types.Bool   `tfsdk:"ha"`
	BillingPeriod     types.String `tfsdk:"billing_period"`
	PodCIDR           types.String `tfsdk:"pod_cidr"`
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
			"vpc_uri_ref": schema.StringAttribute{
				MarkdownDescription: "VPC URI reference for the KaaS resource (e.g., /projects/{project-id}/providers/Aruba.Network/vpcs/{vpc-id})",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subnet_uri_ref": schema.StringAttribute{
				MarkdownDescription: "Subnet URI reference for the KaaS resource (e.g., /projects/{project-id}/providers/Aruba.Network/subnets/{subnet-id})",
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
						MarkdownDescription: "Node CIDR address in CIDR notation (e.g., 10.0.0.0/16)",
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
			"kubernetes_version": schema.StringAttribute{
				MarkdownDescription: "Kubernetes version (e.g., 1.28.0)",
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
							MarkdownDescription: "Instance configuration name for nodes",
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
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period (Hour, Month, Year)",
				Optional:            true,
			},
			"pod_cidr": schema.StringAttribute{
				MarkdownDescription: "Pod CIDR",
				Optional:            true,
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

	// Use VPC and Subnet URIs directly from the plan
	vpcURI := data.VpcUriRef.ValueString()
	subnetURI := data.SubnetUriRef.ValueString()

	// Extract Node CIDR
	var nodeCIDRModel KaaSNodeCIDRModel
	diags := data.NodeCIDR.As(ctx, &nodeCIDRModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract Node Pools
	var nodePoolModels []KaaSNodePoolModel
	if !data.NodePools.IsNull() && !data.NodePools.IsUnknown() {
		diags := data.NodePools.ElementsAs(ctx, &nodePoolModels, false)
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
				Name: data.SecurityGroupName.ValueString(),
			},
			KubernetesVersion: sdktypes.KubernetesVersionInfo{
				Value: data.KubernetesVersion.ValueString(),
			},
			NodePools: nodePools,
		},
	}

	// Add optional fields
	if !data.PodCIDR.IsNull() && !data.PodCIDR.IsUnknown() {
		podCIDR := data.PodCIDR.ValueString()
		createRequest.Properties.PodCIDR = &podCIDR
	}

	if !data.HA.IsNull() && !data.HA.IsUnknown() {
		ha := data.HA.ValueBool()
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
		errorMsg := "Failed to create KaaS cluster"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
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
		errorMsg := "Failed to read KaaS cluster"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
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
		if kaas.Properties.KubernetesVersion.Value != nil {
			data.KubernetesVersion = types.StringValue(*kaas.Properties.KubernetesVersion.Value)
		}
		if kaas.Properties.VPC.URI != nil && *kaas.Properties.VPC.URI != "" {
			data.VpcUriRef = types.StringValue(*kaas.Properties.VPC.URI)
		}
		if kaas.Properties.Subnet.URI != nil && *kaas.Properties.Subnet.URI != "" {
			data.SubnetUriRef = types.StringValue(*kaas.Properties.Subnet.URI)
		}
		if kaas.Properties.SecurityGroup.Name != nil && *kaas.Properties.SecurityGroup.Name != "" {
			data.SecurityGroupName = types.StringValue(*kaas.Properties.SecurityGroup.Name)
		}
		if kaas.Properties.NodeCIDR.Address != nil && *kaas.Properties.NodeCIDR.Address != "" {
			nodeCIDRObj, diags := types.ObjectValue(map[string]attr.Type{
				"address": types.StringType,
				"name":    types.StringType,
			}, map[string]attr.Value{
				"address": types.StringValue(*kaas.Properties.NodeCIDR.Address),
				"name": types.StringValue(func() string {
					if kaas.Properties.NodeCIDR.Name != nil {
						return *kaas.Properties.NodeCIDR.Name
					}
					return ""
				}()),
			})
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.NodeCIDR = nodeCIDRObj
			}
		}
		if kaas.Properties.HA != nil {
			data.HA = types.BoolValue(*kaas.Properties.HA)
		}
		if kaas.Properties.PodCIDR != nil && kaas.Properties.PodCIDR.Address != nil {
			data.PodCIDR = types.StringValue(*kaas.Properties.PodCIDR.Address)
		}
		if kaas.Properties.BillingPlan != nil && kaas.Properties.BillingPlan.BillingPeriod != nil {
			data.BillingPeriod = types.StringValue(*kaas.Properties.BillingPlan.BillingPeriod)
		}

		// Update node pools
		if kaas.Properties.NodePools != nil {
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
				data.NodePools = nodePoolsList
			}
		}

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
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		tags = current.Metadata.Tags
	}

	// Extract Node Pools
	var nodePoolModels []KaaSNodePoolModel
	if !data.NodePools.IsNull() && !data.NodePools.IsUnknown() {
		diags := data.NodePools.ElementsAs(ctx, &nodePoolModels, false)
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
	kubernetesVersionValue := data.KubernetesVersion.ValueString()
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
	if !data.HA.IsNull() && !data.HA.IsUnknown() {
		ha := data.HA.ValueBool()
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
		errorMsg := "Failed to update KaaS cluster"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.Uri = state.Uri // Preserve URI from state
	data.ProjectID = state.ProjectID
	data.VpcUriRef = state.VpcUriRef
	data.SubnetUriRef = state.SubnetUriRef

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
		// Update VPC and Subnet URI refs
		if kaas.Properties.VPC.URI != nil {
			data.VpcUriRef = types.StringValue(*kaas.Properties.VPC.URI)
		} else {
			data.VpcUriRef = state.VpcUriRef // Fallback to state if not available
		}
		if kaas.Properties.Subnet.URI != nil {
			data.SubnetUriRef = types.StringValue(*kaas.Properties.Subnet.URI)
		} else {
			data.SubnetUriRef = state.SubnetUriRef // Fallback to state if not available
		}
		// Update other fields from re-read to ensure consistency
		if kaas.Metadata.Name != nil {
			data.Name = types.StringValue(*kaas.Metadata.Name)
		}
		if kaas.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(kaas.Metadata.LocationResponse.Value)
		}
		if kaas.Properties.KubernetesVersion.Value != nil {
			data.KubernetesVersion = types.StringValue(*kaas.Properties.KubernetesVersion.Value)
		}
		if kaas.Properties.SecurityGroup.Name != nil {
			data.SecurityGroupName = types.StringValue(*kaas.Properties.SecurityGroup.Name)
		}
		if kaas.Properties.NodeCIDR.Address != nil && kaas.Properties.NodeCIDR.Name != nil {
			nodeCIDRObj, diags := types.ObjectValue(map[string]attr.Type{
				"address": types.StringType,
				"name":    types.StringType,
			}, map[string]attr.Value{
				"address": types.StringValue(*kaas.Properties.NodeCIDR.Address),
				"name":    types.StringValue(*kaas.Properties.NodeCIDR.Name),
			})
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.NodeCIDR = nodeCIDRObj
			}
		}
		if kaas.Properties.HA != nil {
			data.HA = types.BoolValue(*kaas.Properties.HA)
		} else {
			data.HA = types.BoolNull()
		}
		if kaas.Properties.PodCIDR != nil && kaas.Properties.PodCIDR.Address != nil {
			data.PodCIDR = types.StringValue(*kaas.Properties.PodCIDR.Address)
		} else {
			data.PodCIDR = types.StringNull()
		}
		if kaas.Properties.BillingPlan != nil && kaas.Properties.BillingPlan.BillingPeriod != nil {
			data.BillingPeriod = types.StringValue(*kaas.Properties.BillingPlan.BillingPeriod)
		} else {
			data.BillingPeriod = types.StringNull()
		}
		// Update node pools from re-read
		if kaas.Properties.NodePools != nil {
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
				data.NodePools = nodePoolsList
			}
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
		// If re-read fails, preserve immutable fields from state
		data.Uri = state.Uri
		data.VpcUriRef = state.VpcUriRef
		data.SubnetUriRef = state.SubnetUriRef
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
