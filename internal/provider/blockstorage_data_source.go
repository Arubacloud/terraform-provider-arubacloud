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

var _ datasource.DataSource = &BlockStorageDataSource{}

func NewBlockStorageDataSource() datasource.DataSource {
	return &BlockStorageDataSource{}
}

type BlockStorageDataSource struct {
	client *ArubaCloudClient
}

type BlockStorageDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	ProjectId     types.String `tfsdk:"project_id"`
	Location      types.String `tfsdk:"location"`
	SizeGB        types.Int64  `tfsdk:"size_gb"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	Zone          types.String `tfsdk:"zone"`
	Type          types.String `tfsdk:"type"`
	Tags          types.List   `tfsdk:"tags"`
	SnapshotId    types.String `tfsdk:"snapshot_id"`
	Bootable      types.Bool   `tfsdk:"bootable"`
	Image         types.String `tfsdk:"image"`
}

type BlockStoragePropertiesModel struct {
	SizeGB        types.Int64  `tfsdk:"size_gb"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	Zone          types.String `tfsdk:"zone"`
	Type          types.String `tfsdk:"type"`
	SnapshotId    types.String `tfsdk:"snapshot_id"`
	Bootable      types.Bool   `tfsdk:"bootable"`
	Image         types.String `tfsdk:"image"`
}

func (d *BlockStorageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blockstorage"
}

func (d *BlockStorageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Block Storage data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Block Storage identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Block Storage name",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this Block Storage belongs to",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Location of the block storage",
				Computed:            true,
			},
			"size_gb": schema.Int64Attribute{
				MarkdownDescription: "Size of the block storage in GB",
				Computed:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period of the block storage",
				Computed:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone of the block storage",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of block storage (Standard, Performance)",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the block storage",
				Computed:            true,
			},
			"snapshot_id": schema.StringAttribute{
				MarkdownDescription: "Snapshot ID for the block storage",
				Computed:            true,
			},
			"bootable": schema.BoolAttribute{
				MarkdownDescription: "Whether the block storage is bootable",
				Computed:            true,
			},
			"image": schema.StringAttribute{
				MarkdownDescription: "Image for the block storage",
				Computed:            true,
			},
		},
	}
}

func (d *BlockStorageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BlockStorageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BlockStorageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate all fields with example data
	data.Name = types.StringValue("example-blockstorage")
	data.ProjectId = types.StringValue("68398923fb2cb026400d4d31")
	data.Location = types.StringValue("ITBG-Bergamo")
	data.SizeGB = types.Int64Value(50)
	data.BillingPeriod = types.StringValue("Hour")
	data.Zone = types.StringValue("ITBG-1")
	data.Type = types.StringValue("Standard")
	data.Tags = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("storage"),
		types.StringValue("data"),
		types.StringValue("test"),
	})
	data.SnapshotId = types.StringNull()
	data.Bootable = types.BoolValue(false)
	data.Image = types.StringNull()

	tflog.Trace(ctx, "read a Block Storage data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
