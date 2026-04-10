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
		MarkdownDescription: "DBaaS data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "DBaaS identifier",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "DBaaS URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "DBaaS name",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "DBaaS location",
				Computed:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the DBaaS resource",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this DBaaS belongs to",
				Required:            true,
			},
			"engine_id": schema.StringAttribute{
				MarkdownDescription: "Database engine ID (e.g., mysql-8.0, mssql-2022-web)",
				Computed:            true,
			},
			"flavor": schema.StringAttribute{
				MarkdownDescription: "DBaaS flavor name (e.g., DBO2A4)",
				Computed:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period (Hour, Month, Year)",
				Computed:            true,
			},
			"storage_size_gb": schema.Int64Attribute{
				MarkdownDescription: "Storage size in GB",
				Computed:            true,
			},
			"autoscaling_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable autoscaling",
				Computed:            true,
			},
			"autoscaling_available_space": schema.Int64Attribute{
				MarkdownDescription: "Minimum available space threshold in GB for autoscaling",
				Computed:            true,
			},
			"autoscaling_step_size": schema.Int64Attribute{
				MarkdownDescription: "Step size for autoscaling (in GB)",
				Computed:            true,
			},
			"vpc_uri_ref": schema.StringAttribute{
				MarkdownDescription: "VPC URI reference",
				Computed:            true,
			},
			"subnet_uri_ref": schema.StringAttribute{
				MarkdownDescription: "Subnet URI reference",
				Computed:            true,
			},
			"security_group_uri_ref": schema.StringAttribute{
				MarkdownDescription: "Security Group URI reference",
				Computed:            true,
			},
			"elastic_ip_uri_ref": schema.StringAttribute{
				MarkdownDescription: "Elastic IP URI reference",
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

	response, err := d.client.Client.FromDatabase().DBaaS().Get(ctx, projectID, dbaasID, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading DBaaS instance", NewTransportError("read", "Dbaas", err).Error())
		return
	}
	if apiErr := CheckResponse("read", "Dbaas", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError("No data returned", "DBaaS Get returned no data")
		return
	}

	dbaas := response.Data
	if dbaas.Metadata.ID != nil {
		data.Id = types.StringValue(*dbaas.Metadata.ID)
	}
	if dbaas.Metadata.URI != nil {
		data.Uri = types.StringValue(*dbaas.Metadata.URI)
	} else {
		data.Uri = types.StringNull()
	}
	if dbaas.Metadata.Name != nil {
		data.Name = types.StringValue(*dbaas.Metadata.Name)
	}
	if dbaas.Metadata.LocationResponse != nil {
		data.Location = types.StringValue(dbaas.Metadata.LocationResponse.Value)
	} else {
		data.Location = types.StringNull()
	}
	data.ProjectID = types.StringValue(projectID)
	// Zone is not returned by the API
	data.Zone = types.StringNull()

	if dbaas.Properties.Engine != nil && dbaas.Properties.Engine.ID != nil {
		data.EngineID = types.StringValue(*dbaas.Properties.Engine.ID)
	} else {
		data.EngineID = types.StringNull()
	}
	if dbaas.Properties.Flavor != nil && dbaas.Properties.Flavor.Name != nil {
		data.Flavor = types.StringValue(*dbaas.Properties.Flavor.Name)
	} else {
		data.Flavor = types.StringNull()
	}
	if dbaas.Properties.BillingPlan != nil && dbaas.Properties.BillingPlan.BillingPeriod != nil {
		data.BillingPeriod = types.StringValue(*dbaas.Properties.BillingPlan.BillingPeriod)
	} else {
		data.BillingPeriod = types.StringNull()
	}

	if dbaas.Properties.Storage != nil && dbaas.Properties.Storage.SizeGB != nil {
		data.StorageSizeGB = types.Int64Value(int64(*dbaas.Properties.Storage.SizeGB))
	} else {
		data.StorageSizeGB = types.Int64Null()
	}
	// Autoscaling fields are not reliably returned by the API
	data.AutoscalingEnabled = types.BoolNull()
	data.AutoscalingAvailableSpace = types.Int64Null()
	data.AutoscalingStepSize = types.Int64Null()

	// Network URIs are not returned by the API
	data.VpcUriRef = types.StringNull()
	data.SubnetUriRef = types.StringNull()
	data.SecurityGroupUriRef = types.StringNull()
	data.ElasticIpUriRef = types.StringNull()

	if len(dbaas.Metadata.Tags) > 0 {
		tagValues := make([]attr.Value, len(dbaas.Metadata.Tags))
		for i, tag := range dbaas.Metadata.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		data.Tags = types.ListValueMust(types.StringType, tagValues)
	} else {
		data.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	tflog.Trace(ctx, "read a DBaaS data source", map[string]interface{}{"dbaas_id": dbaasID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
