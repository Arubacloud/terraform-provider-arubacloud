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
				Computed:            true,
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
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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

	// Populate all fields with example data
	data.Name = types.StringValue("example-vpc")
	data.Location = types.StringValue("ITBG-Bergamo")
	data.ProjectId = types.StringValue("68398923fb2cb026400d4d31")
	data.Tags = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("network"),
		types.StringValue("production"),
		types.StringValue("vpc-main"),
	})

	tflog.Trace(ctx, "read a VPC data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
