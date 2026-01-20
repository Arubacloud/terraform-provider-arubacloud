// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

var _ datasource.DataSource = &ContainerRegistryDataSource{}

func NewContainerRegistryDataSource() datasource.DataSource {
	return &ContainerRegistryDataSource{}
}

type ContainerRegistryDataSource struct {
	client *ArubaCloudClient
}

type ContainerRegistryDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	ProjectID     types.String `tfsdk:"project_id"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	// Network fields (flattened)
	PublicIpUriRef      types.String `tfsdk:"public_ip_uri_ref"`
	VpcUriRef           types.String `tfsdk:"vpc_uri_ref"`
	SubnetUriRef        types.String `tfsdk:"subnet_uri_ref"`
	SecurityGroupUriRef types.String `tfsdk:"security_group_uri_ref"`
	// Storage fields (flattened)
	BlockStorageUriRef types.String `tfsdk:"block_storage_uri_ref"`
	// Settings fields (flattened)
	AdminUser             types.String `tfsdk:"admin_user"`
	ConcurrentUsersFlavor types.String `tfsdk:"concurrent_users_flavor"`
}

func (d *ContainerRegistryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_containerregistry"
}

func (d *ContainerRegistryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Container Registry data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Container Registry identifier",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Container Registry URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Container Registry name",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Container Registry location",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the Container Registry resource",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this Container Registry belongs to",
				Computed:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period (Hour, Month, Year)",
				Computed:            true,
			},
			"public_ip_uri_ref": schema.StringAttribute{
				MarkdownDescription: "Public IP URI reference",
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
			"block_storage_uri_ref": schema.StringAttribute{
				MarkdownDescription: "Block Storage URI reference",
				Computed:            true,
			},
			"admin_user": schema.StringAttribute{
				MarkdownDescription: "Administrator username",
				Computed:            true,
			},
			"concurrent_users_flavor": schema.StringAttribute{
				MarkdownDescription: "Concurrent users flavor size (Small, Medium, HighPerf)",
				Computed:            true,
			},
		},
	}
}

func (d *ContainerRegistryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ArubaCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *ContainerRegistryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ContainerRegistryDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Populate all fields with example data
	data.Uri = types.StringValue("/v2/containerregistries/cr-68398923fb2cb026400d4d31")
	data.Name = types.StringValue("example-containerregistry")
	data.Location = types.StringValue("ITBG-Bergamo")
	data.Tags = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("registry"),
		types.StringValue("docker"),
	})
	data.ProjectID = types.StringValue("68398923fb2cb026400d4d31")
	data.BillingPeriod = types.StringValue("Hour")
	// Network fields
	data.PublicIpUriRef = types.StringValue("/v2/elasticips/eip-68398923fb2cb026400d4d32")
	data.VpcUriRef = types.StringValue("/v2/vpcs/vpc-68398923fb2cb026400d4d33")
	data.SubnetUriRef = types.StringValue("/v2/subnets/subnet-68398923fb2cb026400d4d34")
	data.SecurityGroupUriRef = types.StringValue("/v2/securitygroups/sg-68398923fb2cb026400d4d35")
	// Storage fields
	data.BlockStorageUriRef = types.StringValue("/v2/blockstorages/bs-68398923fb2cb026400d4d36")
	// Settings fields
	data.AdminUser = types.StringValue("admin")
	data.ConcurrentUsersFlavor = types.StringValue("Medium")
	tflog.Trace(ctx, "read a Container Registry data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
