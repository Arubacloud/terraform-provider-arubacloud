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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &VpcPeeringResource{}
var _ resource.ResourceWithImportState = &VpcPeeringResource{}

func NewVpcPeeringResource() resource.Resource {
	return &VpcPeeringResource{}
}

type VpcPeeringResource struct {
	client *ArubaCloudClient
}

type VpcPeeringResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	Tags      types.List   `tfsdk:"tags"`
	ProjectId types.String `tfsdk:"project_id"`
	VpcId     types.String `tfsdk:"vpc_id"`
	PeerVpc   types.String `tfsdk:"peer_vpc"`
}

func (r *VpcPeeringResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpcpeering"
}

func (r *VpcPeeringResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VPC Peering resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VPC Peering identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "VPC Peering URI",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "VPC Peering name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "VPC Peering location",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the VPC Peering",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this VPC Peering belongs to",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPC this peering belongs to",
				Required:            true,
			},
			"peer_vpc": schema.StringAttribute{
				MarkdownDescription: "ID or URI of the peer VPC to connect to",
				Required:            true,
			},
		},
	}
}

func (r *VpcPeeringResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VpcPeeringResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VpcPeeringResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()

	if projectID == "" || vpcID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and VPC ID are required to create a VPC peering",
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

	// Build peer VPC URI
	peerVPCURI := data.PeerVpc.ValueString()
	if !strings.HasPrefix(peerVPCURI, "/") {
		peerVPCURI = fmt.Sprintf("/projects/%s/providers/Aruba.Network/vpcs/%s", projectID, peerVPCURI)
	}

	// Build the create request
	createRequest := sdktypes.VPCPeeringRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.VPCPeeringPropertiesRequest{
			RemoteVPC: &sdktypes.ReferenceResource{
				URI: peerVPCURI,
			},
		},
	}

	// Create the VPC peering using the SDK
	response, err := r.client.Client.FromNetwork().VPCPeerings().Create(ctx, projectID, vpcID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating VPC peering",
			fmt.Sprintf("Unable to create VPC peering: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create VPC peering"
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
			"VPC peering created but no data returned from API",
		)
		return
	}

	// Wait for VPC Peering to be active before returning (VpcPeering is referenced by VpcPeeringRoute)
	// This ensures Terraform doesn't proceed to create dependent resources until VPC Peering is ready
	peeringID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromNetwork().VPCPeerings().Get(ctx, projectID, vpcID, peeringID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for VPC Peering to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "VpcPeering", peeringID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"VPC Peering Not Active",
			fmt.Sprintf("VPC peering was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a VPC Peering resource", map[string]interface{}{
		"vpcpeering_id":   data.Id.ValueString(),
		"vpcpeering_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpcPeeringResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VpcPeeringResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	peeringID := data.Id.ValueString()

	if projectID == "" || vpcID == "" || peeringID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, and Peering ID are required to read the VPC peering",
		)
		return
	}

	// Get VPC peering details using the SDK
	response, err := r.client.Client.FromNetwork().VPCPeerings().Get(ctx, projectID, vpcID, peeringID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading VPC peering",
			fmt.Sprintf("Unable to read VPC peering: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read VPC peering"
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
		peering := response.Data

		if peering.Metadata.ID != nil {
			data.Id = types.StringValue(*peering.Metadata.ID)
		}
		if peering.Metadata.URI != nil {
			data.Uri = types.StringValue(*peering.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if peering.Metadata.Name != nil {
			data.Name = types.StringValue(*peering.Metadata.Name)
		}
		if peering.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(peering.Metadata.LocationResponse.Value)
		}
		if peering.Properties.RemoteVPC != nil {
			data.PeerVpc = types.StringValue(peering.Properties.RemoteVPC.URI)
		}

		// Update tags
		if len(peering.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(peering.Metadata.Tags))
			for i, tag := range peering.Metadata.Tags {
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

func (r *VpcPeeringResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VpcPeeringResourceModel
	var state VpcPeeringResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get IDs from state (not plan) - IDs are immutable and should always be in state
	projectID := state.ProjectId.ValueString()
	vpcID := state.VpcId.ValueString()
	peeringID := state.Id.ValueString()

	if projectID == "" || vpcID == "" || peeringID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, and Peering ID are required to update the VPC peering",
		)
		return
	}

	// Get current VPC peering details
	getResponse, err := r.client.Client.FromNetwork().VPCPeerings().Get(ctx, projectID, vpcID, peeringID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current VPC peering",
			fmt.Sprintf("Unable to get current VPC peering: %s", err),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"VPC Peering Not Found",
			"VPC peering not found or no data returned",
		)
		return
	}

	current := getResponse.Data

	// Check if VPC peering is in InCreation state
	if current.Status.State != nil && *current.Status.State == "InCreation" {
		resp.Diagnostics.AddError(
			"Cannot Update VPC Peering",
			"Cannot update VPC peering while it is in 'InCreation' state. Please wait until the VPC peering is fully created.",
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
			"Unable to determine region value for VPC peering",
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

	// Build update request - only name and tags can be updated, peer VPC must remain unchanged
	updateRequest := sdktypes.VPCPeeringRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.VPCPeeringPropertiesRequest{
			// Peer VPC cannot be updated - use current value
			RemoteVPC: current.Properties.RemoteVPC,
		},
	}

	// Update the VPC peering using the SDK
	response, err := r.client.Client.FromNetwork().VPCPeerings().Update(ctx, projectID, vpcID, peeringID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating VPC peering",
			fmt.Sprintf("Unable to update VPC peering: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to update VPC peering"
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
	data.ProjectId = state.ProjectId
	data.VpcId = state.VpcId

	if response != nil && response.Data != nil {
		// Update from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpcPeeringResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VpcPeeringResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	peeringID := data.Id.ValueString()

	if projectID == "" || vpcID == "" || peeringID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, and Peering ID are required to delete the VPC peering",
		)
		return
	}

	// Delete the VPC peering using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromNetwork().VPCPeerings().Delete(ctx, projectID, vpcID, peeringID, nil)
		},
		ExtractSDKError,
		"VPCPeering",
		peeringID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting VPC peering",
			fmt.Sprintf("Unable to delete VPC peering: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a VPC Peering resource", map[string]interface{}{
		"vpcpeering_id": peeringID,
	})
}

func (r *VpcPeeringResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
