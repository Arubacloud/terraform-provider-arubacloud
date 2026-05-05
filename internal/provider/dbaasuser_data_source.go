package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type DBaaSUserDataSourceModel struct {
	Id        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	DBaaSID   types.String `tfsdk:"dbaas_id"`
	Username  types.String `tfsdk:"username"`
	Password  types.String `tfsdk:"password"`
}

type DBaaSUserDataSource struct {
	client *ArubaCloudClient
}

var _ datasource.DataSource = &DBaaSUserDataSource{}

func NewDBaaSUserDataSource() datasource.DataSource {
	return &DBaaSUserDataSource{}
}

func (d *DBaaSUserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dbaasuser"
}

func (d *DBaaSUserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves read-only information about an existing ArubaCloud DBaaS user.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "Username of the DBaaS user to look up.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource (same as the username).",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"dbaas_id": schema.StringAttribute{
				MarkdownDescription: "ID of the parent DBaaS cluster this user belongs to.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for the DBaaS user. Write-only — this value is sent to the API but is not returned in subsequent read responses.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (d *DBaaSUserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ArubaCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *ArubaCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *DBaaSUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DBaaSUserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()
	username := data.Username.ValueString()
	if projectID == "" || dbaasID == "" || username == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID, DBaaS ID, and Username are required to read the DBaaS user")
		return
	}

	response, err := d.client.Client.FromDatabase().Users().Get(ctx, projectID, dbaasID, username, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading DBaaS user", NewTransportError("read", "Dbaasuser", err).Error())
		return
	}
	if apiErr := CheckResponse("read", "Dbaasuser", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError("No data returned", "DBaaS User Get returned no data")
		return
	}

	user := response.Data
	data.Id = types.StringValue(user.Username)
	data.Username = types.StringValue(user.Username)
	data.ProjectID = types.StringValue(projectID)
	data.DBaaSID = types.StringValue(dbaasID)
	// Password is not returned by the API
	data.Password = types.StringNull()

	tflog.Trace(ctx, "read a DBaaS User data source", map[string]interface{}{"username": username})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
