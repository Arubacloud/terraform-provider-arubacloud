package provider

import (
	"context"
	"fmt"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ContainerRegistryResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	ProjectID     types.String `tfsdk:"project_id"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	Network       types.Object `tfsdk:"network"`
	Storage       types.Object `tfsdk:"storage"`
	Settings      types.Object `tfsdk:"settings"`
}

type ContainerRegistryNetworkModel struct {
	PublicIpUriRef      types.String `tfsdk:"public_ip_uri_ref"`
	VpcUriRef           types.String `tfsdk:"vpc_uri_ref"`
	SubnetUriRef        types.String `tfsdk:"subnet_uri_ref"`
	SecurityGroupUriRef types.String `tfsdk:"security_group_uri_ref"`
}

type ContainerRegistryStorageModel struct {
	BlockStorageUriRef types.String `tfsdk:"block_storage_uri_ref"`
}

type ContainerRegistrySettingsModel struct {
	AdminUser             types.String `tfsdk:"admin_user"`
	ConcurrentUsersFlavor types.String `tfsdk:"concurrent_users_flavor"`
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
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period (Hour, Month, Year)",
				Optional:            true,
			},
			"network": schema.SingleNestedAttribute{
				MarkdownDescription: "Network configuration for the container registry",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"public_ip_uri_ref": schema.StringAttribute{
						MarkdownDescription: "Public IP URI reference (e.g., arubacloud_elasticip.example.uri)",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"vpc_uri_ref": schema.StringAttribute{
						MarkdownDescription: "VPC URI reference (e.g., arubacloud_vpc.example.uri)",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"subnet_uri_ref": schema.StringAttribute{
						MarkdownDescription: "Subnet URI reference (e.g., arubacloud_subnet.example.uri)",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"security_group_uri_ref": schema.StringAttribute{
						MarkdownDescription: "Security Group URI reference (e.g., arubacloud_securitygroup.example.uri)",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"storage": schema.SingleNestedAttribute{
				MarkdownDescription: "Storage configuration for the container registry",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"block_storage_uri_ref": schema.StringAttribute{
						MarkdownDescription: "Block Storage URI reference (e.g., arubacloud_blockstorage.example.uri)",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Container registry settings",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"admin_user": schema.StringAttribute{
						MarkdownDescription: "Administrator username",
						Optional:            true,
					},
					"concurrent_users_flavor": schema.StringAttribute{
						MarkdownDescription: "Concurrent users flavor size. Must be one of: Small, Medium, HighPerf",
						Optional:            true,
					},
				},
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

	// Extract network configuration
	var networkModel ContainerRegistryNetworkModel
	diags := data.Network.As(ctx, &networkModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract storage configuration
	var storageModel ContainerRegistryStorageModel
	diags = data.Storage.As(ctx, &storageModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
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
				URI: networkModel.PublicIpUriRef.ValueString(),
			},
			VPC: sdktypes.ReferenceResource{
				URI: networkModel.VpcUriRef.ValueString(),
			},
			Subnet: sdktypes.ReferenceResource{
				URI: networkModel.SubnetUriRef.ValueString(),
			},
			SecurityGroup: sdktypes.ReferenceResource{
				URI: networkModel.SecurityGroupUriRef.ValueString(),
			},
			BlockStorage: sdktypes.ReferenceResource{
				URI: storageModel.BlockStorageUriRef.ValueString(),
			},
		},
	}

	// Add optional fields
	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		createRequest.Properties.BillingPlan = &sdktypes.BillingPeriodResource{
			BillingPeriod: data.BillingPeriod.ValueString(),
		}
	}

	// Extract settings if provided
	if !data.Settings.IsNull() && !data.Settings.IsUnknown() {
		var settingsModel ContainerRegistrySettingsModel
		diags = data.Settings.As(ctx, &settingsModel, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !settingsModel.AdminUser.IsNull() && !settingsModel.AdminUser.IsUnknown() {
			createRequest.Properties.AdminUser = &sdktypes.UserCredential{
				Username: settingsModel.AdminUser.ValueString(),
			}
		}

		if !settingsModel.ConcurrentUsersFlavor.IsNull() && !settingsModel.ConcurrentUsersFlavor.IsUnknown() {
			concurrentUsersFlavor := settingsModel.ConcurrentUsersFlavor.ValueString()
			createRequest.Properties.ConcurrentUsers = &concurrentUsersFlavor
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

	response, err := registryClient.Create(ctx, projectID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating container registry",
			fmt.Sprintf("Unable to create container registry: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		logContext := map[string]interface{}{
			"project_id": projectID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to create container registry", logContext)
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
		logContext := map[string]interface{}{
			"project_id":  projectID,
			"registry_id": registryID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to read container registry", logContext)
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

		// Build network object
		networkAttrs := map[string]attr.Value{
			"public_ip_uri_ref":      types.StringValue(registry.Properties.PublicIp.URI),
			"vpc_uri_ref":            types.StringValue(registry.Properties.VPC.URI),
			"subnet_uri_ref":         types.StringValue(registry.Properties.Subnet.URI),
			"security_group_uri_ref": types.StringValue(registry.Properties.SecurityGroup.URI),
		}
		networkObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"public_ip_uri_ref":      types.StringType,
				"vpc_uri_ref":            types.StringType,
				"subnet_uri_ref":         types.StringType,
				"security_group_uri_ref": types.StringType,
			},
			networkAttrs,
		)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Network = networkObj
		}

		// Build storage object
		storageAttrs := map[string]attr.Value{
			"block_storage_uri_ref": types.StringValue(registry.Properties.BlockStorage.URI),
		}
		storageObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"block_storage_uri_ref": types.StringType,
			},
			storageAttrs,
		)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Storage = storageObj
		}

		// Build settings object if there are any settings
		var settingsAdminUser types.String
		var settingsConcurrentUsersFlavor types.String

		if registry.Properties.AdminUser.Username != "" {
			settingsAdminUser = types.StringValue(registry.Properties.AdminUser.Username)
		} else {
			settingsAdminUser = types.StringNull()
		}

		if registry.Properties.ConcurrentUsers != nil && *registry.Properties.ConcurrentUsers != "" {
			settingsConcurrentUsersFlavor = types.StringValue(*registry.Properties.ConcurrentUsers)
		} else {
			settingsConcurrentUsersFlavor = types.StringNull()
		}

		settingsAttrs := map[string]attr.Value{
			"admin_user":              settingsAdminUser,
			"concurrent_users_flavor": settingsConcurrentUsersFlavor,
		}
		settingsObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"admin_user":              types.StringType,
				"concurrent_users_flavor": types.StringType,
			},
			settingsAttrs,
		)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Settings = settingsObj
		}

		if registry.Properties.BillingPlan.BillingPeriod != "" {
			data.BillingPeriod = types.StringValue(registry.Properties.BillingPlan.BillingPeriod)
		} else {
			data.BillingPeriod = types.StringNull()
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

	// Extract network configuration from plan
	var networkModel ContainerRegistryNetworkModel
	var diags diag.Diagnostics
	diags = data.Network.As(ctx, &networkModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract storage configuration from plan
	var storageModel ContainerRegistryStorageModel
	diags = data.Storage.As(ctx, &storageModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
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
			// Use network configuration from plan
			PublicIp: sdktypes.ReferenceResource{
				URI: networkModel.PublicIpUriRef.ValueString(),
			},
			VPC: sdktypes.ReferenceResource{
				URI: networkModel.VpcUriRef.ValueString(),
			},
			Subnet: sdktypes.ReferenceResource{
				URI: networkModel.SubnetUriRef.ValueString(),
			},
			SecurityGroup: sdktypes.ReferenceResource{
				URI: networkModel.SecurityGroupUriRef.ValueString(),
			},
			BlockStorage: sdktypes.ReferenceResource{
				URI: storageModel.BlockStorageUriRef.ValueString(),
			},
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

	// Extract and update settings if provided
	if !data.Settings.IsNull() && !data.Settings.IsUnknown() {
		var settingsModel ContainerRegistrySettingsModel
		diags = data.Settings.As(ctx, &settingsModel, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !settingsModel.AdminUser.IsNull() && !settingsModel.AdminUser.IsUnknown() {
			updateRequest.Properties.AdminUser = &sdktypes.UserCredential{
				Username: settingsModel.AdminUser.ValueString(),
			}
		} else if current.Properties.AdminUser.Username != "" {
			updateRequest.Properties.AdminUser = &sdktypes.UserCredential{
				Username: current.Properties.AdminUser.Username,
			}
		}

		if !settingsModel.ConcurrentUsersFlavor.IsNull() && !settingsModel.ConcurrentUsersFlavor.IsUnknown() {
			concurrentUsersFlavor := settingsModel.ConcurrentUsersFlavor.ValueString()
			updateRequest.Properties.ConcurrentUsers = &concurrentUsersFlavor
		} else if current.Properties.ConcurrentUsers != nil {
			updateRequest.Properties.ConcurrentUsers = current.Properties.ConcurrentUsers
		}
	} else {
		// Preserve current settings if not in plan
		if current.Properties.AdminUser.Username != "" {
			updateRequest.Properties.AdminUser = &sdktypes.UserCredential{
				Username: current.Properties.AdminUser.Username,
			}
		}
		if current.Properties.ConcurrentUsers != nil {
			updateRequest.Properties.ConcurrentUsers = current.Properties.ConcurrentUsers
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
		logContext := map[string]interface{}{
			"project_id":  projectID,
			"registry_id": registryID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to update container registry", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.Uri = state.Uri // Preserve URI from state
	data.ProjectID = state.ProjectID
	data.Network = state.Network
	data.Storage = state.Storage
	data.Settings = state.Settings

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

		// Build network object from response
		networkAttrs := map[string]attr.Value{
			"public_ip_uri_ref":      types.StringValue(registry.Properties.PublicIp.URI),
			"vpc_uri_ref":            types.StringValue(registry.Properties.VPC.URI),
			"subnet_uri_ref":         types.StringValue(registry.Properties.Subnet.URI),
			"security_group_uri_ref": types.StringValue(registry.Properties.SecurityGroup.URI),
		}
		networkObj, diagsNetwork := types.ObjectValue(
			map[string]attr.Type{
				"public_ip_uri_ref":      types.StringType,
				"vpc_uri_ref":            types.StringType,
				"subnet_uri_ref":         types.StringType,
				"security_group_uri_ref": types.StringType,
			},
			networkAttrs,
		)
		resp.Diagnostics.Append(diagsNetwork...)
		if !resp.Diagnostics.HasError() {
			data.Network = networkObj
		}

		// Build storage object from response
		storageAttrs := map[string]attr.Value{
			"block_storage_uri_ref": types.StringValue(registry.Properties.BlockStorage.URI),
		}
		storageObj, diagsStorage := types.ObjectValue(
			map[string]attr.Type{
				"block_storage_uri_ref": types.StringType,
			},
			storageAttrs,
		)
		resp.Diagnostics.Append(diagsStorage...)
		if !resp.Diagnostics.HasError() {
			data.Storage = storageObj
		}

		// Build settings object from response
		var settingsAdminUser types.String
		var settingsConcurrentUsersFlavor types.String

		if registry.Properties.AdminUser != nil && registry.Properties.AdminUser.Username != "" {
			settingsAdminUser = types.StringValue(registry.Properties.AdminUser.Username)
		} else {
			settingsAdminUser = types.StringNull()
		}

		if registry.Properties.ConcurrentUsers != nil && *registry.Properties.ConcurrentUsers != "" {
			settingsConcurrentUsersFlavor = types.StringValue(*registry.Properties.ConcurrentUsers)
		} else {
			settingsConcurrentUsersFlavor = types.StringNull()
		}

		settingsAttrs := map[string]attr.Value{
			"admin_user":              settingsAdminUser,
			"concurrent_users_flavor": settingsConcurrentUsersFlavor,
		}
		settingsObj, diagsSettings := types.ObjectValue(
			map[string]attr.Type{
				"admin_user":              types.StringType,
				"concurrent_users_flavor": types.StringType,
			},
			settingsAttrs,
		)
		resp.Diagnostics.Append(diagsSettings...)
		if !resp.Diagnostics.HasError() {
			data.Settings = settingsObj
		}

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

		// Update tags from re-read
		if len(registry.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(registry.Metadata.Tags))
			for i, tag := range registry.Metadata.Tags {
				tagValues[i] = types.StringValue(tag)
			}
			tagsList, diagsTags := types.ListValueFrom(ctx, types.StringType, tagValues)
			resp.Diagnostics.Append(diagsTags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = tagsList
			}
		} else {
			emptyList, diagsTags := types.ListValue(types.StringType, []attr.Value{})
			resp.Diagnostics.Append(diagsTags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = emptyList
			}
		}
	} else {
		// If re-read fails, preserve immutable fields from state
		data.Uri = state.Uri
		data.Network = state.Network
		data.Storage = state.Storage
		data.Settings = state.Settings
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
