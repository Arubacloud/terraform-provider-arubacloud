package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
		MarkdownDescription: "Manages an ArubaCloud VPC Peering connection between two VPCs, enabling private IP routing between them without traversing the public internet. Both VPCs must exist in the same region. Use `arubacloud_vpcpeeringroute` to configure the routes within each peered VPC.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource.",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources (e.g., as a `*_uri_ref` attribute).",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the VPC peering.",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier for the resource (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the local VPC initiating this peering connection.",
				Required:            true,
			},
			"peer_vpc": schema.StringAttribute{
				MarkdownDescription: "ID or URI of the remote peer VPC to connect to.",
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

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
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
			NewTransportError("create", "Vpcpeering", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("create", "Vpcpeering", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
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
		ReportWaitResult(&resp.Diagnostics, err, "VpcPeering", peeringID)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

	if data.Id.IsUnknown() || data.Id.IsNull() || peeringID == "" {
		tflog.Debug(ctx, "VPC Peering ID is empty, removing resource from state", map[string]interface{}{"peering_id": peeringID})
		resp.State.RemoveResource(ctx)
		return
	}
	if projectID == "" || vpcID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and VPC ID are required to read the VPC peering",
		)
		return
	}

	// Get VPC peering details using the SDK
	response, err := r.client.Client.FromNetwork().VPCPeerings().Get(ctx, projectID, vpcID, peeringID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading VPC peering",
			NewTransportError("read", "Vpcpeering", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("read", "Vpcpeering", response); apiErr != nil {
		if IsNotFound(apiErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}

	// If the resource is still provisioning (e.g. after a Create timeout that saved
	// partial state), resume the wait so the next terraform apply reconciles correctly.
	if response.Data.Status.State != nil {
		switch st := *response.Data.Status.State; {
		case isFailedState(st):
			resp.Diagnostics.AddError(
				"Resource in Failed State",
				fmt.Sprintf("VpcPeering %q reached a terminal failure state (%s) and will not recover on its own. "+
					"Use `terraform apply -replace=<address>` to recreate it.", peeringID, st),
			)
			return
		case IsCreatingState(st):
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
			if err := WaitForResourceActive(ctx, checker, "VpcPeering", peeringID, r.client.ResourceTimeout); err != nil {
				ReportWaitResult(&resp.Diagnostics, err, "VpcPeering", peeringID)
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
			// Re-read to get the final active state.
			response, err = r.client.Client.FromNetwork().VPCPeerings().Get(ctx, projectID, vpcID, peeringID, nil)
			if err != nil {
				resp.Diagnostics.AddError("Error reading VpcPeering after provisioning wait",
					NewTransportError("read", "Vpcpeering", err).Error())
				return
			}
			if apiErr := CheckResponse("read", "Vpcpeering", response); apiErr != nil {
				if IsNotFound(apiErr) {
					resp.State.RemoveResource(ctx)
					return
				}
				resp.Diagnostics.AddError("API Error", apiErr.Error())
				return
			}
		}
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

		data.Tags = TagsToListPreserveNull(peering.Metadata.Tags, data.Tags)
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
			NewTransportError("read", "Vpcpeering", err).Error(),
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

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if tags == nil {
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
			NewTransportError("update", "Vpcpeering", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("update", "Vpcpeering", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
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
	deletionChecker := func(ctx context.Context) (bool, error) {
		getResp, getErr := r.client.Client.FromNetwork().VPCPeerings().Get(ctx, projectID, vpcID, peeringID, nil)
		if getErr != nil {
			return false, NewTransportError("get", "VPCPeering", getErr)
		}
		if provErr := CheckResponse("get", "VPCPeering", getResp); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(
		ctx,
		func() error {
			resp, err := r.client.Client.FromNetwork().VPCPeerings().Delete(ctx, projectID, vpcID, peeringID, nil)
			if err != nil {
				return NewTransportError("delete", "VPCPeering", err)
			}
			return CheckResponse("delete", "VPCPeering", resp)
		},
		"VPCPeering",
		peeringID,
		r.client.ResourceTimeout,
		deletionChecker,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting VPC peering",
			NewTransportError("delete", "Vpcpeering", err).Error(),
		)
		return
	}

	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "VPCPeering", peeringID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for VPCPeering deletion", waitErr.Error())
		return
	}

	tflog.Trace(ctx, "deleted a VPC Peering resource", map[string]interface{}{
		"vpcpeering_id": peeringID,
	})
}

func (r *VpcPeeringResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
