package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &DatabaseGrantDataSource{}

func NewDatabaseGrantDataSource() datasource.DataSource {
	return &DatabaseGrantDataSource{}
}

type DatabaseGrantDataSource struct {
	client *ArubaCloudClient
}

type DatabaseGrantDataSourceModel struct {
	Id        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	DBaaSID   types.String `tfsdk:"dbaas_id"`
	Database  types.String `tfsdk:"database"`
	UserID    types.String `tfsdk:"user_id"`
	Role      types.String `tfsdk:"role"`
}

func (d *DatabaseGrantDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databasegrant"
}

func (d *DatabaseGrantDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves read-only information about an existing ArubaCloud database grant.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the grant (composite key: `project_id/dbaas_id/database/user_id`).",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"dbaas_id": schema.StringAttribute{
				MarkdownDescription: "ID of the parent DBaaS cluster this grant belongs to.",
				Required:            true,
			},
			"database": schema.StringAttribute{
				MarkdownDescription: "ID of the database this grant applies to.",
				Required:            true,
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "Name or ID of the DBaaS user whose grant is being looked up.",
				Required:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Privilege level granted to the user on the database.",
				Computed:            true,
			},
		},
	}
}

func (d *DatabaseGrantDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DatabaseGrantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DatabaseGrantDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()
	database := data.Database.ValueString()
	userID := data.UserID.ValueString()
	if projectID == "" || dbaasID == "" || database == "" || userID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID, DBaaS ID, Database, and User ID are required to read the database grant")
		return
	}

	response, err := d.client.Client.FromDatabase().Grants().Get(ctx, projectID, dbaasID, database, userID, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading database grant", NewTransportError("read", "Databasegrant", err).Error())
		return
	}
	if apiErr := CheckResponse("read", "Databasegrant", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError("No data returned", "Database Grant Get returned no data")
		return
	}

	data.Id = types.StringValue(fmt.Sprintf("%s/%s/%s/%s", projectID, dbaasID, database, userID))
	data.ProjectID = types.StringValue(projectID)
	data.DBaaSID = types.StringValue(dbaasID)
	data.Database = types.StringValue(database)
	data.UserID = types.StringValue(userID)
	data.Role = types.StringValue(response.Data.Role.Name)

	tflog.Trace(ctx, "read a Database Grant data source", map[string]interface{}{"database": database, "user_id": userID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
