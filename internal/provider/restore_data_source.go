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

var _ datasource.DataSource = &RestoreDataSource{}

func NewRestoreDataSource() datasource.DataSource {
	return &RestoreDataSource{}
}

type RestoreDataSource struct {
	client *ArubaCloudClient
}

type RestoreDataSourceModel struct {
	Id        types.String   `tfsdk:"id"`
	Uri       types.String   `tfsdk:"uri"`
	Name      types.String   `tfsdk:"name"`
	Location  types.String   `tfsdk:"location"`
	Tags      []types.String `tfsdk:"tags"`
	ProjectId types.String   `tfsdk:"project_id"`
	BackupId  types.String   `tfsdk:"backup_id"`
	VolumeId  types.String   `tfsdk:"volume_id"`
}

func (d *RestoreDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_restore"
}

func (d *RestoreDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves read-only information about an existing ArubaCloud Block Storage Restore operation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the restore operation to look up.",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI. Use this value in `*_uri_ref` attributes of other resources.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the restore operation.",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"backup_id": schema.StringAttribute{
				MarkdownDescription: "ID of the backup this restore operation belongs to.",
				Required:            true,
			},
			"volume_id": schema.StringAttribute{
				MarkdownDescription: "ID of the target block storage volume that was restored.",
				Computed:            true,
			},
		},
	}
}

func (d *RestoreDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RestoreDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RestoreDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	backupID := data.BackupId.ValueString()
	restoreID := data.Id.ValueString()
	if projectID == "" || backupID == "" || restoreID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID, Backup ID, and Restore ID are required to read the restore")
		return
	}

	restore, err := d.client.Client.FromStorage().Restores().Get(ctx,
		aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/backups/"+backupID+"/restores/"+restoreID))
	if provErr := CheckResponseErr("read", "Restore", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(restore.ID())
	data.Uri = strVal(restore.URI())
	data.Name = types.StringValue(restore.Name())
	data.ProjectId = types.StringValue(projectID)
	data.BackupId = types.StringValue(backupID)
	if restore.Region() != "" {
		data.Location = types.StringValue(string(restore.Region()))
	} else {
		data.Location = types.StringNull()
	}
	// VolumeId is not directly available in the restore response.
	data.VolumeId = types.StringNull()

	tags := restore.Tags()
	if len(tags) > 0 {
		tagValues := make([]types.String, len(tags))
		for i, t := range tags {
			tagValues[i] = types.StringValue(t)
		}
		data.Tags = tagValues
	} else {
		data.Tags = []types.String{}
	}

	tflog.Trace(ctx, "read a Restore data source", map[string]interface{}{"restore_id": restoreID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
