package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type KMSDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Endpoint    types.String `tfsdk:"endpoint"`
}

type KMSDataSource struct {
	client *ArubaCloudClient
}

var _ datasource.DataSource = &KMSDataSource{}

func NewKMSDataSource() datasource.DataSource {
	return &KMSDataSource{}
}

func (d *KMSDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kms"
}

func (d *KMSDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "KMS data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "KMS identifier",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this KMS belongs to",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the KMS resource",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the KMS resource",
				Computed:            true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "KMS endpoint URL",
				Computed:            true,
			},
		},
	}
}

func (d *KMSDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KMSDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data KMSDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	kmsID := data.Id.ValueString()
	if projectID == "" || kmsID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID and KMS ID are required to read the KMS")
		return
	}

	response, err := d.client.Client.FromSecurity().KMS().Get(ctx, projectID, kmsID, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading KMS", fmt.Sprintf("Unable to read KMS: %s", err))
		return
	}
	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.Diagnostics.AddError("KMS not found", fmt.Sprintf("No KMS found with ID %q in project %q", kmsID, projectID))
			return
		}
		resp.Diagnostics.AddError("API Error", FormatAPIError(ctx, response.Error, "Failed to read KMS", map[string]interface{}{"project_id": projectID, "kms_id": kmsID}))
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError("No data returned", "KMS Get returned no data")
		return
	}

	kms := response.Data
	if kms.Metadata.ID != nil {
		data.Id = types.StringValue(*kms.Metadata.ID)
	}
	if kms.Metadata.Name != nil {
		data.Name = types.StringValue(*kms.Metadata.Name)
	}
	data.ProjectID = types.StringValue(projectID)
	// description and endpoint are not returned by the API
	data.Description = types.StringNull()
	data.Endpoint = types.StringNull()

	tflog.Trace(ctx, "read a KMS data source", map[string]interface{}{"kms_id": kmsID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
