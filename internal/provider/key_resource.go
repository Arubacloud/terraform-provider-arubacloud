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

type KeyResourceModel struct {
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

type KeyResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &KeyResource{}
var _ resource.ResourceWithImportState = &KeyResource{}

func NewKeyResource() resource.Resource {
	return &KeyResource{}
}

func (r *KeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_key"
}

func (r *KeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Key resource for KMS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Key identifier",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Key URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Key name",
				Required:            true,
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
				MarkdownDescription: "Encryption algorithm (e.g., Aes, Rsa)",
				Required:            true,
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "Key size in bits (e.g., 256, 2048) - optional, not returned by API",
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Key description - optional, not returned by API",
				Optional:            true,
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Key status - computed, not returned by API in v0.1.18",
				Computed:            true,
			},
		},
	}
}

func (r *KeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	kmsID := data.KMSID.ValueString()

	if projectID == "" || kmsID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and KMS ID are required to create a Key",
		)
		return
	}

	// Build the create request
	createRequest := sdktypes.KeyRequest{
		Name:      data.Name.ValueString(),
		Algorithm: sdktypes.KeyAlgorithm(data.Algorithm.ValueString()),
	}

	// Create the Key using the SDK
	response, err := r.client.Client.FromSecurity().KMS().Keys().Create(ctx, projectID, kmsID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Key",
			fmt.Sprintf("Unable to create Key: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		logContext := map[string]interface{}{
			"project_id": projectID,
			"kms_id":     kmsID,
			"key_name":   data.Name.ValueString(),
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to create Key", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		// Debug log the response
		tflog.Debug(ctx, "Key create response received", map[string]interface{}{
			"has_keyid":        response.Data.KeyID != nil,
			"has_name":         response.Data.Name != nil,
			"has_privatekeyid": response.Data.PrivateKeyID != nil,
		})

		if response.Data.KeyID != nil {
			data.Id = types.StringValue(*response.Data.KeyID)
		} else if response.Data.Name != nil && *response.Data.Name != "" {
			// Fallback to using name as ID if KeyID not available
			data.Id = types.StringValue(*response.Data.Name)
			tflog.Warn(ctx, "Using Key name as ID (KeyID not returned by API)", map[string]interface{}{
				"key_name": *response.Data.Name,
			})
		} else {
			resp.Diagnostics.AddError(
				"Invalid API Response",
				"Key created but no ID returned from API",
			)
			return
		}

		// URI not available in KeyResponse from SDK v0.1.20
		data.Uri = types.StringNull()

		// Size and description are not returned by API - set to null or preserve from plan
		if data.Size.IsNull() || data.Size.IsUnknown() {
			data.Size = types.Int64Null()
		}
		if data.Description.IsNull() || data.Description.IsUnknown() {
			data.Description = types.StringNull()
		}

		if response.Data.Status != nil {
			data.Status = types.StringValue(string(*response.Data.Status))
		} else {
			data.Status = types.StringNull()
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Key created but no data returned from API",
		)
		return
	}

	// Wait for Key to be active before returning
	keyID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromSecurity().KMS().Keys().Get(ctx, projectID, kmsID, keyID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil {
			// Keys don't have a status field in v0.1.18, if we can get it, it's active
			return "Active", nil
		}
		return "Unknown", nil
	}

	// Wait for Key to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "Key", keyID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Key Not Active",
			fmt.Sprintf("Key was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a Key resource", map[string]interface{}{
		"key_id":   data.Id.ValueString(),
		"key_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
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
	response, err := r.client.Client.FromSecurity().KMS().Keys().Get(ctx, projectID, kmsID, keyID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Key",
			fmt.Sprintf("Unable to read Key: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
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
		if !data.Size.IsNull() {
			// Keep the size from plan since it's not returned by API
		} else {
			data.Size = types.Int64Null()
		}

		if !data.Description.IsNull() {
			// Keep the description from plan since it's not returned by API
		} else {
			data.Description = types.StringNull()
		}

		if key.Status != nil {
			data.Status = types.StringValue(string(*key.Status))
		} else {
			data.Status = types.StringNull()
		}
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KeyResourceModel
	var state KeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Keys cannot be updated in SDK v0.1.18 - they must be recreated
	// For now, just read the current state
	resp.Diagnostics.AddWarning(
		"Key Update Not Supported",
		"Keys cannot be updated in the current SDK version. To modify a key, it must be destroyed and recreated.",
	)

	// Perform a read to get current state
	projectID := data.ProjectID.ValueString()
	kmsID := data.KMSID.ValueString()
	keyID := state.Id.ValueString()

	response, err := r.client.Client.FromSecurity().KMS().Keys().Get(ctx, projectID, kmsID, keyID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Key",
			fmt.Sprintf("Unable to read Key after update attempt: %s", err),
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
		if response.Data.KeyID != nil {
			data.Id = types.StringValue(*response.Data.KeyID)
		}
		data.Uri = types.StringNull()

		if response.Data.Status != nil {
			data.Status = types.StringValue(string(*response.Data.Status))
		} else {
			data.Status = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	kmsID := data.KMSID.ValueString()
	keyID := data.Id.ValueString()

	// If ID is unknown or empty, the resource doesn't exist (e.g., during plan or if never created)
	// Return early without error - this is expected behavior
	if data.Id.IsUnknown() || data.Id.IsNull() || keyID == "" {
		tflog.Debug(ctx, "Key ID is unknown or empty, skipping delete", map[string]interface{}{
			"key_id": keyID,
		})
		return
	}

	// Project ID and KMS ID should always be set, but check to be safe
	if projectID == "" || kmsID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and KMS ID are required to delete the Key",
		)
		return
	}

	// Delete the Key using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromSecurity().KMS().Keys().Delete(ctx, projectID, kmsID, keyID, nil)
		},
		ExtractSDKError,
		"Key",
		keyID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Key",
			fmt.Sprintf("Unable to delete Key: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a Key resource", map[string]interface{}{
		"key_id": keyID,
	})
}

func (r *KeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
