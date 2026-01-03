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

var _ resource.Resource = &SubnetResource{}
var _ resource.ResourceWithImportState = &SubnetResource{}

func NewSubnetResource() resource.Resource {
	return &SubnetResource{}
}

type SubnetResource struct {
	client *ArubaCloudClient
}

type SubnetResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	Tags      types.List   `tfsdk:"tags"`
	ProjectId types.String `tfsdk:"project_id"`
	VpcId     types.String `tfsdk:"vpc_id"`
	Type      types.String `tfsdk:"type"`
	Network   types.Object `tfsdk:"network"`
	Dhcp      types.Object `tfsdk:"dhcp"`
	Routes    types.List   `tfsdk:"routes"`
	Dns       types.List   `tfsdk:"dns"`
}

type NetworkModel struct {
	Address types.String `tfsdk:"address"`
}

type DhcpModel struct {
	Enabled types.Bool   `tfsdk:"enabled"`
	Range   types.Object `tfsdk:"range"`
}

type DhcpRangeModel struct {
	Start types.String `tfsdk:"start"`
	Count types.Int64  `tfsdk:"count"`
}

type RouteModel struct {
	Address types.String `tfsdk:"address"`
	Gateway types.String `tfsdk:"gateway"`
}

func (r *SubnetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (r *SubnetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Subnet resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Subnet identifier",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Subnet URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Subnet name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Subnet location",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the subnet",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this subnet belongs to",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPC this subnet belongs to",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Subnet type (Basic or Advanced)",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"network": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						MarkdownDescription: "Address of the network in CIDR notation (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"dhcp": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable DHCP",
						Optional:            true,
					},
					"range": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"start": schema.StringAttribute{
								MarkdownDescription: "Starting IP address",
								Optional:            true,
							},
							"count": schema.Int64Attribute{
								MarkdownDescription: "Number of available IP addresses",
								Optional:            true,
							},
						},
						Optional: true,
					},
				},
				Optional: true,
			},
			"routes": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							MarkdownDescription: "IP address of the route",
							Optional:            true,
						},
						"gateway": schema.StringAttribute{
							MarkdownDescription: "Gateway",
							Optional:            true,
						},
					},
				},
				Optional: true,
			},
			"dns": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of DNS IP addresses",
				Optional:            true,
			},
		},
	}
}

func (r *SubnetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	if projectID == "" || vpcID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and VPC ID are required to create a subnet",
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

	// Extract network CIDR if provided
	var network *sdktypes.SubnetNetwork
	if !data.Network.IsNull() && !data.Network.IsUnknown() {
		networkObj, diags := data.Network.ToObjectValue(ctx)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			attrs := networkObj.Attributes()
			if addressAttr, ok := attrs["address"]; ok && addressAttr != nil {
				if addressStr, ok := addressAttr.(types.String); ok && !addressStr.IsNull() {
					addressValue := addressStr.ValueString()
					if addressValue != "" {
						network = &sdktypes.SubnetNetwork{
							Address: addressValue,
						}
					}
				}
			}
		}
	}

	// Determine SubnetType: Advanced if CIDR is provided, Basic otherwise
	subnetType := sdktypes.SubnetTypeBasic
	if network != nil && network.Address != "" {
		subnetType = sdktypes.SubnetTypeAdvanced
	} else if data.Type.ValueString() == "Advanced" {
		subnetType = sdktypes.SubnetTypeAdvanced
	}

	// Build the create request
	createRequest := sdktypes.SubnetRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.SubnetPropertiesRequest{
			Type:    subnetType,
			Network: network,
		},
	}

	// Create the subnet using the SDK
	response, err := r.client.Client.FromNetwork().Subnets().Create(ctx, projectID, vpcID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating subnet",
			fmt.Sprintf("Unable to create subnet: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create subnet"
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
			"Subnet created but no data returned from API",
		)
		return
	}

	// Wait for Subnet to be active before returning (Subnet is referenced by CloudServer)
	// This ensures Terraform doesn't proceed to create dependent resources until Subnet is ready
	subnetID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromNetwork().Subnets().Get(ctx, projectID, vpcID, subnetID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for Subnet to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "Subnet", subnetID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Subnet Not Active",
			fmt.Sprintf("Subnet was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a Subnet resource", map[string]interface{}{
		"subnet_id": data.Id.ValueString(),
		"subnet_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	subnetID := data.Id.ValueString()

	if projectID == "" || vpcID == "" || subnetID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, and Subnet ID are required to read the subnet",
		)
		return
	}

	// Get subnet details using the SDK
	response, err := r.client.Client.FromNetwork().Subnets().Get(ctx, projectID, vpcID, subnetID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading subnet",
			fmt.Sprintf("Unable to read subnet: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read subnet"
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
		subnet := response.Data

		if subnet.Metadata.ID != nil {
			data.Id = types.StringValue(*subnet.Metadata.ID)
		}
		if subnet.Metadata.URI != nil {
			data.Uri = types.StringValue(*subnet.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if subnet.Metadata.Name != nil {
			data.Name = types.StringValue(*subnet.Metadata.Name)
		}
		if subnet.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(subnet.Metadata.LocationResponse.Value)
		}
		data.Type = types.StringValue(string(subnet.Properties.Type))

		// Update network if available
		if subnet.Properties.Network != nil && subnet.Properties.Network.Address != "" {
			networkModel := NetworkModel{
				Address: types.StringValue(subnet.Properties.Network.Address),
			}
			networkObj, diags := types.ObjectValue(map[string]attr.Type{
				"address": types.StringType,
			}, map[string]attr.Value{
				"address": networkModel.Address,
			})
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Network = networkObj
			}
		}

		// Update tags from response
		if len(subnet.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(subnet.Metadata.Tags))
			for i, tag := range subnet.Metadata.Tags {
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

func (r *SubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SubnetResourceModel
	var state SubnetResourceModel

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
	subnetID := state.Id.ValueString()

	if projectID == "" || vpcID == "" || subnetID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, and Subnet ID are required to update the subnet",
		)
		return
	}

	// Get current subnet details
	getResponse, err := r.client.Client.FromNetwork().Subnets().Get(ctx, projectID, vpcID, subnetID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current subnet",
			fmt.Sprintf("Unable to get current subnet: %s", err),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"Subnet Not Found",
			"Subnet not found or no data returned",
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
			"Unable to determine region value for subnet",
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

	// Preserve network from current state
	network := current.Properties.Network

	// Build the update request
	updateRequest := sdktypes.SubnetRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.SubnetPropertiesRequest{
			Type:    current.Properties.Type,
			Network: network,
		},
	}

	// Update the subnet using the SDK
	response, err := r.client.Client.FromNetwork().Subnets().Update(ctx, projectID, vpcID, subnetID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating subnet",
			fmt.Sprintf("Unable to update subnet: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to update subnet"
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

func (r *SubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	subnetID := data.Id.ValueString()

	if projectID == "" || vpcID == "" || subnetID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, and Subnet ID are required to delete the subnet",
		)
		return
	}

	// Delete the subnet using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromNetwork().Subnets().Delete(ctx, projectID, vpcID, subnetID, nil)
		},
		ExtractSDKError,
		"Subnet",
		subnetID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting subnet",
			fmt.Sprintf("Unable to delete subnet: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a Subnet resource", map[string]interface{}{
		"subnet_id": subnetID,
	})
}

func (r *SubnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
