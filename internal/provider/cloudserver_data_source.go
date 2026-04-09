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

type CloudServerDataSourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	ProjectID types.String `tfsdk:"project_id"`
	Zone      types.String `tfsdk:"zone"`
	Tags      types.List   `tfsdk:"tags"`
	// Network fields (flattened from Network object)
	VpcUriRef            types.String `tfsdk:"vpc_uri_ref"`
	ElasticIpUriRef      types.String `tfsdk:"elastic_ip_uri_ref"`
	SubnetUriRefs        types.List   `tfsdk:"subnet_uri_refs"`
	SecurityGroupUriRefs types.List   `tfsdk:"securitygroup_uri_refs"`
	// Settings fields (flattened from Settings object)
	FlavorName    types.String `tfsdk:"flavor_name"`
	KeyPairUriRef types.String `tfsdk:"key_pair_uri_ref"`
	UserData      types.String `tfsdk:"user_data"`
	// Storage fields (flattened from Storage object)
	BootVolumeUriRef types.String `tfsdk:"boot_volume_uri_ref"`
}

type CloudServerDataSource struct {
	client *ArubaCloudClient
}

var _ datasource.DataSource = &CloudServerDataSource{}

func NewCloudServerDataSource() datasource.DataSource {
	return &CloudServerDataSource{}
}

func (d *CloudServerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudserver"
}

func (d *CloudServerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CloudServer data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "CloudServer identifier",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "CloudServer URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "CloudServer name",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "CloudServer location",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID",
				Required:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the Cloud Server",
				Computed:            true,
			},
			"vpc_uri_ref": schema.StringAttribute{
				MarkdownDescription: "VPC URI reference",
				Computed:            true,
			},
			"elastic_ip_uri_ref": schema.StringAttribute{
				MarkdownDescription: "Elastic IP URI reference",
				Computed:            true,
			},
			"subnet_uri_refs": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of subnet URI references",
				Computed:            true,
			},
			"securitygroup_uri_refs": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of security group URI references",
				Computed:            true,
			},
			"flavor_name": schema.StringAttribute{
				MarkdownDescription: "Flavor name (e.g., CSO4A8 for 4 CPU, 8GB RAM)",
				Computed:            true,
			},
			"key_pair_uri_ref": schema.StringAttribute{
				MarkdownDescription: "Key Pair URI reference",
				Computed:            true,
			},
			"user_data": schema.StringAttribute{
				MarkdownDescription: "Cloud-Init user data",
				Computed:            true,
			},
			"boot_volume_uri_ref": schema.StringAttribute{
				MarkdownDescription: "Boot volume URI reference",
				Computed:            true,
			},
		},
	}
}

func (d *CloudServerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CloudServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CloudServerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	serverID := data.Id.ValueString()
	if projectID == "" || serverID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID and CloudServer ID are required to read the cloud server")
		return
	}

	response, err := d.client.Client.FromCompute().CloudServers().Get(ctx, projectID, serverID, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading cloud server", fmt.Sprintf("Unable to read cloud server: %s", err))
		return
	}
	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.Diagnostics.AddError("CloudServer not found", fmt.Sprintf("No cloud server found with ID %q in project %q", serverID, projectID))
			return
		}
		resp.Diagnostics.AddError("API Error", FormatAPIError(ctx, response.Error, "Failed to read cloud server", map[string]interface{}{"project_id": projectID, "server_id": serverID}))
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError("No data returned", "CloudServer Get returned no data")
		return
	}

	server := response.Data
	if server.Metadata.ID != nil {
		data.Id = types.StringValue(*server.Metadata.ID)
	}
	if server.Metadata.URI != nil {
		data.Uri = types.StringValue(*server.Metadata.URI)
	} else {
		data.Uri = types.StringNull()
	}
	if server.Metadata.Name != nil {
		data.Name = types.StringValue(*server.Metadata.Name)
	}
	if server.Metadata.LocationResponse != nil {
		data.Location = types.StringValue(server.Metadata.LocationResponse.Value)
	} else {
		data.Location = types.StringNull()
	}
	data.ProjectID = types.StringValue(projectID)
	// Zone is not returned by the API
	data.Zone = types.StringNull()
	// Flavor name is returned by the API
	data.FlavorName = types.StringValue(server.Properties.Flavor.Name)
	// Network/settings/storage URI refs are not returned by the API
	data.VpcUriRef = types.StringNull()
	data.ElasticIpUriRef = types.StringNull()
	data.SubnetUriRefs = types.ListValueMust(types.StringType, []attr.Value{})
	data.SecurityGroupUriRefs = types.ListValueMust(types.StringType, []attr.Value{})
	data.KeyPairUriRef = types.StringNull()
	data.UserData = types.StringNull()
	data.BootVolumeUriRef = types.StringNull()

	if len(server.Metadata.Tags) > 0 {
		tagValues := make([]attr.Value, len(server.Metadata.Tags))
		for i, tag := range server.Metadata.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		data.Tags = types.ListValueMust(types.StringType, tagValues)
	} else {
		data.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	tflog.Trace(ctx, "read a CloudServer data source", map[string]interface{}{"server_id": serverID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
