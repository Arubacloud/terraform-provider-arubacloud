package provider

import (
	"context"
	"fmt"
	"strings"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &SnapshotDataSource{}

func NewSnapshotDataSource() datasource.DataSource {
	return &SnapshotDataSource{}
}

type SnapshotDataSource struct {
	client *ArubaCloudClient
}

type SnapshotDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	ProjectId     types.String `tfsdk:"project_id"`
	Location      types.String `tfsdk:"location"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	VolumeId      types.String `tfsdk:"volume_id"`
}

func (d *SnapshotDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshot"
}

func (d *SnapshotDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves read-only information about an existing ArubaCloud Snapshot.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the snapshot to look up.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the snapshot.",
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
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle. Accepted values: `Hour`, `Month`, `Year`.",
				Computed:            true,
			},
			"volume_id": schema.StringAttribute{
				MarkdownDescription: "ID of the block storage volume this snapshot was taken from.",
				Computed:            true,
			},
		},
	}
}

func (d *SnapshotDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SnapshotDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SnapshotDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	snapshotID := data.Id.ValueString()
	if projectID == "" || snapshotID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID and Snapshot ID are required to read the snapshot")
		return
	}

	snap, err := d.client.Client.FromStorage().Snapshots().Get(ctx,
		aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/snapshots/"+snapshotID))
	if provErr := CheckResponseErr("read", "Snapshot", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(snap.ID())
	data.Name = types.StringValue(snap.Name())
	data.ProjectId = types.StringValue(projectID)
	if snap.Region() != "" {
		data.Location = types.StringValue(string(snap.Region()))
	} else {
		data.Location = types.StringNull()
	}
	// billing_period is not returned by the API.
	data.BillingPeriod = types.StringNull()
	// Extract volume ID from volume URI.
	if volURI := snap.VolumeURI(); volURI != "" {
		parts := strings.Split(volURI, "/")
		data.VolumeId = types.StringValue(parts[len(parts)-1])
	} else {
		data.VolumeId = types.StringNull()
	}

	tflog.Trace(ctx, "read a Snapshot data source", map[string]interface{}{"snapshot_id": snapshotID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
