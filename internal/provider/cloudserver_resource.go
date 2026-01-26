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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type CloudServerResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	ProjectID types.String `tfsdk:"project_id"`
	Zone      types.String `tfsdk:"zone"`
	Tags      types.List   `tfsdk:"tags"`
	Network   types.Object `tfsdk:"network"`
	Settings  types.Object `tfsdk:"settings"`
	Storage   types.Object `tfsdk:"storage"`
}

type CloudServerNetworkModel struct {
	VpcUriRef            types.String `tfsdk:"vpc_uri_ref"`
	ElasticIpUriRef      types.String `tfsdk:"elastic_ip_uri_ref"`
	SubnetUriRefs        types.List   `tfsdk:"subnet_uri_refs"`
	SecurityGroupUriRefs types.List   `tfsdk:"securitygroup_uri_refs"`
}

type CloudServerSettingsModel struct {
	FlavorName    types.String `tfsdk:"flavor_name"`
	KeyPairUriRef types.String `tfsdk:"key_pair_uri_ref"`
	UserData      types.String `tfsdk:"user_data"`
}

type CloudServerStorageModel struct {
	BootVolumeUriRef types.String `tfsdk:"boot_volume_uri_ref"`
}

type CloudServerResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &CloudServerResource{}
var _ resource.ResourceWithImportState = &CloudServerResource{}

func NewCloudServerResource() resource.Resource {
	return &CloudServerResource{}
}

func (r *CloudServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudserver"
}

func (r *CloudServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CloudServer resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "CloudServer identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "CloudServer URI",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "CloudServer name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "CloudServer location",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID",
				Required:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the Cloud Server",
				Optional:            true,
			},
			"network": schema.SingleNestedAttribute{
				MarkdownDescription: "Network configuration for the cloud server",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"vpc_uri_ref": schema.StringAttribute{
						MarkdownDescription: "VPC URI reference (e.g., arubacloud_vpc.example.uri)",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"elastic_ip_uri_ref": schema.StringAttribute{
						MarkdownDescription: "Elastic IP URI reference (e.g., arubacloud_elasticip.example.uri)",
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"subnet_uri_refs": schema.ListAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "List of subnet URI references (e.g., [arubacloud_subnet.example.uri])",
						Required:            true,
					},
					"securitygroup_uri_refs": schema.ListAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "List of security group URI references (e.g., [arubacloud_securitygroup.example.uri])",
						Required:            true,
					},
				},
			},
			"settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Cloud server settings",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"flavor_name": schema.StringAttribute{
						MarkdownDescription: "Flavor name (e.g., CSO4A8 for 4 CPU, 8GB RAM). See https://api.arubacloud.com/docs/metadata/#cloudserver-flavors",
						Required:            true,
					},
					"key_pair_uri_ref": schema.StringAttribute{
						MarkdownDescription: "Key Pair URI reference (e.g., arubacloud_keypair.example.uri)",
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"user_data": schema.StringAttribute{
						MarkdownDescription: "Cloud-Init user data (raw YAML content)",
						Optional:            true,
						Sensitive:           true,
					},
				},
			},
			"storage": schema.SingleNestedAttribute{
				MarkdownDescription: "Storage configuration for the cloud server",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"boot_volume_uri_ref": schema.StringAttribute{
						MarkdownDescription: "Boot volume URI reference (e.g., arubacloud_blockstorage.example.uri)",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
		},
	}
}

func (r *CloudServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CloudServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CloudServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"Project ID is required to create a cloud server",
		)
		return
	}

	// Extract network model
	var networkModel CloudServerNetworkModel
	diags := data.Network.As(ctx, &networkModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract settings model
	var settingsModel CloudServerSettingsModel
	diags = data.Settings.As(ctx, &settingsModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract storage model
	var storageModel CloudServerStorageModel
	diags = data.Storage.As(ctx, &storageModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract subnets from network model
	var subnetRefs []sdktypes.ReferenceResource
	if !networkModel.SubnetUriRefs.IsNull() && !networkModel.SubnetUriRefs.IsUnknown() {
		var subnetURIs []string
		diags := networkModel.SubnetUriRefs.ElementsAs(ctx, &subnetURIs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, subnetURI := range subnetURIs {
			subnetRefs = append(subnetRefs, sdktypes.ReferenceResource{
				URI: subnetURI,
			})
		}
	}

	// Extract security groups from network model
	var securityGroupRefs []sdktypes.ReferenceResource
	if !networkModel.SecurityGroupUriRefs.IsNull() && !networkModel.SecurityGroupUriRefs.IsUnknown() {
		var sgURIs []string
		diags := networkModel.SecurityGroupUriRefs.ElementsAs(ctx, &sgURIs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, sgURI := range sgURIs {
			securityGroupRefs = append(securityGroupRefs, sdktypes.ReferenceResource{
				URI: sgURI,
			})
		}
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
	flavorName := settingsModel.FlavorName.ValueString()
	zone := data.Zone.ValueString()
	createRequest := sdktypes.CloudServerRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.CloudServerPropertiesRequest{
			FlavorName: &flavorName,
			Zone:       zone,
			BootVolume: sdktypes.ReferenceResource{
				URI: storageModel.BootVolumeUriRef.ValueString(),
			},
			VPC: sdktypes.ReferenceResource{
				URI: networkModel.VpcUriRef.ValueString(),
			},
			Subnets:        subnetRefs,
			SecurityGroups: securityGroupRefs,
		},
	}

	// Add keypair if provided
	if !settingsModel.KeyPairUriRef.IsNull() && !settingsModel.KeyPairUriRef.IsUnknown() {
		createRequest.Properties.KeyPair = sdktypes.ReferenceResource{
			URI: settingsModel.KeyPairUriRef.ValueString(),
		}
	}

	// Add elastic IP if provided
	if !networkModel.ElasticIpUriRef.IsNull() && !networkModel.ElasticIpUriRef.IsUnknown() {
		createRequest.Properties.ElastcIP = sdktypes.ReferenceResource{
			URI: networkModel.ElasticIpUriRef.ValueString(),
		}
	}

	// Add user data if provided
	if !settingsModel.UserData.IsNull() && !settingsModel.UserData.IsUnknown() {
		userDataRaw := settingsModel.UserData.ValueString()
		createRequest.Properties.UserData = &userDataRaw
	}

	// Create the cloud server using the SDK
	response, err := r.client.Client.FromCompute().CloudServers().Create(ctx, projectID, createRequest, nil)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating cloud server",
			fmt.Sprintf("Unable to create cloud server: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		logContext := map[string]interface{}{
			"project_id": projectID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to create cloud server", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	// Extract ID and URI from the Create response (SDK v0.1.13+ has ResourceMetadataResponse with ID and URI)
	if response == nil || response.Data == nil {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Cloud server create response is missing data",
		)
		return
	}

	// Get the server ID from the response
	var serverID string
	if response.Data.Metadata.ID != nil {
		serverID = *response.Data.Metadata.ID
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Cloud server create response is missing ID in metadata",
		)
		return
	}

	data.Id = types.StringValue(serverID)

	// Set URI if available
	if response.Data.Metadata.URI != nil {
		data.Uri = types.StringValue(*response.Data.Metadata.URI)
	} else {
		data.Uri = types.StringNull()
	}

	// Wait for Cloud Server to be active before returning
	// This ensures Terraform doesn't proceed until CloudServer is fully ready
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromCompute().CloudServers().Get(ctx, projectID, serverID, nil)
		if err != nil {
			return "", err
		}
		// Check if response indicates an error state
		if getResp != nil && getResp.IsError() {
			if getResp.Error != nil {
				errMsg := "Unknown error"
				if getResp.Error.Title != nil {
					errMsg = *getResp.Error.Title
				}
				if getResp.Error.Detail != nil {
					errMsg = fmt.Sprintf("%s: %s", errMsg, *getResp.Error.Detail)
				}
				return "", fmt.Errorf("cloud server in error state: %s", errMsg)
			}
			return "", fmt.Errorf("cloud server API returned error response")
		}
		// If we can successfully get the resource, check its status
		if getResp != nil && getResp.Data != nil {
			if getResp.Data.Status.State != nil {
				state := *getResp.Data.Status.State
				return state, nil
			}
		}
		// If Status.State is nil, assume it's still creating
		return "InCreation", nil
	}

	// Wait for Cloud Server to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "CloudServer", serverID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Cloud Server Not Active",
			fmt.Sprintf("Cloud server was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	// Re-read the Cloud Server to ensure all fields are properly set
	getResp, err := r.client.Client.FromCompute().CloudServers().Get(ctx, projectID, serverID, nil)
	if err == nil && getResp != nil && getResp.Data != nil {
		// Update with values from Get response
		if getResp.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*getResp.Data.Metadata.ID)
		}
		if getResp.Data.Metadata.Name != nil {
			data.Name = types.StringValue(*getResp.Data.Metadata.Name)
		}
		if getResp.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*getResp.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
	} else if err != nil {
		// If Get fails, log but don't fail - we already have the ID from create response
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh Cloud Server after creation: %v", err))
	}

	tflog.Trace(ctx, "created a CloudServer resource", map[string]interface{}{
		"cloudserver_id":   data.Id.ValueString(),
		"cloudserver_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CloudServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CloudServerResourceModel
	var originalState CloudServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &originalState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := originalState.ProjectID.ValueString()
	serverID := originalState.Id.ValueString()

	if projectID == "" || serverID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Server ID are required to read the cloud server",
		)
		return
	}

	// Get cloud server details using the SDK
	response, err := r.client.Client.FromCompute().CloudServers().Get(ctx, projectID, serverID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading cloud server",
			fmt.Sprintf("Unable to read cloud server: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		// If cloud server not found, mark as removed
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		logContext := map[string]interface{}{
			"project_id": projectID,
			"server_id":  serverID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to read cloud server", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response == nil || response.Data == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	server := response.Data

	// Extract original nested models to preserve fields not returned by API
	var originalNetworkModel CloudServerNetworkModel
	diags := originalState.Network.As(ctx, &originalNetworkModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var originalSettingsModel CloudServerSettingsModel
	diags = originalState.Settings.As(ctx, &originalSettingsModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var originalStorageModel CloudServerStorageModel
	diags = originalState.Storage.As(ctx, &originalStorageModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update basic fields from API response
	if server.Metadata.ID != nil {
		data.Id = types.StringValue(*server.Metadata.ID)
	}

	if server.Metadata.URI != nil {
		data.Uri = types.StringValue(*server.Metadata.URI)
	} else {
		data.Uri = types.StringNull()
	}

	if server.Metadata.Name != nil {
		data.Name = types.StringValue(*server.Metadata.Name)
	}

	if server.Metadata.LocationResponse != nil && server.Metadata.LocationResponse.Value != "" {
		data.Location = types.StringValue(server.Metadata.LocationResponse.Value)
	}

	data.ProjectID = originalState.ProjectID
	data.Zone = originalState.Zone

	// Update tags from response
	if len(server.Metadata.Tags) > 0 {
		tagValues := make([]types.String, len(server.Metadata.Tags))
		for i, tag := range server.Metadata.Tags {
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

	// Build network object (preserve fields not returned by API)
	networkAttrs := map[string]attr.Value{
		"vpc_uri_ref":            originalNetworkModel.VpcUriRef,
		"elastic_ip_uri_ref":     originalNetworkModel.ElasticIpUriRef,
		"subnet_uri_refs":        originalNetworkModel.SubnetUriRefs,
		"securitygroup_uri_refs": originalNetworkModel.SecurityGroupUriRefs,
	}
	networkObj, diags := types.ObjectValue(
		map[string]attr.Type{
			"vpc_uri_ref":            types.StringType,
			"elastic_ip_uri_ref":     types.StringType,
			"subnet_uri_refs":        types.ListType{ElemType: types.StringType},
			"securitygroup_uri_refs": types.ListType{ElemType: types.StringType},
		},
		networkAttrs,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Network = networkObj

	// Build settings object
	settingsAttrs := map[string]attr.Value{
		"flavor_name":      types.StringValue(server.Properties.Flavor.Name),
		"key_pair_uri_ref": originalSettingsModel.KeyPairUriRef,
		"user_data":        originalSettingsModel.UserData,
	}
	settingsObj, diags := types.ObjectValue(
		map[string]attr.Type{
			"flavor_name":      types.StringType,
			"key_pair_uri_ref": types.StringType,
			"user_data":        types.StringType,
		},
		settingsAttrs,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Settings = settingsObj

	// Build storage object (preserve from state)
	storageAttrs := map[string]attr.Value{
		"boot_volume_uri_ref": originalStorageModel.BootVolumeUriRef,
	}
	storageObj, diags := types.ObjectValue(
		map[string]attr.Type{
			"boot_volume_uri_ref": types.StringType,
		},
		storageAttrs,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Storage = storageObj

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CloudServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CloudServerResourceModel
	var state CloudServerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read current state to preserve values
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract nested models from plan to preserve for inconsistent result check
	var planNetworkModel CloudServerNetworkModel
	diags := data.Network.As(ctx, &planNetworkModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var planSettingsModel CloudServerSettingsModel
	diags = data.Settings.As(ctx, &planSettingsModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var planStorageModel CloudServerStorageModel
	diags = data.Storage.As(ctx, &planStorageModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get IDs from state (not plan) - IDs are immutable and should always be in state
	projectID := state.ProjectID.ValueString()
	serverID := state.Id.ValueString()

	if projectID == "" || serverID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Server ID are required to update the cloud server",
		)
		return
	}

	// Get current cloud server details to preserve existing values
	getResponse, err := r.client.Client.FromCompute().CloudServers().Get(ctx, projectID, serverID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current cloud server",
			fmt.Sprintf("Unable to get current cloud server: %s", err),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"Cloud Server Not Found",
			"Cloud server not found or no data returned",
		)
		return
	}

	current := getResponse.Data

	// Extract tags from Terraform list
	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		// Preserve existing tags if not provided
		tags = current.Metadata.Tags
	}

	// Build the update request, preserving existing values
	flavorName := current.Properties.Flavor.Name
	updateRequest := sdktypes.CloudServerRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: current.Metadata.LocationResponse.Value,
			},
		},
		Properties: sdktypes.CloudServerPropertiesRequest{
			FlavorName: &flavorName,
			BootVolume: current.Properties.BootVolume,
		},
	}

	// Preserve keypair if it exists
	if current.Properties.KeyPair.URI != "" {
		updateRequest.Properties.KeyPair = current.Properties.KeyPair
	}

	// Update the cloud server using the SDK
	response, err := r.client.Client.FromCompute().CloudServers().Update(ctx, projectID, serverID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating cloud server",
			fmt.Sprintf("Unable to update cloud server: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		logContext := map[string]interface{}{
			"project_id": projectID,
			"server_id":  serverID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to update cloud server", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	// Update state with response data if available
	if response != nil && response.Data != nil {
		if response.Data.Metadata.Name != nil && *response.Data.Metadata.Name != "" {
			data.Name = types.StringValue(*response.Data.Metadata.Name)
		}
		// Update tags from response
		if len(response.Data.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(response.Data.Metadata.Tags))
			for i, tag := range response.Data.Metadata.Tags {
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
	}

	// Ensure immutable fields are set from state/plan before saving
	data.Id = state.Id
	data.ProjectID = state.ProjectID
	data.Zone = state.Zone

	// Rebuild nested objects from plan to avoid inconsistent result
	// Network object (preserve from plan)
	networkAttrs := map[string]attr.Value{
		"vpc_uri_ref":            planNetworkModel.VpcUriRef,
		"elastic_ip_uri_ref":     planNetworkModel.ElasticIpUriRef,
		"subnet_uri_refs":        planNetworkModel.SubnetUriRefs,
		"securitygroup_uri_refs": planNetworkModel.SecurityGroupUriRefs,
	}
	networkObj, diags := types.ObjectValue(
		map[string]attr.Type{
			"vpc_uri_ref":            types.StringType,
			"elastic_ip_uri_ref":     types.StringType,
			"subnet_uri_refs":        types.ListType{ElemType: types.StringType},
			"securitygroup_uri_refs": types.ListType{ElemType: types.StringType},
		},
		networkAttrs,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Network = networkObj

	// Settings object (preserve from plan)
	settingsAttrs := map[string]attr.Value{
		"flavor_name":      planSettingsModel.FlavorName,
		"key_pair_uri_ref": planSettingsModel.KeyPairUriRef,
		"user_data":        planSettingsModel.UserData,
	}
	settingsObj, diags := types.ObjectValue(
		map[string]attr.Type{
			"flavor_name":      types.StringType,
			"key_pair_uri_ref": types.StringType,
			"user_data":        types.StringType,
		},
		settingsAttrs,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Settings = settingsObj

	// Storage object (preserve from plan)
	storageAttrs := map[string]attr.Value{
		"boot_volume_uri_ref": planStorageModel.BootVolumeUriRef,
	}
	storageObj, diags := types.ObjectValue(
		map[string]attr.Type{
			"boot_volume_uri_ref": types.StringType,
		},
		storageAttrs,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Storage = storageObj

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CloudServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CloudServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	serverID := data.Id.ValueString()

	if projectID == "" || serverID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Server ID are required to delete the cloud server",
		)
		return
	}

	// Delete the cloud server using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromCompute().CloudServers().Delete(ctx, projectID, serverID, nil)
		},
		ExtractSDKError,
		"CloudServer",
		serverID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting cloud server",
			fmt.Sprintf("Unable to delete cloud server: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a CloudServer resource", map[string]interface{}{
		"cloudserver_id": serverID,
	})
}

func (r *CloudServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
