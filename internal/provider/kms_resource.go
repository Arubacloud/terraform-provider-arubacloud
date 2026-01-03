// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type KMSResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	ProjectID     types.String `tfsdk:"project_id"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	BillingPeriod types.String `tfsdk:"billing_period"`
}

type KMSResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &KMSResource{}
var _ resource.ResourceWithImportState = &KMSResource{}

func NewKMSResource() resource.Resource {
	return &KMSResource{}
}

func (r *KMSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kms"
}

func (r *KMSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "KMS resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "KMS identifier",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "KMS URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "KMS name",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this KMS belongs to",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Location for the KMS",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the KMS",
				Optional:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period for the KMS",
				Required:            true,
			},
		},
	}
}

func (r *KMSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KMSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KMSResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()

	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID is required to create a KMS",
		)
		return
	}

	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Build the create request
	createRequest := sdktypes.KmsRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.KmsPropertiesRequest{
			BillingPeriod: sdktypes.BillingPeriodResource{
				BillingPeriod: data.BillingPeriod.ValueString(),
			},
		},
	}

	// Create the KMS using the SDK
	response, err := r.client.Client.FromSecurity().KMSKeys().Create(ctx, projectID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating KMS",
			fmt.Sprintf("Unable to create KMS: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create KMS"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil && response.Data.Metadata.ID != nil {
		data.Id = types.StringValue(*response.Data.Metadata.ID)
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"KMS created but no ID returned from API",
		)
		return
	}

	// Wait for KMS to be active before returning
	// This ensures Terraform doesn't proceed until KMS is ready
	kmsID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromSecurity().KMSKeys().Get(ctx, projectID, kmsID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for KMS to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "KMS", kmsID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"KMS Not Active",
			fmt.Sprintf("KMS was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a KMS resource", map[string]interface{}{
		"kms_id": data.Id.ValueString(),
		"kms_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KMSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KMSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	kmsID := data.Id.ValueString()

	if projectID == "" || kmsID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and KMS ID are required to read the KMS",
		)
		return
	}

	// Get KMS details using the SDK
	response, err := r.client.Client.FromSecurity().KMSKeys().Get(ctx, projectID, kmsID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading KMS",
			fmt.Sprintf("Unable to read KMS: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read KMS"
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
		kms := response.Data
		if kms.Metadata.ID != nil {
			data.Id = types.StringValue(*kms.Metadata.ID)
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		}
		if kms.Metadata.Name != nil {
			data.Name = types.StringValue(*kms.Metadata.Name)
		}
		if kms.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(kms.Metadata.LocationResponse.Value)
		}
		if kms.Metadata.Tags != nil {
			tagValues := make([]attr.Value, len(kms.Metadata.Tags))
			for i, tag := range kms.Metadata.Tags {
				tagValues[i] = types.StringValue(tag)
			}
			tagsList, diags := types.ListValue(types.StringType, tagValues)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = tagsList
			}
		} else {
			data.Tags = types.ListNull(types.StringType)
		}
		if kms.Properties.BillingPeriod.BillingPeriod != "" {
			data.BillingPeriod = types.StringValue(kms.Properties.BillingPeriod.BillingPeriod)
		}
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KMSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KMSResourceModel
	var state KMSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get IDs from state (not plan) - IDs are immutable and should always be in state
	projectID := state.ProjectID.ValueString()
	kmsID := state.Id.ValueString()

	if projectID == "" || kmsID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and KMS ID are required to update the KMS",
		)
		return
	}

	// Get current KMS to preserve fields
	getResp, err := r.client.Client.FromSecurity().KMSKeys().Get(ctx, projectID, kmsID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting KMS",
			fmt.Sprintf("Unable to get KMS: %s", err),
		)
		return
	}

	if getResp == nil || getResp.Data == nil {
		resp.Diagnostics.AddError(
			"KMS Not Found",
			"KMS not found",
		)
		return
	}

	current := getResp.Data
	regionValue := ""
	if current.Metadata.LocationResponse != nil {
		regionValue = current.Metadata.LocationResponse.Value
	}

	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		tags = current.Metadata.Tags
	}

	// Build update request
	updateRequest := sdktypes.KmsRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.KmsPropertiesRequest{
			BillingPeriod: sdktypes.BillingPeriodResource{
				BillingPeriod: data.BillingPeriod.ValueString(),
			},
		},
	}

	// Update the KMS using the SDK
	response, err := r.client.Client.FromSecurity().KMSKeys().Update(ctx, projectID, kmsID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating KMS",
			fmt.Sprintf("Unable to update KMS: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to update KMS"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil && response.Data.Metadata.ID != nil {
		data.Id = types.StringValue(*response.Data.Metadata.ID)
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
	}

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.ProjectID = state.ProjectID

	if response != nil && response.Data != nil {
		// Update from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KMSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KMSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	kmsID := data.Id.ValueString()

	if projectID == "" || kmsID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and KMS ID are required to delete the KMS",
		)
		return
	}

	// Delete the KMS using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromSecurity().KMSKeys().Delete(ctx, projectID, kmsID, nil)
		},
		ExtractSDKError,
		"KMS",
		kmsID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting KMS",
			fmt.Sprintf("Unable to delete KMS: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a KMS resource", map[string]interface{}{
		"kms_id": kmsID,
	})
}

func (r *KMSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
