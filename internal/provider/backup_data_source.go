package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &BackupDataSource{}

func NewBackupDataSource() datasource.DataSource {
	return &BackupDataSource{}
}

type BackupDataSource struct {
	client *ArubaCloudClient
}

type BackupDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	ProjectID     types.String `tfsdk:"project_id"`
	Type          types.String `tfsdk:"type"`
	VolumeID      types.String `tfsdk:"volume_id"`
	RetentionDays types.Int64  `tfsdk:"retention_days"`
	BillingPeriod types.String `tfsdk:"billing_period"`
}

func (d *BackupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup"
}

func (d *BackupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves read-only information about an existing ArubaCloud Block Storage Backup.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the backup to look up.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the backup.",
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
			"type": schema.StringAttribute{
				MarkdownDescription: "Backup type. Accepted values: `Full`, `Incremental`.",
				Computed:            true,
			},
			"volume_id": schema.StringAttribute{
				MarkdownDescription: "ID of the block storage volume this backup was taken from.",
				Computed:            true,
			},
			"retention_days": schema.Int64Attribute{
				MarkdownDescription: "Number of days to retain the backup before automatic deletion. Optional — if omitted, the backup is retained indefinitely.",
				Computed:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle. Accepted values: `Hour`, `Month`, `Year`.",
				Computed:            true,
			},
		},
	}
}

func (d *BackupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BackupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BackupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	backupID := data.Id.ValueString()
	if projectID == "" || backupID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Backup ID (id) are required to read the backup",
		)
		return
	}

	response, err := d.client.Client.FromStorage().Backups().Get(ctx, projectID, backupID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading backup",
			NewTransportError("read", "Backup", err).Error(),
		)
		return
	}
	if apiErr := CheckResponse("read", "Backup", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError(
			"No data returned",
			"Backup Get returned no data",
		)
		return
	}

	backup := response.Data

	if backup.Metadata.ID != nil {
		data.Id = types.StringValue(*backup.Metadata.ID)
	}
	if backup.Metadata.Name != nil {
		data.Name = types.StringValue(*backup.Metadata.Name)
	}
	if backup.Metadata.LocationResponse != nil {
		data.Location = types.StringValue(backup.Metadata.LocationResponse.Value)
	} else {
		data.Location = types.StringNull()
	}
	data.ProjectID = types.StringValue(projectID)
	if backup.Properties.Type != "" {
		data.Type = types.StringValue(string(backup.Properties.Type))
	} else {
		data.Type = types.StringNull()
	}
	if backup.Properties.Origin.URI != "" {
		parts := strings.Split(backup.Properties.Origin.URI, "/")
		data.VolumeID = types.StringValue(parts[len(parts)-1])
	} else {
		data.VolumeID = types.StringNull()
	}
	if backup.Properties.RetentionDays != nil {
		data.RetentionDays = types.Int64Value(int64(*backup.Properties.RetentionDays))
	} else {
		data.RetentionDays = types.Int64Null()
	}
	if backup.Properties.BillingPeriod != nil {
		data.BillingPeriod = types.StringValue(*backup.Properties.BillingPeriod)
	} else {
		data.BillingPeriod = types.StringNull()
	}

	data.Tags = TagsToListPreserveNull(backup.Metadata.Tags, data.Tags)

	tflog.Trace(ctx, "read a Backup data source", map[string]interface{}{"backup_id": backupID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
