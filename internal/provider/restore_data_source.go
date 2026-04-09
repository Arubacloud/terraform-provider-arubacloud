package provider

import (
	"context"
	"fmt"

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
		MarkdownDescription: "Restore data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Restore identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Restore name",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Restore location",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the restore resource",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this restore belongs to",
				Required:            true,
			},
			"backup_id": schema.StringAttribute{
				MarkdownDescription: "ID of the backup this restore belongs to",
				Required:            true,
			},
			"volume_id": schema.StringAttribute{
				MarkdownDescription: "Volume ID to restore",
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

	response, err := d.client.Client.FromStorage().Restores().Get(ctx, projectID, backupID, restoreID, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading restore", fmt.Sprintf("Unable to read restore: %s", err))
		return
	}
	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.Diagnostics.AddError("Restore not found", fmt.Sprintf("No restore found with ID %q in backup %q", restoreID, backupID))
			return
		}
		resp.Diagnostics.AddError("API Error", FormatAPIError(ctx, response.Error, "Failed to read restore", map[string]interface{}{"project_id": projectID, "backup_id": backupID, "restore_id": restoreID}))
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError("No data returned", "Restore Get returned no data")
		return
	}

	restore := response.Data
	if restore.Metadata.ID != nil {
		data.Id = types.StringValue(*restore.Metadata.ID)
	}
	if restore.Metadata.Name != nil {
		data.Name = types.StringValue(*restore.Metadata.Name)
	}
	if restore.Metadata.LocationResponse != nil {
		data.Location = types.StringValue(restore.Metadata.LocationResponse.Value)
	} else {
		data.Location = types.StringNull()
	}
	data.ProjectId = types.StringValue(projectID)
	data.BackupId = types.StringValue(backupID)
	// VolumeId is not directly available in the restore response
	data.VolumeId = types.StringNull()

	if len(restore.Metadata.Tags) > 0 {
		tagValues := make([]types.String, len(restore.Metadata.Tags))
		for i, tag := range restore.Metadata.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		data.Tags = tagValues
	} else {
		data.Tags = []types.String{}
	}

	tflog.Trace(ctx, "read a Restore data source", map[string]interface{}{"restore_id": restoreID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
