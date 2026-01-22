// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type KMIPResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	ProjectID types.String `tfsdk:"project_id"`
	KMSID     types.String `tfsdk:"kms_id"`
}

type KMIPResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &KMIPResource{}
var _ resource.ResourceWithImportState = &KMIPResource{}

func NewKMIPResource() resource.Resource {
	return &KMIPResource{}
}

func (r *KMIPResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kmip"
}

func (r *KMIPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "KMIP resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "KMIP identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "KMIP name",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this KMIP belongs to",
				Required:            true,
			},
			"kms_id": schema.StringAttribute{
				MarkdownDescription: "ID of the associated KMS",
				Required:            true,
			},
		},
	}
}

func (r *KMIPResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ArubaCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ArubaCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *KMIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KMIPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	kmsID := data.KMSID.ValueString()

	if projectID == "" || kmsID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and KMS ID are required to create a KMIP",
		)
		return
	}

	// KMIP is a child resource of KMS, similar to Subnet being a child of VPC
	// Terraform handles the dependency automatically through kms_id reference
	// No need to validate KMS state - if KMS was just created, it's already active

	// Build the create request
	createRequest := sdktypes.KmipRequest{
		Name: data.Name.ValueString(),
	}

	// Debug log the request
	tflog.Debug(ctx, "Creating KMIP with request", map[string]interface{}{
		"kmip_name":  createRequest.Name,
		"project_id": projectID,
		"kms_id":     kmsID,
	})

	// Create the KMIP using the SDK - nested under KMS
	response, err := r.client.Client.FromSecurity().KMS().Kmips().Create(ctx, projectID, kmsID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating KMIP",
			fmt.Sprintf("Unable to create KMIP: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		logContext := map[string]interface{}{
			"project_id": projectID,
			"kms_id":     kmsID,
			"kmip_name":  data.Name.ValueString(),
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to create KMIP", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		// KMIP uses name as ID or has a separate ID field
		if response.Data.ID != nil {
			data.Id = types.StringValue(*response.Data.ID)
		} else if response.Data.Name != nil && *response.Data.Name != "" {
			data.Id = types.StringValue(*response.Data.Name)
		} else {
			resp.Diagnostics.AddError(
				"Invalid API Response",
				"KMIP created but no ID returned from API",
			)
			return
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"KMIP created but no data returned from API",
		)
		return
	}

	// Wait for KMIP to be active before returning
	// This ensures Terraform doesn't proceed until KMIP is ready
	kmipID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromSecurity().KMS().Kmips().Get(ctx, projectID, kmsID, kmipID, nil)
		if err != nil {
			return "", err
		}
		// KMIP may not have a status field, so if we can get it, it's ready
		if getResp != nil && getResp.Data != nil {
			return "Active", nil
		}
		return "Unknown", nil
	}

	// Wait for KMIP to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "KMIP", kmipID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"KMIP Not Active",
			fmt.Sprintf("KMIP was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a KMIP resource", map[string]interface{}{
		"kmip_id":   data.Id.ValueString(),
		"kmip_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KMIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KMIPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	kmsID := data.KMSID.ValueString()
	kmipID := data.Id.ValueString()

	if projectID == "" || kmsID == "" || kmipID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, KMS ID, and KMIP ID are required to read the KMIP",
		)
		return
	}

	// Get KMIP details using the SDK
	response, err := r.client.Client.FromSecurity().KMS().Kmips().Get(ctx, projectID, kmsID, kmipID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading KMIP",
			fmt.Sprintf("Unable to read KMIP: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
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
		if kmip.ID != nil {
			data.Id = types.StringValue(*kmip.ID)
		} else if kmip.Name != nil && *kmip.Name != "" {
			data.Id = types.StringValue(*kmip.Name)
		}
		if kmip.Name != nil && *kmip.Name != "" {
			data.Name = types.StringValue(*kmip.Name)
		}
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KMIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KMIPResourceModel
	var state KMIPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	kmsID := data.KMSID.ValueString()
	kmipID := state.Id.ValueString()

	if projectID == "" || kmsID == "" || kmipID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, KMS ID, and KMIP ID are required to update the KMIP",
		)
		return
	}

	// Get KMS to validate it exists
	kmsResp, err := r.client.Client.FromSecurity().KMS().Get(ctx, projectID, kmsID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting KMS",
			fmt.Sprintf("Unable to get KMS: %s", err),
		)
		return
	}

	if kmsResp == nil || kmsResp.Data == nil || kmsResp.Data.Metadata.URI == nil {
		resp.Diagnostics.AddError(
			"KMS Not Found",
			"KMS not found or missing URI",
		)
		return
	}

	// Build update request
	updateRequest := sdktypes.KmipRequest{
		Name: data.Name.ValueString(),
	}

	// KMIP doesn't support direct updates - delete and recreate
	// First delete the existing KMIP
	delResp, err := r.client.Client.FromSecurity().KMS().Kmips().Delete(ctx, projectID, kmsID, kmipID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating KMIP",
			fmt.Sprintf("Unable to update KMIP: %s", err),
		)
		return
	}

	if delResp != nil && delResp.IsError() && delResp.Error != nil {
		errorMsg := "Failed to delete KMIP during update"
		if delResp.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *delResp.Error.Title)
		}
		if delResp.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *delResp.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	// Recreate the KMIP with updated values
	createResp, err := r.client.Client.FromSecurity().KMS().Kmips().Create(ctx, projectID, kmsID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error recreating KMIP",
			fmt.Sprintf("Unable to recreate KMIP: %s", err),
		)
		return
	}

	if createResp != nil && createResp.IsError() && createResp.Error != nil {
		errorMsg := "Failed to recreate KMIP"
		if createResp.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *createResp.Error.Title)
		}
		if createResp.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *createResp.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if createResp != nil && createResp.Data != nil {
		if createResp.Data.ID != nil {
			data.Id = types.StringValue(*createResp.Data.ID)
		} else if createResp.Data.Name != nil && *createResp.Data.Name != "" {
			data.Id = types.StringValue(*createResp.Data.Name)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KMIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KMIPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	kmsID := data.KMSID.ValueString()
	kmipID := data.Id.ValueString()

	// If ID is unknown or empty, the resource doesn't exist (e.g., during plan or if never created)
	// Return early without error - this is expected behavior
	if data.Id.IsUnknown() || data.Id.IsNull() || kmipID == "" {
		tflog.Debug(ctx, "KMIP ID is unknown or empty, skipping delete", map[string]interface{}{
			"kmip_id": kmipID,
		})
		return
	}

	// Project ID and KMS ID should always be set, but check to be safe
	if projectID == "" || kmsID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and KMS ID are required to delete the KMIP",
		)
		return
	}

	// Delete the KMIP using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromSecurity().KMS().Kmips().Delete(ctx, projectID, kmsID, kmipID, nil)
		},
		ExtractSDKError,
		"KMIP",
		kmipID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting KMIP",
			fmt.Sprintf("Unable to delete KMIP: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a KMIP resource", map[string]interface{}{
		"kmip_id": kmipID,
	})
}

func (r *KMIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
