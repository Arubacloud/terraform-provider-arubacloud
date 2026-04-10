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

type KeypairDataSourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	ProjectID types.String `tfsdk:"project_id"`
	Value     types.String `tfsdk:"value"`
	Tags      types.List   `tfsdk:"tags"`
}

type KeypairDataSource struct {
	client *ArubaCloudClient
}

var _ datasource.DataSource = &KeypairDataSource{}

func NewKeypairDataSource() datasource.DataSource {
	return &KeypairDataSource{}
}

func (d *KeypairDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keypair"
}

func (d *KeypairDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Keypair data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Keypair identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Keypair name",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Keypair location",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this keypair belongs to",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Keypair value (public key)",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the keypair",
				Computed:            true,
			},
		},
	}
}

func (d *KeypairDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KeypairDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data KeypairDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	keypairID := data.Id.ValueString()
	if projectID == "" || keypairID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Keypair ID (id) are required to read the keypair",
		)
		return
	}

	response, err := d.client.Client.FromCompute().KeyPairs().Get(ctx, projectID, keypairID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading keypair",
			NewTransportError("read", "Keypair", err).Error(),
		)
		return
	}
	if apiErr := CheckResponse("read", "Keypair", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError(
			"No data returned",
			"Keypair Get returned no data",
		)
		return
	}

	keypair := response.Data

	if keypair.Metadata.ID != nil {
		data.Id = types.StringValue(*keypair.Metadata.ID)
	}
	if keypair.Metadata.Name != nil {
		data.Name = types.StringValue(*keypair.Metadata.Name)
	}
	if keypair.Metadata.LocationResponse != nil {
		data.Location = types.StringValue(keypair.Metadata.LocationResponse.Value)
	} else {
		data.Location = types.StringNull()
	}
	data.ProjectID = types.StringValue(projectID)
	if keypair.Properties.Value != "" {
		data.Value = types.StringValue(keypair.Properties.Value)
	} else {
		data.Value = types.StringNull()
	}

	if len(keypair.Metadata.Tags) > 0 {
		tagValues := make([]attr.Value, len(keypair.Metadata.Tags))
		for i, tag := range keypair.Metadata.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		data.Tags = types.ListValueMust(types.StringType, tagValues)
	} else {
		data.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	tflog.Trace(ctx, "read a Keypair data source", map[string]interface{}{"keypair_id": keypairID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
