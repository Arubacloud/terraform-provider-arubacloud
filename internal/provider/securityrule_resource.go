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

// RuleDirection represents the direction of a security rule
type RuleDirection string

const (
	RuleDirectionIngress RuleDirection = "Ingress"
	RuleDirectionEgress  RuleDirection = "Egress"
)

type SecurityRuleResource struct {
	client *http.Client
}

// EndpointTypeDto represents the type of target endpoint
// ...existing code...
type EndpointTypeDto string

const (
	EndpointTypeIP EndpointTypeDto = "Ip"
)

// RuleTarget represents the target of the rule (source or destination according to the direction)
type RuleTarget struct {
	Kind  EndpointTypeDto `tfsdk:"kind"`
	Value string          `tfsdk:"value"`
}

// SecurityRuleProperties contains the properties of a security rule
type SecurityRulePropertiesRequest struct {
	Direction RuleDirection `tfsdk:"direction"`
	Protocol  string        `tfsdk:"protocol"`
	Port      string        `tfsdk:"port"`
	Target    *RuleTarget   `tfsdk:"target"`
}

var _ resource.Resource = &SecurityRuleResource{}
var _ resource.ResourceWithImportState = &SecurityRuleResource{}

func NewSecurityRuleResource() resource.Resource {
	return &SecurityRuleResource{}
}

type SecurityRuleResourceModel struct {
	Id              types.String `tfsdk:"id"`
	VpcId           types.String `tfsdk:"vpc_id"`
	SecurityGroupId types.String `tfsdk:"security_group_id"`
	Properties      types.Object `tfsdk:"properties"`
}

func (r *SecurityRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_securityrule"
}

func (r *SecurityRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Security Rule resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Security Rule identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Security Rule name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Security Rule location",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this Security Rule belongs to",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPC this Security Rule belongs to",
				Required:            true,
			},
			"security_group_id": schema.StringAttribute{
				MarkdownDescription: "ID of the Security Group this rule belongs to",
				Required:            true,
			},
			"properties": schema.SingleNestedAttribute{
				MarkdownDescription: "Properties of the security rule",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"direction": schema.StringAttribute{
						MarkdownDescription: "Direction of the rule (Ingress/Egress)",
						Required:            true,
						// Validators removed for v1.16.1 compatibility
					},
					"protocol": schema.StringAttribute{
						MarkdownDescription: "Protocol (ANY, TCP, UDP, ICMP)",
						Required:            true,
						// Validators removed for v1.16.1 compatibility
					},
					"port": schema.StringAttribute{
						MarkdownDescription: "Port or port range (for TCP/UDP)",
						Optional:            true,
					},
					"target": schema.SingleNestedAttribute{
						MarkdownDescription: "Target of the rule (source or destination)",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"kind": schema.StringAttribute{
								MarkdownDescription: "Type of the target (Ip/SecurityGroup)",
								Required:            true,
								// Validators removed for v1.16.1 compatibility
							},
							"value": schema.StringAttribute{
								MarkdownDescription: "Value of the target (CIDR or SecurityGroup URI)",
								Required:            true,
							},
						},
					},
				},
			},
		},
	}
}

func (r *SecurityRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SecurityRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecurityRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Simulate API response
	data.Id = types.StringValue("securityrule-id")
	tflog.Trace(ctx, "created a Security Rule resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SecurityRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SecurityRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SecurityRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
