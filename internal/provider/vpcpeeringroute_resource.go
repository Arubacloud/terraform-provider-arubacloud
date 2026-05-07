package provider

import (
	"context"
	"fmt"
	"time"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &VpcPeeringRouteResource{}
var _ resource.ResourceWithImportState = &VpcPeeringRouteResource{}

func NewVpcPeeringRouteResource() resource.Resource {
	return &VpcPeeringRouteResource{}
}

type VpcPeeringRouteResource struct {
	client *ArubaCloudClient
}

type VpcPeeringRouteResourceModel struct {
	Id                   types.String `tfsdk:"id"`
	Uri                  types.String `tfsdk:"uri"`
	Name                 types.String `tfsdk:"name"`
	Tags                 types.List   `tfsdk:"tags"`
	ProjectId            types.String `tfsdk:"project_id"`
	VpcId                types.String `tfsdk:"vpc_id"`
	VpcPeeringId         types.String `tfsdk:"vpc_peering_id"`
	LocalNetworkAddress  types.String `tfsdk:"local_network_address"`
	RemoteNetworkAddress types.String `tfsdk:"remote_network_address"`
	BillingPeriod        types.String `tfsdk:"billing_period"`
}

func (r *VpcPeeringRouteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpcpeeringroute"
}

func (r *VpcPeeringRouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a route entry within an ArubaCloud VPC Peering connection. Each route directs traffic destined for a specific CIDR block over the peering link. Routes must be created on both sides of the peering connection.",
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
				MarkdownDescription: "Display name for the VPC peering route.",
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
				MarkdownDescription: "ID of the VPC this peering route belongs to.",
				Required:            true,
			},
			"vpc_peering_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPC peering connection this route belongs to.",
				Required:            true,
			},
			"local_network_address": schema.StringAttribute{
				MarkdownDescription: "Local network CIDR that is reachable on this side of the peering (e.g., `10.0.1.0/24`).",
				Required:            true,
			},
			"remote_network_address": schema.StringAttribute{
				MarkdownDescription: "Remote network CIDR reachable through the peering connection (e.g., `10.0.2.0/24`).",
				Required:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle for the resource. Accepted values: `Hour`, `Month`, `Year`.",
				Required:            true,
			},
		},
	}
}

func (r *VpcPeeringRouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VpcPeeringRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VpcPeeringRouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	peeringID := data.VpcPeeringId.ValueString()

	if projectID == "" || vpcID == "" || peeringID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, and VPC Peering ID are required to create a VPC peering route",
		)
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the create request
	createRequest := sdktypes.VPCPeeringRouteRequest{
		Metadata: sdktypes.ResourceMetadataRequest{
			Name: data.Name.ValueString(),
			Tags: tags,
		},
		Properties: sdktypes.VPCPeeringRoutePropertiesRequest{
			LocalNetworkAddress:  data.LocalNetworkAddress.ValueString(),
			RemoteNetworkAddress: data.RemoteNetworkAddress.ValueString(),
			BillingPlan: sdktypes.BillingPeriodResource{
				BillingPeriod: data.BillingPeriod.ValueString(),
			},
		},
	}

	// Create the VPC peering route using the SDK
	response, err := r.client.Client.FromNetwork().VPCPeeringRoutes().Create(ctx, projectID, vpcID, peeringID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating VPC peering route",
			NewTransportError("create", "Vpcpeeringroute", err).Error(),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create VPC peering route"
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
		// VPC Peering Route uses name as ID
		data.Id = types.StringValue(response.Data.Metadata.Name)
		// VPC Peering Route uses RegionalResourceMetadataRequest which doesn't have URI
		data.Uri = types.StringNull()
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"VPC peering route created but no data returned from API",
		)
		return
	}

	// Wait for VPC Peering Route to be active before returning
	// This ensures Terraform doesn't proceed until VPC Peering Route is ready
	routeID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromNetwork().VPCPeeringRoutes().Get(ctx, projectID, vpcID, peeringID, routeID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for VPC Peering Route to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "VpcPeeringRoute", routeID, r.client.ResourceTimeout); err != nil {
		ReportWaitResult(&resp.Diagnostics, err, "VpcPeeringRoute", routeID)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	tflog.Trace(ctx, "created a VPC Peering Route resource", map[string]interface{}{
		"vpcpeeringroute_id":   data.Id.ValueString(),
		"vpcpeeringroute_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpcPeeringRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VpcPeeringRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	peeringID := data.VpcPeeringId.ValueString()
	routeID := data.Id.ValueString()

	if data.Id.IsUnknown() || data.Id.IsNull() || routeID == "" {
		tflog.Debug(ctx, "VPC Peering Route ID is empty, removing resource from state", map[string]interface{}{"route_id": routeID})
		resp.State.RemoveResource(ctx)
		return
	}
	if projectID == "" || vpcID == "" || peeringID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, and VPC Peering ID are required to read the VPC peering route",
		)
		return
	}

	// Get VPC peering route details using the SDK
	response, err := r.client.Client.FromNetwork().VPCPeeringRoutes().Get(ctx, projectID, vpcID, peeringID, routeID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading VPC peering route",
			NewTransportError("read", "Vpcpeeringroute", err).Error(),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read VPC peering route"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	// If the resource is still provisioning (e.g. after a Create timeout that saved
	// partial state), resume the wait so the next terraform apply reconciles correctly.
	if response != nil && response.Data != nil && response.Data.Status.State != nil {
		switch st := *response.Data.Status.State; {
		case isFailedState(st):
			resp.Diagnostics.AddError(
				"Resource in Failed State",
				fmt.Sprintf("VpcPeeringRoute %q reached a terminal failure state (%s) and will not recover on its own. "+
					"Use `terraform apply -replace=<address>` to recreate it.", routeID, st),
			)
			return
		case IsCreatingState(st):
			checker := func(ctx context.Context) (string, error) {
				getResp, err := r.client.Client.FromNetwork().VPCPeeringRoutes().Get(ctx, projectID, vpcID, peeringID, routeID, nil)
				if err != nil {
					return "", err
				}
				if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
					return *getResp.Data.Status.State, nil
				}
				return "Unknown", nil
			}
			if err := WaitForResourceActive(ctx, checker, "VpcPeeringRoute", routeID, r.client.ResourceTimeout); err != nil {
				ReportWaitResult(&resp.Diagnostics, err, "VpcPeeringRoute", routeID)
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
			// Re-read to get the final active state.
			response, err = r.client.Client.FromNetwork().VPCPeeringRoutes().Get(ctx, projectID, vpcID, peeringID, routeID, nil)
			if err != nil {
				resp.Diagnostics.AddError("Error reading VpcPeeringRoute after provisioning wait",
					NewTransportError("read", "Vpcpeeringroute", err).Error())
				return
			}
			if response != nil && response.IsError() && response.Error != nil {
				if response.StatusCode == 404 {
					resp.State.RemoveResource(ctx)
					return
				}
				resp.Diagnostics.AddError("API Error", "Failed to read VPC peering route after provisioning wait")
				return
			}
		}
	}

	if response != nil && response.Data != nil {
		route := response.Data

		data.Id = types.StringValue(route.Metadata.Name)
		// VPC Peering Route uses RegionalResourceMetadataRequest which doesn't have URI
		data.Uri = types.StringNull()
		data.Name = types.StringValue(route.Metadata.Name)
		data.LocalNetworkAddress = types.StringValue(route.Properties.LocalNetworkAddress)
		data.RemoteNetworkAddress = types.StringValue(route.Properties.RemoteNetworkAddress)
		if route.Properties.BillingPlan.BillingPeriod != "" {
			data.BillingPeriod = types.StringValue(route.Properties.BillingPlan.BillingPeriod)
		}

		data.Tags = TagsToListPreserveNull(route.Metadata.Tags, data.Tags)
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpcPeeringRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VpcPeeringRouteResourceModel
	var state VpcPeeringRouteResourceModel

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
	peeringID := state.VpcPeeringId.ValueString()
	routeID := state.Id.ValueString()

	if projectID == "" || vpcID == "" || peeringID == "" || routeID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, VPC Peering ID, and Route ID are required to update the VPC peering route",
		)
		return
	}

	// Get current VPC peering route details
	getResponse, err := r.client.Client.FromNetwork().VPCPeeringRoutes().Get(ctx, projectID, vpcID, peeringID, routeID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current VPC peering route",
			NewTransportError("read", "Vpcpeeringroute", err).Error(),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"VPC Peering Route Not Found",
			"VPC peering route not found or no data returned",
		)
		return
	}

	current := getResponse.Data

	// Check if VPC peering route is in InCreation state
	if current.Status.State != nil && *current.Status.State == "InCreation" {
		resp.Diagnostics.AddError(
			"Cannot Update VPC Peering Route",
			"Cannot update VPC peering route while it is in 'InCreation' state. Please wait until the VPC peering route is fully created.",
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

	// Build update request
	updateRequest := sdktypes.VPCPeeringRouteRequest{
		Metadata: sdktypes.ResourceMetadataRequest{
			Name: data.Name.ValueString(),
			Tags: tags,
		},
		Properties: sdktypes.VPCPeeringRoutePropertiesRequest{
			LocalNetworkAddress:  data.LocalNetworkAddress.ValueString(),
			RemoteNetworkAddress: data.RemoteNetworkAddress.ValueString(),
			BillingPlan: sdktypes.BillingPeriodResource{
				BillingPeriod: data.BillingPeriod.ValueString(),
			},
		},
	}

	// Update the VPC peering route using the SDK
	response, err := r.client.Client.FromNetwork().VPCPeeringRoutes().Update(ctx, projectID, vpcID, peeringID, routeID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating VPC peering route",
			NewTransportError("update", "Vpcpeeringroute", err).Error(),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to update VPC peering route"
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
	data.VpcPeeringId = state.VpcPeeringId

	// Note: VpcPeeringRoute uses name as ID and response doesn't have Metadata.ID
	// ID is already set from state above

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpcPeeringRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VpcPeeringRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	peeringID := data.VpcPeeringId.ValueString()
	routeID := data.Id.ValueString()

	if projectID == "" || vpcID == "" || peeringID == "" || routeID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, VPC Peering ID, and Route ID are required to delete the VPC peering route",
		)
		return
	}

	// Delete the VPC peering route using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	deletionChecker := func(ctx context.Context) (bool, error) {
		getResp, getErr := r.client.Client.FromNetwork().VPCPeeringRoutes().Get(ctx, projectID, vpcID, peeringID, routeID, nil)
		if getErr != nil {
			return false, NewTransportError("get", "VPCPeeringRoute", getErr)
		}
		if provErr := CheckResponse("get", "VPCPeeringRoute", getResp); provErr != nil {
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
			resp, err := r.client.Client.FromNetwork().VPCPeeringRoutes().Delete(ctx, projectID, vpcID, peeringID, routeID, nil)
			if err != nil {
				return NewTransportError("delete", "VPCPeeringRoute", err)
			}
			return CheckResponse("delete", "VPCPeeringRoute", resp)
		},
		"VPCPeeringRoute",
		routeID,
		r.client.ResourceTimeout,
		deletionChecker,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting VPC peering route",
			NewTransportError("delete", "Vpcpeeringroute", err).Error(),
		)
		return
	}

	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "VPCPeeringRoute", routeID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for VPCPeeringRoute deletion", waitErr.Error())
		return
	}

	tflog.Trace(ctx, "deleted a VPC Peering Route resource", map[string]interface{}{
		"vpcpeeringroute_id": routeID,
	})
}

func (r *VpcPeeringRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
