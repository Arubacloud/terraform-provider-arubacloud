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

var _ datasource.DataSource = &ElasticIPDataSource{}

func NewElasticIPDataSource() datasource.DataSource {
	return &ElasticIPDataSource{}
}

type ElasticIPDataSource struct {
	client *ArubaCloudClient
}

type ElasticIPDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	ProjectId     types.String `tfsdk:"project_id"`
	Address       types.String `tfsdk:"address"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	Tags          types.List   `tfsdk:"tags"`
}

func (d *ElasticIPDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticip"
}

func (d *ElasticIPDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves read-only information about an existing ArubaCloud Elastic IP.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the Elastic IP.",
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
			"address": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Public IPv4 address allocated for this Elastic IP.",
				Computed:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle for the resource. Accepted values: `Hour`, `Month`, `Year`.",
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

func (d *ElasticIPDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ElasticIPDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ElasticIPDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	eipID := data.Id.ValueString()
	if projectID == "" || eipID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Elastic IP ID (id) are required to read the Elastic IP",
		)
		return
	}

	response, err := d.client.Client.FromNetwork().ElasticIPs().Get(ctx, projectID, eipID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Elastic IP",
			NewTransportError("read", "Elasticip", err).Error(),
		)
		return
	}
	if apiErr := CheckResponse("read", "Elasticip", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError(
			"No data returned",
			"Elastic IP Get returned no data",
		)
		return
	}

	eip := response.Data

	if eip.Metadata.ID != nil {
		data.Id = types.StringValue(*eip.Metadata.ID)
	}
	if eip.Metadata.Name != nil {
		data.Name = types.StringValue(*eip.Metadata.Name)
	}
	if eip.Metadata.LocationResponse != nil {
		data.Location = types.StringValue(eip.Metadata.LocationResponse.Value)
	} else {
		data.Location = types.StringNull()
	}
	data.ProjectId = types.StringValue(projectID)
	if eip.Properties.Address != nil {
		data.Address = types.StringValue(*eip.Properties.Address)
	} else {
		data.Address = types.StringNull()
	}
	if eip.Properties.BillingPlan.BillingPeriod != "" {
		data.BillingPeriod = types.StringValue(billingPeriodFromAPI(eip.Properties.BillingPlan.BillingPeriod))
	} else {
		data.BillingPeriod = types.StringNull()
	}

	if len(eip.Metadata.Tags) > 0 {
		tagValues := make([]attr.Value, len(eip.Metadata.Tags))
		for i, tag := range eip.Metadata.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		data.Tags = types.ListValueMust(types.StringType, tagValues)
	} else {
		data.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	tflog.Trace(ctx, "read an Elastic IP data source", map[string]interface{}{"eip_id": eipID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
