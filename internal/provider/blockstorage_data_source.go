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
		MarkdownDescription: "Retrieves read-only information about an existing ArubaCloud Block Storage volume.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the block storage volume to look up.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the block storage volume.",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
				Computed:            true,
			},
			"size_gb": schema.Int64Attribute{
				MarkdownDescription: "Size of the block storage volume in GiB. Must be a positive integer.",
				Computed:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle. Accepted values: `Hour`, `Month`, `Year`.",
				Computed:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Availability zone within the region. If omitted the volume is regional (accessible across all zones).",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Storage type. Accepted values: `Standard`, `Performance`.",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Computed:            true,
			},
			"snapshot_id": schema.StringAttribute{
				MarkdownDescription: "ID of the snapshot this volume was created from, if any.",
				Computed:            true,
			},
			"bootable": schema.BoolAttribute{
				MarkdownDescription: "Whether this volume can be used as a boot volume for an `arubacloud_cloudserver`. Must be `true` when `image` is set.",
				Computed:            true,
			},
			"image": schema.StringAttribute{
				MarkdownDescription: "Image ID to use when creating a bootable volume. Required when `bootable` is `true`. See the [available images](https://api.arubacloud.com/docs/metadata/#cloud-server-bootvolume).",
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
			fmt.Sprintf("Expected *ArubaCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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

	projectID := data.ProjectId.ValueString()
	volumeID := data.Id.ValueString()
	if projectID == "" || volumeID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID and Block Storage ID are required to read the block storage")
		return
	}

	vol, err := d.client.Client.FromStorage().Volumes().Get(ctx,
		aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/blockStorages/"+volumeID))
	if provErr := CheckResponseErr("read", "BlockStorage", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(vol.ID())
	data.Name = types.StringValue(vol.Name())
	data.ProjectId = types.StringValue(projectID)
	if vol.Region() != "" {
		data.Location = types.StringValue(string(vol.Region()))
	} else {
		data.Location = types.StringNull()
	}
	data.SizeGB = types.Int64Value(int64(vol.SizeGB()))
	data.Type = types.StringValue(string(vol.Type()))
	if z := string(vol.Zone()); z != "" {
		data.Zone = types.StringValue(z)
	} else {
		data.Zone = types.StringNull()
	}
	if vol.IsBootable() {
		data.Bootable = types.BoolValue(true)
	} else {
		data.Bootable = types.BoolNull()
	}
	if img := vol.Image(); img != "" {
		data.Image = types.StringValue(img)
	} else {
		data.Image = types.StringNull()
	}
	// billing_period and snapshot_id are not returned by the API.
	data.BillingPeriod = types.StringNull()
	data.SnapshotId = types.StringNull()

	data.Tags = TagsToListPreserveNull(vol.Tags(), data.Tags)

	tflog.Trace(ctx, "read a Block Storage data source", map[string]interface{}{"volume_id": volumeID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
