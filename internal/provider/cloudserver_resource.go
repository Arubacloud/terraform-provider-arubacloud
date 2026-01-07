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

type CloudServerResourceModel struct {
	Id                   types.String `tfsdk:"id"`
	Uri                  types.String `tfsdk:"uri"`
	Name                 types.String `tfsdk:"name"`
	Location             types.String `tfsdk:"location"`
	ProjectID            types.String `tfsdk:"project_id"`
	Zone                 types.String `tfsdk:"zone"`
	VpcUriRef            types.String `tfsdk:"vpc_uri_ref"`
	FlavorName           types.String `tfsdk:"flavor_name"`
	ElasticIpUriRef      types.String `tfsdk:"elastic_ip_uri_ref"`
	BootVolumeUriRef     types.String `tfsdk:"boot_volume_uri_ref"`
	KeyPairUriRef        types.String `tfsdk:"key_pair_uri_ref"`
	SubnetUriRefs        types.List   `tfsdk:"subnet_uri_refs"`
	SecurityGroupUriRefs types.List   `tfsdk:"securitygroup_uri_refs"`
	Tags                 types.List   `tfsdk:"tags"`
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
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "CloudServer URI",
				Computed:            true,
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
			"vpc_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI reference to the VPC. Should be the VPC URI (e.g., `/projects/{project_id}/providers/Aruba.Network/vpcs/{vpc_id}`). You can reference the `uri` attribute from an `arubacloud_vpc` resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"flavor_name": schema.StringAttribute{
				MarkdownDescription: "Flavor name. Available flavors are described in the [ArubaCloud API documentation](https://api.arubacloud.com/docs/metadata/#cloudserver-flavors). For example, `CSO4A8` means 4 CPU and 8GB RAM.",
				Required:            true,
			},
			"elastic_ip_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI reference to the Elastic IP. Should be the Elastic IP URI. You can reference the `uri` attribute from an `arubacloud_elasticip` resource.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"boot_volume_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI reference to the boot volume (block storage). Should be the block storage URI. You can reference the `uri` attribute from an `arubacloud_blockstorage` resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key_pair_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI reference to the Key Pair. Should be the Key Pair URI. You can reference the `uri` attribute from an `arubacloud_keypair` resource.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subnet_uri_refs": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of subnet URI references. Should be subnet URIs. You can reference the `uri` attribute from `arubacloud_subnet` resources like `[arubacloud_subnet.example.uri]`",
				Required:            true,
			},
			"securitygroup_uri_refs": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of security group URI references. Should be security group URIs. You can reference the `uri` attribute from `arubacloud_securitygroup` resources like `[arubacloud_securitygroup.example.uri]`",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the Cloud Server",
				Optional:            true,
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

	// Extract subnets from Terraform list (these are URIs)
	var subnetRefs []sdktypes.ReferenceResource
	if !data.SubnetUriRefs.IsNull() && !data.SubnetUriRefs.IsUnknown() {
		var subnetURIs []string
		diags := data.SubnetUriRefs.ElementsAs(ctx, &subnetURIs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Subnet URIs should already be full URIs from resource references
		for _, subnetURI := range subnetURIs {
			subnetRefs = append(subnetRefs, sdktypes.ReferenceResource{
				URI: subnetURI,
			})
		}
	}

	// Extract security groups from Terraform list (these are URIs)
	var securityGroupRefs []sdktypes.ReferenceResource
	if !data.SecurityGroupUriRefs.IsNull() && !data.SecurityGroupUriRefs.IsUnknown() {
		var sgURIs []string
		diags := data.SecurityGroupUriRefs.ElementsAs(ctx, &sgURIs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Security group URIs should already be full URIs from resource references
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

	// Boot volume URI should already be a full URI from resource reference
	bootVolumeURI := data.BootVolumeUriRef.ValueString()

	// VPC URI should already be a full URI from resource reference
	vpcURI := data.VpcUriRef.ValueString()

	// Note: Elastic IP might be handled differently in the API
	// For now, we'll include VPC, subnets, security groups, and zone which are required

	// Build the create request
	flavorName := data.FlavorName.ValueString()
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
				URI: bootVolumeURI,
			},
			VPC: sdktypes.ReferenceResource{
				URI: vpcURI,
			},
			Subnets:        subnetRefs,
			SecurityGroups: securityGroupRefs,
		},
	}

	// Add keypair if provided (URI should already be a full URI from resource reference)
	if !data.KeyPairUriRef.IsNull() && !data.KeyPairUriRef.IsUnknown() {
		createRequest.Properties.KeyPair = sdktypes.ReferenceResource{
			URI: data.KeyPairUriRef.ValueString(),
		}
	}

	// Note: Elastic IP attachment might be handled separately after server creation
	// or through a different API endpoint. The elastic_ip_id is stored in state for reference.

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
		errorMsg := "Failed to create cloud server"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	// Create response may not have ID in metadata (it's a request type)
	// Use name as ID initially, then get actual ID after wait
	serverID := data.Name.ValueString()
	data.Id = types.StringValue(serverID)
	// CloudServer response uses RegionalResourceMetadataRequest which doesn't have URI
	// Set URI to null for now
	data.Uri = types.StringNull()

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
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
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
	// Note: CloudServer uses name as ID and RegionalResourceMetadataRequest doesn't have URI
	getResp, err := r.client.Client.FromCompute().CloudServers().Get(ctx, projectID, serverID, nil)
	if err == nil && getResp != nil && getResp.Data != nil {
		// CloudServer uses name as ID
		if getResp.Data.Metadata.Name != "" {
			data.Id = types.StringValue(getResp.Data.Metadata.Name)
			data.Name = types.StringValue(getResp.Data.Metadata.Name)
		}
		// CloudServer response uses RegionalResourceMetadataRequest which doesn't have URI
		// Set URI to null (as per Read function)
		data.Uri = types.StringNull()
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
	var state CloudServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Preserve immutable URI refs from state (they're not returned by API)
	data = state

	projectID := data.ProjectID.ValueString()
	serverID := data.Id.ValueString()

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
		errorMsg := "Failed to read cloud server"
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
		server := response.Data

		// Update data from API response
		// CloudServer uses name as ID (Metadata doesn't have ID field)
		data.Id = types.StringValue(server.Metadata.Name)
		// CloudServer response uses RegionalResourceMetadataRequest which doesn't have URI
		// Set URI to null for now
		data.Uri = types.StringNull()
		data.Name = types.StringValue(server.Metadata.Name)
		data.Location = types.StringValue(server.Metadata.Location.Value)
		data.FlavorName = types.StringValue(server.Properties.Flavor.Name)

		if server.Properties.BootVolume.URI != "" {
			data.BootVolumeUriRef = types.StringValue(server.Properties.BootVolume.URI)
		}

		// Update VPC URI reference
		if server.Properties.VPC.URI != "" {
			data.VpcUriRef = types.StringValue(server.Properties.VPC.URI)
		}

		// Update KeyPair URI reference
		if server.Properties.KeyPair.URI != "" {
			data.KeyPairUriRef = types.StringValue(server.Properties.KeyPair.URI)
		} else {
			data.KeyPairUriRef = types.StringNull()
		}

		// Note: Subnets and SecurityGroups are not returned in the API response
		// They are immutable after creation, so we preserve them from state
		// These will be set from state in the Update function

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

		// Note: Subnets and SecurityGroups might need to be extracted from LinkedResources
		// This is a simplified version - you may need to adjust based on actual API response structure
	} else {
		// Cloud server not found, mark as removed
		resp.State.RemoveResource(ctx)
		return
	}

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
				Value: current.Metadata.Location.Value,
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
		errorMsg := "Failed to update cloud server"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	// Update state with response data if available
	// Note: Update response may have different structure
	if response != nil && response.Data != nil {
		// Response may be a request type, so handle carefully
		if response.Data.Metadata.Name != "" {
			data.Name = types.StringValue(response.Data.Metadata.Name)
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

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.ProjectID = state.ProjectID
	data.VpcUriRef = state.VpcUriRef
	data.Zone = state.Zone
	data.FlavorName = state.FlavorName
	data.BootVolumeUriRef = state.BootVolumeUriRef
	data.KeyPairUriRef = state.KeyPairUriRef
	data.SubnetUriRefs = state.SubnetUriRefs
	data.SecurityGroupUriRefs = state.SecurityGroupUriRefs
	// Preserve elastic_ip_uri_ref from state if not in plan
	if data.ElasticIpUriRef.IsNull() || data.ElasticIpUriRef.IsUnknown() {
		data.ElasticIpUriRef = state.ElasticIpUriRef
	}

	// Note: CloudServer uses name as ID and response doesn't have Metadata.ID
	// ID is already set from state above

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
