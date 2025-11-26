// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type KeypairDataSourceModel struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Location types.String `tfsdk:"location"`
	Tags     types.List   `tfsdk:"tags"`
	Value    types.String `tfsdk:"value"`
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
			   "tags": schema.ListAttribute{
				   ElementType:         types.StringType,
				   MarkdownDescription: "List of tags for the keypair",
				   Computed:            true,
			   },
			   "value": schema.StringAttribute{
				   MarkdownDescription: "Keypair value",
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
	// Simulate API response
	data.Id = types.StringValue("keypair-id")
	data.Value = types.StringValue("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC...")
	tflog.Trace(ctx, "read a Keypair data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
