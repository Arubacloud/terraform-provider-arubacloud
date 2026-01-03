// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ContainerRegistryResourceModel struct {
	Id              types.String `tfsdk:"id"`
	Uri             types.String `tfsdk:"uri"`
	Name            types.String `tfsdk:"name"`
	Location        types.String `tfsdk:"location"`
	Tags            types.List   `tfsdk:"tags"`
	ProjectID       types.String `tfsdk:"project_id"`
	ElasticIPID     types.String `tfsdk:"elasticip_id"`
	VpcID           types.String `tfsdk:"vpc_id"`
	SubnetID        types.String `tfsdk:"subnet_id"`
	SecurityGroupID types.String `tfsdk:"security_group_id"`
	BlockStorageID  types.String `tfsdk:"block_storage_id"`
	BillingPeriod   types.String `tfsdk:"billing_period"`
	AdminUser       types.String `tfsdk:"admin_user"`
	ConcurrentUsers types.String `tfsdk:"concurrent_users"`
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
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Container Registry URI",
				Computed:            true,
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
			"elasticip_id": schema.StringAttribute{
				MarkdownDescription: "Elastic IP ID or URI",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "VPC ID or URI",
				Required:            true,
			},
			"subnet_id": schema.StringAttribute{
				MarkdownDescription: "Subnet ID or URI",
				Required:            true,
			},
			"security_group_id": schema.StringAttribute{
				MarkdownDescription: "Security Group ID or URI",
				Required:            true,
			},
			"block_storage_id": schema.StringAttribute{
				MarkdownDescription: "Block Storage ID or URI",
				Required:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period (Hour, Month, Year)",
				Optional:            true,
			},
			"admin_user": schema.StringAttribute{
				MarkdownDescription: "Administrator username",
				Optional:            true,
			},
			"concurrent_users": schema.StringAttribute{
				MarkdownDescription: "Number of concurrent users",
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

	// Build URIs for referenced resources
	publicIPURI := data.ElasticIPID.ValueString()
	if !strings.HasPrefix(publicIPURI, "/") {
		publicIPURI = fmt.Sprintf("/projects/%s/providers/Aruba.Network/elasticIps/%s", projectID, publicIPURI)
	}

	vpcURI := data.VpcID.ValueString()
	if !strings.HasPrefix(vpcURI, "/") {
		vpcURI = fmt.Sprintf("/projects/%s/providers/Aruba.Network/vpcs/%s", projectID, vpcURI)
	}

	subnetURI := data.SubnetID.ValueString()
	if !strings.HasPrefix(subnetURI, "/") {
		// Extract VPC ID from vpcURI
		parts := strings.Split(vpcURI, "/")
		if len(parts) > 0 {
			vpcIDFromURI := parts[len(parts)-1]
			subnetURI = fmt.Sprintf("/projects/%s/providers/Aruba.Network/vpcs/%s/subnets/%s", projectID, vpcIDFromURI, subnetURI)
		}
	}

	securityGroupURI := data.SecurityGroupID.ValueString()
	if !strings.HasPrefix(securityGroupURI, "/") {
		parts := strings.Split(vpcURI, "/")
		if len(parts) > 0 {
			vpcIDFromURI := parts[len(parts)-1]
			securityGroupURI = fmt.Sprintf("/projects/%s/providers/Aruba.Network/vpcs/%s/securityGroups/%s", projectID, vpcIDFromURI, securityGroupURI)
		}
	}

	blockStorageURI := data.BlockStorageID.ValueString()
	if !strings.HasPrefix(blockStorageURI, "/") {
		blockStorageURI = fmt.Sprintf("/projects/%s/providers/Aruba.Storage/blockStorages/%s", projectID, blockStorageURI)
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

	// Add optional fields
	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		createRequest.Properties.BillingPlan = &sdktypes.BillingPeriodResource{
			BillingPeriod: data.BillingPeriod.ValueString(),
		}
	}

	if !data.AdminUser.IsNull() && !data.AdminUser.IsUnknown() {
		createRequest.Properties.AdminUser = &sdktypes.UserCredential{
			Username: data.AdminUser.ValueString(),
		}
	}

	if !data.ConcurrentUsers.IsNull() && !data.ConcurrentUsers.IsUnknown() {
		concurrentUsers := data.ConcurrentUsers.ValueString()
		createRequest.Properties.ConcurrentUsers = &concurrentUsers
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
		"containerregistry_id": data.Id.ValueString(),
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
		if registry.Properties.PublicIp.URI != "" {
			// Extract Elastic IP ID from URI
			parts := strings.Split(registry.Properties.PublicIp.URI, "/")
			if len(parts) > 0 {
				data.ElasticIPID = types.StringValue(parts[len(parts)-1])
			}
		}
		if registry.Properties.VPC.URI != "" {
			parts := strings.Split(registry.Properties.VPC.URI, "/")
			if len(parts) > 0 {
				data.VpcID = types.StringValue(parts[len(parts)-1])
			}
		}
		if registry.Properties.Subnet.URI != "" {
			parts := strings.Split(registry.Properties.Subnet.URI, "/")
			if len(parts) > 0 {
				data.SubnetID = types.StringValue(parts[len(parts)-1])
			}
		}
		if registry.Properties.SecurityGroup.URI != "" {
			parts := strings.Split(registry.Properties.SecurityGroup.URI, "/")
			if len(parts) > 0 {
				data.SecurityGroupID = types.StringValue(parts[len(parts)-1])
			}
		}
		if registry.Properties.BlockStorage.URI != "" {
			parts := strings.Split(registry.Properties.BlockStorage.URI, "/")
			if len(parts) > 0 {
				data.BlockStorageID = types.StringValue(parts[len(parts)-1])
			}
		}
		if registry.Properties.BillingPlan.BillingPeriod != "" {
			data.BillingPeriod = types.StringValue(registry.Properties.BillingPlan.BillingPeriod)
		}
		if registry.Properties.AdminUser.Username != "" {
			data.AdminUser = types.StringValue(registry.Properties.AdminUser.Username)
		}
		if registry.Properties.ConcurrentUsers != nil {
			data.ConcurrentUsers = types.StringValue(*registry.Properties.ConcurrentUsers)
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
			// Preserve existing required properties
			PublicIp:      current.Properties.PublicIp,
			VPC:           current.Properties.VPC,
			Subnet:        current.Properties.Subnet,
			SecurityGroup: current.Properties.SecurityGroup,
			BlockStorage:  current.Properties.BlockStorage,
		},
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

	// Update concurrent users if provided, otherwise preserve current
	if !data.ConcurrentUsers.IsNull() && !data.ConcurrentUsers.IsUnknown() {
		concurrentUsers := data.ConcurrentUsers.ValueString()
		updateRequest.Properties.ConcurrentUsers = &concurrentUsers
	} else if current.Properties.ConcurrentUsers != nil {
		updateRequest.Properties.ConcurrentUsers = current.Properties.ConcurrentUsers
	}

	// Preserve admin user if it exists
	if current.Properties.AdminUser.Username != "" {
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
	data.ProjectID = state.ProjectID

	if response != nil && response.Data != nil {
		// Update from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
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
