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

var _ resource.Resource = &VPNRouteResource{}
var _ resource.ResourceWithImportState = &VPNRouteResource{}

func NewVPNRouteResource() resource.Resource {
	return &VPNRouteResource{}
}

type VPNRouteResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	Name        types.String `tfsdk:"name"`
	Location    types.String `tfsdk:"location"`
	Tags        types.List   `tfsdk:"tags"`
	ProjectId   types.String `tfsdk:"project_id"`
	VPNTunnelId types.String `tfsdk:"vpn_tunnel_id"`
	Properties  types.Object `tfsdk:"properties"`
}

type VPNRouteResource struct {
	client *ArubaCloudClient
}

func (r *VPNRouteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpnroute"
}

func (r *VPNRouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VPN Route resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VPN Route identifier",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "VPN Route URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "VPN Route name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "VPN Route location",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the VPN Route",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this VPN Route belongs to",
				Required:            true,
			},
			"vpn_tunnel_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPN Tunnel this route belongs to",
				Required:            true,
			},
			"properties": schema.SingleNestedAttribute{
				MarkdownDescription: "Properties of the VPN Route",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"cloud_subnet": schema.StringAttribute{
						MarkdownDescription: "CIDR of the cloud subnet",
						Required:            true,
					},
					"on_prem_subnet": schema.StringAttribute{
						MarkdownDescription: "CIDR of the on-prem subnet",
						Required:            true,
					},
				},
			},
		},
	}
}

func (r *VPNRouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VPNRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VPNRouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpnTunnelID := data.VPNTunnelId.ValueString()

	if projectID == "" || vpnTunnelID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and VPN Tunnel ID are required to create a VPN route",
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

	// Extract properties from Terraform object
	propertiesObj, diags := data.Properties.ToObjectValue(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	propertiesAttrs := propertiesObj.Attributes()
	cloudSubnet := propertiesAttrs["cloud_subnet"].(types.String).ValueString()
	onPremSubnet := propertiesAttrs["on_prem_subnet"].(types.String).ValueString()

	// Build the create request
	createRequest := sdktypes.VPNRouteRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.VPNRoutePropertiesRequest{
			CloudSubnet:  cloudSubnet,
			OnPremSubnet: onPremSubnet,
		},
	}

	// Create the VPN route using the SDK
	response, err := r.client.Client.FromNetwork().VPNRoutes().Create(ctx, projectID, vpnTunnelID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating VPN route",
			fmt.Sprintf("Unable to create VPN route: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create VPN route"
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
			"VPN route created but no data returned from API",
		)
		return
	}

	// Wait for VPN Route to be active before returning (VPNRoute references VPNTunnel)
	// This ensures Terraform doesn't proceed to create dependent resources until VPN Route is ready
	routeID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromNetwork().VPNRoutes().Get(ctx, projectID, vpnTunnelID, routeID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for VPN Route to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "VPNRoute", routeID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"VPN Route Not Active",
			fmt.Sprintf("VPN route was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a VPN Route resource", map[string]interface{}{
		"vpnroute_id": data.Id.ValueString(),
		"vpnroute_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPNRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VPNRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpnTunnelID := data.VPNTunnelId.ValueString()
	routeID := data.Id.ValueString()

	if projectID == "" || vpnTunnelID == "" || routeID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPN Tunnel ID, and Route ID are required to read the VPN route",
		)
		return
	}

	// Get VPN route details using the SDK
	response, err := r.client.Client.FromNetwork().VPNRoutes().Get(ctx, projectID, vpnTunnelID, routeID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading VPN route",
			fmt.Sprintf("Unable to read VPN route: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read VPN route"
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
		route := response.Data

		if route.Metadata.ID != nil {
			data.Id = types.StringValue(*route.Metadata.ID)
		}
		if route.Metadata.URI != nil {
			data.Uri = types.StringValue(*route.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if route.Metadata.Name != nil {
			data.Name = types.StringValue(*route.Metadata.Name)
		}
		if route.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(route.Metadata.LocationResponse.Value)
		}

		// Update properties from response
		propertiesMap := map[string]attr.Value{
			"cloud_subnet":  types.StringValue(route.Properties.CloudSubnet),
			"on_prem_subnet": types.StringValue(route.Properties.OnPremSubnet),
		}

		propertiesObj, diags := types.ObjectValue(map[string]attr.Type{
			"cloud_subnet":  types.StringType,
			"on_prem_subnet": types.StringType,
		}, propertiesMap)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Properties = propertiesObj
		}

		// Update tags
		if len(route.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(route.Metadata.Tags))
			for i, tag := range route.Metadata.Tags {
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

func (r *VPNRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VPNRouteResourceModel
	var state VPNRouteResourceModel

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
	vpnTunnelID := state.VPNTunnelId.ValueString()
	routeID := state.Id.ValueString()

	if projectID == "" || vpnTunnelID == "" || routeID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPN Tunnel ID, and Route ID are required to update the VPN route",
		)
		return
	}

	// Get current VPN route details
	getResponse, err := r.client.Client.FromNetwork().VPNRoutes().Get(ctx, projectID, vpnTunnelID, routeID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current VPN route",
			fmt.Sprintf("Unable to get current VPN route: %s", err),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"VPN Route Not Found",
			"VPN route not found or no data returned",
		)
		return
	}

	current := getResponse.Data

	// Check if VPN route is in InCreation state
	if current.Status.State != nil && *current.Status.State == "InCreation" {
		resp.Diagnostics.AddError(
			"Cannot Update VPN Route",
			"Cannot update VPN route while it is in 'InCreation' state. Please wait until the VPN route is fully created.",
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
			"Unable to determine region value for VPN route",
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

	// Extract properties
	propertiesObj, diags := data.Properties.ToObjectValue(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	propertiesAttrs := propertiesObj.Attributes()
	cloudSubnet := propertiesAttrs["cloud_subnet"].(types.String).ValueString()
	onPremSubnet := propertiesAttrs["on_prem_subnet"].(types.String).ValueString()

	// Build update request
	updateRequest := sdktypes.VPNRouteRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.VPNRoutePropertiesRequest{
			CloudSubnet:  cloudSubnet,
			OnPremSubnet: onPremSubnet,
		},
	}

	// Update the VPN route using the SDK
	response, err := r.client.Client.FromNetwork().VPNRoutes().Update(ctx, projectID, vpnTunnelID, routeID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating VPN route",
			fmt.Sprintf("Unable to update VPN route: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to update VPN route"
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
	data.VPNTunnelId = state.VPNTunnelId

	if response != nil && response.Data != nil {
		// Update from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPNRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VPNRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpnTunnelID := data.VPNTunnelId.ValueString()
	routeID := data.Id.ValueString()

	if projectID == "" || vpnTunnelID == "" || routeID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPN Tunnel ID, and Route ID are required to delete the VPN route",
		)
		return
	}

	// Delete the VPN route using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromNetwork().VPNRoutes().Delete(ctx, projectID, vpnTunnelID, routeID, nil)
		},
		ExtractSDKError,
		"VPNRoute",
		routeID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting VPN route",
			fmt.Sprintf("Unable to delete VPN route: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a VPN Route resource", map[string]interface{}{
		"vpnroute_id": routeID,
	})
}

func (r *VPNRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
