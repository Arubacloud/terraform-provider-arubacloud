package provider

import (
	"context"
	"fmt"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &VpcPeeringRouteResource{}
var _ resource.ResourceWithImportState = &VpcPeeringRouteResource{}

func NewVpcPeeringRouteResource() resource.Resource {
	return &VpcPeeringRouteResource{}
}

type VpcPeeringRouteResource struct {
	client *ArubaCloudClient
}

type VpcPeeringRouteResourceModel struct {
	Id                   types.String `tfsdk:"id"`
	Uri                  types.String `tfsdk:"uri"`
	Name                 types.String `tfsdk:"name"`
	Tags                 types.List   `tfsdk:"tags"`
	ProjectId            types.String `tfsdk:"project_id"`
	VpcId                types.String `tfsdk:"vpc_id"`
	VpcPeeringId         types.String `tfsdk:"vpc_peering_id"`
	LocalNetworkAddress  types.String `tfsdk:"local_network_address"`
	RemoteNetworkAddress types.String `tfsdk:"remote_network_address"`
	BillingPeriod        types.String `tfsdk:"billing_period"`
}

func (r *VpcPeeringRouteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpcpeeringroute"
}

func (r *VpcPeeringRouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a route entry within an ArubaCloud VPC Peering connection. Each route directs traffic destined for a specific CIDR block over the peering link. Routes must be created on both sides of the peering connection.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources (e.g., as a `*_uri_ref` attribute).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the VPC peering route.",
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
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPC this peering route belongs to.",
				Required:            true,
			},
			"vpc_peering_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPC peering connection this route belongs to.",
				Required:            true,
			},
			"local_network_address": schema.StringAttribute{
				MarkdownDescription: "Local network CIDR that is reachable on this side of the peering (e.g., `10.0.1.0/24`).",
				Required:            true,
			},
			"remote_network_address": schema.StringAttribute{
				MarkdownDescription: "Remote network CIDR reachable through the peering connection (e.g., `10.0.2.0/24`).",
				Required:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle for the resource. Accepted values: `Hour`, `Month`, `Year`.",
				Required:            true,
			},
		},
	}
}

func (r *VpcPeeringRouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func vpcPeeringRouteRef(data *VpcPeeringRouteResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.VPCPeeringRouteRef(
		data.ProjectId.ValueString(),
		data.VpcId.ValueString(),
		data.VpcPeeringId.ValueString(),
		data.Id.ValueString(),
	)
}

func applyVPCPeeringRouteToModel(route *aruba.VPCPeeringRoute, data *VpcPeeringRouteResourceModel) {
	// VPCPeeringRoute uses name as ID (no separate UUID from API).
	data.Id = types.StringValue(route.Name())
	data.Uri = strVal(route.URI())
	data.Name = types.StringValue(route.Name())
	data.Tags = TagsToListPreserveNull(route.Tags(), data.Tags)
	if route.LocalCIDR() != "" {
		data.LocalNetworkAddress = types.StringValue(route.LocalCIDR())
	}
	if route.RemoteCIDR() != "" {
		data.RemoteNetworkAddress = types.StringValue(route.RemoteCIDR())
	}
	if bp := billingPeriodFromAPI(string(route.BillingPeriod())); bp != "" {
		data.BillingPeriod = types.StringValue(bp)
	}
}

func (r *VpcPeeringRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VpcPeeringRouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	peeringID := data.VpcPeeringId.ValueString()

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	peeringRef := aruba.VPCPeeringRef(projectID, vpcID, peeringID)
	route, err := r.client.Client.FromNetwork().VPCPeeringRoutes().Create(ctx,
		aruba.NewVPCPeeringRoute().
			Named(data.Name.ValueString()).
			InVPCPeering(peeringRef).
			Tagged(tags...).
			WithLocalCIDR(data.LocalNetworkAddress.ValueString()).
			WithRemoteCIDR(data.RemoteNetworkAddress.ValueString()).
			BilledBy(aruba.BillingPeriod(data.BillingPeriod.ValueString())),
	)
	if provErr := CheckResponseErr("create", "VPCPeeringRoute", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	// VPCPeeringRoute uses name as ID.
	data.Id = types.StringValue(route.Name())
	data.Uri = strVal(route.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if waitErr := route.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "VPCPeeringRoute", data.Id.ValueString())
		return
	}

	tflog.Trace(ctx, "created a VPC Peering Route resource", map[string]interface{}{
		"vpcpeeringroute_id":   data.Id.ValueString(),
		"vpcpeeringroute_name": data.Name.ValueString(),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpcPeeringRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VpcPeeringRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	route, err := r.client.Client.FromNetwork().VPCPeeringRoutes().Get(ctx, vpcPeeringRouteRef(&data))
	if provErr := CheckResponseErr("read", "VPCPeeringRoute", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	st := string(route.State())
	switch {
	case isFailedState(st):
		resp.Diagnostics.AddWarning("Resource in Failed State",
			fmt.Sprintf("VPCPeeringRoute %q is in a terminal failure state (%s). "+
				"Run `terraform destroy` to clean it up, or `terraform apply -replace=<address>` to recreate it.", data.Id.ValueString(), st))
	case IsCreatingState(st):
		if waitErr := route.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "VPCPeeringRoute", data.Id.ValueString())
			return
		}
		route, err = r.client.Client.FromNetwork().VPCPeeringRoutes().Get(ctx, vpcPeeringRouteRef(&data))
		if provErr := CheckResponseErr("read", "VPCPeeringRoute", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	projectID := data.ProjectId
	vpcID := data.VpcId
	peeringID := data.VpcPeeringId
	applyVPCPeeringRouteToModel(route, &data)
	data.ProjectId = projectID
	data.VpcId = vpcID
	data.VpcPeeringId = peeringID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpcPeeringRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VpcPeeringRouteResourceModel
	var state VpcPeeringRouteResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	route, err := r.client.Client.FromNetwork().VPCPeeringRoutes().Get(ctx, vpcPeeringRouteRef(&state))
	if provErr := CheckResponseErr("read", "VPCPeeringRoute", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	route.Named(data.Name.ValueString())
	if tags != nil {
		route.RetaggedAs(tags...)
	} else {
		route.RetaggedAs(route.Tags()...)
	}

	updated, err := r.client.Client.FromNetwork().VPCPeeringRoutes().Update(ctx, route)
	if provErr := CheckResponseErr("update", "VPCPeeringRoute", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = state.Id
	data.Uri = state.Uri
	data.ProjectId = state.ProjectId
	data.VpcId = state.VpcId
	data.VpcPeeringId = state.VpcPeeringId
	data.LocalNetworkAddress = state.LocalNetworkAddress
	data.RemoteNetworkAddress = state.RemoteNetworkAddress
	data.BillingPeriod = state.BillingPeriod
	data.Name = types.StringValue(updated.Name())
	data.Tags = TagsToListPreserveNull(updated.Tags(), data.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpcPeeringRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VpcPeeringRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := vpcPeeringRouteRef(&data)
	routeID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromNetwork().VPCPeeringRoutes().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "VPCPeeringRoute", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErrAsError("delete", "VPCPeeringRoute",
			r.client.Client.FromNetwork().VPCPeeringRoutes().Delete(ctx, ref))
	}, "VPCPeeringRoute", routeID, r.client.ResourceTimeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting VPCPeeringRoute", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "VPCPeeringRoute", routeID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for VPCPeeringRoute deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a VPC Peering Route resource", map[string]interface{}{"vpcpeeringroute_id": routeID})
}

func (r *VpcPeeringRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
