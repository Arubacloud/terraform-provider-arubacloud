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

type KeyDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Uri         types.String `tfsdk:"uri"`
	Name        types.String `tfsdk:"name"`
	ProjectID   types.String `tfsdk:"project_id"`
	KMSID       types.String `tfsdk:"kms_id"`
	Algorithm   types.String `tfsdk:"algorithm"`
	Size        types.Int64  `tfsdk:"size"`
	Description types.String `tfsdk:"description"`
	Status      types.String `tfsdk:"status"`
}

type KeyDataSource struct {
	client *ArubaCloudClient
}

var _ datasource.DataSource = &KeyDataSource{}

func NewKeyDataSource() datasource.DataSource {
	return &KeyDataSource{}
}

func (d *KeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_key"
}

func (d *KeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Key data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Key identifier",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Key URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Key",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this Key belongs to",
				Required:            true,
			},
			"kms_id": schema.StringAttribute{
				MarkdownDescription: "ID of the associated KMS",
				Required:            true,
			},
			"algorithm": schema.StringAttribute{
				MarkdownDescription: "Encryption algorithm",
				Computed:            true,
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "Key size in bits",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Key description",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Key status",
				Computed:            true,
			},
		},
	}
}

func (d *KeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data KeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	kmsID := data.KMSID.ValueString()
	keyID := data.Id.ValueString()

	if projectID == "" || kmsID == "" || keyID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, KMS ID, and Key ID are required to read the Key",
		)
		return
	}

	// Get Key details using the SDK
	response, err := d.client.Client.FromSecurity().KMS().Keys().Get(ctx, projectID, kmsID, keyID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Key",
			fmt.Sprintf("Unable to read Key: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to read Key"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		key := response.Data
		if key.KeyID != nil {
			data.Id = types.StringValue(*key.KeyID)
		}
		// URI not available in KeyResponse
		data.Uri = types.StringNull()

		if key.Name != nil && *key.Name != "" {
			data.Name = types.StringValue(*key.Name)
		}
		if key.Algorithm != nil {
			data.Algorithm = types.StringValue(string(*key.Algorithm))
		}
		// Size field not available in KeyResponse from SDK v0.1.18
		data.Size = types.Int64Null()
		// Description field not available in KeyResponse from SDK v0.1.18
		data.Description = types.StringNull()

		if key.Status != nil {
			data.Status = types.StringValue(string(*key.Status))
		} else {
			data.Status = types.StringNull()
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Key not found or no data returned from API",
		)
		return
	}

	tflog.Trace(ctx, "read a Key data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
