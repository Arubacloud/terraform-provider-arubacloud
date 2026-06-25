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
		MarkdownDescription: "Retrieves information about an existing ArubaCloud KMS (Key Management Service) instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the KMS instance to look up.",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the KMS instance.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional human-readable description of the KMS instance.",
				Computed:            true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Endpoint URL used to interact with the KMS service.",
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

	ref := aruba.URI("/projects/" + projectID + "/providers/Aruba.Security/kms/" + kmsID)
	kms, err := d.client.Client.FromSecurity().KMS().Get(ctx, ref)
	if provErr := CheckResponseErr("read", "KMS", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(kms.ID())
	data.Name = types.StringValue(kms.Name())
	data.ProjectID = types.StringValue(projectID)
	// description and endpoint are not returned by the API
	data.Description = types.StringNull()
	data.Endpoint = types.StringNull()

	tflog.Trace(ctx, "read a KMS data source", map[string]interface{}{"kms_id": kmsID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
