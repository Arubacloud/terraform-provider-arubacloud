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

type KeypairResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	ProjectID types.String `tfsdk:"project_id"`
	Value     types.String `tfsdk:"value"`
}

type KeypairResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &KeypairResource{}
var _ resource.ResourceWithImportState = &KeypairResource{}

func NewKeypairResource() resource.Resource {
	return &KeypairResource{}
}

func (r *KeypairResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keypair"
}

func (r *KeypairResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Keypair resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Keypair identifier (name)",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Keypair name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Keypair location",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Public key value",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *KeypairResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KeypairResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KeypairResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"Project ID is required to create a keypair",
		)
		return
	}

	// Build the create request
	createRequest := sdktypes.KeyPairRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.KeyPairPropertiesRequest{
			Value: data.Value.ValueString(),
		},
	}

	// Create the keypair using the SDK
	response, err := r.client.Client.FromCompute().KeyPairs().Create(ctx, projectID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating keypair",
			fmt.Sprintf("Unable to create keypair: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create keypair"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		// Include status code for better debugging
		if response.StatusCode > 0 {
			errorMsg = fmt.Sprintf("%s (HTTP %d)", errorMsg, response.StatusCode)
		}
		
		// Log the full error for debugging
		tflog.Error(ctx, "Keypair creation failed", map[string]interface{}{
			"error_title":  response.Error.Title,
			"error_detail":  response.Error.Detail,
			"status_code":   response.StatusCode,
			"keypair_name":  data.Name.ValueString(),
			"project_id":    projectID,
		})
		
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		if response.Data.Metadata.Name != nil {
			data.Id = types.StringValue(*response.Data.Metadata.Name)
		} else {
			data.Id = types.StringValue(data.Name.ValueString())
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Keypair created but no data returned from API",
		)
		return
	}

	// Wait for Keypair to be active before returning (Keypair is referenced by CloudServer)
	// This ensures Terraform doesn't proceed to create dependent resources until Keypair is ready
	keypairName := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromCompute().KeyPairs().Get(ctx, projectID, keypairName, nil)
		if err != nil {
			return "", err
		}
		// Keypairs don't have a Status field - if we can get it, it's ready
		if getResp != nil && getResp.Data != nil {
			return "Active", nil
		}
		return "Unknown", nil
	}

	// Wait for Keypair to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "Keypair", keypairName, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Keypair Not Active",
			fmt.Sprintf("Keypair was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a Keypair resource", map[string]interface{}{
		"keypair_id": data.Id.ValueString(),
		"keypair_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeypairResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KeypairResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	keypairName := data.Id.ValueString()
	if keypairName == "" {
		keypairName = data.Name.ValueString()
	}

	if projectID == "" || keypairName == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Keypair name are required to read the keypair",
		)
		return
	}

	// Get keypair details using the SDK
	response, err := r.client.Client.FromCompute().KeyPairs().Get(ctx, projectID, keypairName, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading keypair",
			fmt.Sprintf("Unable to read keypair: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read keypair"
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
		keypair := response.Data
		if keypair.Metadata.Name != nil {
			data.Id = types.StringValue(*keypair.Metadata.Name)
			data.Name = types.StringValue(*keypair.Metadata.Name)
		}
		if keypair.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(keypair.Metadata.LocationResponse.Value)
		}
		// Note: Public key value is not returned by the API for security reasons
		// We keep the existing value from state
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeypairResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KeypairResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Keypair update is not supported by the API
	// If the public key changes, we need to delete and recreate
	var state KeypairResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if public key changed
	if !data.Value.Equal(state.Value) {
		resp.Diagnostics.AddError(
			"Keypair Update Not Supported",
			"Changing the public key value is not supported. Please delete and recreate the keypair.",
		)
		return
	}

	// If only name changed, that's also not supported
	if !data.Name.Equal(state.Name) {
		resp.Diagnostics.AddError(
			"Keypair Name Update Not Supported",
			"Changing the keypair name is not supported. Please delete and recreate the keypair.",
		)
		return
	}

	// No changes or only project_id changed (which shouldn't happen)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeypairResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KeypairResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	keypairName := data.Id.ValueString()
	if keypairName == "" {
		keypairName = data.Name.ValueString()
	}

	if projectID == "" || keypairName == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Keypair name are required to delete the keypair",
		)
		return
	}

	// Delete the keypair using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromCompute().KeyPairs().Delete(ctx, projectID, keypairName, nil)
		},
		ExtractSDKError,
		"Keypair",
		keypairName,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting keypair",
			fmt.Sprintf("Unable to delete keypair: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a Keypair resource", map[string]interface{}{
		"keypair_name": keypairName,
	})
}

func (r *KeypairResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
