package provider

import (
	"context"
	"encoding/base64"
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
				Required:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period (Hour, Month, Year)",
				Computed:            true,
			},
			"management_ip": schema.StringAttribute{
				MarkdownDescription: "Management IP address",
				Computed:            true,
			},
			"kubeconfig": schema.StringAttribute{
				MarkdownDescription: "Kubeconfig YAML for kubectl access (downloaded when KaaS is ready). Sensitive.",
				Computed:            true,
				Sensitive:           true,
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

	projectID := data.ProjectID.ValueString()
	kaasID := data.Id.ValueString()
	if projectID == "" || kaasID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and KaaS ID (id) are required to read the KaaS cluster",
		)
		return
	}

	response, err := d.client.Client.FromContainer().KaaS().Get(ctx, projectID, kaasID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading KaaS cluster",
			fmt.Sprintf("Unable to read KaaS cluster: %s", err),
		)
		return
	}
	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.Diagnostics.AddError(
				"KaaS cluster not found",
				fmt.Sprintf("No KaaS cluster found with ID %q in project %q", kaasID, projectID),
			)
			return
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to read KaaS cluster", map[string]interface{}{
			"project_id": projectID,
			"kaas_id":    kaasID,
		})
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError(
			"No data returned",
			"KaaS cluster Get returned no data",
		)
		return
	}

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
	data.ProjectID = types.StringValue(projectID)
	if kaas.Properties.BillingPlan != nil && kaas.Properties.BillingPlan.BillingPeriod != nil {
		data.BillingPeriod = types.StringValue(*kaas.Properties.BillingPlan.BillingPeriod)
	} else {
		data.BillingPeriod = types.StringNull()
	}
	if kaas.Properties.ManagementIP != nil && *kaas.Properties.ManagementIP != "" {
		data.ManagementIP = types.StringValue(*kaas.Properties.ManagementIP)
	} else {
		data.ManagementIP = types.StringNull()
	}

	// Network (flattened)
	if kaas.Properties.VPC.URI != nil {
		data.VpcUriRef = types.StringValue(*kaas.Properties.VPC.URI)
	} else {
		data.VpcUriRef = types.StringNull()
	}
	if kaas.Properties.Subnet.URI != nil {
		data.SubnetUriRef = types.StringValue(*kaas.Properties.Subnet.URI)
	} else {
		data.SubnetUriRef = types.StringNull()
	}
	if kaas.Properties.NodeCIDR.Address != nil {
		data.NodeCIDRAddress = types.StringValue(*kaas.Properties.NodeCIDR.Address)
	} else {
		data.NodeCIDRAddress = types.StringNull()
	}
	if kaas.Properties.NodeCIDR.Name != nil {
		data.NodeCIDRName = types.StringValue(*kaas.Properties.NodeCIDR.Name)
	} else {
		data.NodeCIDRName = types.StringNull()
	}
	if kaas.Properties.SecurityGroup.Name != nil {
		data.SecurityGroupName = types.StringValue(*kaas.Properties.SecurityGroup.Name)
	} else {
		data.SecurityGroupName = types.StringNull()
	}
	if kaas.Properties.PodCIDR != nil && kaas.Properties.PodCIDR.Address != nil {
		data.PodCIDR = types.StringValue(*kaas.Properties.PodCIDR.Address)
	} else {
		data.PodCIDR = types.StringNull()
	}

	// Settings (flattened)
	if kaas.Properties.KubernetesVersion.Value != nil {
		data.KubernetesVersion = types.StringValue(*kaas.Properties.KubernetesVersion.Value)
	} else {
		data.KubernetesVersion = types.StringNull()
	}
	if kaas.Properties.HA != nil {
		data.HA = types.BoolValue(*kaas.Properties.HA)
	} else {
		data.HA = types.BoolNull()
	}

	// Node pools
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
	if kaas.Properties.NodePools != nil && len(*kaas.Properties.NodePools) > 0 {
		nodePoolValues := make([]attr.Value, 0, len(*kaas.Properties.NodePools))
		for _, np := range *kaas.Properties.NodePools {
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
			nodePoolMap := map[string]attr.Value{
				"name":        types.StringValue(ptrToString(np.Name)),
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
			obj, diags := types.ObjectValue(nodePoolType.AttrTypes, nodePoolMap)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				nodePoolValues = append(nodePoolValues, obj)
			}
		}
		if !resp.Diagnostics.HasError() {
			data.NodePools = types.ListValueMust(nodePoolType, nodePoolValues)
		}
	} else {
		data.NodePools = types.ListValueMust(nodePoolType, []attr.Value{})
	}

	// Tags
	if len(kaas.Metadata.Tags) > 0 {
		tagValues := make([]attr.Value, len(kaas.Metadata.Tags))
		for i, tag := range kaas.Metadata.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		data.Tags = types.ListValueMust(types.StringType, tagValues)
	} else {
		data.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Download kubeconfig (API returns base64-encoded content), when cluster has management IP
	if kaas.Properties.ManagementIP != nil && *kaas.Properties.ManagementIP != "" {
		kubeconfigResp, kerr := d.client.Client.FromContainer().KaaS().DownloadKubeconfig(ctx, projectID, kaasID, nil)
		if kerr == nil && kubeconfigResp != nil && !kubeconfigResp.IsError() && kubeconfigResp.Data != nil && kubeconfigResp.Data.Content != "" {
			if decoded, decErr := base64.StdEncoding.DecodeString(kubeconfigResp.Data.Content); decErr == nil {
				data.Kubeconfig = types.StringValue(string(decoded))
			} else {
				tflog.Warn(ctx, "Failed to decode kubeconfig base64", map[string]interface{}{"error": decErr.Error()})
				data.Kubeconfig = types.StringNull()
			}
		} else {
			data.Kubeconfig = types.StringNull()
		}
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
