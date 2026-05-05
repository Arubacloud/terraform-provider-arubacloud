package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &VPCPeeringRouteDataSource{}

func NewVPCPeeringRouteDataSource() datasource.DataSource {
	return &VPCPeeringRouteDataSource{}
}

type VPCPeeringRouteDataSource struct {
	client *ArubaCloudClient
}

type VPCPeeringRouteDataSourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ProjectId    types.String `tfsdk:"project_id"`
	VpcId        types.String `tfsdk:"vpc_id"`
	VpcPeeringId types.String `tfsdk:"vpc_peering_id"`
}

func (d *VPCPeeringRouteDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpcpeeringroute"
}

func (d *VPCPeeringRouteDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves read-only information about an existing `arubacloud_vpcpeeringroute`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the VPC peering route to look up (same as the route name).",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the VPC peering route.",
				Computed:            true,
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
		},
	}
}

func (d *VPCPeeringRouteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VPCPeeringRouteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VPCPeeringRouteDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	peeringID := data.VpcPeeringId.ValueString()
	routeID := data.Id.ValueString()
	if projectID == "" || vpcID == "" || peeringID == "" || routeID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID, VPC ID, VPC Peering ID, and Route ID are required to read the VPC peering route")
		return
	}

	response, err := d.client.Client.FromNetwork().VPCPeeringRoutes().Get(ctx, projectID, vpcID, peeringID, routeID, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading VPC peering route", NewTransportError("read", "Vpcpeeringroute", err).Error())
		return
	}
	if apiErr := CheckResponse("read", "Vpcpeeringroute", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError("No data returned", "VPC Peering Route Get returned no data")
		return
	}

	route := response.Data
	// VPC Peering Route uses name as ID
	data.Id = types.StringValue(route.Metadata.Name)
	data.Name = types.StringValue(route.Metadata.Name)
	data.ProjectId = types.StringValue(projectID)
	data.VpcId = types.StringValue(vpcID)
	data.VpcPeeringId = types.StringValue(peeringID)

	tflog.Trace(ctx, "read a VPC Peering Route data source", map[string]interface{}{"route_id": routeID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
