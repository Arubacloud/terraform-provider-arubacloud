package provider

import (
	"context"
	"fmt"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &VPNRouteResource{}
var _ resource.ResourceWithImportState = &VPNRouteResource{}

func NewVPNRouteResource() resource.Resource {
	return &VPNRouteResource{}
}

type VPNRouteResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Uri         types.String `tfsdk:"uri"`
	Name        types.String `tfsdk:"name"`
	Location    types.String `tfsdk:"location"`
	Tags        types.List   `tfsdk:"tags"`
	ProjectId   types.String `tfsdk:"project_id"`
	VPNTunnelId types.String `tfsdk:"vpn_tunnel_id"`
	Properties  types.Object `tfsdk:"properties"`
}

type VPNRouteResource struct {
	client *ArubaCloudClient
}

func (r *VPNRouteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpnroute"
}

func (r *VPNRouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a static route associated with an ArubaCloud VPN Tunnel. The route instructs the ArubaCloud gateway to forward traffic for a specified CIDR over the parent VPN tunnel.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource.",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources (e.g., as a `*_uri_ref` attribute).",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the VPN route.",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier for the resource (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
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
			"vpn_tunnel_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPN tunnel this route is associated with.",
				Required:            true,
			},
			"properties": schema.SingleNestedAttribute{
				MarkdownDescription: "Routing properties for the VPN route.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"cloud_subnet": schema.StringAttribute{
						MarkdownDescription: "CIDR of the ArubaCloud-side subnet to route over this tunnel (e.g., `10.0.1.0/24`).",
						Required:            true,
					},
					"on_prem_subnet": schema.StringAttribute{
						MarkdownDescription: "CIDR of the on-premises subnet reachable through this tunnel (e.g., `192.168.1.0/24`).",
						Required:            true,
					},
				},
			},
		},
	}
}

func (r *VPNRouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func vpnRouteRef(data *VPNRouteResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.VPNRouteRef(data.ProjectId.ValueString(), data.VPNTunnelId.ValueString(), data.Id.ValueString())
}

func extractVPNRouteSubnets(data *VPNRouteResourceModel) (cloudSubnet, onPremSubnet string) {
	if data.Properties.IsNull() || data.Properties.IsUnknown() {
		return
	}
	attrs := data.Properties.Attributes()
	if v, ok := attrs["cloud_subnet"]; ok {
		if s, ok := v.(types.String); ok && !s.IsNull() {
			cloudSubnet = s.ValueString()
		}
	}
	if v, ok := attrs["on_prem_subnet"]; ok {
		if s, ok := v.(types.String); ok && !s.IsNull() {
			onPremSubnet = s.ValueString()
		}
	}
	return
}

func (r *VPNRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VPNRouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpnTunnelID := data.VPNTunnelId.ValueString()

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	cloudSubnet, onPremSubnet := extractVPNRouteSubnets(&data)

	tunnelRef := aruba.VPNTunnelRef(projectID, vpnTunnelID)
	route, err := r.client.Client.FromNetwork().VPNRoutes().Create(ctx,
		aruba.NewVPNRoute().
			Named(data.Name.ValueString()).
			InVPNTunnel(tunnelRef).
			InRegion(aruba.Region(data.Location.ValueString())).
			Tagged(tags...).
			WithCloudSubnet(cloudSubnet).
			WithOnPremSubnet(onPremSubnet),
	)
	if provErr := CheckResponseErr("create", "VPNRoute", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(route.ID())
	data.Uri = strVal(route.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if waitErr := route.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "VPNRoute", data.Id.ValueString())
		return
	}

	tflog.Trace(ctx, "created a VPN Route resource", map[string]interface{}{
		"vpnroute_id":   data.Id.ValueString(),
		"vpnroute_name": data.Name.ValueString(),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPNRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VPNRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	route, err := r.client.Client.FromNetwork().VPNRoutes().Get(ctx, vpnRouteRef(&data))
	if provErr := CheckResponseErr("read", "VPNRoute", err); provErr != nil {
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
			fmt.Sprintf("VPNRoute %q is in a terminal failure state (%s). "+
				"Run `terraform destroy` to clean it up, or `terraform apply -replace=<address>` to recreate it.", data.Id.ValueString(), st))
	case IsCreatingState(st):
		if waitErr := route.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "VPNRoute", data.Id.ValueString())
			return
		}
		route, err = r.client.Client.FromNetwork().VPNRoutes().Get(ctx, vpnRouteRef(&data))
		if provErr := CheckResponseErr("read", "VPNRoute", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	data.Id = types.StringValue(route.ID())
	data.Uri = strVal(route.URI())
	data.Name = types.StringValue(route.Name())
	data.Tags = TagsToListPreserveNull(route.Tags(), data.Tags)
	if route.Region() != "" {
		data.Location = types.StringValue(string(route.Region()))
	}

	propertiesObj, diags := types.ObjectValue(
		map[string]attr.Type{
			"cloud_subnet":   types.StringType,
			"on_prem_subnet": types.StringType,
		},
		map[string]attr.Value{
			"cloud_subnet":   types.StringValue(route.CloudSubnet()),
			"on_prem_subnet": types.StringValue(route.OnPremSubnet()),
		},
	)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		data.Properties = propertiesObj
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPNRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VPNRouteResourceModel
	var state VPNRouteResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	route, err := r.client.Client.FromNetwork().VPNRoutes().Get(ctx, vpnRouteRef(&state))
	if provErr := CheckResponseErr("read", "VPNRoute", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	route.Named(data.Name.ValueString())
	if tags != nil {
		route.RetaggedAs(tags...)
	} else {
		route.RetaggedAs(route.Tags()...)
	}
	// Update subnets from plan.
	cloudSubnet, onPremSubnet := extractVPNRouteSubnets(&data)
	if cloudSubnet != "" {
		route.WithCloudSubnet(cloudSubnet)
	}
	if onPremSubnet != "" {
		route.WithOnPremSubnet(onPremSubnet)
	}

	updated, err := r.client.Client.FromNetwork().VPNRoutes().Update(ctx, route)
	if provErr := CheckResponseErr("update", "VPNRoute", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = state.Id
	data.Uri = state.Uri
	data.ProjectId = state.ProjectId
	data.VPNTunnelId = state.VPNTunnelId
	data.Location = state.Location
	data.Name = types.StringValue(updated.Name())
	data.Tags = TagsToListPreserveNull(updated.Tags(), data.Tags)

	propertiesObj, diags := types.ObjectValue(
		map[string]attr.Type{
			"cloud_subnet":   types.StringType,
			"on_prem_subnet": types.StringType,
		},
		map[string]attr.Value{
			"cloud_subnet":   types.StringValue(updated.CloudSubnet()),
			"on_prem_subnet": types.StringValue(updated.OnPremSubnet()),
		},
	)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		data.Properties = propertiesObj
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPNRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VPNRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := vpnRouteRef(&data)
	routeID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromNetwork().VPNRoutes().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "VPNRoute", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErrAsError("delete", "VPNRoute",
			r.client.Client.FromNetwork().VPNRoutes().Delete(ctx, ref))
	}, "VPNRoute", routeID, r.client.ResourceTimeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting VPNRoute", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "VPNRoute", routeID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for VPNRoute deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a VPN Route resource", map[string]interface{}{"vpnroute_id": routeID})
}

func (r *VPNRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
