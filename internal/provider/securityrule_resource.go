package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// protocolNormalizePlanModifier normalizes protocol values during planning
// to prevent case-sensitivity issues.
type protocolNormalizePlanModifier struct{}

func (m protocolNormalizePlanModifier) Description(ctx context.Context) string {
	return "Normalizes protocol values to prevent case-sensitivity issues"
}

func (m protocolNormalizePlanModifier) MarkdownDescription(ctx context.Context) string {
	return "Normalizes protocol values to prevent case-sensitivity issues"
}

func (m protocolNormalizePlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsNull() && !req.PlanValue.IsUnknown() {
		resp.PlanValue = types.StringValue(strings.ToUpper(req.PlanValue.ValueString()))
		return
	}
	if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
		resp.PlanValue = req.StateValue
		return
	}
}

// normalizeProtocol normalizes protocol strings to the canonical API form.
func normalizeProtocol(p string) string {
	switch strings.ToLower(p) {
	case "any":
		return "Any"
	case "tcp":
		return "TCP"
	case "udp":
		return "UDP"
	case "icmp":
		return "ICMP"
	case "":
		return ""
	default:
		if len(p) == 0 {
			return ""
		}
		return strings.ToUpper(p[:1]) + p[1:]
	}
}

// normalizeTargetKind maps the wire "Ip" value to the user-facing "IP".
func normalizeTargetKind(kind string) string {
	if strings.EqualFold(kind, "Ip") || strings.EqualFold(kind, "ip") {
		return "IP"
	}
	return kind
}

type SecurityRuleResource struct {
	client *ArubaCloudClient
}

type SecurityRuleResourceModel struct {
	Id              types.String `tfsdk:"id"`
	Uri             types.String `tfsdk:"uri"`
	Name            types.String `tfsdk:"name"`
	Location        types.String `tfsdk:"location"`
	Tags            types.List   `tfsdk:"tags"`
	ProjectId       types.String `tfsdk:"project_id"`
	VpcId           types.String `tfsdk:"vpc_id"`
	SecurityGroupId types.String `tfsdk:"security_group_id"`
	Properties      types.Object `tfsdk:"properties"`
}

var _ resource.Resource = &SecurityRuleResource{}
var _ resource.ResourceWithImportState = &SecurityRuleResource{}

func NewSecurityRuleResource() resource.Resource {
	return &SecurityRuleResource{}
}

func (r *SecurityRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_securityrule"
}

func (r *SecurityRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an individual ArubaCloud Security Rule within an `arubacloud_securitygroup`. Each rule defines allowed or denied traffic for one direction (inbound or outbound), protocol, port range, and source/destination CIDR. Most rule attributes are immutable after creation; to change them, destroy and re-create the rule.",
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
				MarkdownDescription: "Display name for the security rule.",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier for the resource (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center). (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPC this security rule belongs to. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"security_group_id": schema.StringAttribute{
				MarkdownDescription: "ID of the security group this rule belongs to. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
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
			"properties": schema.SingleNestedAttribute{
				MarkdownDescription: "Traffic-matching properties of the security rule. Most fields are immutable after creation.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"direction": schema.StringAttribute{
						MarkdownDescription: "Traffic direction the rule applies to. Accepted values: `Ingress`, `Egress`. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("Ingress", "Egress"),
						},
					},
					"protocol": schema.StringAttribute{
						MarkdownDescription: "IP protocol. Accepted values: `TCP`, `UDP`, `ICMP`, `ANY` (case-insensitive).",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							protocolNormalizePlanModifier{},
						},
					},
					"port": schema.StringAttribute{
						MarkdownDescription: "Port or port range for TCP/UDP (e.g., `80` or `8080-8090`). Use `0` for ICMP or ANY.",
						Optional:            true,
					},
					"target": schema.SingleNestedAttribute{
						MarkdownDescription: "Source (inbound) or destination (outbound) endpoint for this rule.",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"kind": schema.StringAttribute{
								MarkdownDescription: "Type of the target endpoint. Accepted values: `IP`, `SecurityGroup`.",
								Required:            true,
							},
							"value": schema.StringAttribute{
								MarkdownDescription: "Source (inbound) or destination (outbound) CIDR in notation like `0.0.0.0/0`, or SecurityGroup URI.",
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

func sgRuleRef(data *SecurityRuleResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.SecurityRuleRef(
		data.ProjectId.ValueString(),
		data.VpcId.ValueString(),
		data.SecurityGroupId.ValueString(),
		data.Id.ValueString(),
	)
}

func applySecurityRuleToModel(ctx context.Context, rule *aruba.SecurityRule, data *SecurityRuleResourceModel) {
	data.Id = types.StringValue(rule.ID())
	data.Uri = strVal(rule.URI())
	data.Name = types.StringValue(rule.Name())
	data.Tags = TagsToListPreserveNull(rule.Tags(), data.Tags)

	if rule.Region() != "" {
		data.Location = types.StringValue(string(rule.Region()))
	}

	targetKindStr := normalizeTargetKind(string(rule.TargetKind()))
	targetObj, _ := types.ObjectValue(
		map[string]attr.Type{
			"kind":  types.StringType,
			"value": types.StringType,
		},
		map[string]attr.Value{
			"kind":  types.StringValue(targetKindStr),
			"value": types.StringValue(rule.TargetValue()),
		},
	)

	propertiesObj, _ := types.ObjectValue(
		map[string]attr.Type{
			"direction": types.StringType,
			"protocol":  types.StringType,
			"port":      types.StringType,
			"target": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"kind":  types.StringType,
					"value": types.StringType,
				},
			},
		},
		map[string]attr.Value{
			"direction": types.StringValue(string(rule.Direction())),
			"protocol":  types.StringValue(strings.ToUpper(string(rule.Protocol()))),
			"port":      types.StringValue(rule.Port()),
			"target":    targetObj,
		},
	)
	data.Properties = propertiesObj
}

// extractRuleProperties pulls direction, protocol, port, targetKind, and targetValue
// from the plan's Properties object attribute.
func extractRuleProperties(ctx context.Context, data *SecurityRuleResourceModel) (direction, protocol, port, targetKind, targetValue string, ok bool) {
	if data.Properties.IsNull() || data.Properties.IsUnknown() {
		return
	}
	attrs := data.Properties.Attributes()

	if v, exists := attrs["direction"]; exists {
		if s, isStr := v.(types.String); isStr && !s.IsNull() {
			direction = s.ValueString()
		}
	}
	if v, exists := attrs["protocol"]; exists {
		if s, isStr := v.(types.String); isStr && !s.IsNull() {
			protocol = strings.ToUpper(s.ValueString())
		}
	}
	if v, exists := attrs["port"]; exists {
		if s, isStr := v.(types.String); isStr && !s.IsNull() {
			port = s.ValueString()
		}
	}
	if protocol == "ANY" || protocol == "ICMP" {
		port = ""
	}
	if v, exists := attrs["target"]; exists {
		if targetObjAttr, isObj := v.(types.Object); isObj && !targetObjAttr.IsNull() {
			tAttrs := targetObjAttr.Attributes()
			if k, exists := tAttrs["kind"]; exists {
				if s, isStr := k.(types.String); isStr && !s.IsNull() {
					targetKind = normalizeTargetKind(s.ValueString())
				}
			}
			if val, exists := tAttrs["value"]; exists {
				if s, isStr := val.(types.String); isStr && !s.IsNull() {
					targetValue = s.ValueString()
				}
			}
		}
	}
	ok = direction != "" && protocol != "" && targetKind != ""
	return
}

func (r *SecurityRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecurityRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	securityGroupID := data.SecurityGroupId.ValueString()

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	direction, protocol, port, targetKind, targetValue, ok := extractRuleProperties(ctx, &data)
	if !ok {
		resp.Diagnostics.AddError("Invalid Properties", "direction, protocol, and target are required")
		return
	}

	sgURI := aruba.SecurityGroupRef(projectID, vpcID, securityGroupID)
	builder := aruba.NewSecurityRule().
		Named(data.Name.ValueString()).
		InSecurityGroup(sgURI).
		InRegion(aruba.Region(data.Location.ValueString())).
		Tagged(tags...).
		WithDirection(aruba.RuleDirection(direction)).
		WithProtocol(aruba.RuleProtocol(protocol))

	if port != "" {
		builder = builder.WithPort(port)
	}

	if targetKind == "IP" {
		builder = builder.TargetingCIDR(targetValue)
	} else {
		builder = builder.TargetingSecurityGroup(aruba.URI(targetValue))
	}

	rule, err := r.client.Client.FromNetwork().SecurityGroupRules().Create(ctx, builder)
	if provErr := CheckResponseErr("create", "SecurityRule", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(rule.ID())
	data.Uri = strVal(rule.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if waitErr := rule.WaitUntilReady(ctx, aruba.WithTimeout(r.client.ResourceTimeout)); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "SecurityRule", data.Id.ValueString())
		return
	}

	fresh, freshErr := r.client.Client.FromNetwork().SecurityGroupRules().Get(ctx, sgRuleRef(&data))
	if freshErr == nil {
		applySecurityRuleToModel(ctx, fresh, &data)
	} else {
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh SecurityRule after creation: %v", freshErr))
	}

	tflog.Trace(ctx, "created a SecurityRule resource", map[string]interface{}{
		"securityrule_id":   data.Id.ValueString(),
		"securityrule_name": data.Name.ValueString(),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	rule, err := r.client.Client.FromNetwork().SecurityGroupRules().Get(ctx, sgRuleRef(&data))
	if provErr := CheckResponseErr("read", "SecurityRule", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	st := string(rule.State())
	switch {
	case isFailedState(st):
		resp.Diagnostics.AddWarning("Resource in Failed State",
			fmt.Sprintf("SecurityRule %q is in a terminal failure state (%s). "+
				"Run `terraform destroy` to clean it up, or `terraform apply -replace=<address>` to recreate it.", data.Id.ValueString(), st))
	case IsCreatingState(st):
		if waitErr := rule.WaitUntilReady(ctx, aruba.WithTimeout(r.client.ResourceTimeout)); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "SecurityRule", data.Id.ValueString())
			return
		}
		rule, err = r.client.Client.FromNetwork().SecurityGroupRules().Get(ctx, sgRuleRef(&data))
		if provErr := CheckResponseErr("read", "SecurityRule", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	projectID := data.ProjectId
	vpcID := data.VpcId
	sgID := data.SecurityGroupId
	applySecurityRuleToModel(ctx, rule, &data)
	data.ProjectId = projectID
	data.VpcId = vpcID
	data.SecurityGroupId = sgID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SecurityRuleResourceModel
	var state SecurityRuleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.Client.FromNetwork().SecurityGroupRules().Get(ctx, sgRuleRef(&state))
	if provErr := CheckResponseErr("read", "SecurityRule", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	// Properties are immutable; reject if they differ from current.
	planDirection, planProtocol, planPort, planTargetKind, planTargetValue, _ := extractRuleProperties(ctx, &data)
	var propertiesChanged bool
	if planDirection != "" && string(rule.Direction()) != planDirection {
		propertiesChanged = true
	}
	if planProtocol != "" && strings.ToUpper(string(rule.Protocol())) != strings.ToUpper(planProtocol) {
		propertiesChanged = true
	}
	if planPort != "" && rule.Port() != planPort {
		propertiesChanged = true
	}
	if planTargetKind != "" && normalizeTargetKind(string(rule.TargetKind())) != planTargetKind {
		propertiesChanged = true
	}
	if planTargetValue != "" && rule.TargetValue() != planTargetValue {
		propertiesChanged = true
	}
	if propertiesChanged {
		resp.Diagnostics.AddError(
			"Cannot Update Security Rule Properties",
			"Security rule properties (direction, protocol, port, target) cannot be updated. To change properties, delete and recreate the security rule.",
		)
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	rule.Named(data.Name.ValueString())
	if tags != nil {
		rule.RetaggedAs(tags...)
	} else {
		rule.RetaggedAs(rule.Tags()...)
	}

	updated, err := r.client.Client.FromNetwork().SecurityGroupRules().Update(ctx, rule)
	if provErr := CheckResponseErr("update", "SecurityRule", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	// Preserve immutable fields from state.
	data.Id = state.Id
	data.Uri = state.Uri
	data.ProjectId = state.ProjectId
	data.VpcId = state.VpcId
	data.SecurityGroupId = state.SecurityGroupId
	data.Name = types.StringValue(updated.Name())
	data.Tags = TagsToListPreserveNull(updated.Tags(), data.Tags)
	// Properties remain unchanged.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SecurityRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := sgRuleRef(&data)
	ruleID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromNetwork().SecurityGroupRules().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "SecurityRule", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErr("delete", "SecurityRule",
			r.client.Client.FromNetwork().SecurityGroupRules().Delete(ctx, ref))
	}, "SecurityRule", ruleID, r.client.ResourceTimeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting SecurityRule", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "SecurityRule", ruleID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for SecurityRule deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a SecurityRule resource", map[string]interface{}{"securityrule_id": ruleID})
}

func (r *SecurityRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
