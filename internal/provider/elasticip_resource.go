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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	Uri           types.String `tfsdk:"uri"`
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
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				// Validators removed for v1.16.1 compatibility
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this Elastic IP belongs to",
				Required:            true,
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "Elastic IP address (computed from ElasticIpPropertiesResponse)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Elastic IP Identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Elastic IP URI",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				BillingPeriod: func() string {
					if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
						return data.BillingPeriod.ValueString()
					}
					// Default to "hourly" if not provided
					return "hourly"
				}(),
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
		logContext := map[string]interface{}{
			"project_id": projectID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to create Elastic IP", logContext)
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
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if response.Data.Properties.Address != nil {
			data.Address = types.StringValue(*response.Data.Properties.Address)
		}
		// Set billing_period from create response if available
		if response.Data.Properties.BillingPlan.BillingPeriod != "" {
			data.BillingPeriod = types.StringValue(response.Data.Properties.BillingPlan.BillingPeriod)
		} else if data.BillingPeriod.IsNull() || data.BillingPeriod.IsUnknown() {
			// Fallback to plan value or default
			if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
				// Keep plan value
			} else {
				data.BillingPeriod = types.StringValue("hourly")
			}
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
		if getResp.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*getResp.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		// Update address from re-read response (computed field from ElasticIpPropertiesResponse)
		if getResp.Data.Properties.Address != nil {
			data.Address = types.StringValue(*getResp.Data.Properties.Address)
		} else {
			// Address might not be available yet, set to null
			data.Address = types.StringNull()
		}
		// Update billing_period from the re-read response
		if getResp.Data.Properties.BillingPlan.BillingPeriod != "" {
			data.BillingPeriod = types.StringValue(getResp.Data.Properties.BillingPlan.BillingPeriod)
		} else {
			// Fallback to plan value or default if not available
			if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
				// Keep the plan value
			} else {
				data.BillingPeriod = types.StringValue("hourly")
			}
		}
		// Also update other fields that might have changed
		if getResp.Data.Metadata.Name != nil {
			data.Name = types.StringValue(*getResp.Data.Metadata.Name)
		}
	} else if err != nil {
		// If Get fails, log but don't fail - we already have the ID from create response
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh Elastic IP after creation: %v", err))
		// Ensure billing_period is set even if re-read fails
		if data.BillingPeriod.IsNull() || data.BillingPeriod.IsUnknown() {
			data.BillingPeriod = types.StringValue("hourly")
		}
	}

	tflog.Trace(ctx, "created an Elastic IP resource", map[string]interface{}{
		"elasticip_id":   data.Id.ValueString(),
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
		logContext := map[string]interface{}{
			"project_id": projectID,
			"eip_id":     eipID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to read Elastic IP", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		eip := response.Data

		if eip.Metadata.ID != nil {
			data.Id = types.StringValue(*eip.Metadata.ID)
		}
		if eip.Metadata.URI != nil {
			data.Uri = types.StringValue(*eip.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if eip.Metadata.Name != nil {
			data.Name = types.StringValue(*eip.Metadata.Name)
		}
		if eip.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(eip.Metadata.LocationResponse.Value)
		}
		// Update address from response (computed field from ElasticIpPropertiesResponse)
		if eip.Properties.Address != nil {
			data.Address = types.StringValue(*eip.Properties.Address)
		} else {
			// Address might not be available yet, set to null
			data.Address = types.StringNull()
		}
		// Always set billing_period from API response (it's always available from API)
		// This ensures it's always in state, preventing false changes in plan
		if eip.Properties.BillingPlan.BillingPeriod != "" {
			data.BillingPeriod = types.StringValue(eip.Properties.BillingPlan.BillingPeriod)
		} else {
			// Fallback to "hourly" if API doesn't return it (shouldn't happen)
			data.BillingPeriod = types.StringValue("hourly")
		}

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
		logContext := map[string]interface{}{
			"project_id": projectID,
			"eip_id":     eipID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to update Elastic IP", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.ProjectId = state.ProjectId
	data.Uri = state.Uri         // Preserve URI from state
	data.Address = state.Address // Preserve address from state (computed field)

	if response != nil && response.Data != nil {
		// Update from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			// If no URI in response, re-read the Elastic IP to get the latest state
			getResp, err := r.client.Client.FromNetwork().ElasticIPs().Get(ctx, projectID, eipID, nil)
			if err == nil && getResp != nil && getResp.Data != nil {
				if getResp.Data.Metadata.URI != nil {
					data.Uri = types.StringValue(*getResp.Data.Metadata.URI)
				} else {
					data.Uri = state.Uri // Fallback to state if not available
				}
				// Also update address from re-read if available
				if getResp.Data.Properties.Address != nil {
					data.Address = types.StringValue(*getResp.Data.Properties.Address)
				}
			} else {
				data.Uri = state.Uri // Fallback to state if re-read fails
			}
		}
		// Update address from response if available (computed field from ElasticIpPropertiesResponse)
		if response.Data.Properties.Address != nil {
			data.Address = types.StringValue(*response.Data.Properties.Address)
		}
		// Always set billing_period from response if available
		if response.Data.Properties.BillingPlan.BillingPeriod != "" {
			data.BillingPeriod = types.StringValue(response.Data.Properties.BillingPlan.BillingPeriod)
		} else {
			// If not in response, re-read to get it
			getResp, err := r.client.Client.FromNetwork().ElasticIPs().Get(ctx, projectID, eipID, nil)
			if err == nil && getResp != nil && getResp.Data != nil {
				if getResp.Data.Properties.BillingPlan.BillingPeriod != "" {
					data.BillingPeriod = types.StringValue(getResp.Data.Properties.BillingPlan.BillingPeriod)
				} else {
					// Fallback to state or default
					if !state.BillingPeriod.IsNull() && !state.BillingPeriod.IsUnknown() {
						data.BillingPeriod = state.BillingPeriod
					} else {
						data.BillingPeriod = types.StringValue("hourly")
					}
				}
				// Also update address from re-read if available (computed field from ElasticIpPropertiesResponse)
				if getResp.Data.Properties.Address != nil {
					data.Address = types.StringValue(*getResp.Data.Properties.Address)
				}
			} else {
				// Fallback to state or default
				if !state.BillingPeriod.IsNull() && !state.BillingPeriod.IsUnknown() {
					data.BillingPeriod = state.BillingPeriod
				} else {
					data.BillingPeriod = types.StringValue("hourly")
				}
			}
		}
	} else {
		// If no response, preserve URI from state and set billing_period from API or default
		data.Uri = state.Uri
		// Re-read to get billing_period
		getResp, err := r.client.Client.FromNetwork().ElasticIPs().Get(ctx, projectID, eipID, nil)
		if err == nil && getResp != nil && getResp.Data != nil {
			if getResp.Data.Properties.BillingPlan.BillingPeriod != "" {
				data.BillingPeriod = types.StringValue(getResp.Data.Properties.BillingPlan.BillingPeriod)
			} else {
				if !state.BillingPeriod.IsNull() && !state.BillingPeriod.IsUnknown() {
					data.BillingPeriod = state.BillingPeriod
				} else {
					data.BillingPeriod = types.StringValue("hourly")
				}
			}
		} else {
			// Fallback to state or default
			if !state.BillingPeriod.IsNull() && !state.BillingPeriod.IsUnknown() {
				data.BillingPeriod = state.BillingPeriod
			} else {
				data.BillingPeriod = types.StringValue("hourly")
			}
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
