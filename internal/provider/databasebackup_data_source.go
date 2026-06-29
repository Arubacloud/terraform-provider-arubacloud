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

var _ datasource.DataSource = &DatabaseBackupDataSource{}

func NewDatabaseBackupDataSource() datasource.DataSource {
	return &DatabaseBackupDataSource{}
}

type DatabaseBackupDataSource struct {
	client *ArubaCloudClient
}

type DatabaseBackupDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	Zone          types.String `tfsdk:"zone"`
	ProjectID     types.String `tfsdk:"project_id"`
	DBaaSID       types.String `tfsdk:"dbaas_id"`
	Database      types.String `tfsdk:"database"`
	BillingPeriod types.String `tfsdk:"billing_period"`
}

func (d *DatabaseBackupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databasebackup"
}

func (d *DatabaseBackupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves read-only information about an existing ArubaCloud database backup.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the database backup to look up.",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI. Use this value in `*_uri_ref` attributes of other resources.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the database backup.",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier where the backup is stored.",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Computed:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Availability zone within the region where the backup is stored.",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"dbaas_id": schema.StringAttribute{
				MarkdownDescription: "ID of the DBaaS cluster this backup belongs to.",
				Computed:            true,
			},
			"database": schema.StringAttribute{
				MarkdownDescription: "Name or ID of the logical database this backup was taken from.",
				Computed:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle. Accepted values: `Hour`, `Month`, `Year`.",
				Computed:            true,
			},
		},
	}
}

func (d *DatabaseBackupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DatabaseBackupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DatabaseBackupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	backupID := data.Id.ValueString()
	if projectID == "" || backupID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID and Database Backup ID are required to read the database backup")
		return
	}

	backup, err := d.client.Client.FromDatabase().Backups().Get(ctx,
		aruba.URI("/projects/"+projectID+"/providers/Aruba.Database/backups/"+backupID))
	if provErr := CheckResponseErr("read", "DBaaSBackup", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(backup.ID())
	data.Uri = strVal(backup.URI())
	data.Name = types.StringValue(backup.Name())
	data.ProjectID = types.StringValue(projectID)
	data.Tags = TagsToListPreserveNull(backup.Tags(), data.Tags)

	if backup.Region() != "" {
		data.Location = types.StringValue(string(backup.Region()))
	} else {
		data.Location = types.StringNull()
	}
	if z := string(backup.Zone()); z != "" {
		data.Zone = types.StringValue(z)
	} else {
		data.Zone = types.StringNull()
	}
	if bp := string(backup.BillingPeriod()); bp != "" {
		data.BillingPeriod = types.StringValue(bp)
	} else {
		data.BillingPeriod = types.StringNull()
	}
	// DBaaSID and Database are not directly available in the response.
	data.DBaaSID = types.StringNull()
	data.Database = types.StringNull()

	tflog.Trace(ctx, "read a Database Backup data source", map[string]interface{}{"backup_id": backupID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
