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

var _ resource.Resource = &ElasticIPResource{}
var _ resource.ResourceWithImportState = &ElasticIPResource{}

func NewElasticIPResource() resource.Resource {
	return &ElasticIPResource{}
}

type ElasticIPResource struct {
	client *ArubaCloudClient
}

type ElasticIPResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	Address       types.String `tfsdk:"address"`
	ProjectId     types.String `tfsdk:"project_id"`
}

func (r *ElasticIPResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticip"
}

func (r *ElasticIPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Project resource",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Elastic IP name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Elastic IP location",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the Elastic IP",
				Optional:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period for the Elastic IP (only 'hourly' allowed)",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this Elastic IP belongs to",
				Required:            true,
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "Elastic IP address",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Project Identifier",
				Computed:            true,
			},
		},
	}
}

func (r *ElasticIPResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ArubaCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *ElasticIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ElasticIPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"Project ID is required to create an Elastic IP",
		)
		return
	}

	// Extract tags from Terraform list
	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Build the create request
	createRequest := sdktypes.ElasticIPRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.ElasticIPPropertiesRequest{
			BillingPlan: sdktypes.BillingPeriodResource{
				BillingPeriod: data.BillingPeriod.ValueString(),
			},
		},
	}

	// Create the Elastic IP using the SDK
	response, err := r.client.Client.FromNetwork().ElasticIPs().Create(ctx, projectID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Elastic IP",
			fmt.Sprintf("Unable to create Elastic IP: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create Elastic IP"
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
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		} else {
			resp.Diagnostics.AddError(
				"Invalid API Response",
				"Elastic IP created but ID not returned from API",
			)
			return
		}
		if response.Data.Properties.Address != nil {
			data.Address = types.StringValue(*response.Data.Properties.Address)
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Elastic IP created but no data returned from API",
		)
		return
	}

	// Wait for Elastic IP to be active before returning (ElasticIP is referenced by CloudServer)
	// This ensures Terraform doesn't proceed to create dependent resources until ElasticIP is ready
	eipID := data.Id.ValueString()
	if eipID == "" {
		resp.Diagnostics.AddError(
			"Missing Elastic IP ID",
			"Elastic IP ID is required but was not set",
		)
		return
	}
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromNetwork().ElasticIPs().Get(ctx, projectID, eipID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for Elastic IP to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "ElasticIP", eipID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Elastic IP Not Active",
			fmt.Sprintf("Elastic IP was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	// Read the Elastic IP again to get the address and ensure ID is set (it might not be available immediately after creation)
	getResp, err := r.client.Client.FromNetwork().ElasticIPs().Get(ctx, projectID, eipID, nil)
	if err == nil && getResp != nil && getResp.Data != nil {
		// Ensure ID is set (should already be set, but double-check)
		if getResp.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*getResp.Data.Metadata.ID)
		}
		if getResp.Data.Properties.Address != nil {
			data.Address = types.StringValue(*getResp.Data.Properties.Address)
		}
		// Also update other fields that might have changed
		if getResp.Data.Metadata.Name != nil {
			data.Name = types.StringValue(*getResp.Data.Metadata.Name)
		}
	} else if err != nil {
		// If Get fails, log but don't fail - we already have the ID from create response
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh Elastic IP after creation: %v", err))
	}

	tflog.Trace(ctx, "created an Elastic IP resource", map[string]interface{}{
		"elasticip_id": data.Id.ValueString(),
		"elasticip_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ElasticIPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	eipID := data.Id.ValueString()

	// If ID is unknown or empty, the resource doesn't exist yet (e.g., during plan for new resources)
	// Return early without error - this is expected behavior
	if data.Id.IsUnknown() || data.Id.IsNull() || eipID == "" {
		tflog.Debug(ctx, "Elastic IP ID is unknown or empty, skipping read", map[string]interface{}{
			"eip_id": eipID,
		})
		return
	}

	// Project ID should always be set, but check to be safe
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID is required to read the Elastic IP",
		)
		return
	}

	// Get Elastic IP details using the SDK
	response, err := r.client.Client.FromNetwork().ElasticIPs().Get(ctx, projectID, eipID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Elastic IP",
			fmt.Sprintf("Unable to read Elastic IP: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read Elastic IP"
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
		eip := response.Data

		if eip.Metadata.ID != nil {
			data.Id = types.StringValue(*eip.Metadata.ID)
		}
		if eip.Metadata.Name != nil {
			data.Name = types.StringValue(*eip.Metadata.Name)
		}
		if eip.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(eip.Metadata.LocationResponse.Value)
		}
		if eip.Properties.Address != nil {
			data.Address = types.StringValue(*eip.Properties.Address)
		} else {
			// Address might not be available yet, set to null
			data.Address = types.StringNull()
		}
		data.BillingPeriod = types.StringValue(eip.Properties.BillingPlan.BillingPeriod)

		// Update tags from response
		if len(eip.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(eip.Metadata.Tags))
			for i, tag := range eip.Metadata.Tags {
				tagValues[i] = types.StringValue(tag)
			}
			tagsList, diags := types.ListValueFrom(ctx, types.StringType, tagValues)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = tagsList
			}
		} else {
			emptyList, diags := types.ListValue(types.StringType, []attr.Value{})
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = emptyList
			}
		}
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ElasticIPResourceModel
	var state ElasticIPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use IDs from state (they are immutable)
	projectID := state.ProjectId.ValueString()
	eipID := state.Id.ValueString()

	if projectID == "" || eipID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Elastic IP ID are required to update the Elastic IP",
		)
		return
	}

	// Get current Elastic IP details
	getResponse, err := r.client.Client.FromNetwork().ElasticIPs().Get(ctx, projectID, eipID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current Elastic IP",
			fmt.Sprintf("Unable to get current Elastic IP: %s", err),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"Elastic IP Not Found",
			"Elastic IP not found or no data returned",
		)
		return
	}

	current := getResponse.Data

	// Check if Elastic IP is in InCreation state
	if current.Status.State != nil && *current.Status.State == "InCreation" {
		resp.Diagnostics.AddError(
			"Cannot Update During Creation",
			"Cannot update Elastic IP while it is in 'InCreation' state. Please wait until the Elastic IP is fully created.",
		)
		return
	}

	// Get region value
	regionValue := ""
	if current.Metadata.LocationResponse != nil {
		regionValue = current.Metadata.LocationResponse.Value
	}
	if regionValue == "" {
		resp.Diagnostics.AddError(
			"Missing Region",
			"Unable to determine region value for Elastic IP",
		)
		return
	}

	// Extract tags from Terraform list
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

	// Build the update request
	updateRequest := sdktypes.ElasticIPRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.ElasticIPPropertiesRequest{
			BillingPlan: current.Properties.BillingPlan,
		},
	}

	// Update the Elastic IP using the SDK
	response, err := r.client.Client.FromNetwork().ElasticIPs().Update(ctx, projectID, eipID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Elastic IP",
			fmt.Sprintf("Unable to update Elastic IP: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to update Elastic IP"
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
		if response.Data.Properties.Address != nil {
			data.Address = types.StringValue(*response.Data.Properties.Address)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ElasticIPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	eipID := data.Id.ValueString()

	// If ID is unknown or empty, the resource doesn't exist (e.g., during plan or if never created)
	// Return early without error - this is expected behavior
	if data.Id.IsUnknown() || data.Id.IsNull() || eipID == "" {
		tflog.Debug(ctx, "Elastic IP ID is unknown or empty, skipping delete", map[string]interface{}{
			"eip_id": eipID,
		})
		return
	}

	// Project ID should always be set, but check to be safe
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID is required to delete the Elastic IP",
		)
		return
	}

	// Delete the Elastic IP using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromNetwork().ElasticIPs().Delete(ctx, projectID, eipID, nil)
		},
		ExtractSDKError,
		"ElasticIP",
		eipID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Elastic IP",
			fmt.Sprintf("Unable to delete Elastic IP: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted an Elastic IP resource", map[string]interface{}{
		"elasticip_id": eipID,
	})
}

func (r *ElasticIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
