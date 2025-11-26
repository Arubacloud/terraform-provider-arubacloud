// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

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
	Name            types.String `tfsdk:"name"`
	Location        types.String `tfsdk:"location"`
	ProjectId       types.String `tfsdk:"project_id"`
	VpcId           types.String `tfsdk:"vpc_id"`
	SecurityGroupId types.String `tfsdk:"security_group_id"`
	Properties      types.Object `tfsdk:"properties"`
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
				Computed:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPC this Security Rule belongs to",
				Computed:            true,
			},
			"security_group_id": schema.StringAttribute{
				MarkdownDescription: "ID of the Security Group this rule belongs to",
				Computed:            true,
			},
			"properties": schema.SingleNestedAttribute{
				MarkdownDescription: "Properties of the security rule",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
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
					"target": schema.SingleNestedAttribute{
						MarkdownDescription: "Target of the rule (source or destination)",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"kind": schema.StringAttribute{
								MarkdownDescription: "Type of the target (Ip/SecurityGroup)",
								Computed:            true,
							},
							"value": schema.StringAttribute{
								MarkdownDescription: "Value of the target (CIDR or SecurityGroup URI)",
								Computed:            true,
							},
						},
					},
				},
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
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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
	// Simulate API response for all attributes
	data.Name = types.StringValue("example-securityrule")
	data.Location = types.StringValue("example-location")
	data.ProjectId = types.StringValue("example-project-id")
	data.VpcId = types.StringValue("example-vpc-id")
	data.SecurityGroupId = types.StringValue("example-security-group-id")

	// Build nested target object
	targetType := map[string]attr.Type{
		"kind":  types.StringType,
		"value": types.StringType,
	}
	targetValue := map[string]attr.Value{
		"kind":  types.StringValue("Ip"),
		"value": types.StringValue("192.168.1.1/32"),
	}
	targetObj, _ := types.ObjectValue(targetType, targetValue)

	// Build properties object
	propertiesType := map[string]attr.Type{
		"direction": types.StringType,
		"protocol":  types.StringType,
		"port":      types.StringType,
		"target":    types.ObjectType{AttrTypes: targetType},
	}
	propertiesValue := map[string]attr.Value{
		"direction": types.StringValue("Ingress"),
		"protocol":  types.StringValue("TCP"),
		"port":      types.StringValue("80"),
		"target":    targetObj,
	}
	propertiesObj, _ := types.ObjectValue(propertiesType, propertiesValue)
	data.Properties = propertiesObj

	tflog.Trace(ctx, "read a Security Rule data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
