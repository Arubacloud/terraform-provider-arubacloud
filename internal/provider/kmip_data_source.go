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

type KMIPDataSourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ProjectID    types.String `tfsdk:"project_id"`
	KMSID        types.String `tfsdk:"kms_id"`
	Type         types.String `tfsdk:"type"`
	Status       types.String `tfsdk:"status"`
	CreationDate types.String `tfsdk:"creation_date"`
	DeletionDate types.String `tfsdk:"deletion_date"`
}

type KMIPDataSource struct {
	client *ArubaCloudClient
}

var _ datasource.DataSource = &KMIPDataSource{}

func NewKMIPDataSource() datasource.DataSource {
	return &KMIPDataSource{}
}

func (d *KMIPDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kmip"
}

func (d *KMIPDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "KMIP data source - retrieves information about a KMIP service within a KMS instance",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "KMIP identifier (typically the name)",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the KMIP service",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this KMIP belongs to",
				Required:            true,
			},
			"kms_id": schema.StringAttribute{
				MarkdownDescription: "ID of the associated KMS instance",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "KMIP service type",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "KMIP service status",
				Computed:            true,
			},
			"creation_date": schema.StringAttribute{
				MarkdownDescription: "KMIP service creation date",
				Computed:            true,
			},
			"deletion_date": schema.StringAttribute{
				MarkdownDescription: "KMIP service deletion date",
				Computed:            true,
			},
		},
	}
}

func (d *KMIPDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KMIPDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data KMIPDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	kmsID := data.KMSID.ValueString()
	kmipID := data.Id.ValueString()

	if projectID == "" || kmsID == "" || kmipID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, KMS ID, and KMIP ID are required to read the KMIP service",
		)
		return
	}

	// Get KMIP details using the SDK
	response, err := d.client.Client.FromSecurity().KMS().Kmips().Get(ctx, projectID, kmsID, kmipID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading KMIP",
			fmt.Sprintf("Unable to read KMIP service: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to read KMIP"
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
		kmip := response.Data

		// Set the ID (KMIP typically uses name as ID)
		if kmip.ID != nil {
			data.Id = types.StringValue(*kmip.ID)
		} else if kmip.Name != nil && *kmip.Name != "" {
			data.Id = types.StringValue(*kmip.Name)
		}

		// Set name
		if kmip.Name != nil && *kmip.Name != "" {
			data.Name = types.StringValue(*kmip.Name)
		} else {
			data.Name = types.StringNull()
		}

		// Set type
		if kmip.Type != nil && *kmip.Type != "" {
			data.Type = types.StringValue(*kmip.Type)
		} else {
			data.Type = types.StringNull()
		}

		// Set status (ServiceStatus is a string type alias)
		if kmip.Status != nil {
			data.Status = types.StringValue(string(*kmip.Status))
		} else {
			data.Status = types.StringNull()
		}

		// Set creation_date
		if kmip.CreationDate != nil && *kmip.CreationDate != "" {
			data.CreationDate = types.StringValue(*kmip.CreationDate)
		} else {
			data.CreationDate = types.StringNull()
		}

		// Set deletion_date
		if kmip.DeletionDate != nil && *kmip.DeletionDate != "" {
			data.DeletionDate = types.StringValue(*kmip.DeletionDate)
		} else {
			data.DeletionDate = types.StringNull()
		}

		tflog.Trace(ctx, fmt.Sprintf("Successfully read KMIP data source: %s", data.Id.ValueString()))
	} else {
		resp.Diagnostics.AddError(
			"Empty Response",
			"Received empty response from KMIP API",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
