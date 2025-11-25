package provider

import (
	"context"
	"fmt"
	"net/http"

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
	client *http.Client
}

type BlockStorageDataSourceModel struct {
	Id         types.String                `tfsdk:"id"`
	Name       types.String                `tfsdk:"name"`
	ProjectId  types.String                `tfsdk:"project_id"`
	Properties BlockStoragePropertiesModel `tfsdk:"properties"`
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
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Block Storage name",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this Block Storage belongs to",
				Required:            true,
			},
			"properties": schema.SingleNestedAttribute{
				MarkdownDescription: "Properties of the Block Storage",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"size_gb": schema.Int64Attribute{
						MarkdownDescription: "Size of the block storage in GB",
						Required:            true,
					},
					"billing_period": schema.StringAttribute{
						MarkdownDescription: "Billing period of the block storage (only 'Hour' allowed)",
						Required:            true,
					},
					"zone": schema.StringAttribute{
						MarkdownDescription: "Zone of the block storage",
						Required:            true,
					},
					"type": schema.StringAttribute{
						MarkdownDescription: "Type of block storage (Standard, Performance)",
						Required:            true,
					},
					"snapshot_id": schema.StringAttribute{
						MarkdownDescription: "Snapshot ID for the block storage",
						Optional:            true,
					},
					"bootable": schema.BoolAttribute{
						MarkdownDescription: "Whether the block storage is bootable",
						Optional:            true,
					},
					"image": schema.StringAttribute{
						MarkdownDescription: "Image for the block storage",
						Optional:            true,
					},
				},
			},
		},
	}
}

func (d *BlockStorageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*http.Client)
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
	data.Name = types.StringValue("example-blockstorage")
	tflog.Trace(ctx, "read a Block Storage data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
