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

var _ datasource.DataSource = &SecurityGroupDataSource{}

func NewSecurityGroupDataSource() datasource.DataSource {
	return &SecurityGroupDataSource{}
}

type SecurityGroupDataSource struct {
	client *ArubaCloudClient
}

type SecurityGroupDataSourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	Tags      types.List   `tfsdk:"tags"`
	ProjectId types.String `tfsdk:"project_id"`
	VpcId     types.String `tfsdk:"vpc_id"`
}

func (d *SecurityGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_securitygroup"
}

func (d *SecurityGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Security Group data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Security Group identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Security Group name",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Security Group location",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the Security Group",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this Security Group belongs to",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPC this Security Group belongs to",
				Required:            true,
			},
		},
	}
}

func (d *SecurityGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SecurityGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	sgID := data.Id.ValueString()
	if projectID == "" || vpcID == "" || sgID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID, VPC ID, and Security Group ID are required to read the security group")
		return
	}

	response, err := d.client.Client.FromNetwork().SecurityGroups().Get(ctx, projectID, vpcID, sgID, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading security group", fmt.Sprintf("Unable to read security group: %s", err))
		return
	}
	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.Diagnostics.AddError("Security Group not found", fmt.Sprintf("No security group found with ID %q in VPC %q", sgID, vpcID))
			return
		}
		resp.Diagnostics.AddError("API Error", FormatAPIError(ctx, response.Error, "Failed to read security group", map[string]interface{}{"project_id": projectID, "vpc_id": vpcID, "security_group_id": sgID}))
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError("No data returned", "Security Group Get returned no data")
		return
	}

	sg := response.Data
	if sg.Metadata.ID != nil {
		data.Id = types.StringValue(*sg.Metadata.ID)
	}
	if sg.Metadata.Name != nil {
		data.Name = types.StringValue(*sg.Metadata.Name)
	}
	if sg.Metadata.LocationResponse != nil {
		data.Location = types.StringValue(sg.Metadata.LocationResponse.Value)
	} else {
		data.Location = types.StringNull()
	}
	data.ProjectId = types.StringValue(projectID)
	data.VpcId = types.StringValue(vpcID)

	if len(sg.Metadata.Tags) > 0 {
		tagValues := make([]attr.Value, len(sg.Metadata.Tags))
		for i, tag := range sg.Metadata.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		data.Tags = types.ListValueMust(types.StringType, tagValues)
	} else {
		data.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	tflog.Trace(ctx, "read a Security Group data source", map[string]interface{}{"security_group_id": sgID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
