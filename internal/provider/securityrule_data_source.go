package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &SecurityRuleDataSource{}

func NewSecurityRuleDataSource() datasource.DataSource {
	return &SecurityRuleDataSource{}
}

type SecurityRuleDataSource struct {
	client *ArubaCloudClient
}

type SecurityRuleDataSourceModel struct {
	Id              types.String `tfsdk:"id"`
	Uri             types.String `tfsdk:"uri"`
	Name            types.String `tfsdk:"name"`
	Location        types.String `tfsdk:"location"`
	ProjectId       types.String `tfsdk:"project_id"`
	VpcId           types.String `tfsdk:"vpc_id"`
	SecurityGroupId types.String `tfsdk:"security_group_id"`
	Tags            types.List   `tfsdk:"tags"`
	// Properties fields (flattened)
	Direction types.String `tfsdk:"direction"`
	Protocol  types.String `tfsdk:"protocol"`
	Port      types.String `tfsdk:"port"`
	// Target fields (flattened)
	TargetKind  types.String `tfsdk:"target_kind"`
	TargetValue types.String `tfsdk:"target_value"`
}

func (d *SecurityRuleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_securityrule"
}

func (d *SecurityRuleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Security Rule data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Security Rule identifier",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Security Rule URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Security Rule name",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Security Rule location",
				Computed:            true,
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
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the Security Rule",
				Computed:            true,
			},
			"direction": schema.StringAttribute{
				MarkdownDescription: "Direction of the rule (Ingress/Egress)",
				Computed:            true,
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Protocol (ANY, TCP, UDP, ICMP)",
				Computed:            true,
			},
			"port": schema.StringAttribute{
				MarkdownDescription: "Port or port range (for TCP/UDP)",
				Computed:            true,
			},
			"target_kind": schema.StringAttribute{
				MarkdownDescription: "Type of the target (IP/SecurityGroup)",
				Computed:            true,
			},
			"target_value": schema.StringAttribute{
				MarkdownDescription: "Value of the target (CIDR or SecurityGroup URI)",
				Computed:            true,
			},
		},
	}
}

func (d *SecurityRuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ArubaCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ArubaCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *SecurityRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityRuleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	securityGroupID := data.SecurityGroupId.ValueString()
	ruleID := data.Id.ValueString()
	if projectID == "" || vpcID == "" || securityGroupID == "" || ruleID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID, VPC ID, Security Group ID, and Rule ID are required to read the security rule")
		return
	}

	response, err := d.client.Client.FromNetwork().SecurityGroupRules().Get(ctx, projectID, vpcID, securityGroupID, ruleID, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading security rule", fmt.Sprintf("Unable to read security rule: %s", err))
		return
	}
	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.Diagnostics.AddError("Security Rule not found", fmt.Sprintf("No security rule found with ID %q in security group %q", ruleID, securityGroupID))
			return
		}
		resp.Diagnostics.AddError("API Error", FormatAPIError(ctx, response.Error, "Failed to read security rule", map[string]interface{}{"project_id": projectID, "vpc_id": vpcID, "security_group_id": securityGroupID, "rule_id": ruleID}))
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError("No data returned", "Security Rule Get returned no data")
		return
	}

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
	} else {
		data.Location = types.StringNull()
	}
	data.ProjectId = types.StringValue(projectID)
	data.VpcId = types.StringValue(vpcID)
	data.SecurityGroupId = types.StringValue(securityGroupID)

	data.Direction = types.StringValue(string(rule.Properties.Direction))
	data.Protocol = types.StringValue(strings.ToUpper(rule.Properties.Protocol))
	data.Port = types.StringValue(rule.Properties.Port)

	if rule.Properties.Target != nil {
		data.TargetKind = types.StringValue(string(rule.Properties.Target.Kind))
		data.TargetValue = types.StringValue(rule.Properties.Target.Value)
	} else {
		data.TargetKind = types.StringNull()
		data.TargetValue = types.StringNull()
	}

	if len(rule.Metadata.Tags) > 0 {
		tagValues := make([]attr.Value, len(rule.Metadata.Tags))
		for i, tag := range rule.Metadata.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		data.Tags = types.ListValueMust(types.StringType, tagValues)
	} else {
		data.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	tflog.Trace(ctx, "read a Security Rule data source", map[string]interface{}{"rule_id": ruleID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
