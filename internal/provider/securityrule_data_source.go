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
	// Populate all fields with example data
	data.Uri = types.StringValue("/v2/securityrules/sr-68398923fb2cb026400d4d31")
	data.Name = types.StringValue("example-securityrule")
	data.Location = types.StringValue("ITBG-Bergamo")
	data.ProjectId = types.StringValue("68398923fb2cb026400d4d31")
	data.VpcId = types.StringValue("vpc-68398923fb2cb026400d4d32")
	data.SecurityGroupId = types.StringValue("sg-68398923fb2cb026400d4d33")
	data.Tags = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("security"),
		types.StringValue("firewall"),
	})
	// Properties fields
	data.Direction = types.StringValue("Ingress")
	data.Protocol = types.StringValue("TCP")
	data.Port = types.StringValue("80")
	// Target fields
	data.TargetKind = types.StringValue("IP")
	data.TargetValue = types.StringValue("192.168.1.0/24")

	tflog.Trace(ctx, "read a Security Rule data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
