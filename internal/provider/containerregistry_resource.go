// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ContainerRegistryResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	ProjectID     types.String `tfsdk:"project_id"`
	Network       types.Object `tfsdk:"network"`
	Storage       types.Object `tfsdk:"storage"`
	Settings      types.Object `tfsdk:"settings"`
	BillingPeriod types.String `tfsdk:"billing_period"`
}

type ContainerRegistryResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &ContainerRegistryResource{}
var _ resource.ResourceWithImportState = &ContainerRegistryResource{}

func NewContainerRegistryResource() resource.Resource {
	return &ContainerRegistryResource{}
}

func (r *ContainerRegistryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_containerregistry"
}

func (r *ContainerRegistryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Container Registry resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Container Registry identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Container Registry URI",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Container Registry name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Container Registry location",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the Container Registry resource",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this Container Registry belongs to",
				Required:            true,
			},
			"network": schema.SingleNestedAttribute{
				MarkdownDescription: "Network configuration for the Container Registry",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"public_ip_uri_ref": schema.StringAttribute{
						MarkdownDescription: "Public IP URI reference (e.g., `arubacloud_elasticip.example.uri`)",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"vpc_uri_ref": schema.StringAttribute{
						MarkdownDescription: "VPC URI reference (e.g., `arubacloud_vpc.example.uri`)",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"subnet_uri_ref": schema.StringAttribute{
						MarkdownDescription: "Subnet URI reference (e.g., `arubacloud_subnet.example.uri`)",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"security_group_uri_ref": schema.StringAttribute{
						MarkdownDescription: "Security Group URI reference (e.g., `arubacloud_securitygroup.example.uri`)",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"storage": schema.SingleNestedAttribute{
				MarkdownDescription: "Storage configuration for the Container Registry",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"block_storage_uri_ref": schema.StringAttribute{
						MarkdownDescription: "Block Storage URI reference (e.g., `arubacloud_blockstorage.example.uri`)",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Settings configuration for the Container Registry",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"concurrent_users_flavor": schema.StringAttribute{
						MarkdownDescription: "Concurrent users flavor: small, medium, or large",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("small", "medium", "large"),
						},
					},
					"admin_user": schema.StringAttribute{
						MarkdownDescription: "Administrator username",
						Optional:            true,
					},
				},
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period (Hour, Month, Year)",
				Optional:            true,
			},
		},
	}
}

func (r *ContainerRegistryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ContainerRegistryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ContainerRegistryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"Project ID is required to create a container registry",
		)
		return
	}

	// Extract tags
	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Extract network configuration from nested object
	if data.Network.IsNull() || data.Network.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing Network Configuration",
			"Network configuration is required to create a container registry",
		)
		return
	}

	networkObj, diags := data.Network.ToObjectValue(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	networkAttrs := networkObj.Attributes()
	publicIPURI := ""
	vpcURI := ""
	subnetURI := ""
	securityGroupURI := ""

	if publicIpAttr, ok := networkAttrs["public_ip_uri_ref"]; ok && publicIpAttr != nil {
		if publicIpStr, ok := publicIpAttr.(types.String); ok && !publicIpStr.IsNull() && !publicIpStr.IsUnknown() {
			publicIPURI = publicIpStr.ValueString()
		}
	}
	if vpcAttr, ok := networkAttrs["vpc_uri_ref"]; ok && vpcAttr != nil {
		if vpcStr, ok := vpcAttr.(types.String); ok && !vpcStr.IsNull() && !vpcStr.IsUnknown() {
			vpcURI = vpcStr.ValueString()
		}
	}
	if subnetAttr, ok := networkAttrs["subnet_uri_ref"]; ok && subnetAttr != nil {
		if subnetStr, ok := subnetAttr.(types.String); ok && !subnetStr.IsNull() && !subnetStr.IsUnknown() {
			subnetURI = subnetStr.ValueString()
		}
	}
	if securityGroupAttr, ok := networkAttrs["security_group_uri_ref"]; ok && securityGroupAttr != nil {
		if securityGroupStr, ok := securityGroupAttr.(types.String); ok && !securityGroupStr.IsNull() && !securityGroupStr.IsUnknown() {
			securityGroupURI = securityGroupStr.ValueString()
		}
	}

	// Extract storage configuration from nested object
	if data.Storage.IsNull() || data.Storage.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing Storage Configuration",
			"Storage configuration is required to create a container registry",
		)
		return
	}

	storageObj, diags := data.Storage.ToObjectValue(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	storageAttrs := storageObj.Attributes()
	blockStorageURI := ""

	if blockStorageAttr, ok := storageAttrs["block_storage_uri_ref"]; ok && blockStorageAttr != nil {
		if blockStorageStr, ok := blockStorageAttr.(types.String); ok && !blockStorageStr.IsNull() && !blockStorageStr.IsUnknown() {
			blockStorageURI = blockStorageStr.ValueString()
		}
	}

	// Build the create request
	createRequest := sdktypes.ContainerRegistryRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.ContainerRegistryPropertiesRequest{
			PublicIp: sdktypes.ReferenceResource{
				URI: publicIPURI,
			},
			VPC: sdktypes.ReferenceResource{
				URI: vpcURI,
			},
			Subnet: sdktypes.ReferenceResource{
				URI: subnetURI,
			},
			SecurityGroup: sdktypes.ReferenceResource{
				URI: securityGroupURI,
			},
			BlockStorage: sdktypes.ReferenceResource{
				URI: blockStorageURI,
			},
		},
	}

	// Extract settings configuration (concurrent_users_flavor and admin_user)
	if data.Settings.IsNull() || data.Settings.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing Settings Configuration",
			"Settings configuration is required to create a container registry",
		)
		return
	}

	settingsObj, diags := data.Settings.ToObjectValue(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	settingsAttrs := settingsObj.Attributes()
	var concurrentUsersFlavor string
	var adminUser string

	if concurrentUsersAttr, ok := settingsAttrs["concurrent_users_flavor"]; ok && concurrentUsersAttr != nil {
		if cuStr, ok := concurrentUsersAttr.(types.String); ok && !cuStr.IsNull() && !cuStr.IsUnknown() {
			concurrentUsersFlavor = cuStr.ValueString()
		}
	}

	if adminUserAttr, ok := settingsAttrs["admin_user"]; ok && adminUserAttr != nil {
		if auStr, ok := adminUserAttr.(types.String); ok && !auStr.IsNull() && !auStr.IsUnknown() {
			adminUser = auStr.ValueString()
		}
	}

	// Debug logging
	tflog.Debug(ctx, "Container Registry settings extracted", map[string]interface{}{
		"concurrent_users_flavor": concurrentUsersFlavor,
		"admin_user":              adminUser,
	})

	// Add concurrent_users_flavor field (required, maps to 'size' in API)
	// Note: SDK uses ConcurrentUsers field which maps to 'size' JSON field
	if concurrentUsersFlavor != "" {
		createRequest.Properties.ConcurrentUsers = &concurrentUsersFlavor
		tflog.Debug(ctx, "Setting ConcurrentUsers in request", map[string]interface{}{
			"value": concurrentUsersFlavor,
		})
	} else {
		tflog.Warn(ctx, "concurrent_users_flavor is empty, size will not be set in request")
	}

	// Add optional fields
	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		createRequest.Properties.BillingPlan = &sdktypes.BillingPeriodResource{
			BillingPeriod: data.BillingPeriod.ValueString(),
		}
	}

	if adminUser != "" {
		createRequest.Properties.AdminUser = &sdktypes.UserCredential{
			Username: adminUser,
		}
	}

	// Create the container registry using the SDK
	containerClient := r.client.Client.FromContainer()
	if containerClient == nil {
		resp.Diagnostics.AddError(
			"Container Client Not Available",
			"Container client is not available",
		)
		return
	}

	registryClient := containerClient.ContainerRegistry()
	if registryClient == nil {
		resp.Diagnostics.AddError(
			"Container Registry Client Not Available",
			"Container Registry client is not available",
		)
		return
	}

	// Debug: Log the full request
	tflog.Debug(ctx, "Creating container registry with request", map[string]interface{}{
		"project_id":       projectID,
		"name":             createRequest.Metadata.Name,
		"location":         createRequest.Metadata.Location.Value,
		"concurrent_users": fmt.Sprintf("%v", createRequest.Properties.ConcurrentUsers),
		"full_request":     fmt.Sprintf("%+v", createRequest),
	})

	response, err := registryClient.Create(ctx, projectID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating container registry",
			fmt.Sprintf("Unable to create container registry: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create container registry"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}

		// Log detailed error information for debugging
		errorDetails := map[string]interface{}{
			"project_id":  projectID,
			"status_code": response.StatusCode,
		}
		if response.Error.Title != nil {
			errorDetails["error_title"] = *response.Error.Title
		}
		if response.Error.Detail != nil {
			errorDetails["error_detail"] = *response.Error.Detail
		}
		if response.Error.Status != nil {
			errorDetails["error_status"] = *response.Error.Status
		}
		if response.Error.Type != nil {
			errorDetails["error_type"] = *response.Error.Type
		}

		// Log full request and error response JSON for debugging
		if requestJSON, jsonErr := json.MarshalIndent(createRequest, "", "  "); jsonErr == nil {
			tflog.Debug(ctx, "Full container registry create request JSON (error case)", map[string]interface{}{
				"request_json": string(requestJSON),
			})
		}
		if errorJSON, jsonErr := json.MarshalIndent(response.Error, "", "  "); jsonErr == nil {
			tflog.Debug(ctx, "Full API error response JSON", map[string]interface{}{
				"error_json": string(errorJSON),
			})
		}

		tflog.Error(ctx, "Container registry create request failed", errorDetails)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Container registry created but no data returned from API",
		)
		return
	}

	// Wait for Container Registry to be active before returning
	// This ensures Terraform doesn't proceed until Container Registry is ready
	registryID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := registryClient.Get(ctx, projectID, registryID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for Container Registry to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "ContainerRegistry", registryID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Container Registry Not Active",
			fmt.Sprintf("Container registry was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a Container Registry resource", map[string]interface{}{
		"containerregistry_id":   data.Id.ValueString(),
		"containerregistry_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContainerRegistryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ContainerRegistryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	registryID := data.Id.ValueString()

	if projectID == "" || registryID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Registry ID are required to read the container registry",
		)
		return
	}

	containerClient := r.client.Client.FromContainer()
	if containerClient == nil {
		resp.Diagnostics.AddError(
			"Container Client Not Available",
			"Container client is not available",
		)
		return
	}

	registryClient := containerClient.ContainerRegistry()
	if registryClient == nil {
		resp.Diagnostics.AddError(
			"Container Registry Client Not Available",
			"Container Registry client is not available",
		)
		return
	}

	// Get container registry details using the SDK
	response, err := registryClient.Get(ctx, projectID, registryID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading container registry",
			fmt.Sprintf("Unable to read container registry: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read container registry"
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
		registry := response.Data

		if registry.Metadata.ID != nil {
			data.Id = types.StringValue(*registry.Metadata.ID)
		}
		if registry.Metadata.URI != nil {
			data.Uri = types.StringValue(*registry.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if registry.Metadata.Name != nil {
			data.Name = types.StringValue(*registry.Metadata.Name)
		}
		if registry.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(registry.Metadata.LocationResponse.Value)
		}

		// Network and storage are preserved from state (they're immutable and not fully returned by API)
		// The network and storage objects are already set from req.State.Get above

		if registry.Properties.BillingPlan.BillingPeriod != "" {
			data.BillingPeriod = types.StringValue(registry.Properties.BillingPlan.BillingPeriod)
		}

		// Parse settings from API (concurrent_users_flavor from ConcurrentUsers field, admin_user from AdminUser field)
		var concurrentUsersFlavorValue types.String
		var adminUserValue types.String

		if registry.Properties.ConcurrentUsers != nil {
			concurrentUsersFlavorValue = types.StringValue(*registry.Properties.ConcurrentUsers)
		} else {
			concurrentUsersFlavorValue = types.StringNull()
		}

		if registry.Properties.AdminUser.Username != "" {
			adminUserValue = types.StringValue(registry.Properties.AdminUser.Username)
		} else {
			adminUserValue = types.StringNull()
		}

		// Create settings object
		settingsAttrs := map[string]attr.Value{
			"concurrent_users_flavor": concurrentUsersFlavorValue,
			"admin_user":              adminUserValue,
		}
		settingsObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"concurrent_users_flavor": types.StringType,
				"admin_user":              types.StringType,
			},
			settingsAttrs,
		)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Settings = settingsObj
		}

		// Update tags
		if len(registry.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(registry.Metadata.Tags))
			for i, tag := range registry.Metadata.Tags {
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

func (r *ContainerRegistryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ContainerRegistryResourceModel
	var state ContainerRegistryResourceModel

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
	registryID := state.Id.ValueString()

	if projectID == "" || registryID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Registry ID are required to update the container registry",
		)
		return
	}

	containerClient := r.client.Client.FromContainer()
	if containerClient == nil {
		resp.Diagnostics.AddError(
			"Container Client Not Available",
			"Container client is not available",
		)
		return
	}

	registryClient := containerClient.ContainerRegistry()
	if registryClient == nil {
		resp.Diagnostics.AddError(
			"Container Registry Client Not Available",
			"Container Registry client is not available",
		)
		return
	}

	// Get current container registry details
	getResponse, err := registryClient.Get(ctx, projectID, registryID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current container registry",
			fmt.Sprintf("Unable to get current container registry: %s", err),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"Container Registry Not Found",
			"Container registry not found or no data returned",
		)
		return
	}

	current := getResponse.Data

	// Get region value
	regionValue := ""
	if current.Metadata.LocationResponse != nil {
		regionValue = current.Metadata.LocationResponse.Value
	}
	if regionValue == "" {
		resp.Diagnostics.AddError(
			"Missing Region",
			"Unable to determine region value for container registry",
		)
		return
	}

	// Extract tags
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

	// Extract network configuration from plan or state (network is immutable)
	var publicIPURI, vpcURI, subnetURI, securityGroupURI string
	if !data.Network.IsNull() && !data.Network.IsUnknown() {
		networkObj, diags := data.Network.ToObjectValue(ctx)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			networkAttrs := networkObj.Attributes()
			if publicIpAttr, ok := networkAttrs["public_ip_uri_ref"]; ok && publicIpAttr != nil {
				if publicIpStr, ok := publicIpAttr.(types.String); ok && !publicIpStr.IsNull() && !publicIpStr.IsUnknown() {
					publicIPURI = publicIpStr.ValueString()
				}
			}
			if vpcAttr, ok := networkAttrs["vpc_uri_ref"]; ok && vpcAttr != nil {
				if vpcStr, ok := vpcAttr.(types.String); ok && !vpcStr.IsNull() && !vpcStr.IsUnknown() {
					vpcURI = vpcStr.ValueString()
				}
			}
			if subnetAttr, ok := networkAttrs["subnet_uri_ref"]; ok && subnetAttr != nil {
				if subnetStr, ok := subnetAttr.(types.String); ok && !subnetStr.IsNull() && !subnetStr.IsUnknown() {
					subnetURI = subnetStr.ValueString()
				}
			}
			if securityGroupAttr, ok := networkAttrs["security_group_uri_ref"]; ok && securityGroupAttr != nil {
				if securityGroupStr, ok := securityGroupAttr.(types.String); ok && !securityGroupStr.IsNull() && !securityGroupStr.IsUnknown() {
					securityGroupURI = securityGroupStr.ValueString()
				}
			}
		}
	}
	// Fallback to current state if not in plan
	if publicIPURI == "" {
		publicIPURI = current.Properties.PublicIp.URI
	}
	if vpcURI == "" {
		vpcURI = current.Properties.VPC.URI
	}
	if subnetURI == "" {
		subnetURI = current.Properties.Subnet.URI
	}
	if securityGroupURI == "" {
		securityGroupURI = current.Properties.SecurityGroup.URI
	}

	// Extract storage configuration from plan or state (storage is immutable)
	var blockStorageURI string
	if !data.Storage.IsNull() && !data.Storage.IsUnknown() {
		storageObj, diags := data.Storage.ToObjectValue(ctx)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			storageAttrs := storageObj.Attributes()
			if blockStorageAttr, ok := storageAttrs["block_storage_uri_ref"]; ok && blockStorageAttr != nil {
				if blockStorageStr, ok := blockStorageAttr.(types.String); ok && !blockStorageStr.IsNull() && !blockStorageStr.IsUnknown() {
					blockStorageURI = blockStorageStr.ValueString()
				}
			}
		}
	}
	// Fallback to current state if not in plan
	if blockStorageURI == "" {
		blockStorageURI = current.Properties.BlockStorage.URI
	}

	// Build update request - use ContainerRegistryRequest for updates (same as create)
	updateRequest := sdktypes.ContainerRegistryRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.ContainerRegistryPropertiesRequest{
			PublicIp: sdktypes.ReferenceResource{
				URI: publicIPURI,
			},
			VPC: sdktypes.ReferenceResource{
				URI: vpcURI,
			},
			Subnet: sdktypes.ReferenceResource{
				URI: subnetURI,
			},
			SecurityGroup: sdktypes.ReferenceResource{
				URI: securityGroupURI,
			},
			BlockStorage: sdktypes.ReferenceResource{
				URI: blockStorageURI,
			},
		},
	}

	// Extract settings configuration (concurrent_users_flavor and admin_user)
	var concurrentUsersFlavor string
	var adminUser string

	if !data.Settings.IsNull() && !data.Settings.IsUnknown() {
		settingsObj, diags := data.Settings.ToObjectValue(ctx)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			settingsAttrs := settingsObj.Attributes()

			if concurrentUsersAttr, ok := settingsAttrs["concurrent_users_flavor"]; ok && concurrentUsersAttr != nil {
				if cuStr, ok := concurrentUsersAttr.(types.String); ok && !cuStr.IsNull() && !cuStr.IsUnknown() {
					concurrentUsersFlavor = cuStr.ValueString()
				}
			}

			if adminUserAttr, ok := settingsAttrs["admin_user"]; ok && adminUserAttr != nil {
				if auStr, ok := adminUserAttr.(types.String); ok && !auStr.IsNull() && !auStr.IsUnknown() {
					adminUser = auStr.ValueString()
				}
			}
		}
	}

	// Update billing period if provided, otherwise preserve current
	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		updateRequest.Properties.BillingPlan = &sdktypes.BillingPeriodResource{
			BillingPeriod: data.BillingPeriod.ValueString(),
		}
	} else if current.Properties.BillingPlan.BillingPeriod != "" {
		updateRequest.Properties.BillingPlan = &sdktypes.BillingPeriodResource{
			BillingPeriod: current.Properties.BillingPlan.BillingPeriod,
		}
	}

	// Update concurrent_users_flavor if provided (maps to 'size' in API)
	// Note: SDK uses ConcurrentUsers field which maps to 'size' JSON field
	if concurrentUsersFlavor != "" {
		updateRequest.Properties.ConcurrentUsers = &concurrentUsersFlavor
	} else if current.Properties.ConcurrentUsers != nil {
		updateRequest.Properties.ConcurrentUsers = current.Properties.ConcurrentUsers
	}

	// Update admin user if provided, otherwise preserve current
	if adminUser != "" {
		updateRequest.Properties.AdminUser = &sdktypes.UserCredential{
			Username: adminUser,
		}
	} else if current.Properties.AdminUser.Username != "" {
		updateRequest.Properties.AdminUser = &sdktypes.UserCredential{
			Username: current.Properties.AdminUser.Username,
		}
	}

	// Update the container registry using the SDK
	response, err := registryClient.Update(ctx, projectID, registryID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating container registry",
			fmt.Sprintf("Unable to update container registry: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to update container registry"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.Uri = state.Uri // Preserve URI from state
	data.ProjectID = state.ProjectID
	// Preserve network and storage from state (they're immutable)
	data.Network = state.Network
	data.Storage = state.Storage

	if response != nil && response.Data != nil {
		// Update from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		}
	} else {
		// If no response, re-read the container registry to get the latest state including URI
		getResp, err := registryClient.Get(ctx, projectID, registryID, nil)
		if err == nil && getResp != nil && getResp.Data != nil {
			if getResp.Data.Metadata.URI != nil {
				data.Uri = types.StringValue(*getResp.Data.Metadata.URI)
			} else {
				data.Uri = state.Uri // Fallback to state if not in response
			}
		} else {
			// If re-read fails, preserve from state
			data.Uri = state.Uri
		}
	}

	// Re-read to get the latest state and update all fields
	getResp, err := registryClient.Get(ctx, projectID, registryID, nil)
	if err == nil && getResp != nil && getResp.Data != nil {
		registry := getResp.Data
		// Update URI if available
		if registry.Metadata.URI != nil {
			data.Uri = types.StringValue(*registry.Metadata.URI)
		} else {
			data.Uri = state.Uri // Fallback to state if not available
		}
		// Network and storage are preserved from state (they're immutable)
		// No need to update them from API response
		// Update other fields from re-read to ensure consistency
		if registry.Metadata.Name != nil {
			data.Name = types.StringValue(*registry.Metadata.Name)
		}
		if registry.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(registry.Metadata.LocationResponse.Value)
		}
		if registry.Properties.BillingPlan != nil && registry.Properties.BillingPlan.BillingPeriod != "" {
			data.BillingPeriod = types.StringValue(registry.Properties.BillingPlan.BillingPeriod)
		} else {
			data.BillingPeriod = types.StringNull()
		}

		// Parse settings from API (concurrent_users from ConcurrentUsers field, admin_user from AdminUser field)
		var concurrentUsersValue types.Int64
		var adminUserValue types.String

		if registry.Properties.ConcurrentUsers != nil {
			var cu int64
			if _, err := fmt.Sscanf(*registry.Properties.ConcurrentUsers, "%d", &cu); err == nil {
				concurrentUsersValue = types.Int64Value(cu)
			} else {
				concurrentUsersValue = types.Int64Null()
			}
		} else {
			concurrentUsersValue = types.Int64Null()
		}

		if registry.Properties.AdminUser != nil && registry.Properties.AdminUser.Username != "" {
			adminUserValue = types.StringValue(registry.Properties.AdminUser.Username)
		} else {
			adminUserValue = types.StringNull()
		}

		// Create settings object
		settingsAttrs := map[string]attr.Value{
			"concurrent_users": concurrentUsersValue,
			"admin_user":       adminUserValue,
		}
		settingsObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"concurrent_users": types.Int64Type,
				"admin_user":       types.StringType,
			},
			settingsAttrs,
		)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Settings = settingsObj
		}

		// Update tags from re-read
		if len(registry.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(registry.Metadata.Tags))
			for i, tag := range registry.Metadata.Tags {
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
		// If re-read fails, preserve immutable fields from state
		data.Uri = state.Uri
		data.Network = state.Network
		data.Storage = state.Storage
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContainerRegistryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ContainerRegistryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	registryID := data.Id.ValueString()

	if projectID == "" || registryID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Registry ID are required to delete the container registry",
		)
		return
	}

	containerClient := r.client.Client.FromContainer()
	if containerClient == nil {
		resp.Diagnostics.AddError(
			"Container Client Not Available",
			"Container client is not available",
		)
		return
	}

	registryClient := containerClient.ContainerRegistry()
	if registryClient == nil {
		resp.Diagnostics.AddError(
			"Container Registry Client Not Available",
			"Container Registry client is not available",
		)
		return
	}

	// Delete the container registry using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return registryClient.Delete(ctx, projectID, registryID, nil)
		},
		ExtractSDKError,
		"ContainerRegistry",
		registryID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting container registry",
			fmt.Sprintf("Unable to delete container registry: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a Container Registry resource", map[string]interface{}{
		"containerregistry_id": registryID,
	})
}

func (r *ContainerRegistryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
