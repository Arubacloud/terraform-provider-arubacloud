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
				Required:            true,
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

	response, err := d.client.Client.FromStorage().Volumes().Get(ctx, projectID, volumeID, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading block storage", fmt.Sprintf("Unable to read block storage: %s", err))
		return
	}
	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.Diagnostics.AddError("Block Storage not found", fmt.Sprintf("No block storage found with ID %q in project %q", volumeID, projectID))
			return
		}
		resp.Diagnostics.AddError("API Error", FormatAPIError(ctx, response.Error, "Failed to read block storage", map[string]interface{}{"project_id": projectID, "volume_id": volumeID}))
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError("No data returned", "Block Storage Get returned no data")
		return
	}

	volume := response.Data
	if volume.Metadata.ID != nil {
		data.Id = types.StringValue(*volume.Metadata.ID)
	}
	if volume.Metadata.Name != nil {
		data.Name = types.StringValue(*volume.Metadata.Name)
	}
	if volume.Metadata.LocationResponse != nil {
		data.Location = types.StringValue(volume.Metadata.LocationResponse.Value)
	} else {
		data.Location = types.StringNull()
	}
	data.ProjectId = types.StringValue(projectID)
	data.SizeGB = types.Int64Value(int64(volume.Properties.SizeGB))
	data.Type = types.StringValue(string(volume.Properties.Type))
	if volume.Properties.Zone != "" {
		data.Zone = types.StringValue(volume.Properties.Zone)
	} else {
		data.Zone = types.StringNull()
	}
	if volume.Properties.Bootable != nil {
		data.Bootable = types.BoolValue(*volume.Properties.Bootable)
	} else {
		data.Bootable = types.BoolNull()
	}
	if volume.Properties.Image != nil {
		data.Image = types.StringValue(*volume.Properties.Image)
	} else {
		data.Image = types.StringNull()
	}
	// billing_period and snapshot_id are not returned by the API
	data.BillingPeriod = types.StringNull()
	data.SnapshotId = types.StringNull()

	if len(volume.Metadata.Tags) > 0 {
		tagValues := make([]attr.Value, len(volume.Metadata.Tags))
		for i, tag := range volume.Metadata.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		data.Tags = types.ListValueMust(types.StringType, tagValues)
	} else {
		data.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	tflog.Trace(ctx, "read a Block Storage data source", map[string]interface{}{"volume_id": volumeID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
