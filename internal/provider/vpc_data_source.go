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

var _ datasource.DataSource = &VPCDataSource{}

func NewVPCDataSource() datasource.DataSource {
	return &VPCDataSource{}
}

type VPCDataSource struct {
	client *ArubaCloudClient
}

type VPCDataSourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	ProjectId types.String `tfsdk:"project_id"`
	Tags      types.List   `tfsdk:"tags"`
}

func (d *VPCDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (d *VPCDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VPC data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VPC identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "VPC name",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "VPC location",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this VPC belongs to",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the VPC",
				Computed:            true,
			},
		},
	}
}

func (d *VPCDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VPCDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VPCDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.Id.ValueString()
	if projectID == "" || vpcID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID and VPC ID are required to read the VPC")
		return
	}

	response, err := d.client.Client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading VPC", fmt.Sprintf("Unable to read VPC: %s", err))
		return
	}
	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.Diagnostics.AddError("VPC not found", fmt.Sprintf("No VPC found with ID %q in project %q", vpcID, projectID))
			return
		}
		resp.Diagnostics.AddError("API Error", FormatAPIError(ctx, response.Error, "Failed to read VPC", map[string]interface{}{"project_id": projectID, "vpc_id": vpcID}))
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError("No data returned", "VPC Get returned no data")
		return
	}

	vpc := response.Data
	if vpc.Metadata.ID != nil {
		data.Id = types.StringValue(*vpc.Metadata.ID)
	}
	if vpc.Metadata.Name != nil {
		data.Name = types.StringValue(*vpc.Metadata.Name)
	}
	if vpc.Metadata.LocationResponse != nil {
		data.Location = types.StringValue(vpc.Metadata.LocationResponse.Value)
	} else {
		data.Location = types.StringNull()
	}
	data.ProjectId = types.StringValue(projectID)

	if len(vpc.Metadata.Tags) > 0 {
		tagValues := make([]attr.Value, len(vpc.Metadata.Tags))
		for i, tag := range vpc.Metadata.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		data.Tags = types.ListValueMust(types.StringType, tagValues)
	} else {
		data.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	tflog.Trace(ctx, "read a VPC data source", map[string]interface{}{"vpc_id": vpcID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
