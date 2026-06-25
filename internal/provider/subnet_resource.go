package provider

import (
	"context"
	"fmt"
	"net"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &SubnetResource{}
var _ resource.ResourceWithImportState = &SubnetResource{}
var _ resource.ResourceWithConfigValidators = &SubnetResource{}

func NewSubnetResource() resource.Resource {
	return &SubnetResource{}
}

type SubnetResource struct {
	client *ArubaCloudClient
}

type SubnetResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	Tags      types.List   `tfsdk:"tags"`
	ProjectId types.String `tfsdk:"project_id"`
	VpcId     types.String `tfsdk:"vpc_id"`
	Type      types.String `tfsdk:"type"`
	Network   types.Object `tfsdk:"network"`
}

type NetworkModel struct {
	Address types.String `tfsdk:"address"`
	Dhcp    types.Object `tfsdk:"dhcp"`
}

type DhcpModel struct {
	Enabled types.Bool   `tfsdk:"enabled"`
	Range   types.Object `tfsdk:"range"`
	Routes  types.List   `tfsdk:"routes"`
	Dns     types.List   `tfsdk:"dns"`
}

type DhcpRangeModel struct {
	Start types.String `tfsdk:"start"`
	Count types.Int64  `tfsdk:"count"`
}

type RouteModel struct {
	Address types.String `tfsdk:"address"`
	Gateway types.String `tfsdk:"gateway"`
}

func (r *SubnetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (r *SubnetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an ArubaCloud Subnet within a VPC. Subnets can be Basic (default networking) or Advanced (custom CIDR with configurable DHCP pools, routes, and DNS). Changing `type`, `location`, `project_id`, or `vpc_id` destroys and re-creates the subnet. For `Advanced` subnets the `network` block is mandatory and must include a valid RFC-1918 CIDR.",
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
				MarkdownDescription: "Display name for the subnet.",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier for the resource (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center). (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the parent VPC this subnet belongs to. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Subnet type. Accepted values: `Basic` (no custom CIDR), `Advanced` (requires the `network` block). (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("Basic", "Advanced"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network": schema.SingleNestedAttribute{
				MarkdownDescription: "Network configuration block. Required when `type` is `Advanced`.",
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						MarkdownDescription: "Subnet CIDR in RFC-1918 notation (e.g., `10.0.1.0/24`). Must fall within the parent VPC CIDR.",
						Optional:            true,
					},
					"dhcp": schema.SingleNestedAttribute{
						MarkdownDescription: "DHCP configuration for the subnet.",
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								MarkdownDescription: "Whether DHCP is enabled on this subnet.",
								Optional:            true,
							},
							"range": schema.SingleNestedAttribute{
								MarkdownDescription: "IP address range allocated to DHCP clients.",
								Attributes: map[string]schema.Attribute{
									"start": schema.StringAttribute{
										MarkdownDescription: "First IP address in the DHCP allocation range.",
										Optional:            true,
									},
									"count": schema.Int64Attribute{
										MarkdownDescription: "Number of consecutive IP addresses in the DHCP pool.",
										Optional:            true,
									},
								},
								Optional: true,
							},
							"routes": schema.ListNestedAttribute{
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"address": schema.StringAttribute{
											MarkdownDescription: "Destination network in CIDR notation. Must be within the subnet's `network.address` CIDR block (e.g., `10.0.1.128/25` when the subnet is `10.0.1.0/24`).",
											Optional:            true,
										},
										"gateway": schema.StringAttribute{
											MarkdownDescription: "Gateway IP address for this route.",
											Optional:            true,
										},
									},
								},
								MarkdownDescription: "Static routes distributed to DHCP clients.",
								Optional:            true,
							},
							"dns": schema.ListAttribute{
								ElementType:         types.StringType,
								MarkdownDescription: "List of DNS server IP addresses distributed to DHCP clients.",
								Optional:            true,
							},
						},
						Optional: true,
					},
				},
				Optional: true,
			},
		},
	}
}

func (r *SubnetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func subnetRef(data *SubnetResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.SubnetRef(data.ProjectId.ValueString(), data.VpcId.ValueString(), data.Id.ValueString())
}

// buildSubnetDHCP builds a *aruba.SubnetDHCPCommon from a dhcp Object attribute.
func buildSubnetDHCP(ctx context.Context, dhcpAttr types.Object, diags *diag.Diagnostics) *aruba.SubnetDHCPCommon {
	if dhcpAttr.IsNull() || dhcpAttr.IsUnknown() {
		return nil
	}
	dhcpAttrs := dhcpAttr.Attributes()
	dhcp := aruba.NewSubnetDHCP()

	if v, ok := dhcpAttrs["enabled"]; ok && v != nil {
		if b, ok := v.(types.Bool); ok && !b.IsNull() && b.ValueBool() {
			dhcp.Enabled()
		}
	}

	if v, ok := dhcpAttrs["range"]; ok && v != nil {
		if rangeObj, ok := v.(types.Object); ok && !rangeObj.IsNull() {
			rangeAttrs := rangeObj.Attributes()
			start := ""
			count := 0
			if sv, ok := rangeAttrs["start"]; ok && sv != nil {
				if s, ok := sv.(types.String); ok && !s.IsNull() {
					start = s.ValueString()
				}
			}
			if cv, ok := rangeAttrs["count"]; ok && cv != nil {
				if c, ok := cv.(types.Int64); ok && !c.IsNull() {
					count = int(c.ValueInt64())
				}
			}
			if start != "" || count > 0 {
				dhcp.WithRange(start, count)
			}
		}
	}

	if v, ok := dhcpAttrs["routes"]; ok && v != nil {
		if routesList, ok := v.(types.List); ok && !routesList.IsNull() {
			var routeObjs []types.Object
			d := routesList.ElementsAs(ctx, &routeObjs, false)
			diags.Append(d...)
			if !diags.HasError() {
				for _, routeObj := range routeObjs {
					routeAttrs := routeObj.Attributes()
					addr, gw := "", ""
					if av, ok := routeAttrs["address"]; ok && av != nil {
						if s, ok := av.(types.String); ok && !s.IsNull() {
							addr = s.ValueString()
						}
					}
					if gv, ok := routeAttrs["gateway"]; ok && gv != nil {
						if s, ok := gv.(types.String); ok && !s.IsNull() {
							gw = s.ValueString()
						}
					}
					if addr != "" || gw != "" {
						dhcp.WithRoutes(aruba.SubnetDHCPRouteCommon{Address: addr, Gateway: gw})
					}
				}
			}
		}
	}

	if v, ok := dhcpAttrs["dns"]; ok && v != nil {
		if dnsList, ok := v.(types.List); ok && !dnsList.IsNull() {
			var dnsServers []string
			d := dnsList.ElementsAs(ctx, &dnsServers, false)
			diags.Append(d...)
			if !diags.HasError() && len(dnsServers) > 0 {
				dhcp.WithDNSServers(dnsServers...)
			}
		}
	}

	return dhcp
}

// applySubnetToModel hydrates data from the wrapper response.
// It preserves project_id and vpc_id from the caller (those aren't in the API response).
func applySubnetToModel(_ context.Context, subnet *aruba.Subnet, data *SubnetResourceModel, diags *diag.Diagnostics) {
	data.Id = types.StringValue(subnet.ID())
	data.Uri = strVal(subnet.URI())
	data.Name = types.StringValue(subnet.Name())
	data.Tags = TagsToListPreserveNull(subnet.Tags(), data.Tags)
	if subnet.Region() != "" {
		data.Location = types.StringValue(string(subnet.Region()))
	}
	data.Type = types.StringValue(string(subnet.Type()))

	networkAttrTypes := subnetNetworkAttrTypes()
	networkWasInState := !data.Network.IsNull() && !data.Network.IsUnknown()
	shouldSetNetwork := string(subnet.Type()) == "Advanced" || networkWasInState

	if !shouldSetNetwork {
		data.Network = types.ObjectNull(networkAttrTypes)
		return
	}

	networkAttrs := make(map[string]attr.Value)
	if cidr := subnet.CIDR(); cidr != "" {
		networkAttrs["address"] = types.StringValue(cidr)
	} else {
		networkAttrs["address"] = types.StringNull()
	}

	dhcpAttrTypes := subnetDHCPAttrTypes()
	dhcp := subnet.DHCP()
	if dhcp != nil {
		dhcpAttrs := map[string]attr.Value{
			"enabled": types.BoolValue(dhcp.IsEnabled()),
		}

		if dhcp.RangeStart() != "" || dhcp.RangeCount() > 0 {
			rangeObj, d := types.ObjectValue(
				map[string]attr.Type{"start": types.StringType, "count": types.Int64Type},
				map[string]attr.Value{
					"start": types.StringValue(dhcp.RangeStart()),
					"count": types.Int64Value(int64(dhcp.RangeCount())),
				},
			)
			diags.Append(d...)
			dhcpAttrs["range"] = rangeObj
		} else {
			dhcpAttrs["range"] = types.ObjectNull(map[string]attr.Type{
				"start": types.StringType, "count": types.Int64Type,
			})
		}

		routes := dhcp.Routes()
		if len(routes) > 0 {
			routeObjType := routeObjectType()
			routeObjs := make([]attr.Value, 0, len(routes))
			for _, route := range routes {
				routeObj, d := types.ObjectValue(routeObjType.AttrTypes, map[string]attr.Value{
					"address": types.StringValue(route.Address),
					"gateway": types.StringValue(route.Gateway),
				})
				diags.Append(d...)
				routeObjs = append(routeObjs, routeObj)
			}
			routesList, d := types.ListValue(routeObjType, routeObjs)
			diags.Append(d...)
			dhcpAttrs["routes"] = routesList
		} else {
			dhcpAttrs["routes"] = types.ListNull(routeObjectType())
		}

		dns := dhcp.DNS()
		if len(dns) > 0 {
			dnsVals := make([]attr.Value, len(dns))
			for i, s := range dns {
				dnsVals[i] = types.StringValue(s)
			}
			dnsList, d := types.ListValue(types.StringType, dnsVals)
			diags.Append(d...)
			dhcpAttrs["dns"] = dnsList
		} else {
			dhcpAttrs["dns"] = types.ListNull(types.StringType)
		}

		dhcpObj, d := types.ObjectValue(dhcpAttrTypes, dhcpAttrs)
		diags.Append(d...)
		networkAttrs["dhcp"] = dhcpObj
	} else {
		networkAttrs["dhcp"] = types.ObjectNull(dhcpAttrTypes)
	}

	networkObj, d := types.ObjectValue(networkAttrTypes, networkAttrs)
	diags.Append(d...)
	data.Network = networkObj
}

func subnetNetworkAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"address": types.StringType,
		"dhcp":    types.ObjectType{AttrTypes: subnetDHCPAttrTypes()},
	}
}

func subnetDHCPAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled": types.BoolType,
		"range": types.ObjectType{
			AttrTypes: map[string]attr.Type{"start": types.StringType, "count": types.Int64Type},
		},
		"routes": types.ListType{ElemType: routeObjectType()},
		"dns":    types.ListType{ElemType: types.StringType},
	}
}

func routeObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{"address": types.StringType, "gateway": types.StringType},
	}
}

func (r *SubnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	subnetTypeStr := data.Type.ValueString()

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate Advanced subnet requirements.
	if subnetTypeStr == "Advanced" {
		if data.Network.IsNull() || data.Network.IsUnknown() {
			resp.Diagnostics.AddError("Missing Required Field",
				"The 'network' block is required when subnet type is 'Advanced'")
			return
		}
	}

	vpcURI := aruba.URI("/projects/" + projectID + "/network/vpcs/" + vpcID)
	subnetType := aruba.SubnetTypeBasic
	if subnetTypeStr == "Advanced" {
		subnetType = aruba.SubnetTypeAdvanced
	}

	builder := aruba.NewSubnet().
		Named(data.Name.ValueString()).
		InVPC(vpcURI).
		InRegion(aruba.Region(data.Location.ValueString())).
		OfType(subnetType).
		Tagged(tags...)

	if !data.Network.IsNull() && !data.Network.IsUnknown() {
		networkObj, d := data.Network.ToObjectValue(ctx)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		attrs := networkObj.Attributes()

		addressValue := ""
		if av, ok := attrs["address"]; ok && av != nil {
			if s, ok := av.(types.String); ok && !s.IsNull() {
				addressValue = s.ValueString()
			}
		}

		if subnetTypeStr == "Advanced" && addressValue == "" {
			resp.Diagnostics.AddError("Missing Required Field",
				"The 'network.address' field is required when subnet type is 'Advanced'")
			return
		}

		if addressValue != "" {
			builder = builder.WithCIDR(addressValue)
		}

		if dhcpAttrVal, ok := attrs["dhcp"]; ok && dhcpAttrVal != nil {
			if dhcpObj, ok := dhcpAttrVal.(types.Object); ok && !dhcpObj.IsNull() {
				// Validate Advanced DHCP requirements.
				if subnetTypeStr == "Advanced" {
					dhcpAttrs := dhcpObj.Attributes()
					dhcpEnabledSet := false
					rangeStart, rangeCount := "", 0
					if ev, ok := dhcpAttrs["enabled"]; ok && ev != nil {
						if b, ok := ev.(types.Bool); ok && !b.IsNull() {
							dhcpEnabledSet = true
							_ = dhcpEnabledSet
						}
					}
					if rv, ok := dhcpAttrs["range"]; ok && rv != nil {
						if rangeObj, ok := rv.(types.Object); ok && !rangeObj.IsNull() {
							if sv, ok := rangeObj.Attributes()["start"]; ok {
								if s, ok := sv.(types.String); ok && !s.IsNull() {
									rangeStart = s.ValueString()
								}
							}
							if cv, ok := rangeObj.Attributes()["count"]; ok {
								if c, ok := cv.(types.Int64); ok && !c.IsNull() {
									rangeCount = int(c.ValueInt64())
								}
							}
						}
					}
					if rangeStart == "" || rangeCount == 0 {
						resp.Diagnostics.AddError("Missing Required Fields",
							"The 'network.dhcp.range' block with 'start' and 'count' fields is required when subnet type is 'Advanced'")
						return
					}
				}
				dhcp := buildSubnetDHCP(ctx, dhcpObj, &resp.Diagnostics)
				if resp.Diagnostics.HasError() {
					return
				}
				if dhcp != nil {
					builder = builder.WithDHCP(dhcp)
				}
			} else if subnetTypeStr == "Advanced" {
				resp.Diagnostics.AddError("Missing Required Field",
					"The 'network.dhcp' block is required when subnet type is 'Advanced'")
				return
			}
		} else if subnetTypeStr == "Advanced" {
			resp.Diagnostics.AddError("Missing Required Field",
				"The 'network.dhcp' block is required when subnet type is 'Advanced'")
			return
		}
	}

	subnet, err := r.client.Client.FromNetwork().Subnets().Create(ctx, builder)
	if provErr := CheckResponseErr("create", "Subnet", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(subnet.ID())
	data.Uri = strVal(subnet.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if waitErr := subnet.WaitUntilReady(ctx, aruba.WithTimeout(r.client.ResourceTimeout)); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "Subnet", data.Id.ValueString())
		return
	}

	fresh, freshErr := r.client.Client.FromNetwork().Subnets().Get(ctx, subnetRef(&data))
	if freshErr == nil {
		projectID := data.ProjectId
		vpcID := data.VpcId
		applySubnetToModel(ctx, fresh, &data, &resp.Diagnostics)
		data.ProjectId = projectID
		data.VpcId = vpcID
	} else {
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh Subnet after creation: %v", freshErr))
	}

	tflog.Trace(ctx, "created a Subnet resource", map[string]interface{}{
		"subnet_id":   data.Id.ValueString(),
		"subnet_name": data.Name.ValueString(),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	subnet, err := r.client.Client.FromNetwork().Subnets().Get(ctx, subnetRef(&data))
	if provErr := CheckResponseErr("read", "Subnet", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	st := string(subnet.State())
	switch {
	case isFailedState(st):
		resp.Diagnostics.AddWarning("Resource in Failed State",
			fmt.Sprintf("Subnet %q is in a terminal failure state (%s). "+
				"Run `terraform destroy` to clean it up, or `terraform apply -replace=<address>` to recreate it.", data.Id.ValueString(), st))
	case IsCreatingState(st):
		if waitErr := subnet.WaitUntilReady(ctx, aruba.WithTimeout(r.client.ResourceTimeout)); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "Subnet", data.Id.ValueString())
			return
		}
		subnet, err = r.client.Client.FromNetwork().Subnets().Get(ctx, subnetRef(&data))
		if provErr := CheckResponseErr("read", "Subnet", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	projectID := data.ProjectId
	vpcID := data.VpcId
	applySubnetToModel(ctx, subnet, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ProjectId = projectID
	data.VpcId = vpcID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SubnetResourceModel
	var state SubnetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subnet, err := r.client.Client.FromNetwork().Subnets().Get(ctx, subnetRef(&state))
	if provErr := CheckResponseErr("read", "Subnet", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	subnet.Named(data.Name.ValueString())
	if tags != nil {
		subnet.RetaggedAs(tags...)
	} else {
		subnet.RetaggedAs(subnet.Tags()...)
	}

	// Update DHCP if the plan provides a network block.
	if !data.Network.IsNull() && !data.Network.IsUnknown() {
		networkObj, d := data.Network.ToObjectValue(ctx)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		attrs := networkObj.Attributes()
		if dhcpAttrVal, ok := attrs["dhcp"]; ok && dhcpAttrVal != nil {
			if dhcpObj, ok := dhcpAttrVal.(types.Object); ok && !dhcpObj.IsNull() {
				newDHCP := buildSubnetDHCP(ctx, dhcpObj, &resp.Diagnostics)
				if resp.Diagnostics.HasError() {
					return
				}
				if newDHCP != nil {
					subnet.WithDHCP(newDHCP)
				}
			}
		}
	}

	updated, err := r.client.Client.FromNetwork().Subnets().Update(ctx, subnet)
	if provErr := CheckResponseErr("update", "Subnet", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = state.Id
	data.ProjectId = state.ProjectId
	data.VpcId = state.VpcId
	data.Uri = state.Uri

	projectID := data.ProjectId
	vpcID := data.VpcId
	applySubnetToModel(ctx, updated, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ProjectId = projectID
	data.VpcId = vpcID
	data.Id = state.Id
	data.Uri = state.Uri

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := subnetRef(&data)
	subnetID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromNetwork().Subnets().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Subnet", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErrAsError("delete", "Subnet",
			r.client.Client.FromNetwork().Subnets().Delete(ctx, ref))
	}, "Subnet", subnetID, r.client.ResourceTimeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Subnet", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "Subnet", subnetID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for Subnet deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a Subnet resource", map[string]interface{}{"subnet_id": subnetID})
}

func (r *SubnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *SubnetResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		subnetRouteAddressValidator{},
	}
}

type subnetRouteAddressValidator struct{}

func (v subnetRouteAddressValidator) Description(_ context.Context) string {
	return "Route addresses must be within the subnet's network CIDR block."
}

func (v subnetRouteAddressValidator) MarkdownDescription(_ context.Context) string {
	return "Route addresses must be within the subnet's network CIDR block."
}

func (v subnetRouteAddressValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Type.IsNull() || data.Type.IsUnknown() || data.Type.ValueString() != "Advanced" {
		return
	}
	if data.Network.IsNull() || data.Network.IsUnknown() {
		return
	}

	var networkModel NetworkModel
	resp.Diagnostics.Append(data.Network.As(ctx, &networkModel, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if networkModel.Address.IsNull() || networkModel.Address.IsUnknown() {
		return
	}
	networkAddr := networkModel.Address.ValueString()
	if networkAddr == "" {
		return
	}
	_, subnetNet, err := net.ParseCIDR(networkAddr)
	if err != nil {
		return
	}

	if networkModel.Dhcp.IsNull() || networkModel.Dhcp.IsUnknown() {
		return
	}

	var dhcpModel DhcpModel
	resp.Diagnostics.Append(networkModel.Dhcp.As(ctx, &dhcpModel, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if dhcpModel.Routes.IsNull() || dhcpModel.Routes.IsUnknown() {
		return
	}

	var routeObjs []types.Object
	resp.Diagnostics.Append(dhcpModel.Routes.ElementsAs(ctx, &routeObjs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for i, routeObj := range routeObjs {
		addrAttr, ok := routeObj.Attributes()["address"]
		if !ok {
			continue
		}
		addrStr, ok := addrAttr.(types.String)
		if !ok || addrStr.IsNull() || addrStr.IsUnknown() {
			continue
		}
		routeAddr := addrStr.ValueString()
		if routeAddr == "" {
			continue
		}
		_, routeNet, err := net.ParseCIDR(routeAddr)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("network").AtName("dhcp").AtName("routes").AtListIndex(i).AtName("address"),
				"Invalid Route Address",
				fmt.Sprintf("Route address %q is not valid CIDR notation.", routeAddr),
			)
			continue
		}
		if !cidrContains(subnetNet, routeNet) {
			resp.Diagnostics.AddAttributeError(
				path.Root("network").AtName("dhcp").AtName("routes").AtListIndex(i).AtName("address"),
				"Route Address Outside Subnet CIDR",
				fmt.Sprintf("Route address %q must be within the subnet CIDR %q. "+
					"ArubaCloud requires all route addresses to be covered by the subnet's address block.", routeAddr, networkAddr),
			)
		}
	}
}

// cidrContains reports whether child is a subnet of (or equal to) parent.
func cidrContains(parent, child *net.IPNet) bool {
	parentOnes, parentBits := parent.Mask.Size()
	childOnes, childBits := child.Mask.Size()
	if parentBits != childBits {
		return false
	}
	return childOnes >= parentOnes && parent.Contains(child.IP)
}
