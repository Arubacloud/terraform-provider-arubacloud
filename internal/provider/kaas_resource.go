package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type NodeCIDRProperties struct {
	Address string `json:"address"`
	Name    string `json:"subnetName"`
}

type NodePoolProperties struct {
	Name     string `tfsdk:"nodePoolName"`
	Nodes    int32  `tfsdk:"replicas"`
	Instance string `tfsdk:"type"`
	Zone     string `tfsdk:"zone"`
}

type SecurityGroupProperties struct {
	Name string `tfsdk:"name"`
}

type KaaSPropertiesRequest struct {
	Preset            bool                 `tfsdk:"preset"`
	VpcID             ReferenceResource    `tfsdk:"vpc_id"`
	SubnetID          types.String         `tfsdk:"project_id"`
	NodeCIDR          NodeCIDRProperties   `tfsdk:"nodeCidr"`
	SecurityGroupName types.String         `tfsdk:"securityGroupName"`
	KubernetesVersion types.String         `tfsdk:"version"`
	NodePools         []NodePoolProperties `tfsdk:"nodePools"`
	HA                bool                 `tfsdk:"ha"`
	BillingPeriod     types.String         `tfsdk:"billing_period"`
}

type KaaSResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	Tags      types.List   `tfsdk:"tags"`
	ProjectID types.String `tfsdk:"project_id"`
}

type KaaSResource struct {
	client *http.Client
}

var _ resource.Resource = &KaaSResource{}
var _ resource.ResourceWithImportState = &KaaSResource{}

func NewKaaSResource() resource.Resource {
	return &KaaSResource{}
}

func (r *KaaSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kaas"
}

func (r *KaaSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "KaaS resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "KaaS identifier",
				Computed:            true,
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
			"preset": schema.BoolAttribute{
				MarkdownDescription: "Whether to use a preset configuration",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "VPC ID for the KaaS resource",
				Required:            true,
			},
			"subnet_id": schema.StringAttribute{
				MarkdownDescription: "Subnet ID for the KaaS resource",
				Required:            true,
			},
			"node_cidr": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						MarkdownDescription: "Node CIDR address",
						Required:            true,
					},
					"subnet_name": schema.StringAttribute{
						MarkdownDescription: "Node CIDR subnet name",
						Required:            true,
					},
				},
				Required: true,
			},
			"security_group_name": schema.StringAttribute{
				MarkdownDescription: "Security group name",
				Required:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Kubernetes version",
				Required:            true,
			},
			"node_pools": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"node_pool_name": schema.StringAttribute{Required: true},
						"replicas":       schema.Int64Attribute{Required: true},
						"type":           schema.StringAttribute{Required: true},
						"zone":           schema.StringAttribute{Required: true},
					},
				},
				Required: true,
			},
			"ha": schema.BoolAttribute{
				MarkdownDescription: "High availability",
				Required:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period",
				Required:            true,
			},
		},
	}
}

func (r *KaaSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*http.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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
	// Simulate API response
	data.Id = types.StringValue("kaas-id")
	tflog.Trace(ctx, "created a KaaS resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KaaSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KaaSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KaaSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KaaSResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KaaSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KaaSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *KaaSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
