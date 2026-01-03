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

type CloudServerResourceModel struct {
	Id             types.String `tfsdk:"id"`
	Uri            types.String `tfsdk:"uri"`
	Name           types.String `tfsdk:"name"`
	Location       types.String `tfsdk:"location"`
	ProjectID      types.String `tfsdk:"project_id"`
	Zone           types.String `tfsdk:"zone"`
	VpcID          types.String `tfsdk:"vpc_id"`
	FlavorName     types.String `tfsdk:"flavor_name"`
	ElasticIPID    types.String `tfsdk:"elastic_ip_id"`
	BootVolume     types.String `tfsdk:"boot_volume"`
	KeyPairID      types.String `tfsdk:"key_pair_id"`
	Subnets        types.List   `tfsdk:"subnets"`
	SecurityGroups types.List   `tfsdk:"securitygroups"`
	Tags           types.List   `tfsdk:"tags"`
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
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "VPC ID",
				Required:            true,
			},
			"flavor_name": schema.StringAttribute{
				MarkdownDescription: "Flavor name",
				Required:            true,
			},
			"elastic_ip_id": schema.StringAttribute{
				MarkdownDescription: "Elastic IP ID",
				Optional:            true,
			},
			"boot_volume": schema.StringAttribute{
				MarkdownDescription: "Boot volume ID",
				Required:            true,
			},
			"key_pair_id": schema.StringAttribute{
				MarkdownDescription: "Key pair ID",
				Optional:            true,
			},
			"subnets": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of subnet IDs",
				Required:            true,
			},
			"securitygroups": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of security group reference IDs",
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

	// Extract subnets from Terraform list
	var subnetRefs []sdktypes.ReferenceResource
	if !data.Subnets.IsNull() && !data.Subnets.IsUnknown() {
		var subnetIDs []string
		diags := data.Subnets.ElementsAs(ctx, &subnetIDs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Convert subnet IDs to ReferenceResource URIs
		for _, subnetID := range subnetIDs {
			subnetURI := subnetID
			if !strings.HasPrefix(subnetURI, "/") {
				subnetURI = fmt.Sprintf("/projects/%s/providers/Aruba.Network/subnets/%s", projectID, subnetID)
			}
			subnetRefs = append(subnetRefs, sdktypes.ReferenceResource{
				URI: subnetURI,
			})
		}
	}

	// Extract security groups from Terraform list
	var securityGroupRefs []sdktypes.ReferenceResource
	if !data.SecurityGroups.IsNull() && !data.SecurityGroups.IsUnknown() {
		var sgIDs []string
		diags := data.SecurityGroups.ElementsAs(ctx, &sgIDs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Convert security group IDs to ReferenceResource URIs
		for _, sgID := range sgIDs {
			sgURI := sgID
			if !strings.HasPrefix(sgURI, "/") {
				sgURI = fmt.Sprintf("/projects/%s/providers/Aruba.Network/securityGroups/%s", projectID, sgID)
			}
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

	// Build template URI
	templateURI := data.BootVolume.ValueString()
	if !strings.HasPrefix(templateURI, "/") {
		templateURI = fmt.Sprintf("/projects/%s/providers/Aruba.Compute/templates/%s", projectID, data.BootVolume.ValueString())
	}

	// Build VPC URI
	vpcURI := data.VpcID.ValueString()
	if !strings.HasPrefix(vpcURI, "/") {
		vpcURI = fmt.Sprintf("/projects/%s/providers/Aruba.Network/vpcs/%s", projectID, data.VpcID.ValueString())
	}

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
				URI: templateURI,
			},
			VPC: sdktypes.ReferenceResource{
				URI: vpcURI,
			},
			Subnets:        subnetRefs,
			SecurityGroups: securityGroupRefs,
		},
	}

	// Add keypair if provided
	if !data.KeyPairID.IsNull() && !data.KeyPairID.IsUnknown() {
		keypairURI := data.KeyPairID.ValueString()
		if !strings.HasPrefix(keypairURI, "/") {
			keypairURI = fmt.Sprintf("/projects/%s/providers/Aruba.Compute/keyPairs/%s", projectID, data.KeyPairID.ValueString())
		}
		createRequest.Properties.KeyPair = sdktypes.ReferenceResource{
			URI: keypairURI,
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
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
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
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

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
			// Extract template ID from URI if needed
			data.BootVolume = types.StringValue(server.Properties.BootVolume.URI)
		}

		if server.Properties.KeyPair.URI != "" {
			// Extract keypair ID from URI if needed
			data.KeyPairID = types.StringValue(server.Properties.KeyPair.URI)
		} else {
			data.KeyPairID = types.StringNull()
		}

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
	data.VpcID = state.VpcID
	data.Zone = state.Zone
	data.FlavorName = state.FlavorName
	data.BootVolume = state.BootVolume
	data.KeyPairID = state.KeyPairID
	data.Subnets = state.Subnets
	data.SecurityGroups = state.SecurityGroups

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
