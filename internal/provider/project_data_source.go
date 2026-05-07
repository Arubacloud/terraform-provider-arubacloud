package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &ProjectDataSource{}

func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

type ProjectDataSource struct {
	client *ArubaCloudClient
}

type ProjectDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Tags        types.List   `tfsdk:"tags"`
}

func (d *ProjectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *ProjectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves read-only information about an existing ArubaCloud Project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the Project.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional human-readable description of the project.",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Computed:            true,
			},
		},
	}
}

func (d *ProjectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ArubaCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ArubaCloudClient, got: %T. Please report this issue to the provider developers. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.Id.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID (id) is required to read the project")
		return
	}

	response, err := d.client.Client.FromProject().Get(ctx, projectID, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading project", NewTransportError("read", "Project", err).Error())
		return
	}
	if apiErr := CheckResponse("read", "Project", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError("No data returned", "Project Get returned no data")
		return
	}

	project := response.Data
	if project.Metadata.ID != nil {
		data.Id = types.StringValue(*project.Metadata.ID)
	}
	if project.Metadata.Name != nil {
		data.Name = types.StringValue(*project.Metadata.Name)
	}
	if project.Properties.Description != nil {
		data.Description = types.StringValue(*project.Properties.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Tags = TagsToListPreserveNull(project.Metadata.Tags, data.Tags)

	tflog.Trace(ctx, "read project data source", map[string]interface{}{"project_id": projectID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
