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

var _ datasource.DataSource = &DatabaseBackupDataSource{}

func NewDatabaseBackupDataSource() datasource.DataSource {
	return &DatabaseBackupDataSource{}
}

type DatabaseBackupDataSource struct {
	client *ArubaCloudClient
}

type DatabaseBackupDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
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
		MarkdownDescription: "Database Backup data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Database Backup identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Database Backup name",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Database Backup location",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the Database Backup resource",
				Computed:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone for the Database Backup",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this database backup belongs to",
				Required:            true,
			},
			"dbaas_id": schema.StringAttribute{
				MarkdownDescription: "DBaaS ID this backup belongs to",
				Computed:            true,
			},
			"database": schema.StringAttribute{
				MarkdownDescription: "Database to backup (ID or name)",
				Computed:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period",
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

	response, err := d.client.Client.FromDatabase().Backups().Get(ctx, projectID, backupID, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading database backup", NewTransportError("read", "Databasebackup", err).Error())
		return
	}
	if apiErr := CheckResponse("read", "Databasebackup", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError("No data returned", "Database Backup Get returned no data")
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

	if backup.Properties.Zone != "" {
		data.Zone = types.StringValue(backup.Properties.Zone)
	} else {
		data.Zone = types.StringNull()
	}
	if backup.Properties.BillingPlan.BillingPeriod != "" {
		data.BillingPeriod = types.StringValue(backup.Properties.BillingPlan.BillingPeriod)
	} else {
		data.BillingPeriod = types.StringNull()
	}
	// DBaaSID and Database are not directly available in the response
	data.DBaaSID = types.StringNull()
	data.Database = types.StringNull()

	if len(backup.Metadata.Tags) > 0 {
		tagValues := make([]attr.Value, len(backup.Metadata.Tags))
		for i, tag := range backup.Metadata.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		data.Tags = types.ListValueMust(types.StringType, tagValues)
	} else {
		data.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	tflog.Trace(ctx, "read a Database Backup data source", map[string]interface{}{"backup_id": backupID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
