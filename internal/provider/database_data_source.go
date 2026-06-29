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

var _ datasource.DataSource = &DatabaseDataSource{}

func NewDatabaseDataSource() datasource.DataSource {
	return &DatabaseDataSource{}
}

type DatabaseDataSource struct {
	client *ArubaCloudClient
}

type DatabaseDataSourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	ProjectID types.String `tfsdk:"project_id"`
	DBaaSID   types.String `tfsdk:"dbaas_id"`
	Name      types.String `tfsdk:"name"`
}

func (d *DatabaseDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (d *DatabaseDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves read-only information about an existing ArubaCloud database within a DBaaS cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the database to look up (same as the database name).",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI. Use this value in `*_uri_ref` attributes of other resources.",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"dbaas_id": schema.StringAttribute{
				MarkdownDescription: "ID of the parent DBaaS cluster this database belongs to.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the database.",
				Computed:            true,
			},
		},
	}
}

func (d *DatabaseDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DatabaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DatabaseDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()
	databaseName := data.Id.ValueString()
	if projectID == "" || dbaasID == "" || databaseName == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID, DBaaS ID, and Database ID are required to read the database")
		return
	}

	db, err := d.client.Client.FromDatabase().Databases().Get(ctx,
		aruba.URI("/projects/"+projectID+"/providers/Aruba.Database/dbaas/"+dbaasID+"/databases/"+databaseName))
	if provErr := CheckResponseErr("read", "Database", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(db.Name())
	data.Uri = strVal(db.URI())
	data.Name = types.StringValue(db.Name())
	data.ProjectID = types.StringValue(projectID)
	data.DBaaSID = types.StringValue(dbaasID)

	tflog.Trace(ctx, "read a Database data source", map[string]interface{}{"database_name": databaseName})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
