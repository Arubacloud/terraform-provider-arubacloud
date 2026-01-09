// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// RuleDirection represents the direction of a security rule.
type RuleDirection string

const (
	RuleDirectionIngress RuleDirection = "Ingress"
	RuleDirectionEgress  RuleDirection = "Egress"
)

type SecurityRuleResource struct {
	client *ArubaCloudClient
}

// EndpointTypeDto represents the type of target endpoint
// ...existing code...
type EndpointTypeDto string

const (
	EndpointTypeIP EndpointTypeDto = "Ip"
)

// normalizeProtocol normalizes protocol values to match API expectations
// API expects: Any, TCP, UDP, ICMP (case-sensitive)
func normalizeProtocol(protocol string) string {
	protocolUpper := strings.ToUpper(protocol)
	switch protocolUpper {
	case "ANY":
		return "Any"
	case "TCP":
		return "TCP"
	case "UDP":
		return "UDP"
	case "ICMP":
		return "ICMP"
	default:
		// If it's already in the correct format (e.g., "Any"), return as-is
		// Otherwise, try to capitalize first letter
		if len(protocol) > 0 {
			return strings.ToUpper(protocol[:1]) + strings.ToLower(protocol[1:])
		}
		return protocol
	}
}

// normalizeTargetKind normalizes target kind values to match API expectations
// API expects: IP (not Ip, ip, etc.)
func normalizeTargetKind(kind string) string {
	kindUpper := strings.ToUpper(kind)
	switch kindUpper {
	case "IP":
		return "IP"
	case "SECURITYGROUP":
		return "SecurityGroup"
	default:
		// If it's already in the correct format, return as-is
		// For "Ip", convert to "IP"
		if strings.EqualFold(kind, "Ip") {
			return "IP"
		}
		return kind
	}
}

// RuleTarget represents the target of the rule (source or destination according to the direction).
type RuleTarget struct {
	Kind  EndpointTypeDto `tfsdk:"kind"`
	Value string          `tfsdk:"value"`
}

// SecurityRuleProperties contains the properties of a security rule.
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
	Uri             types.String `tfsdk:"uri"`
	Name            types.String `tfsdk:"name"`
	Location        types.String `tfsdk:"location"`
	Tags            types.List   `tfsdk:"tags"`
	ProjectId       types.String `tfsdk:"project_id"`
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Security Rule URI",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"security_group_id": schema.StringAttribute{
				MarkdownDescription: "ID of the Security Group this rule belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the Security Rule",
				Optional:            true,
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
	client, ok := req.ProviderData.(*ArubaCloudClient)
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

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	securityGroupID := data.SecurityGroupId.ValueString()

	if projectID == "" || vpcID == "" || securityGroupID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, and Security Group ID are required to create a security rule",
		)
		return
	}

	// Extract tags - only include if actually provided
	// CLI doesn't include tags in JSON if not provided (SDK omits empty slices with omitempty)
	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	// Don't set tags to empty slice - let SDK handle it (will be omitted with omitempty)
	// This matches CLI behavior where tags are only included if provided

	// Extract properties from Terraform object
	propertiesObj, diags := data.Properties.ToObjectValue(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	propertiesAttrs := propertiesObj.Attributes()
	directionAttr, ok := propertiesAttrs["direction"].(types.String)
	if !ok {
		resp.Diagnostics.AddError("Invalid Type", "direction must be a String")
		return
	}
	direction := directionAttr.ValueString()

	protocolAttr, ok := propertiesAttrs["protocol"].(types.String)
	if !ok {
		resp.Diagnostics.AddError("Invalid Type", "protocol must be a String")
		return
	}
	protocol := protocolAttr.ValueString()

	// Normalize protocol to match API expectations (case-sensitive)
	// API expects: Any, TCP, UDP, ICMP (not ANY, tcp, etc.)
	protocol = normalizeProtocol(protocol)

	port := ""
	if portAttr, ok := propertiesAttrs["port"]; ok && portAttr != nil {
		if portStr, ok := portAttr.(types.String); ok && !portStr.IsNull() {
			port = portStr.ValueString()
		}
	}

	// Adapter: If protocol is Any or ICMP, port must be completely omitted (API requirement)
	// The CLI omits the port field entirely for Any/ICMP protocols (see CLI request JSON)
	// Clear port to empty string so it will be omitted in JSON
	originalPort := port
	if strings.EqualFold(protocol, "Any") || strings.EqualFold(protocol, "ICMP") {
		port = ""
		tflog.Info(ctx, "Clearing port field for Any/ICMP protocol (will be omitted from JSON)", map[string]interface{}{
			"protocol":    protocol,
			"originalPort": originalPort,
			"clearedPort":  port,
		})
	}

	// Extract target
	targetObjAttr, ok := propertiesAttrs["target"].(types.Object)
	if !ok {
		resp.Diagnostics.AddError("Invalid Type", "target must be an Object")
		return
	}
	targetObjValue, diags := targetObjAttr.ToObjectValue(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetAttrs := targetObjValue.Attributes()
	targetKindAttr, ok := targetAttrs["kind"].(types.String)
	if !ok {
		resp.Diagnostics.AddError("Invalid Type", "target.kind must be a String")
		return
	}
	targetKind := targetKindAttr.ValueString()

	// Normalize target kind to match API expectations (case-sensitive)
	// API expects: IP (not Ip, ip, etc.)
	targetKind = normalizeTargetKind(targetKind)

	targetValueAttr, ok := targetAttrs["value"].(types.String)
	if !ok {
		resp.Diagnostics.AddError("Invalid Type", "target.value must be a String")
		return
	}
	targetValue := targetValueAttr.ValueString()

	tflog.Debug(ctx, "Creating security rule request", map[string]interface{}{
		"protocol":     protocol,
		"port":         port,
		"portEmpty":    port == "",
		"direction":    direction,
		"targetKind":   targetKind,
		"targetValue":  targetValue,
	})

	// Build the properties
	// Note: The SDK should handle empty Port strings with omitempty JSON tag
	// If that doesn't work, we'll need to use JSON manipulation
	target := &sdktypes.RuleTarget{
		Kind:  sdktypes.EndpointTypeDto(targetKind),
		Value: targetValue,
	}

	// Build properties - match CLI behavior exactly
	// CLI omits Port field entirely for Any/ICMP protocols (see CLI request JSON)
	// IMPORTANT: For Any/ICMP, port MUST be empty string (cleared above) and Port field MUST be omitted
	// Use JSON manipulation to ensure Port is completely omitted when empty
	var properties sdktypes.SecurityRulePropertiesRequest
	
	// Double-check: If protocol is Any/ICMP, ensure port is empty (safety check)
	if strings.EqualFold(protocol, "Any") || strings.EqualFold(protocol, "ICMP") {
		if port != "" {
			tflog.Warn(ctx, "Port was not cleared for Any/ICMP protocol, forcing it now", map[string]interface{}{
				"protocol": protocol,
				"port":     port,
			})
			port = ""
		}
	}
	
	if port == "" {
		// For Any/ICMP protocols, omit Port field completely (match CLI behavior)
		tflog.Debug(ctx, "Omitting Port field for Any/ICMP protocol (matching CLI behavior)", map[string]interface{}{
			"protocol": protocol,
		})
		propertiesMap := map[string]interface{}{
			"direction": string(sdktypes.RuleDirection(direction)),
			"protocol":  protocol,
			"target": map[string]interface{}{
				"kind":  string(sdktypes.EndpointTypeDto(targetKind)),
				"value": targetValue,
			},
		}
		jsonData, err := json.Marshal(propertiesMap)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error building request",
				fmt.Sprintf("Unable to build security rule properties: %s", err),
			)
			return
		}
		if err := json.Unmarshal(jsonData, &properties); err != nil {
			resp.Diagnostics.AddError(
				"Error building request",
				fmt.Sprintf("Unable to parse security rule properties: %s", err),
			)
			return
		}
		// Ensure Target is set correctly after unmarshal
		properties.Target = target
		// Explicitly clear Port field in struct (double safety)
		properties.Port = ""
	} else {
		// Normal case with Port field
		properties = sdktypes.SecurityRulePropertiesRequest{
			Direction: sdktypes.RuleDirection(direction),
			Protocol:  protocol,
			Port:      port,
			Target:    target,
		}
	}
	
	tflog.Debug(ctx, "Properties built", map[string]interface{}{
		"direction":    properties.Direction,
		"protocol":     properties.Protocol,
		"port":         properties.Port,
		"portEmpty":    properties.Port == "",
		"hasTarget":    properties.Target != nil,
		"originalPort": port,
	})
	
	// Verify properties JSON doesn't include port when it should be omitted
	propsJSON, _ := json.Marshal(properties)
	if port == "" && strings.Contains(string(propsJSON), `"port"`) {
		tflog.Error(ctx, "Port field still present in properties JSON after omitting!", map[string]interface{}{
			"propertiesJSON": string(propsJSON),
		})
	}

	// Build the create request
	// Match CLI behavior exactly - only include tags if they're actually provided
	// CLI doesn't include tags in JSON if not provided (SDK omits empty slices with omitempty)
	metadataRequest := sdktypes.ResourceMetadataRequest{
		Name: data.Name.ValueString(),
	}
	// Only set Tags if we have actual tags (not empty slice)
	// This ensures SDK omits tags field in JSON when empty (matches CLI behavior)
	if len(tags) > 0 {
		metadataRequest.Tags = tags
	}
	
	createRequest := sdktypes.SecurityRuleRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: metadataRequest,
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: properties,
	}

	// Debug: Log the actual JSON that will be sent
	requestJSON, _ := json.MarshalIndent(createRequest, "", "  ")
	requestJSONStr := string(requestJSON)
	
	// Verify Port field is not present in JSON when it should be omitted
	// CRITICAL: If port should be omitted but is still in JSON, rebuild request without it
	if port == "" || strings.EqualFold(protocol, "Any") || strings.EqualFold(protocol, "ICMP") {
		if strings.Contains(requestJSONStr, `"port"`) {
			tflog.Error(ctx, "CRITICAL: Port field found in final JSON when it should be omitted! Rebuilding request...", map[string]interface{}{
				"json":          requestJSONStr,
				"propertiesPort": properties.Port,
				"port":          port,
				"protocol":      protocol,
			})
			// Rebuild the entire request using JSON manipulation to ensure Port is omitted
			requestMap := map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": data.Name.ValueString(),
					"location": map[string]interface{}{
						"value": data.Location.ValueString(),
					},
				},
				"properties": map[string]interface{}{
					"direction": string(sdktypes.RuleDirection(direction)),
					"protocol":  protocol,
					"target": map[string]interface{}{
						"kind":  string(sdktypes.EndpointTypeDto(targetKind)),
						"value": targetValue,
					},
				},
			}
			// Only add tags if they exist
			if len(tags) > 0 {
				metadata := requestMap["metadata"].(map[string]interface{})
				metadata["tags"] = tags
			}
			// Rebuild request from JSON to ensure Port is omitted
			rebuildJSON, err := json.Marshal(requestMap)
			if err == nil {
				if err := json.Unmarshal(rebuildJSON, &createRequest); err == nil {
					requestJSON, _ = json.MarshalIndent(createRequest, "", "  ")
					requestJSONStr = string(requestJSON)
					tflog.Info(ctx, "Rebuilt request without Port field", map[string]interface{}{
						"newJSON": requestJSONStr,
					})
				}
			}
		} else {
			tflog.Debug(ctx, "Port field correctly omitted from final JSON")
		}
	}
	
	tflog.Debug(ctx, "Full request JSON", map[string]interface{}{
		"json": string(requestJSON),
	})
	tflog.Info(ctx, "Creating security rule", map[string]interface{}{
		"name":            data.Name.ValueString(),
		"direction":       direction,
		"protocol":        protocol,
		"port":            port,
		"targetKind":      targetKind,
		"targetValue":     targetValue,
		"portOmitted":     port == "",
		"propertiesPort":  properties.Port,
	})

	// Create the security rule using the SDK
	response, err := r.client.Client.FromNetwork().SecurityGroupRules().Create(ctx, projectID, vpcID, securityGroupID, createRequest, nil)
	if err != nil {
		tflog.Error(ctx, "SDK error creating security rule", map[string]interface{}{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError(
			"Error creating security rule",
			fmt.Sprintf("Unable to create security rule: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create security rule"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		// Log detailed error information including full response
		tflog.Error(ctx, "API error creating security rule", map[string]interface{}{
			"statusCode": response.StatusCode,
			"title":      response.Error.Title,
			"detail":     response.Error.Detail,
			"requestJSON": string(requestJSON),
			"fullError":   fmt.Sprintf("%+v", response.Error),
		})
		// Include request JSON in error message for debugging
		errorMsg = fmt.Sprintf("%s\n\nRequest JSON:\n%s", errorMsg, string(requestJSON))
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Security rule created but no data returned from API",
		)
		return
	}

	// Wait for Security Rule to be active before returning
	// This ensures Terraform doesn't proceed to create dependent resources until Security Rule is ready
	ruleID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromNetwork().SecurityGroupRules().Get(ctx, projectID, vpcID, securityGroupID, ruleID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for Security Rule to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "SecurityRule", ruleID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Security Rule Not Active",
			fmt.Sprintf("Security rule was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	// Re-read the Security Rule to get the latest state including tags
	getResp, err := r.client.Client.FromNetwork().SecurityGroupRules().Get(ctx, projectID, vpcID, securityGroupID, ruleID, nil)
	if err == nil && getResp != nil && getResp.Data != nil {
		rule := getResp.Data
		// Update tags from re-read response
		if len(rule.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(rule.Metadata.Tags))
			for i, tag := range rule.Metadata.Tags {
				tagValues[i] = types.StringValue(tag)
			}
			tagsList, diags := types.ListValueFrom(ctx, types.StringType, tagValues)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = tagsList
			}
		} else {
			data.Tags = types.ListNull(types.StringType)
		}
	}

	tflog.Trace(ctx, "created a Security Rule resource", map[string]interface{}{
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

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	securityGroupID := data.SecurityGroupId.ValueString()
	ruleID := data.Id.ValueString()

	if projectID == "" || vpcID == "" || securityGroupID == "" || ruleID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, Security Group ID, and Rule ID are required to read the security rule",
		)
		return
	}

	// Get security rule details using the SDK
	response, err := r.client.Client.FromNetwork().SecurityGroupRules().Get(ctx, projectID, vpcID, securityGroupID, ruleID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading security rule",
			fmt.Sprintf("Unable to read security rule: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read security rule"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		rule := response.Data

		if rule.Metadata.ID != nil {
			data.Id = types.StringValue(*rule.Metadata.ID)
		}
		if rule.Metadata.URI != nil {
			data.Uri = types.StringValue(*rule.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if rule.Metadata.Name != nil {
			data.Name = types.StringValue(*rule.Metadata.Name)
		}
		if rule.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(rule.Metadata.LocationResponse.Value)
		}

		// Update properties from response
		propertiesMap := map[string]attr.Value{
			"direction": types.StringValue(string(rule.Properties.Direction)),
			"protocol":  types.StringValue(rule.Properties.Protocol),
			"port":      types.StringValue(rule.Properties.Port),
		}

		// Update target
		if rule.Properties.Target != nil {
			targetMap := map[string]attr.Value{
				"kind":  types.StringValue(string(rule.Properties.Target.Kind)),
				"value": types.StringValue(rule.Properties.Target.Value),
			}
			targetObj, diags := types.ObjectValue(map[string]attr.Type{
				"kind":  types.StringType,
				"value": types.StringType,
			}, targetMap)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				propertiesMap["target"] = targetObj
			}
		}

		propertiesObj, diags := types.ObjectValue(map[string]attr.Type{
			"direction": types.StringType,
			"protocol":  types.StringType,
			"port":      types.StringType,
			"target": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"kind":  types.StringType,
					"value": types.StringType,
				},
			},
		}, propertiesMap)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Properties = propertiesObj
		}

		// Update tags from response
		if len(rule.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(rule.Metadata.Tags))
			for i, tag := range rule.Metadata.Tags {
				tagValues[i] = types.StringValue(tag)
			}
			tagsList, diags := types.ListValueFrom(ctx, types.StringType, tagValues)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = tagsList
			}
		} else {
			// API has no tags - set to null
			data.Tags = types.ListNull(types.StringType)
		}

	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SecurityRuleResourceModel
	var state SecurityRuleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use IDs from state (they are immutable)
	projectID := state.ProjectId.ValueString()
	vpcID := state.VpcId.ValueString()
	securityGroupID := state.SecurityGroupId.ValueString()
	ruleID := state.Id.ValueString()

	if projectID == "" || vpcID == "" || securityGroupID == "" || ruleID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, Security Group ID, and Rule ID are required to update the security rule",
		)
		return
	}

	// Get current security rule details
	getResponse, err := r.client.Client.FromNetwork().SecurityGroupRules().Get(ctx, projectID, vpcID, securityGroupID, ruleID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current security rule",
			fmt.Sprintf("Unable to get current security rule: %s", err),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"Security Rule Not Found",
			"Security rule not found or no data returned",
		)
		return
	}

	current := getResponse.Data

	// Check if security rule is in InCreation state
	if current.Status.State != nil && *current.Status.State == "InCreation" {
		resp.Diagnostics.AddError(
			"Cannot Update Security Rule",
			"Cannot update security rule while it is in 'InCreation' state. Please wait until the security rule is fully created.",
		)
		return
	}

	// Check if properties have changed - if so, security rule must be recreated
	// Extract properties from plan using the same method as Create
	var planDirection, planProtocol, planPort, planTargetKind, planTargetValue string
	if !data.Properties.IsNull() && !data.Properties.IsUnknown() {
		propertiesObj, diags := data.Properties.ToObjectValue(ctx)
		if !diags.HasError() {
			propertiesAttrs := propertiesObj.Attributes()
			if dir, ok := propertiesAttrs["direction"]; ok && dir != nil {
				if dirStr, ok := dir.(types.String); ok {
					planDirection = dirStr.ValueString()
				}
			}
			if prot, ok := propertiesAttrs["protocol"]; ok && prot != nil {
				if protStr, ok := prot.(types.String); ok {
					planProtocol = normalizeProtocol(protStr.ValueString())
				}
			}
			if portAttr, ok := propertiesAttrs["port"]; ok && portAttr != nil {
				if portStr, ok := portAttr.(types.String); ok && !portStr.IsNull() {
					planPort = portStr.ValueString()
				}
			}

			// Adapter: If protocol is Any or ICMP, port must not be set (API requirement)
			// Use case-insensitive check for protocol
			if strings.EqualFold(planProtocol, "Any") || strings.EqualFold(planProtocol, "ICMP") {
				planPort = ""
			}

			if targetAttr, ok := propertiesAttrs["target"]; ok && targetAttr != nil {
				if targetObj, ok := targetAttr.(types.Object); ok {
					targetObjValue, targetDiags := targetObj.ToObjectValue(ctx)
					if !targetDiags.HasError() {
						targetAttrs := targetObjValue.Attributes()
						if kind, ok := targetAttrs["kind"]; ok && kind != nil {
							if kindStr, ok := kind.(types.String); ok {
								planTargetKind = normalizeTargetKind(kindStr.ValueString())
							}
						}
						if value, ok := targetAttrs["value"]; ok && value != nil {
							if valueStr, ok := value.(types.String); ok {
								planTargetValue = valueStr.ValueString()
							}
						}
					}
				}
			}
		}
	}

	// Compare with current properties
	var propertiesChanged bool
	if planDirection != "" && string(current.Properties.Direction) != planDirection {
		propertiesChanged = true
	}
	if planProtocol != "" && current.Properties.Protocol != planProtocol {
		propertiesChanged = true
	}
	currentPort := current.Properties.Port
	if planPort != "" && currentPort != planPort {
		propertiesChanged = true
	}
	if planTargetKind != "" && string(current.Properties.Target.Kind) != planTargetKind {
		propertiesChanged = true
	}
	if planTargetValue != "" && current.Properties.Target.Value != planTargetValue {
		propertiesChanged = true
	}

	if propertiesChanged {
		resp.Diagnostics.AddError(
			"Cannot Update Security Rule Properties",
			"Security rule properties (direction, protocol, port, target) cannot be updated. To change properties, delete and recreate the security rule.",
		)
		return
	}

	// Get region value
	regionValue := ""
	if current.Metadata.LocationResponse != nil {
		regionValue = current.Metadata.LocationResponse.Value
	}
	if regionValue == "" {
		// Try to get from VPC
		vpcResp, err := r.client.Client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
		if err == nil && vpcResp != nil && !vpcResp.IsError() && vpcResp.Data != nil {
			if vpcResp.Data.Metadata.LocationResponse != nil {
				regionValue = vpcResp.Data.Metadata.LocationResponse.Value
			}
		}
	}
	if regionValue == "" {
		resp.Diagnostics.AddError(
			"Missing Region",
			"Unable to determine region value for security rule",
		)
		return
	}

	// Extract tags
	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		tags = current.Metadata.Tags
	}

	// Build update request - only name and tags can be updated, properties must remain unchanged
	// Note: Properties must be included in the request but will use current values (they cannot be changed)

	// Build properties conditionally - omit Port field if it's empty (for ANY/ICMP protocols)
	var updateProperties sdktypes.SecurityRulePropertiesRequest
	if current.Properties.Port == "" {
		// Build without Port field by using JSON manipulation
		propertiesMap := map[string]interface{}{
			"direction": current.Properties.Direction,
			"protocol":  current.Properties.Protocol,
			"target": map[string]interface{}{
				"kind":  current.Properties.Target.Kind,
				"value": current.Properties.Target.Value,
			},
		}
		// Marshal to JSON and back to ensure Port field is not present
		jsonData, _ := json.Marshal(propertiesMap)
		json.Unmarshal(jsonData, &updateProperties)
	} else {
		// Normal case with Port field
		updateProperties = sdktypes.SecurityRulePropertiesRequest{
			Direction: current.Properties.Direction,
			Protocol:  current.Properties.Protocol,
			Port:      current.Properties.Port,
			Target:    current.Properties.Target,
		}
	}

	updateRequest := sdktypes.SecurityRuleRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: updateProperties,
	}

	// Log the update request for debugging
	tflog.Debug(ctx, "Updating security rule", map[string]interface{}{
		"rule_id":    ruleID,
		"name":       data.Name.ValueString(),
		"tags":       tags,
		"properties": fmt.Sprintf("Direction=%s, Protocol=%s, Port=%s", current.Properties.Direction, current.Properties.Protocol, current.Properties.Port),
	})

	// Update the security rule using the SDK
	response, err := r.client.Client.FromNetwork().SecurityGroupRules().Update(ctx, projectID, vpcID, securityGroupID, ruleID, updateRequest, nil)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error calling Update API: %v", err))
		resp.Diagnostics.AddError(
			"Error updating security rule",
			fmt.Sprintf("Unable to update security rule: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to update security rule"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		// Include status code if available
		if response.StatusCode > 0 {
			errorMsg = fmt.Sprintf("%s (HTTP %d)", errorMsg, response.StatusCode)
		}
		tflog.Error(ctx, fmt.Sprintf("API returned error: %+v", response.Error))
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	// Verify the update was successful
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Security rule updated but no data returned from API",
		)
		return
	}

	// Update the state with the response data
	// At this point, we know response != nil && response.Data != nil from the check above
	updated := response.Data
	if updated.Metadata.ID != nil {
		data.Id = types.StringValue(*updated.Metadata.ID)
	}
	if updated.Metadata.Name != nil {
		data.Name = types.StringValue(*updated.Metadata.Name)
	}
	if updated.Metadata.LocationResponse != nil {
		data.Location = types.StringValue(updated.Metadata.LocationResponse.Value)
	}

	// Update tags from response
	if len(updated.Metadata.Tags) > 0 {
		tagValues := make([]types.String, len(updated.Metadata.Tags))
		for i, tag := range updated.Metadata.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, tagValues)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Tags = tagsList
		}
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	// Properties remain unchanged - they are immutable
	// Keep the existing state values to ensure Terraform state matches what the user configured

	// Ensure immutable fields are set from state before saving
	data.ProjectId = state.ProjectId
	data.VpcId = state.VpcId
	data.SecurityGroupId = state.SecurityGroupId

	// Update ID from response (should match state)
	// At this point, we know response != nil && response.Data != nil from the check above
	if updated.Metadata.ID != nil {
		data.Id = types.StringValue(*updated.Metadata.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SecurityRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	securityGroupID := data.SecurityGroupId.ValueString()
	ruleID := data.Id.ValueString()

	if projectID == "" || vpcID == "" || securityGroupID == "" || ruleID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, Security Group ID, and Rule ID are required to delete the security rule",
		)
		return
	}

	// Delete the security rule using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromNetwork().SecurityGroupRules().Delete(ctx, projectID, vpcID, securityGroupID, ruleID, nil)
		},
		ExtractSDKError,
		"SecurityRule",
		ruleID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting security rule",
			fmt.Sprintf("Unable to delete security rule: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a Security Rule resource", map[string]interface{}{
		"securityrule_id": ruleID,
	})
}

func (r *SecurityRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
