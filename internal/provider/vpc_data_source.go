package provider

import (
	"context"
	"fmt"

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
		MarkdownDescription: "Retrieves read-only metadata about an existing `arubacloud_vpc`. Use this data source to reference a VPC's URI in subnet or server configurations when the VPC is managed in a separate Terraform root module.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the VPC to look up.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the VPC.",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier for the resource (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
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
		resp.Diagnostics.AddError("Error reading VPC", NewTransportError("read", "Vpc", err).Error())
		return
	}
	if apiErr := CheckResponse("read", "Vpc", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
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

	data.Tags = TagsToList(vpc.Metadata.Tags)

	tflog.Trace(ctx, "read a VPC data source", map[string]interface{}{"vpc_id": vpcID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
