package provider

import (
	"context"
	"fmt"

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
		MarkdownDescription: "KaaS data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "KaaS identifier",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "KaaS URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "KaaS name",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "KaaS location",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the KaaS resource",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this KaaS resource belongs to",
				Computed:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period (Hour, Month, Year)",
				Computed:            true,
			},
			"management_ip": schema.StringAttribute{
				MarkdownDescription: "Management IP address",
				Computed:            true,
			},
			"vpc_uri_ref": schema.StringAttribute{
				MarkdownDescription: "VPC URI reference",
				Computed:            true,
			},
			"subnet_uri_ref": schema.StringAttribute{
				MarkdownDescription: "Subnet URI reference",
				Computed:            true,
			},
			"node_cidr_address": schema.StringAttribute{
				MarkdownDescription: "Node CIDR address in CIDR notation",
				Computed:            true,
			},
			"node_cidr_name": schema.StringAttribute{
				MarkdownDescription: "Node CIDR name",
				Computed:            true,
			},
			"security_group_name": schema.StringAttribute{
				MarkdownDescription: "Security group name",
				Computed:            true,
			},
			"pod_cidr": schema.StringAttribute{
				MarkdownDescription: "Pod CIDR in CIDR notation",
				Computed:            true,
			},
			"kubernetes_version": schema.StringAttribute{
				MarkdownDescription: "Kubernetes version",
				Computed:            true,
			},
			"node_pools": schema.ListNestedAttribute{
				MarkdownDescription: "Node pools configuration",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Node pool name",
							Computed:            true,
						},
						"nodes": schema.Int64Attribute{
							MarkdownDescription: "Number of nodes in the node pool",
							Computed:            true,
						},
						"instance": schema.StringAttribute{
							MarkdownDescription: "KaaS flavor name for nodes",
							Computed:            true,
						},
						"zone": schema.StringAttribute{
							MarkdownDescription: "Datacenter/zone code for nodes",
							Computed:            true,
						},
						"autoscaling": schema.BoolAttribute{
							MarkdownDescription: "Enable autoscaling for node pool",
							Computed:            true,
						},
						"min_count": schema.Int64Attribute{
							MarkdownDescription: "Minimum number of nodes for autoscaling",
							Computed:            true,
						},
						"max_count": schema.Int64Attribute{
							MarkdownDescription: "Maximum number of nodes for autoscaling",
							Computed:            true,
						},
					},
				},
			},
			"ha": schema.BoolAttribute{
				MarkdownDescription: "High availability",
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
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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
	// Populate all fields with example data
	data.Uri = types.StringValue("/v2/kaas/kaas-68398923fb2cb026400d4d31")
	data.Name = types.StringValue("example-kaas")
	data.Location = types.StringValue("ITBG-Bergamo")
	data.Tags = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("kubernetes"),
		types.StringValue("production"),
	})
	data.ProjectID = types.StringValue("68398923fb2cb026400d4d31")
	data.BillingPeriod = types.StringValue("Hour")
	data.ManagementIP = types.StringValue("10.0.2.100")
	// Network fields
	data.VpcUriRef = types.StringValue("/v2/vpcs/vpc-68398923fb2cb026400d4d32")
	data.SubnetUriRef = types.StringValue("/v2/subnets/subnet-68398923fb2cb026400d4d33")
	data.NodeCIDRAddress = types.StringValue("10.0.2.0/24")
	data.NodeCIDRName = types.StringValue("kaas-nodes")
	data.SecurityGroupName = types.StringValue("kaas-security-group")
	data.PodCIDR = types.StringValue("10.0.3.0/24")
	// Settings fields
	data.KubernetesVersion = types.StringValue("1.33.2")
	data.HA = types.BoolValue(true)

	// Create node pools
	nodePoolType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":        types.StringType,
			"nodes":       types.Int64Type,
			"instance":    types.StringType,
			"zone":        types.StringType,
			"autoscaling": types.BoolType,
			"min_count":   types.Int64Type,
			"max_count":   types.Int64Type,
		},
	}
	nodePool1, _ := types.ObjectValue(nodePoolType.AttrTypes, map[string]attr.Value{
		"name":        types.StringValue("pool-1"),
		"nodes":       types.Int64Value(3),
		"instance":    types.StringValue("K2A4"),
		"zone":        types.StringValue("ITBG-1"),
		"autoscaling": types.BoolValue(true),
		"min_count":   types.Int64Value(2),
		"max_count":   types.Int64Value(5),
	})
	nodePool2, _ := types.ObjectValue(nodePoolType.AttrTypes, map[string]attr.Value{
		"name":        types.StringValue("pool-2"),
		"nodes":       types.Int64Value(2),
		"instance":    types.StringValue("K4A8"),
		"zone":        types.StringValue("ITBG-2"),
		"autoscaling": types.BoolValue(false),
		"min_count":   types.Int64Null(),
		"max_count":   types.Int64Null(),
	})
	data.NodePools = types.ListValueMust(nodePoolType, []attr.Value{nodePool1, nodePool2})
	tflog.Trace(ctx, "read a KaaS data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
