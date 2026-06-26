package provider

import (
	"context"
	"fmt"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &VPCPeeringDataSource{}

func NewVPCPeeringDataSource() datasource.DataSource {
	return &VPCPeeringDataSource{}
}

type VPCPeeringDataSource struct {
	client *ArubaCloudClient
}

type VPCPeeringDataSourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	ProjectId types.String `tfsdk:"project_id"`
	VpcId     types.String `tfsdk:"vpc_id"`
}

func (d *VPCPeeringDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpcpeering"
}

func (d *VPCPeeringDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves read-only information about an existing `arubacloud_vpcpeering` connection.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the VPC peering to look up.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the VPC peering.",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the local VPC this peering belongs to.",
				Required:            true,
			},
		},
	}
}

func (d *VPCPeeringDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VPCPeeringDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VPCPeeringDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	peeringID := data.Id.ValueString()
	if projectID == "" || vpcID == "" || peeringID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID, VPC ID, and VPC Peering ID are required to read the VPC peering")
		return
	}

	peering, err := d.client.Client.FromNetwork().VPCPeerings().Get(ctx,
		aruba.VPCPeeringRef(projectID, vpcID, peeringID))
	if provErr := CheckResponseErr("read", "VPCPeering", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(peering.ID())
	data.Name = types.StringValue(peering.Name())
	data.ProjectId = types.StringValue(projectID)
	data.VpcId = types.StringValue(vpcID)

	tflog.Trace(ctx, "read a VPC Peering data source", map[string]interface{}{"peering_id": peeringID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
