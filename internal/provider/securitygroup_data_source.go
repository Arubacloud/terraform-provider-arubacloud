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

var _ datasource.DataSource = &SecurityGroupDataSource{}

func NewSecurityGroupDataSource() datasource.DataSource {
	return &SecurityGroupDataSource{}
}

type SecurityGroupDataSource struct {
	client *ArubaCloudClient
}

type SecurityGroupDataSourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	Tags      types.List   `tfsdk:"tags"`
	ProjectId types.String `tfsdk:"project_id"`
	VpcId     types.String `tfsdk:"vpc_id"`
}

func (d *SecurityGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_securitygroup"
}

func (d *SecurityGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves read-only information about an existing `arubacloud_securitygroup`. Use this data source to look up a security group's URI for use in CloudServer network configurations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the security group to look up.",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI. Use this value in `*_uri_ref` attributes of other resources.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the security group.",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier for the resource (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPC this security group is scoped to.",
				Required:            true,
			},
		},
	}
}

func (d *SecurityGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SecurityGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	sgID := data.Id.ValueString()
	if projectID == "" || vpcID == "" || sgID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID, VPC ID, and Security Group ID are required to read the security group")
		return
	}

	sg, err := d.client.Client.FromNetwork().SecurityGroups().Get(ctx,
		aruba.SecurityGroupRef(projectID, vpcID, sgID))
	if provErr := CheckResponseErr("read", "SecurityGroup", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(sg.ID())
	data.Uri = strVal(sg.URI())
	data.Name = types.StringValue(sg.Name())
	data.ProjectId = types.StringValue(projectID)
	data.VpcId = types.StringValue(vpcID)
	raw := sg.Raw()
	if raw != nil && raw.Metadata.LocationResponse != nil {
		data.Location = types.StringValue(string(raw.Metadata.LocationResponse.Value))
	} else {
		data.Location = types.StringNull()
	}
	data.Tags = TagsToListPreserveNull(sg.Tags(), data.Tags)

	tflog.Trace(ctx, "read a Security Group data source", map[string]interface{}{"security_group_id": sgID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
