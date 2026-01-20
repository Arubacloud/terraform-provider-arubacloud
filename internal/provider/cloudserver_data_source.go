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
				Computed:            true,
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

	// Populate all fields with example data
	data.Uri = types.StringValue("/v2/cloudservers/68398923fb2cb026400d4d31")
	data.Name = types.StringValue("example-cloudserver")
	data.Location = types.StringValue("ITBG-Bergamo")
	data.ProjectID = types.StringValue("68398923fb2cb026400d4d31")
	data.Zone = types.StringValue("ITBG-1")
	data.Tags = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("compute"),
		types.StringValue("production"),
	})
	// Network fields
	data.VpcUriRef = types.StringValue("/v2/vpcs/vpc-68398923fb2cb026400d4d32")
	data.ElasticIpUriRef = types.StringValue("/v2/elasticips/eip-68398923fb2cb026400d4d33")
	data.SubnetUriRefs = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("/v2/subnets/subnet-68398923fb2cb026400d4d35"),
		types.StringValue("/v2/subnets/subnet-68398923fb2cb026400d4d36"),
	})
	data.SecurityGroupUriRefs = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("/v2/securitygroups/sg-68398923fb2cb026400d4d37"),
		types.StringValue("/v2/securitygroups/sg-68398923fb2cb026400d4d38"),
	})
	// Settings fields
	data.FlavorName = types.StringValue("CSO4A8")
	data.KeyPairUriRef = types.StringValue("/v2/keypairs/keypair-example")
	data.UserData = types.StringValue("#cloud-config\npackages:\n  - nginx")
	// Storage fields
	data.BootVolumeUriRef = types.StringValue("/v2/blockstorages/vol-68398923fb2cb026400d4d34")

	tflog.Trace(ctx, "read a CloudServer data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
