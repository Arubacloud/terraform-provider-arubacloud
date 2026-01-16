// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
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
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	ProjectID     types.String `tfsdk:"project_id"`
	Network       types.Object `tfsdk:"network"`
	Settings      types.Object `tfsdk:"settings"`
	BillingPeriod types.String `tfsdk:"billing_period"`
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
			"network": schema.SingleNestedAttribute{
				MarkdownDescription: "Network configuration for the KaaS cluster",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"vpc_uri_ref": schema.StringAttribute{
						MarkdownDescription: "VPC URI reference (e.g., `arubacloud_vpc.example.uri`)",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"subnet_uri_ref": schema.StringAttribute{
						MarkdownDescription: "Subnet URI reference (e.g., `arubacloud_subnet.example.uri`)",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"security_group_name": schema.StringAttribute{
						MarkdownDescription: "Security group name (must match an existing security group)",
						Required:            true,
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
					"pod_cidr": schema.StringAttribute{
						MarkdownDescription: "Pod CIDR in CIDR notation (e.g., 10.0.3.0/24)",
						Optional:            true,
					},
				},
			},
			"settings": schema.SingleNestedAttribute{
				MarkdownDescription: "KaaS cluster settings",
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
					"controlplane_ha": schema.BoolAttribute{
						MarkdownDescription: "Control plane high availability",
						Optional:            true,
					},
				},
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period (Hour, Month, Year)",
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

	// Extract network configuration from nested object
	if data.Network.IsNull() || data.Network.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing Network Configuration",
			"Network configuration is required to create a KaaS cluster",
		)
		return
	}

	networkObj, diags := data.Network.ToObjectValue(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	networkAttrs := networkObj.Attributes()
	vpcURI := ""
	subnetURI := ""
	securityGroupName := ""
	var nodeCIDRModel KaaSNodeCIDRModel
	podCIDR := ""

	if vpcAttr, ok := networkAttrs["vpc_uri_ref"]; ok && vpcAttr != nil {
		if vpcStr, ok := vpcAttr.(types.String); ok && !vpcStr.IsNull() && !vpcStr.IsUnknown() {
			vpcURI = vpcStr.ValueString()
		}
	}
	if subnetAttr, ok := networkAttrs["subnet_uri_ref"]; ok && subnetAttr != nil {
		if subnetStr, ok := subnetAttr.(types.String); ok && !subnetStr.IsNull() && !subnetStr.IsUnknown() {
			subnetURI = subnetStr.ValueString()
		}
	}
	if securityGroupAttr, ok := networkAttrs["security_group_name"]; ok && securityGroupAttr != nil {
		if securityGroupStr, ok := securityGroupAttr.(types.String); ok && !securityGroupStr.IsNull() && !securityGroupStr.IsUnknown() {
			securityGroupName = securityGroupStr.ValueString()
		}
	}
	if nodeCIDRAttr, ok := networkAttrs["node_cidr"]; ok && nodeCIDRAttr != nil {
		if nodeCIDRObj, ok := nodeCIDRAttr.(types.Object); ok && !nodeCIDRObj.IsNull() && !nodeCIDRObj.IsUnknown() {
			diags := nodeCIDRObj.As(ctx, &nodeCIDRModel, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}
	if podCIDRAttr, ok := networkAttrs["pod_cidr"]; ok && podCIDRAttr != nil {
		if podCIDRStr, ok := podCIDRAttr.(types.String); ok && !podCIDRStr.IsNull() && !podCIDRStr.IsUnknown() {
			podCIDR = podCIDRStr.ValueString()
		}
	}

	// Extract settings configuration from nested object
	if data.Settings.IsNull() || data.Settings.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing Settings Configuration",
			"Settings configuration is required to create a KaaS cluster",
		)
		return
	}

	settingsObj, diags := data.Settings.ToObjectValue(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	settingsAttrs := settingsObj.Attributes()
	kubernetesVersion := ""
	var nodePoolModels []KaaSNodePoolModel
	ha := false

	if kubernetesVersionAttr, ok := settingsAttrs["kubernetes_version"]; ok && kubernetesVersionAttr != nil {
		if kubernetesVersionStr, ok := kubernetesVersionAttr.(types.String); ok && !kubernetesVersionStr.IsNull() && !kubernetesVersionStr.IsUnknown() {
			kubernetesVersion = kubernetesVersionStr.ValueString()
		}
	}
	if nodePoolsAttr, ok := settingsAttrs["node_pools"]; ok && nodePoolsAttr != nil {
		if nodePoolsList, ok := nodePoolsAttr.(types.List); ok && !nodePoolsList.IsNull() && !nodePoolsList.IsUnknown() {
			diags := nodePoolsList.ElementsAs(ctx, &nodePoolModels, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}
	if haAttr, ok := settingsAttrs["controlplane_ha"]; ok && haAttr != nil {
		if haBool, ok := haAttr.(types.Bool); ok && !haBool.IsNull() && !haBool.IsUnknown() {
			ha = haBool.ValueBool()
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
				Name: securityGroupName,
			},
			KubernetesVersion: sdktypes.KubernetesVersionInfo{
				Value: kubernetesVersion,
			},
			NodePools: nodePools,
		},
	}

	// Add optional fields
	if podCIDR != "" {
		createRequest.Properties.PodCIDR = &podCIDR
	}

	// Always set HA field (defaults to false if not specified)
	createRequest.Properties.HA = &ha

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

		// Log detailed error information for debugging
		errorDetails := map[string]interface{}{
			"project_id":  projectID,
			"status_code": response.StatusCode,
		}
		if response.Error.Title != nil {
			errorDetails["error_title"] = *response.Error.Title
		}
		if response.Error.Detail != nil {
			errorDetails["error_detail"] = *response.Error.Detail
		}
		if response.Error.Status != nil {
			errorDetails["error_status"] = *response.Error.Status
		}
		if response.Error.Type != nil {
			errorDetails["error_type"] = *response.Error.Type
		}

		// Log full request and error response JSON for debugging
		if requestJSON, jsonErr := json.MarshalIndent(createRequest, "", "  "); jsonErr == nil {
			tflog.Debug(ctx, "Full KaaS create request JSON (error case)", map[string]interface{}{
				"request_json": string(requestJSON),
			})
		}
		if errorJSON, jsonErr := json.MarshalIndent(response.Error, "", "  "); jsonErr == nil {
			tflog.Debug(ctx, "Full API error response JSON", map[string]interface{}{
				"error_json": string(errorJSON),
			})
		}

		tflog.Error(ctx, "KaaS create request failed", errorDetails)
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

		// Network and settings are preserved from state (they're immutable and not fully returned by API)
		// The network and settings objects are already set from req.State.Get above
		if kaas.Properties.BillingPlan != nil && kaas.Properties.BillingPlan.BillingPeriod != nil {
			data.BillingPeriod = types.StringValue(*kaas.Properties.BillingPlan.BillingPeriod)
		}

		// Network and settings are preserved from state (they're immutable and not fully returned by API)
		// No need to update them from API response

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

	// Extract settings configuration from plan or state
	var nodePoolModels []KaaSNodePoolModel
	kubernetesVersionValue := ""
	ha := false

	if !data.Settings.IsNull() && !data.Settings.IsUnknown() {
		settingsObj, diags := data.Settings.ToObjectValue(ctx)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			settingsAttrs := settingsObj.Attributes()
			if kubernetesVersionAttr, ok := settingsAttrs["kubernetes_version"]; ok && kubernetesVersionAttr != nil {
				if kubernetesVersionStr, ok := kubernetesVersionAttr.(types.String); ok && !kubernetesVersionStr.IsNull() && !kubernetesVersionStr.IsUnknown() {
					kubernetesVersionValue = kubernetesVersionStr.ValueString()
				}
			}
			if nodePoolsAttr, ok := settingsAttrs["node_pools"]; ok && nodePoolsAttr != nil {
				if nodePoolsList, ok := nodePoolsAttr.(types.List); ok && !nodePoolsList.IsNull() && !nodePoolsList.IsUnknown() {
					diags := nodePoolsList.ElementsAs(ctx, &nodePoolModels, false)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
				}
			}
			if haAttr, ok := settingsAttrs["controlplane_ha"]; ok && haAttr != nil {
				if haBool, ok := haAttr.(types.Bool); ok && !haBool.IsNull() && !haBool.IsUnknown() {
					ha = haBool.ValueBool()
				}
			}
		}
	}

	// Fallback to current state if not in plan
	if kubernetesVersionValue == "" && current.Properties.KubernetesVersion.Value != nil {
		kubernetesVersionValue = *current.Properties.KubernetesVersion.Value
	}
	if !ha && current.Properties.HA != nil {
		ha = *current.Properties.HA
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
	if ha {
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
	// Preserve network and settings from state (they're immutable)
	data.Network = state.Network
	data.Settings = state.Settings

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

	// Re-read to get the latest state and update mutable fields
	getResp, err := r.client.Client.FromContainer().KaaS().Get(ctx, projectID, kaasID, nil)
	if err == nil && getResp != nil && getResp.Data != nil {
		kaas := getResp.Data
		// Update URI if available
		if kaas.Metadata.URI != nil {
			data.Uri = types.StringValue(*kaas.Metadata.URI)
		} else {
			data.Uri = state.Uri // Fallback to state if not available
		}
		// Update other mutable fields from re-read
		if kaas.Metadata.Name != nil {
			data.Name = types.StringValue(*kaas.Metadata.Name)
		}
		if kaas.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(kaas.Metadata.LocationResponse.Value)
		}
		if kaas.Properties.BillingPlan != nil && kaas.Properties.BillingPlan.BillingPeriod != nil {
			data.BillingPeriod = types.StringValue(*kaas.Properties.BillingPlan.BillingPeriod)
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
		// Network and settings are preserved from state (they're immutable)
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
