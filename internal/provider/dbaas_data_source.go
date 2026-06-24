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

var _ datasource.DataSource = &DBaaSDataSource{}

func NewDBaaSDataSource() datasource.DataSource {
	return &DBaaSDataSource{}
}

type DBaaSDataSource struct {
	client *ArubaCloudClient
}

type DBaaSDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Zone          types.String `tfsdk:"zone"`
	Tags          types.List   `tfsdk:"tags"`
	ProjectID     types.String `tfsdk:"project_id"`
	EngineID      types.String `tfsdk:"engine_id"`
	Flavor        types.String `tfsdk:"flavor"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	// Storage fields (flattened)
	StorageSizeGB             types.Int64 `tfsdk:"storage_size_gb"`
	AutoscalingEnabled        types.Bool  `tfsdk:"autoscaling_enabled"`
	AutoscalingAvailableSpace types.Int64 `tfsdk:"autoscaling_available_space"`
	AutoscalingStepSize       types.Int64 `tfsdk:"autoscaling_step_size"`
	// Network fields (flattened)
	VpcUriRef           types.String `tfsdk:"vpc_uri_ref"`
	SubnetUriRef        types.String `tfsdk:"subnet_uri_ref"`
	SecurityGroupUriRef types.String `tfsdk:"security_group_uri_ref"`
	ElasticIpUriRef     types.String `tfsdk:"elastic_ip_uri_ref"`
}

func (d *DBaaSDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dbaas"
}

func (d *DBaaSDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves read-only information about an existing ArubaCloud DBaaS cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the DBaaS cluster to look up.",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the DBaaS cluster.",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
				Computed:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Availability zone within the region where the DBaaS cluster is deployed.",
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
			"engine_id": schema.StringAttribute{
				MarkdownDescription: "Database engine type and version identifier (e.g., `mysql-8.0`, `postgresql-15`).",
				Computed:            true,
			},
			"flavor": schema.StringAttribute{
				MarkdownDescription: "Compute flavour for the DBaaS cluster nodes (e.g., `DBO2A4`).",
				Computed:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle. Accepted values: `Hour`, `Month`, `Year`.",
				Computed:            true,
			},
			"storage_size_gb": schema.Int64Attribute{
				MarkdownDescription: "Computed by the API. Storage size in GB allocated to the DBaaS instance.",
				Computed:            true,
			},
			"autoscaling_enabled": schema.BoolAttribute{
				MarkdownDescription: "Computed by the API. Whether storage autoscaling is enabled.",
				Computed:            true,
			},
			"autoscaling_available_space": schema.Int64Attribute{
				MarkdownDescription: "Computed by the API. Minimum available space threshold in GB that triggers autoscaling.",
				Computed:            true,
			},
			"autoscaling_step_size": schema.Int64Attribute{
				MarkdownDescription: "Computed by the API. Amount of storage (in GB) added on each autoscaling event.",
				Computed:            true,
			},
			"vpc_uri_ref": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. URI reference to the VPC the DBaaS cluster is attached to.",
				Computed:            true,
			},
			"subnet_uri_ref": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. URI reference to the Subnet the DBaaS cluster is attached to.",
				Computed:            true,
			},
			"security_group_uri_ref": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. URI reference to the Security Group attached to the DBaaS cluster.",
				Computed:            true,
			},
			"elastic_ip_uri_ref": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. URI reference to the Elastic IP attached to the DBaaS cluster, if any.",
				Computed:            true,
			},
		},
	}
}

func (d *DBaaSDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DBaaSDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DBaaSDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.Id.ValueString()
	if projectID == "" || dbaasID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID and DBaaS ID are required to read the DBaaS instance")
		return
	}

	dbaas, err := d.client.Client.FromDatabase().DBaaS().Get(ctx,
		aruba.URI("/projects/"+projectID+"/providers/Aruba.Database/dbaas/"+dbaasID))
	if provErr := CheckResponseErr("read", "DBaaS", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(dbaas.ID())
	data.Uri = strVal(dbaas.URI())
	data.Name = types.StringValue(dbaas.Name())
	data.ProjectID = types.StringValue(projectID)
	data.Tags = TagsToListPreserveNull(dbaas.Tags(), data.Tags)

	if dbaas.Region() != "" {
		data.Location = types.StringValue(string(dbaas.Region()))
	} else {
		data.Location = types.StringNull()
	}
	// Zone is not reliably returned by the API.
	data.Zone = types.StringNull()

	if e := string(dbaas.Engine()); e != "" {
		data.EngineID = types.StringValue(e)
	} else {
		data.EngineID = types.StringNull()
	}
	if f := string(dbaas.Flavor()); f != "" {
		data.Flavor = types.StringValue(f)
	} else {
		data.Flavor = types.StringNull()
	}
	if bp := string(dbaas.BillingPeriod()); bp != "" {
		data.BillingPeriod = types.StringValue(bp)
	} else {
		data.BillingPeriod = types.StringNull()
	}

	if s := dbaas.SizeGB(); s > 0 {
		data.StorageSizeGB = types.Int64Value(int64(s))
	} else {
		data.StorageSizeGB = types.Int64Null()
	}
	if dbaas.AutoscalingEnabled() {
		data.AutoscalingEnabled = types.BoolValue(true)
		data.AutoscalingAvailableSpace = types.Int64Value(int64(dbaas.AutoscalingAvailableSpaceGB()))
		data.AutoscalingStepSize = types.Int64Value(int64(dbaas.AutoscalingStepSizeGB()))
	} else {
		data.AutoscalingEnabled = types.BoolNull()
		data.AutoscalingAvailableSpace = types.Int64Null()
		data.AutoscalingStepSize = types.Int64Null()
	}

	// Network URIs (may be empty if not in API response).
	data.VpcUriRef = strVal(dbaas.VPC())
	data.SubnetUriRef = strVal(dbaas.Subnet())
	data.SecurityGroupUriRef = strVal(dbaas.SecurityGroup())
	data.ElasticIpUriRef = strVal(dbaas.ElasticIP())

	tflog.Trace(ctx, "read a DBaaS data source", map[string]interface{}{"dbaas_id": dbaasID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
