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

var _ resource.Resource = &VpcPeeringRouteResource{}
var _ resource.ResourceWithImportState = &VpcPeeringRouteResource{}

func NewVpcPeeringRouteResource() resource.Resource {
	return &VpcPeeringRouteResource{}
}

type VpcPeeringRouteResource struct {
	client *http.Client
}

type VpcPeeringRouteResourceModel struct {
	Id                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Location             types.String `tfsdk:"location"`
	Tags                 types.List   `tfsdk:"tags"`
	LocalNetworkAddress  types.String `tfsdk:"local_network_address"`
	RemoteNetworkAddress types.String `tfsdk:"remote_network_address"`
	BillingPeriod        types.String `tfsdk:"billing_period"`
	VpcPeeringId         types.String `tfsdk:"vpc_peering_id"`
}

func (r *VpcPeeringRouteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpcpeeringroute"
}

func (r *VpcPeeringRouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VPC Peering Route resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VPC Peering Route identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "VPC Peering Route name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "VPC Peering Route location",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the VPC Peering Route",
				Optional:            true,
			},
			"local_network_address": schema.StringAttribute{
				MarkdownDescription: "Local network address in CIDR notation",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"remote_network_address": schema.StringAttribute{
				MarkdownDescription: "Remote network address in CIDR notation",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period (only 'Hour' allowed)",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"vpc_peering_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPC Peering this route belongs to",
				Required:            true,
			},
		},
	}
}

func (r *VpcPeeringRouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VpcPeeringRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VpcPeeringRouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Simulate API response
	data.Id = types.StringValue("vpcpeeringroute-id")
	tflog.Trace(ctx, "created a VPC Peering Route resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpcPeeringRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VpcPeeringRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
func (r *VpcPeeringRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VpcPeeringRouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *VpcPeeringRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VpcPeeringRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
	}
}

func (r *VpcPeeringRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
