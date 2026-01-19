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

var _ resource.Resource = &SecurityGroupResource{}
var _ resource.ResourceWithImportState = &SecurityGroupResource{}

func NewSecurityGroupResource() resource.Resource {
	return &SecurityGroupResource{}
}

type SecurityGroupResource struct {
	client *ArubaCloudClient
}

type SecurityGroupResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	Tags      types.List   `tfsdk:"tags"`
	ProjectId types.String `tfsdk:"project_id"`
	VpcId     types.String `tfsdk:"vpc_id"`
}

func (r *SecurityGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_securitygroup"
}

func (r *SecurityGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Security Group resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Security Group identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Security Group URI",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Security Group name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Security Group location",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the Security Group",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this Security Group belongs to",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VPC this Security Group belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *SecurityGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SecurityGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecurityGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	if projectID == "" || vpcID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and VPC ID are required to create a security group",
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

	// Build the create request
	createRequest := sdktypes.SecurityGroupRequest{
		Metadata: sdktypes.ResourceMetadataRequest{
			Name: data.Name.ValueString(),
			Tags: tags,
		},
	}

	// Create the security group using the SDK
	response, err := r.client.Client.FromNetwork().SecurityGroups().Create(ctx, projectID, vpcID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating security group",
			fmt.Sprintf("Unable to create security group: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		logContext := map[string]interface{}{}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to create security group", logContext)
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
		if response.Data.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(response.Data.Metadata.LocationResponse.Value)
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Security group created but no data returned from API",
		)
		return
	}

	// Wait for Security Group to be active before returning (SecurityGroup is referenced by CloudServer, SecurityRule)
	// This ensures Terraform doesn't proceed to create dependent resources until SecurityGroup is ready
	sgID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromNetwork().SecurityGroups().Get(ctx, projectID, vpcID, sgID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for Security Group to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "SecurityGroup", sgID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Security Group Not Active",
			fmt.Sprintf("Security group was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	// Re-read the Security Group to get the URI and ensure all fields are properly set
	getResp, err := r.client.Client.FromNetwork().SecurityGroups().Get(ctx, projectID, vpcID, sgID, nil)
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
		// Update tags from response
		if len(getResp.Data.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(getResp.Data.Metadata.Tags))
			for i, tag := range getResp.Data.Metadata.Tags {
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
	} else if err != nil {
		// If Get fails, log but don't fail - we already have the ID from create response
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh Security Group after creation: %v", err))
	}

	tflog.Trace(ctx, "created a Security Group resource", map[string]interface{}{
		"securitygroup_id":   data.Id.ValueString(),
		"securitygroup_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	sgID := data.Id.ValueString()

	if projectID == "" || vpcID == "" || sgID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, and Security Group ID are required to read the security group",
		)
		return
	}

	// Get security group details using the SDK
	response, err := r.client.Client.FromNetwork().SecurityGroups().Get(ctx, projectID, vpcID, sgID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading security group",
			fmt.Sprintf("Unable to read security group: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		logContext := map[string]interface{}{
			"project_id":        projectID,
			"security_group_id": sgID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to read security group", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		sg := response.Data

		if sg.Metadata.ID != nil {
			data.Id = types.StringValue(*sg.Metadata.ID)
		}
		if sg.Metadata.URI != nil {
			data.Uri = types.StringValue(*sg.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if sg.Metadata.Name != nil {
			data.Name = types.StringValue(*sg.Metadata.Name)
		}
		if sg.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(sg.Metadata.LocationResponse.Value)
		}

		// Update tags from response
		if len(sg.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(sg.Metadata.Tags))
			for i, tag := range sg.Metadata.Tags {
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

func (r *SecurityGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SecurityGroupResourceModel
	var state SecurityGroupResourceModel

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
	sgID := state.Id.ValueString()

	if projectID == "" || vpcID == "" || sgID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, and Security Group ID are required to update the security group",
		)
		return
	}

	// Get current security group details
	getResponse, err := r.client.Client.FromNetwork().SecurityGroups().Get(ctx, projectID, vpcID, sgID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current security group",
			fmt.Sprintf("Unable to get current security group: %s", err),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"Security Group Not Found",
			"Security group not found or no data returned",
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
		tags = current.Metadata.Tags
	}

	// Build the update request
	updateRequest := sdktypes.SecurityGroupRequest{
		Metadata: sdktypes.ResourceMetadataRequest{
			Name: data.Name.ValueString(),
			Tags: tags,
		},
	}

	// Update the security group using the SDK
	response, err := r.client.Client.FromNetwork().SecurityGroups().Update(ctx, projectID, vpcID, sgID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating security group",
			fmt.Sprintf("Unable to update security group: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		logContext := map[string]interface{}{}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to update security group", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.ProjectId = state.ProjectId
	data.VpcId = state.VpcId
	data.Uri = state.Uri // Preserve URI from state

	if response != nil && response.Data != nil {
		// Update from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			// If no URI in response, re-read the security group to get the latest state
			getResp, err := r.client.Client.FromNetwork().SecurityGroups().Get(ctx, projectID, vpcID, sgID, nil)
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

func (r *SecurityGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SecurityGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()
	sgID := data.Id.ValueString()

	if projectID == "" || vpcID == "" || sgID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, VPC ID, and Security Group ID are required to delete the security group",
		)
		return
	}

	// Delete the security group using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromNetwork().SecurityGroups().Delete(ctx, projectID, vpcID, sgID, nil)
		},
		ExtractSDKError,
		"SecurityGroup",
		sgID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting security group",
			fmt.Sprintf("Unable to delete security group: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a Security Group resource", map[string]interface{}{
		"securitygroup_id": sgID,
	})
}

func (r *SecurityGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
