package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
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
	BillingPeriod types.String `tfsdk:"billing_period"`
	ManagementIP  types.String `tfsdk:"management_ip"`
	Kubeconfig    types.String `tfsdk:"kubeconfig"`
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
		MarkdownDescription: "Manages an ArubaCloud Kubernetes cluster (KaaS — Kubernetes-as-a-Service).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the KaaS cluster.",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
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
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle. Accepted values: `Hour`, `Month`, `Year`.",
				Optional:            true,
			},
			"management_ip": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Management IP address of the cluster control plane, available once the cluster is active.",
				Computed:            true,
			},
			"kubeconfig": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Kubeconfig YAML for kubectl access, downloaded when the cluster is active. Write-only — this value is sent to the API but is not returned in subsequent read responses.",
				Computed:            true,
				Sensitive:           true,
			},
			"network": schema.SingleNestedAttribute{
				MarkdownDescription: "Network configuration for the KaaS cluster.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"vpc_uri_ref": schema.StringAttribute{
						MarkdownDescription: "URI of the VPC that hosts the cluster (e.g., `arubacloud_vpc.example.uri`).",
						Required:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"subnet_uri_ref": schema.StringAttribute{
						MarkdownDescription: "URI of the subnet within the VPC (e.g., `arubacloud_subnet.example.uri`).",
						Required:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"node_cidr": schema.SingleNestedAttribute{
						MarkdownDescription: "CIDR block assigned to cluster nodes.",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"address": schema.StringAttribute{
								MarkdownDescription: "Node CIDR address in CIDR notation (e.g., `10.0.0.0/24`).",
								Required:            true,
							},
							"name": schema.StringAttribute{
								MarkdownDescription: "Human-readable label for the node CIDR block.",
								Required:            true,
							},
						},
					},
					"security_group_name": schema.StringAttribute{
						MarkdownDescription: "Name of the security group applied to cluster nodes.",
						Required:            true,
					},
					"pod_cidr": schema.StringAttribute{
						MarkdownDescription: "CIDR block used for pod networking within the cluster (e.g., `10.0.3.0/24`).",
						Optional:            true,
					},
				},
			},
			"settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Kubernetes version and node-pool configuration.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"kubernetes_version": schema.StringAttribute{
						MarkdownDescription: "Kubernetes version string (e.g., `1.28`). Available versions are listed in the ArubaCloud metadata API.",
						Required:            true,
					},
					"node_pools": schema.ListNestedAttribute{
						MarkdownDescription: "One or more node pools that make up the cluster worker fleet.",
						Required:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Display name for the node pool.",
									Required:            true,
								},
								"nodes": schema.Int64Attribute{
									MarkdownDescription: "Number of worker nodes in the cluster.",
									Required:            true,
								},
								"instance": schema.StringAttribute{
									MarkdownDescription: "Compute flavour for cluster nodes (e.g., `CSO4A8`). See [available flavours](https://api.arubacloud.com/docs/metadata/#cloudserver-flavors).",
									Required:            true,
								},
								"zone": schema.StringAttribute{
									MarkdownDescription: "Datacenter zone code where the node pool is deployed.",
									Required:            true,
								},
								"autoscaling": schema.BoolAttribute{
									MarkdownDescription: "When `true`, the node pool scales automatically between `min_count` and `max_count`.",
									Optional:            true,
								},
								"min_count": schema.Int64Attribute{
									MarkdownDescription: "Minimum number of nodes when autoscaling is enabled.",
									Optional:            true,
								},
								"max_count": schema.Int64Attribute{
									MarkdownDescription: "Maximum number of nodes when autoscaling is enabled.",
									Optional:            true,
								},
							},
						},
					},
					"ha": schema.BoolAttribute{
						MarkdownDescription: "When `true`, the control plane is deployed in high-availability mode.",
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

func kaasRef(data *KaaSResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.URI("/projects/" + data.ProjectID.ValueString() +
		"/providers/Aruba.Container/kaas/" + data.Id.ValueString())
}

// buildNodePools converts Terraform node pool models to aruba.NewNodePool() builders.
func buildNodePools(nodePoolModels []KaaSNodePoolModel) []*aruba.NodePool {
	pools := make([]*aruba.NodePool, len(nodePoolModels))
	for i, np := range nodePoolModels {
		pool := aruba.NewNodePool().
			Named(np.Name.ValueString()).
			WithCount(int(np.Nodes.ValueInt64())).
			OfInstance(aruba.NodePoolInstance(np.Instance.ValueString())).
			InZone(aruba.Zone(np.Zone.ValueString()))
		if !np.Autoscaling.IsNull() && np.Autoscaling.ValueBool() &&
			!np.MinCount.IsNull() && !np.MaxCount.IsNull() {
			pool = pool.WithAutoscaling(int(np.MinCount.ValueInt64()), int(np.MaxCount.ValueInt64()))
		}
		pools[i] = pool
	}
	return pools
}

// nodePoolAttrTypes returns the attr.Type map for the node_pools list element.
func nodePoolAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":        types.StringType,
		"nodes":       types.Int64Type,
		"instance":    types.StringType,
		"zone":        types.StringType,
		"autoscaling": types.BoolType,
		"min_count":   types.Int64Type,
		"max_count":   types.Int64Type,
	}
}

// downloadKubeconfig downloads and decodes the kubeconfig from the KaaS wrapper.
func downloadKubeconfig(ctx context.Context, kaas *aruba.KaaS) types.String {
	kb, kerr := kaas.DownloadKubeconfig(ctx)
	if kerr != nil || len(kb) == 0 {
		return types.StringNull()
	}
	// API returns base64-encoded content.
	if decoded, decErr := base64.StdEncoding.DecodeString(string(kb)); decErr == nil {
		return types.StringValue(string(decoded))
	}
	// Fall back to raw if not base64.
	return types.StringValue(string(kb))
}

// buildNodePoolAttrValues converts raw API NodePoolPropertiesResponse to Terraform attr.Value slice.
func buildNodePoolAttrValues(raw *aruba.KaaS, originalNodePools types.List) (types.List, bool) {
	resp := raw.Raw()
	if resp == nil || resp.Properties.NodePools == nil || len(*resp.Properties.NodePools) == 0 {
		return originalNodePools, false
	}
	nodePoolObjType := types.ObjectType{AttrTypes: nodePoolAttrTypes()}
	nodePoolValues := make([]attr.Value, 0, len(*resp.Properties.NodePools))
	for _, np := range *resp.Properties.NodePools {
		instanceName := ""
		if np.Instance != nil && np.Instance.Name != nil {
			instanceName = *np.Instance.Name
		}
		zoneCode := ""
		if np.DataCenter != nil && np.DataCenter.Code != nil {
			zoneCode = *np.DataCenter.Code
		}
		nodes := int64(0)
		if np.Nodes != nil {
			nodes = int64(*np.Nodes)
		}
		npName := ""
		if np.Name != nil {
			npName = *np.Name
		}
		nodePoolMap := map[string]attr.Value{
			"name":        types.StringValue(npName),
			"nodes":       types.Int64Value(nodes),
			"instance":    types.StringValue(instanceName),
			"zone":        types.StringValue(zoneCode),
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
		obj, d := types.ObjectValue(nodePoolAttrTypes(), nodePoolMap)
		if !d.HasError() {
			nodePoolValues = append(nodePoolValues, obj)
		}
	}
	list, d := types.ListValue(nodePoolObjType, nodePoolValues)
	if d.HasError() {
		return originalNodePools, false
	}
	return list, true
}

func (r *KaaSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KaaSResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var networkModel KaaSNetworkModel
	resp.Diagnostics.Append(data.Network.As(ctx, &networkModel, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	var nodeCIDRModel KaaSNodeCIDRModel
	resp.Diagnostics.Append(networkModel.NodeCIDR.As(ctx, &nodeCIDRModel, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	var settingsModel KaaSSettingsModel
	resp.Diagnostics.Append(data.Settings.As(ctx, &settingsModel, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	var nodePoolModels []KaaSNodePoolModel
	if !settingsModel.NodePools.IsNull() && !settingsModel.NodePools.IsUnknown() {
		resp.Diagnostics.Append(settingsModel.NodePools.ElementsAs(ctx, &nodePoolModels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if len(nodePoolModels) == 0 {
		resp.Diagnostics.AddError("Missing Node Pools", "At least one node pool is required")
		return
	}

	builder := aruba.NewKaaS().
		Named(data.Name.ValueString()).
		InProject(aruba.URI("/projects/"+projectID)).
		InRegion(aruba.Region(data.Location.ValueString())).
		WithKubernetesVersion(aruba.KubernetesVersion(settingsModel.KubernetesVersion.ValueString())).
		WithVPC(aruba.URI(networkModel.VpcUriRef.ValueString())).
		WithSubnet(aruba.URI(networkModel.SubnetUriRef.ValueString())).
		WithNodeCIDR(nodeCIDRModel.Address.ValueString(), nodeCIDRModel.Name.ValueString()).
		WithSecurityGroupName(networkModel.SecurityGroupName.ValueString()).
		WithNodePools(buildNodePools(nodePoolModels)...).
		Tagged(tags...)

	if !networkModel.PodCIDR.IsNull() && !networkModel.PodCIDR.IsUnknown() {
		builder = builder.WithPodCIDR(networkModel.PodCIDR.ValueString())
	}
	if !settingsModel.HA.IsNull() && !settingsModel.HA.IsUnknown() && settingsModel.HA.ValueBool() {
		builder = builder.HighlyAvailable()
	}
	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		builder = builder.BilledBy(aruba.BillingPeriod(data.BillingPeriod.ValueString()))
	}

	kaas, err := r.client.Client.FromContainer().KaaS().Create(ctx, builder)
	if provErr := CheckResponseErr("create", "KaaS", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(kaas.ID())
	data.Uri = strVal(kaas.URI())

	// Build Network and Settings objects from plan (to preserve in state before wait).
	networkAttrs := map[string]attr.Value{
		"vpc_uri_ref":         types.StringValue(networkModel.VpcUriRef.ValueString()),
		"subnet_uri_ref":      types.StringValue(networkModel.SubnetUriRef.ValueString()),
		"security_group_name": types.StringValue(networkModel.SecurityGroupName.ValueString()),
	}
	if !networkModel.PodCIDR.IsNull() {
		networkAttrs["pod_cidr"] = types.StringValue(networkModel.PodCIDR.ValueString())
	} else {
		networkAttrs["pod_cidr"] = types.StringNull()
	}
	nodeCIDRAttrs := map[string]attr.Value{
		"address": types.StringValue(nodeCIDRModel.Address.ValueString()),
		"name":    types.StringValue(nodeCIDRModel.Name.ValueString()),
	}
	nodeCIDRObj, d := types.ObjectValue(map[string]attr.Type{"address": types.StringType, "name": types.StringType}, nodeCIDRAttrs)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	networkAttrs["node_cidr"] = nodeCIDRObj
	networkObj, d := types.ObjectValue(map[string]attr.Type{
		"vpc_uri_ref": types.StringType, "subnet_uri_ref": types.StringType,
		"node_cidr":           types.ObjectType{AttrTypes: map[string]attr.Type{"address": types.StringType, "name": types.StringType}},
		"security_group_name": types.StringType, "pod_cidr": types.StringType,
	}, networkAttrs)
	resp.Diagnostics.Append(d...)
	if !resp.Diagnostics.HasError() {
		data.Network = networkObj
	}

	settingsAttrs := map[string]attr.Value{
		"kubernetes_version": types.StringValue(settingsModel.KubernetesVersion.ValueString()),
		"node_pools":         settingsModel.NodePools,
	}
	if !settingsModel.HA.IsNull() {
		settingsAttrs["ha"] = settingsModel.HA
	} else {
		settingsAttrs["ha"] = types.BoolNull()
	}
	nodePoolListType := types.ListType{ElemType: types.ObjectType{AttrTypes: nodePoolAttrTypes()}}
	settingsObj, d := types.ObjectValue(map[string]attr.Type{
		"kubernetes_version": types.StringType, "node_pools": nodePoolListType, "ha": types.BoolType,
	}, settingsAttrs)
	resp.Diagnostics.Append(d...)
	if !resp.Diagnostics.HasError() {
		data.Settings = settingsObj
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if waitErr := kaas.WaitUntilReady(ctx, aruba.WithTimeout(r.client.ResourceTimeout)); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "KaaS", data.Id.ValueString())
		data.Kubeconfig = types.StringNull()
		data.ManagementIP = types.StringNull()
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Re-read to get management IP.
	fresh, freshErr := r.client.Client.FromContainer().KaaS().Get(ctx, kaasRef(&data))
	if freshErr == nil {
		raw := fresh.Raw()
		if raw != nil && raw.Properties.ManagementIP != nil && *raw.Properties.ManagementIP != "" {
			data.ManagementIP = types.StringValue(*raw.Properties.ManagementIP)
		} else {
			data.ManagementIP = types.StringNull()
		}
		data.Kubeconfig = downloadKubeconfig(ctx, fresh)
	} else {
		data.ManagementIP = types.StringNull()
		data.Kubeconfig = types.StringNull()
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
	var originalState KaaSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &originalState)...)

	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	kaas, err := r.client.Client.FromContainer().KaaS().Get(ctx, kaasRef(&data))
	if provErr := CheckResponseErr("read", "KaaS", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	st := string(kaas.State())
	switch {
	case isFailedState(st):
		resp.Diagnostics.AddWarning("Resource in Failed State",
			fmt.Sprintf("KaaS %q is in a terminal failure state (%s). "+
				"Run `terraform destroy` to clean it up, or `terraform apply -replace=<address>` to recreate it.", data.Id.ValueString(), st))
	case IsCreatingState(st):
		if waitErr := kaas.WaitUntilReady(ctx, aruba.WithTimeout(r.client.ResourceTimeout)); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "KaaS", data.Id.ValueString())
			return
		}
		kaas, err = r.client.Client.FromContainer().KaaS().Get(ctx, kaasRef(&data))
		if provErr := CheckResponseErr("read", "KaaS", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	raw := kaas.Raw()
	if raw == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Id = types.StringValue(kaas.ID())
	data.Uri = strVal(kaas.URI())
	data.Name = types.StringValue(kaas.Name())
	data.Tags = TagsToListPreserveNull(kaas.Tags(), data.Tags)
	if kaas.Region() != "" {
		data.Location = types.StringValue(string(kaas.Region()))
	}
	if bp := string(kaas.BillingPeriod()); bp != "" {
		data.BillingPeriod = types.StringValue(bp)
	}
	if raw.Properties.ManagementIP != nil && *raw.Properties.ManagementIP != "" {
		data.ManagementIP = types.StringValue(*raw.Properties.ManagementIP)
	} else {
		data.ManagementIP = types.StringNull()
	}

	// Build Network object.
	networkAttrs := map[string]attr.Value{
		"vpc_uri_ref":         types.StringNull(),
		"subnet_uri_ref":      types.StringNull(),
		"node_cidr":           types.ObjectNull(map[string]attr.Type{"address": types.StringType, "name": types.StringType}),
		"security_group_name": types.StringNull(),
		"pod_cidr":            types.StringNull(),
	}
	if v := kaas.VPC(); v != "" {
		networkAttrs["vpc_uri_ref"] = types.StringValue(v)
	}
	if v := kaas.Subnet(); v != "" {
		networkAttrs["subnet_uri_ref"] = types.StringValue(v)
	}
	if v := kaas.SecurityGroupName(); v != "" {
		networkAttrs["security_group_name"] = types.StringValue(v)
	} else if !originalState.Network.IsNull() {
		var origNet KaaSNetworkModel
		if dd := originalState.Network.As(ctx, &origNet, basetypes.ObjectAsOptions{}); !dd.HasError() && !origNet.SecurityGroupName.IsNull() {
			networkAttrs["security_group_name"] = origNet.SecurityGroupName
		}
	}
	nodeCIDRAddress := raw.Properties.NodeCIDR.Address
	if nodeCIDRAddress != nil && *nodeCIDRAddress != "" {
		nodeCIDRName := ""
		if raw.Properties.NodeCIDR.Name != nil {
			nodeCIDRName = *raw.Properties.NodeCIDR.Name
		}
		if nodeCIDRName == "" && !originalState.Network.IsNull() {
			var origNet KaaSNetworkModel
			if dd := originalState.Network.As(ctx, &origNet, basetypes.ObjectAsOptions{}); !dd.HasError() && !origNet.NodeCIDR.IsNull() {
				var origCIDR KaaSNodeCIDRModel
				if dd2 := origNet.NodeCIDR.As(ctx, &origCIDR, basetypes.ObjectAsOptions{}); !dd2.HasError() && !origCIDR.Name.IsNull() {
					nodeCIDRName = origCIDR.Name.ValueString()
				}
			}
		}
		nodeCIDRObj, dCIDR := types.ObjectValue(map[string]attr.Type{"address": types.StringType, "name": types.StringType},
			map[string]attr.Value{"address": types.StringValue(*nodeCIDRAddress), "name": types.StringValue(nodeCIDRName)})
		resp.Diagnostics.Append(dCIDR...)
		if !resp.Diagnostics.HasError() {
			networkAttrs["node_cidr"] = nodeCIDRObj
		}
	}
	if v := kaas.PodCIDR(); v != "" {
		networkAttrs["pod_cidr"] = types.StringValue(v)
	}
	networkObj, dNet := types.ObjectValue(map[string]attr.Type{
		"vpc_uri_ref": types.StringType, "subnet_uri_ref": types.StringType,
		"node_cidr":           types.ObjectType{AttrTypes: map[string]attr.Type{"address": types.StringType, "name": types.StringType}},
		"security_group_name": types.StringType, "pod_cidr": types.StringType,
	}, networkAttrs)
	resp.Diagnostics.Append(dNet...)
	if !resp.Diagnostics.HasError() {
		data.Network = networkObj
	}

	// Build Settings object.
	settingsAttrs := map[string]attr.Value{
		"kubernetes_version": types.StringNull(),
		"node_pools":         types.ListNull(types.ObjectType{AttrTypes: nodePoolAttrTypes()}),
		"ha":                 types.BoolNull(),
	}
	if kv := string(kaas.KubernetesVersion()); kv != "" {
		settingsAttrs["kubernetes_version"] = types.StringValue(kv)
	}
	if raw.Properties.HA != nil {
		settingsAttrs["ha"] = types.BoolValue(*raw.Properties.HA)
	}
	nodePoolList, npOk := buildNodePoolAttrValues(kaas, originalState.Settings.Attributes()["node_pools"].(types.List))
	if npOk {
		settingsAttrs["node_pools"] = nodePoolList
	} else if !originalState.Settings.IsNull() {
		var origSettings KaaSSettingsModel
		if dd := originalState.Settings.As(ctx, &origSettings, basetypes.ObjectAsOptions{}); !dd.HasError() && !origSettings.NodePools.IsNull() {
			settingsAttrs["node_pools"] = origSettings.NodePools
		}
	}
	nodePoolListType := types.ListType{ElemType: types.ObjectType{AttrTypes: nodePoolAttrTypes()}}
	settingsObj, dSett := types.ObjectValue(map[string]attr.Type{
		"kubernetes_version": types.StringType, "node_pools": nodePoolListType, "ha": types.BoolType,
	}, settingsAttrs)
	resp.Diagnostics.Append(dSett...)
	if !resp.Diagnostics.HasError() {
		data.Settings = settingsObj
	}

	// Refresh kubeconfig when cluster has management IP.
	if raw.Properties.ManagementIP != nil && *raw.Properties.ManagementIP != "" {
		kc := downloadKubeconfig(ctx, kaas)
		if kc.IsNull() && !originalState.Kubeconfig.IsNull() {
			data.Kubeconfig = originalState.Kubeconfig
		} else {
			data.Kubeconfig = kc
		}
	} else if !originalState.Kubeconfig.IsNull() {
		data.Kubeconfig = originalState.Kubeconfig
	} else {
		data.Kubeconfig = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KaaSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KaaSResourceModel
	var state KaaSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	kaas, err := r.client.Client.FromContainer().KaaS().Get(ctx, kaasRef(&state))
	if provErr := CheckResponseErr("read", "KaaS", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	var settingsModel KaaSSettingsModel
	resp.Diagnostics.Append(data.Settings.As(ctx, &settingsModel, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	var nodePoolModels []KaaSNodePoolModel
	if !settingsModel.NodePools.IsNull() && !settingsModel.NodePools.IsUnknown() {
		resp.Diagnostics.Append(settingsModel.NodePools.ElementsAs(ctx, &nodePoolModels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	kaas.Named(data.Name.ValueString())
	if tags != nil {
		kaas.RetaggedAs(tags...)
	} else {
		kaas.RetaggedAs(kaas.Tags()...)
	}
	kaas.WithKubernetesVersion(aruba.KubernetesVersion(settingsModel.KubernetesVersion.ValueString()))
	if len(nodePoolModels) > 0 {
		kaas.ReplaceNodePools(buildNodePools(nodePoolModels)...)
	}
	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		kaas.BilledBy(aruba.BillingPeriod(data.BillingPeriod.ValueString()))
	}

	updated, err := r.client.Client.FromContainer().KaaS().Update(ctx, kaas)
	if provErr := CheckResponseErr("update", "KaaS", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = state.Id
	data.ProjectID = state.ProjectID
	data.Uri = strVal(updated.URI())
	data.Network = state.Network // Network is immutable
	data.Name = types.StringValue(updated.Name())
	data.Tags = TagsToListPreserveNull(updated.Tags(), data.Tags)

	// Re-read for full state.
	fresh, freshErr := r.client.Client.FromContainer().KaaS().Get(ctx, kaasRef(&data))
	if freshErr == nil {
		rawFresh := fresh.Raw()
		if rawFresh != nil {
			if rawFresh.Properties.ManagementIP != nil && *rawFresh.Properties.ManagementIP != "" {
				data.ManagementIP = types.StringValue(*rawFresh.Properties.ManagementIP)
			} else {
				data.ManagementIP = types.StringNull()
			}
			if bp := string(fresh.BillingPeriod()); bp != "" {
				data.BillingPeriod = types.StringValue(bp)
			}
		}
		kc := downloadKubeconfig(ctx, fresh)
		if kc.IsNull() && !state.Kubeconfig.IsNull() {
			data.Kubeconfig = state.Kubeconfig
		} else {
			data.Kubeconfig = kc
		}

		nodePoolList, npOk := buildNodePoolAttrValues(fresh, settingsModel.NodePools)
		settingsAttrs := map[string]attr.Value{
			"kubernetes_version": types.StringValue(settingsModel.KubernetesVersion.ValueString()),
			"ha":                 settingsModel.HA,
		}
		if rawFresh != nil && rawFresh.Properties.HA != nil {
			settingsAttrs["ha"] = types.BoolValue(*rawFresh.Properties.HA)
		}
		if npOk {
			settingsAttrs["node_pools"] = nodePoolList
		} else {
			settingsAttrs["node_pools"] = settingsModel.NodePools
		}
		nodePoolListType := types.ListType{ElemType: types.ObjectType{AttrTypes: nodePoolAttrTypes()}}
		settingsObj, d := types.ObjectValue(map[string]attr.Type{
			"kubernetes_version": types.StringType, "node_pools": nodePoolListType, "ha": types.BoolType,
		}, settingsAttrs)
		resp.Diagnostics.Append(d...)
		if !resp.Diagnostics.HasError() {
			data.Settings = settingsObj
		} else {
			data.Settings = state.Settings
		}
	} else {
		data.ManagementIP = state.ManagementIP
		data.Kubeconfig = state.Kubeconfig
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

	ref := kaasRef(&data)
	kaasID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromContainer().KaaS().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "KaaS", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErr("delete", "KaaS",
			r.client.Client.FromContainer().KaaS().Delete(ctx, ref))
	}, "KaaS", kaasID, r.client.ResourceTimeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting KaaS cluster", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "KaaS", kaasID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for KaaS cluster deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a KaaS resource", map[string]interface{}{"kaas_id": kaasID})
}

func (r *KaaSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
