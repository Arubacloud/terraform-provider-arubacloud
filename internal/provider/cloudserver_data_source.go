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
		MarkdownDescription: "Retrieves read-only information about an existing ArubaCloud CloudServer virtual machine.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource.",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources (e.g., as a `*_uri_ref` attribute).",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the CloudServer.",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier for the resource (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Availability zone within the region (e.g., `ITBG-1`). See [available zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Computed:            true,
			},
			"vpc_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI of the VPC attached to this CloudServer.",
				Computed:            true,
			},
			"elastic_ip_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI of the Elastic IP associated with this CloudServer, if any.",
				Computed:            true,
			},
			"subnet_uri_refs": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of subnet URIs attached to this CloudServer.",
				Computed:            true,
			},
			"securitygroup_uri_refs": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of security group URIs applied to this CloudServer.",
				Computed:            true,
			},
			"flavor_name": schema.StringAttribute{
				MarkdownDescription: "Compute flavour name (e.g., `CSO4A8` for 4 vCPU / 8 GB RAM). See [available flavours](https://api.arubacloud.com/docs/metadata/#cloudserver-flavors).",
				Computed:            true,
			},
			"key_pair_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI of the SSH key pair injected at boot.",
				Computed:            true,
			},
			"user_data": schema.StringAttribute{
				MarkdownDescription: "Cloud-Init configuration passed to the instance at first boot.",
				Computed:            true,
			},
			"boot_volume_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI of the bootable block storage volume.",
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
		resp.Diagnostics.AddError("Error reading cloud server", NewTransportError("read", "Cloudserver", err).Error())
		return
	}
	if apiErr := CheckResponse("read", "Cloudserver", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
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

	data.Tags = TagsToList(server.Metadata.Tags)

	tflog.Trace(ctx, "read a CloudServer data source", map[string]interface{}{"server_id": serverID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
