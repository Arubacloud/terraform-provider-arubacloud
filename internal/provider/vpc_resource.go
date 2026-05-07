package provider

import (
	"context"
	"fmt"
	"time"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type VPCResource struct {
	client *ArubaCloudClient
}

type VPCResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	ProjectID types.String `tfsdk:"project_id"`
	Tags      types.List   `tfsdk:"tags"`
}

var _ resource.Resource = &VPCResource{}
var _ resource.ResourceWithImportState = &VPCResource{}

func NewVPCResource() resource.Resource {
	return &VPCResource{}
}

func (r *VPCResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an ArubaCloud VPC (Virtual Private Cloud) — the isolated network boundary within a region where subnets, security groups, and server instances are provisioned.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources (e.g., as a `*_uri_ref` attribute).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the VPC.",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier for the resource (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center). (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Optional:            true,
			},
		},
	}
}

func (r *VPCResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VPCResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VPCResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"Project ID is required to create a VPC",
		)
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the create request
	setDefault := false
	setPreset := false
	createRequest := sdktypes.VPCRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.VPCPropertiesRequest{
			Properties: &sdktypes.VPCProperties{
				Default: &setDefault,
				Preset:  &setPreset,
			},
		},
	}

	// Create the VPC using the SDK
	response, err := r.client.Client.FromNetwork().VPCs().Create(ctx, projectID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating VPC",
			NewTransportError("create", "Vpc", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("create", "Vpc", response); apiErr != nil {
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
			"VPC created but no data returned from API",
		)
		return
	}

	// Wait for VPC to be active before returning (VPC is referenced by Subnets, SecurityGroups, etc.)
	// This ensures Terraform doesn't proceed to create dependent resources until VPC is ready
	vpcID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for VPC to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "VPC", vpcID, r.client.ResourceTimeout); err != nil {
		ReportWaitResult(&resp.Diagnostics, err, "VPC", vpcID)
		// uri is only populated from the GET after the wait; null it out so state has
		// no unknown values (Terraform rejects unknown values in saved state).
		if data.Uri.IsUnknown() {
			data.Uri = types.StringNull()
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Re-read the VPC to get the URI and ensure all fields are properly set
	getResp, err := r.client.Client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
	if err == nil && getResp != nil && getResp.Data != nil {
		// Ensure ID is set from metadata (should already be set, but double-check)
		if getResp.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*getResp.Data.Metadata.ID)
		}
		if getResp.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*getResp.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		// Also update other fields that might have changed
		if getResp.Data.Metadata.Name != nil {
			data.Name = types.StringValue(*getResp.Data.Metadata.Name)
		}
		if getResp.Data.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(getResp.Data.Metadata.LocationResponse.Value)
		}
		data.Tags = TagsToList(getResp.Data.Metadata.Tags)
	} else if err != nil {
		// If Get fails, log but don't fail - we already have the ID from create response
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh VPC after creation: %v", err))
	}

	tflog.Trace(ctx, "created a VPC resource", map[string]interface{}{
		"vpc_id":   data.Id.ValueString(),
		"vpc_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPCResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VPCResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	vpcID := data.Id.ValueString()

	if data.Id.IsUnknown() || data.Id.IsNull() || vpcID == "" {
		tflog.Debug(ctx, "VPC ID is empty, removing resource from state", map[string]interface{}{"vpc_id": vpcID})
		resp.State.RemoveResource(ctx)
		return
	}
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID is required to read the VPC",
		)
		return
	}

	// Get VPC details using the SDK
	response, err := r.client.Client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading VPC",
			NewTransportError("read", "Vpc", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("read", "Vpc", response); apiErr != nil {
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
				fmt.Sprintf("VPC %q reached a terminal failure state (%s) and will not recover on its own. "+
					"Use `terraform apply -replace=<address>` to recreate it.", vpcID, st),
			)
			return
		case IsCreatingState(st):
			checker := func(ctx context.Context) (string, error) {
				getResp, err := r.client.Client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
				if err != nil {
					return "", err
				}
				if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
					return *getResp.Data.Status.State, nil
				}
				return "Unknown", nil
			}
			if err := WaitForResourceActive(ctx, checker, "VPC", vpcID, r.client.ResourceTimeout); err != nil {
				ReportWaitResult(&resp.Diagnostics, err, "VPC", vpcID)
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
			// Re-read to get the final active state.
			response, err = r.client.Client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
			if err != nil {
				resp.Diagnostics.AddError("Error reading VPC after provisioning wait",
					NewTransportError("read", "Vpc", err).Error())
				return
			}
			if apiErr := CheckResponse("read", "Vpc", response); apiErr != nil {
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
		vpc := response.Data

		if vpc.Metadata.ID != nil {
			data.Id = types.StringValue(*vpc.Metadata.ID)
		}
		if vpc.Metadata.URI != nil {
			data.Uri = types.StringValue(*vpc.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if vpc.Metadata.Name != nil {
			data.Name = types.StringValue(*vpc.Metadata.Name)
		}
		if vpc.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(vpc.Metadata.LocationResponse.Value)
		}

		data.Tags = TagsToList(vpc.Metadata.Tags)
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPCResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VPCResourceModel
	var state VPCResourceModel

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
	vpcID := state.Id.ValueString()

	if projectID == "" || vpcID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and VPC ID are required to update the VPC",
		)
		return
	}

	// Get current VPC details
	getResponse, err := r.client.Client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current VPC",
			NewTransportError("read", "Vpc", err).Error(),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"VPC Not Found",
			"VPC not found or no data returned",
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
			"Unable to determine region value for VPC",
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

	// Build the update request
	setDefault := current.Properties.Default
	// Note: Preset field may not be available in VPCPropertiesResponse
	// Only preserve Default if it exists
	updateRequest := sdktypes.VPCRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.VPCPropertiesRequest{
			Properties: &sdktypes.VPCProperties{
				Default: &setDefault,
				// Preset field not available in response type - omit if not needed
			},
		},
	}

	// Update the VPC using the SDK
	response, err := r.client.Client.FromNetwork().VPCs().Update(ctx, projectID, vpcID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating VPC",
			NewTransportError("update", "Vpc", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("update", "Vpc", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.ProjectID = state.ProjectID
	data.Uri = state.Uri // Preserve URI from state

	if response != nil && response.Data != nil {
		// Update from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			// If no URI in response, re-read the VPC to get the latest state
			getResp, err := r.client.Client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
			if err == nil && getResp != nil && getResp.Data != nil {
				if getResp.Data.Metadata.URI != nil {
					data.Uri = types.StringValue(*getResp.Data.Metadata.URI)
				} else {
					data.Uri = state.Uri // Fallback to state if not available
				}
			} else {
				data.Uri = state.Uri // Fallback to state if re-read fails
			}
		}
	} else {
		// If no response, preserve URI from state
		data.Uri = state.Uri
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPCResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VPCResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	vpcID := data.Id.ValueString()

	if projectID == "" || vpcID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and VPC ID are required to delete the VPC",
		)
		return
	}

	deletionChecker := func(ctx context.Context) (bool, error) {
		getResp, getErr := r.client.Client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
		if getErr != nil {
			return false, NewTransportError("get", "VPC", getErr)
		}
		if provErr := CheckResponse("get", "VPC", getResp); provErr != nil {
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
			resp, err := r.client.Client.FromNetwork().VPCs().Delete(ctx, projectID, vpcID, nil)
			if err != nil {
				return NewTransportError("delete", "VPC", err)
			}
			return CheckResponse("delete", "VPC", resp)
		},
		"VPC",
		vpcID,
		r.client.ResourceTimeout,
		deletionChecker,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting VPC",
			NewTransportError("delete", "Vpc", err).Error(),
		)
		return
	}

	// Poll until the VPC is confirmed deleted (async deletion)
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "VPC", vpcID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError(
			"Error waiting for VPC deletion",
			waitErr.Error(),
		)
		return
	}

	tflog.Trace(ctx, "deleted a VPC resource", map[string]interface{}{
		"vpc_id": vpcID,
	})
}

func (r *VPCResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *VPCResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}
