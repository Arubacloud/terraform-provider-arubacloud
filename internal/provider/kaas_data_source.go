package provider

import (
	"context"
	"fmt"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &KaaSDataSource{}

func NewKaaSDataSource() datasource.DataSource {
	return &KaaSDataSource{}
}

type KaaSDataSource struct {
	client *ArubaCloudClient
}

type KaaSDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	ProjectID     types.String `tfsdk:"project_id"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	ManagementIP  types.String `tfsdk:"management_ip"`
	Kubeconfig    types.String `tfsdk:"kubeconfig"`
	// Network fields (flattened)
	VpcUriRef         types.String `tfsdk:"vpc_uri_ref"`
	SubnetUriRef      types.String `tfsdk:"subnet_uri_ref"`
	NodeCIDRAddress   types.String `tfsdk:"node_cidr_address"`
	NodeCIDRName      types.String `tfsdk:"node_cidr_name"`
	SecurityGroupName types.String `tfsdk:"security_group_name"`
	PodCIDR           types.String `tfsdk:"pod_cidr"`
	// Settings fields (flattened)
	KubernetesVersion types.String `tfsdk:"kubernetes_version"`
	NodePools         types.List   `tfsdk:"node_pools"`
	HA                types.Bool   `tfsdk:"ha"`
}

type NodePoolDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	Nodes       types.Int64  `tfsdk:"nodes"`
	Instance    types.String `tfsdk:"instance"`
	Zone        types.String `tfsdk:"zone"`
	Autoscaling types.Bool   `tfsdk:"autoscaling"`
	MinCount    types.Int64  `tfsdk:"min_count"`
	MaxCount    types.Int64  `tfsdk:"max_count"`
}

func (d *KaaSDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kaas"
}

func (d *KaaSDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about an existing ArubaCloud KaaS (Kubernetes-as-a-Service) cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the KaaS cluster to look up.",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the KaaS cluster.",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle. Accepted values: `Hour`, `Month`, `Year`.",
				Computed:            true,
			},
			"management_ip": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Management IP address of the cluster control plane.",
				Computed:            true,
			},
			"kubeconfig": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Kubeconfig YAML for kubectl access. Write-only — this value is sent to the API but is not returned in subsequent read responses.",
				Computed:            true,
				Sensitive:           true,
			},
			"vpc_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI of the VPC that hosts the cluster.",
				Computed:            true,
			},
			"subnet_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI of the subnet within the VPC.",
				Computed:            true,
			},
			"node_cidr_address": schema.StringAttribute{
				MarkdownDescription: "Node CIDR address in CIDR notation.",
				Computed:            true,
			},
			"node_cidr_name": schema.StringAttribute{
				MarkdownDescription: "Human-readable label for the node CIDR block.",
				Computed:            true,
			},
			"security_group_name": schema.StringAttribute{
				MarkdownDescription: "Name of the security group applied to cluster nodes.",
				Computed:            true,
			},
			"pod_cidr": schema.StringAttribute{
				MarkdownDescription: "CIDR block used for pod networking within the cluster.",
				Computed:            true,
			},
			"kubernetes_version": schema.StringAttribute{
				MarkdownDescription: "Kubernetes version string (e.g., `1.28`). Available versions are listed in the ArubaCloud metadata API.",
				Computed:            true,
			},
			"node_pools": schema.ListNestedAttribute{
				MarkdownDescription: "Node pools that make up the cluster worker fleet.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Display name for the node pool.",
							Computed:            true,
						},
						"nodes": schema.Int64Attribute{
							MarkdownDescription: "Number of worker nodes in the cluster.",
							Computed:            true,
						},
						"instance": schema.StringAttribute{
							MarkdownDescription: "Compute flavour for cluster nodes.",
							Computed:            true,
						},
						"zone": schema.StringAttribute{
							MarkdownDescription: "Datacenter zone code where the node pool is deployed.",
							Computed:            true,
						},
						"autoscaling": schema.BoolAttribute{
							MarkdownDescription: "Whether autoscaling is enabled for this node pool.",
							Computed:            true,
						},
						"min_count": schema.Int64Attribute{
							MarkdownDescription: "Minimum number of nodes when autoscaling is enabled.",
							Computed:            true,
						},
						"max_count": schema.Int64Attribute{
							MarkdownDescription: "Maximum number of nodes when autoscaling is enabled.",
							Computed:            true,
						},
					},
				},
			},
			"ha": schema.BoolAttribute{
				MarkdownDescription: "Whether the control plane is deployed in high-availability mode.",
				Computed:            true,
			},
		},
	}
}

func (d *KaaSDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KaaSDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data KaaSDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	kaasID := data.Id.ValueString()
	if projectID == "" || kaasID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and KaaS ID (id) are required to read the KaaS cluster",
		)
		return
	}

	ref := aruba.URI("/projects/" + projectID + "/providers/Aruba.Container/kaas/" + kaasID)
	kaas, err := d.client.Client.FromContainer().KaaS().Get(ctx, ref)
	if provErr := CheckResponseErr("read", "KaaS", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	raw := kaas.Raw()

	data.Id = types.StringValue(kaas.ID())
	data.Uri = strVal(kaas.URI())
	data.Name = types.StringValue(kaas.Name())
	data.ProjectID = types.StringValue(projectID)
	if kaas.Region() != "" {
		data.Location = types.StringValue(string(kaas.Region()))
	} else {
		data.Location = types.StringNull()
	}
	data.Tags = TagsToListPreserveNull(kaas.Tags(), data.Tags)
	if bp := string(kaas.BillingPeriod()); bp != "" {
		data.BillingPeriod = types.StringValue(bp)
	} else {
		data.BillingPeriod = types.StringNull()
	}
	if raw != nil && raw.Properties.ManagementIP != nil && *raw.Properties.ManagementIP != "" {
		data.ManagementIP = types.StringValue(*raw.Properties.ManagementIP)
	} else {
		data.ManagementIP = types.StringNull()
	}

	// Network (flattened)
	if v := kaas.VPC(); v != "" {
		data.VpcUriRef = types.StringValue(v)
	} else {
		data.VpcUriRef = types.StringNull()
	}
	if v := kaas.Subnet(); v != "" {
		data.SubnetUriRef = types.StringValue(v)
	} else {
		data.SubnetUriRef = types.StringNull()
	}
	if v := kaas.SecurityGroupName(); v != "" {
		data.SecurityGroupName = types.StringValue(v)
	} else {
		data.SecurityGroupName = types.StringNull()
	}
	if v := kaas.PodCIDR(); v != "" {
		data.PodCIDR = types.StringValue(v)
	} else {
		data.PodCIDR = types.StringNull()
	}
	if raw != nil && raw.Properties.NodeCIDR.Address != nil {
		data.NodeCIDRAddress = types.StringValue(*raw.Properties.NodeCIDR.Address)
	} else {
		data.NodeCIDRAddress = types.StringNull()
	}
	if raw != nil && raw.Properties.NodeCIDR.Name != nil {
		data.NodeCIDRName = types.StringValue(*raw.Properties.NodeCIDR.Name)
	} else {
		data.NodeCIDRName = types.StringNull()
	}

	// Settings (flattened)
	if kv := string(kaas.KubernetesVersion()); kv != "" {
		data.KubernetesVersion = types.StringValue(kv)
	} else {
		data.KubernetesVersion = types.StringNull()
	}
	if raw != nil && raw.Properties.HA != nil {
		data.HA = types.BoolValue(*raw.Properties.HA)
	} else {
		data.HA = types.BoolNull()
	}

	// Node pools
	nodePoolType := types.ObjectType{AttrTypes: nodePoolAttrTypes()}
	nodePoolList, npOk := buildNodePoolAttrValues(kaas, types.ListNull(nodePoolType))
	if npOk {
		data.NodePools = nodePoolList
	} else {
		data.NodePools = types.ListValueMust(nodePoolType, []attr.Value{})
	}

	// Kubeconfig
	if raw != nil && raw.Properties.ManagementIP != nil && *raw.Properties.ManagementIP != "" {
		data.Kubeconfig = downloadKubeconfig(ctx, kaas)
	} else {
		data.Kubeconfig = types.StringNull()
	}

	tflog.Trace(ctx, "read a KaaS data source", map[string]interface{}{"kaas_id": kaasID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func ptrToString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}
