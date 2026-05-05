package provider

import (
	"context"
	"fmt"
	"time"

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

var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

type ProjectResource struct {
	client *ArubaCloudClient
}
type ProjectResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Tags        types.List   `tfsdk:"tags"`
	Id          types.String `tfsdk:"id"`
}

func (r *ProjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *ProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an ArubaCloud Project — the top-level organisational unit for all ArubaCloud resources.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the Project.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional human-readable description of the project.",
				Optional:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ProjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
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
	createRequest := sdktypes.ProjectRequest{
		Metadata: sdktypes.ResourceMetadataRequest{
			Name: data.Name.ValueString(),
			Tags: tags,
		},
		Properties: sdktypes.ProjectPropertiesRequest{
			Default: false, // Default to false unless specified
		},
	}

	// Add description if provided
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		description := data.Description.ValueString()
		createRequest.Properties.Description = &description
	}

	// Create the project using the SDK
	response, err := r.client.Client.FromProject().Create(ctx, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project",
			NewTransportError("create", "Project", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("create", "Project", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}

	if response != nil && response.Data != nil && response.Data.Metadata.ID != nil {
		data.Id = types.StringValue(*response.Data.Metadata.ID)

		// Update description from response if available
		if response.Data.Properties.Description != nil {
			data.Description = types.StringValue(*response.Data.Properties.Description)
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
			// If the config had null tags, keep it null to avoid drift
			// If it had an empty list, set empty list
			if data.Tags.IsNull() {
				data.Tags = types.ListNull(types.StringType)
			} else {
				// Set empty list if no tags and config had a list
				emptyList, diags := types.ListValue(types.StringType, []attr.Value{})
				resp.Diagnostics.Append(diags...)
				if !resp.Diagnostics.HasError() {
					data.Tags = emptyList
				}
			}
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Project created but no ID returned from API",
		)
		return
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a project resource", map[string]interface{}{
		"project_id":   data.Id.ValueString(),
		"project_name": data.Name.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get project ID from state
	projectID := data.Id.ValueString()

	if data.Id.IsUnknown() || data.Id.IsNull() || projectID == "" {
		tflog.Debug(ctx, "Project ID is empty, removing resource from state", map[string]interface{}{"project_id": projectID})
		resp.State.RemoveResource(ctx)
		return
	}

	// Get project details using the SDK
	response, err := r.client.Client.FromProject().Get(ctx, projectID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading project",
			NewTransportError("read", "Project", err).Error(),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		// If project not found, mark as removed
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		apiErr := newResponseError("read", "Project", response.StatusCode, response.Error, response.RawBody)
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}

	if response != nil && response.Data != nil {
		project := response.Data

		// Update data from API response
		if project.Metadata.ID != nil {
			data.Id = types.StringValue(*project.Metadata.ID)
		}
		if project.Metadata.Name != nil {
			data.Name = types.StringValue(*project.Metadata.Name)
		}
		if project.Properties.Description != nil {
			data.Description = types.StringValue(*project.Properties.Description)
		} else {
			data.Description = types.StringNull()
		}

		// Update tags from response
		if len(project.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(project.Metadata.Tags))
			for i, tag := range project.Metadata.Tags {
				tagValues[i] = types.StringValue(tag)
			}
			tagsList, diags := types.ListValueFrom(ctx, types.StringType, tagValues)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = tagsList
			}
		} else {
			// If the config had null tags, keep it null to avoid drift
			// If it had an empty list, set empty list
			if data.Tags.IsNull() {
				data.Tags = types.ListNull(types.StringType)
			} else {
				// Set empty list if no tags and config had a list
				emptyList, diags := types.ListValue(types.StringType, []attr.Value{})
				resp.Diagnostics.Append(diags...)
				if !resp.Diagnostics.HasError() {
					data.Tags = emptyList
				}
			}
		}
	} else {
		// Project not found, mark as removed
		resp.State.RemoveResource(ctx)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectResourceModel
	var state ProjectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read current state to preserve values
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get project ID from state (not plan) - ID is immutable and should always be in state
	// If state doesn't have ID, try to get it from plan as fallback (shouldn't happen but be defensive)
	var projectID string
	if !state.Id.IsUnknown() && !state.Id.IsNull() && state.Id.ValueString() != "" {
		projectID = state.Id.ValueString()
	} else if !data.Id.IsUnknown() && !data.Id.IsNull() && data.Id.ValueString() != "" {
		// Fallback to plan if state doesn't have it (shouldn't happen for existing resources)
		tflog.Warn(ctx, "Project ID not found in state, using plan ID as fallback")
		projectID = data.Id.ValueString()
	}

	if projectID == "" {
		tflog.Error(ctx, "Project ID is missing from both state and plan", map[string]interface{}{
			"state_id_unknown": state.Id.IsUnknown(),
			"state_id_null":    state.Id.IsNull(),
			"state_id_value":   state.Id.ValueString(),
			"plan_id_unknown":  data.Id.IsUnknown(),
			"plan_id_null":     data.Id.IsNull(),
			"plan_id_value":    data.Id.ValueString(),
			"state_name":       state.Name.ValueString(),
		})
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"Project ID is required to update the project. The resource exists in state but the ID is missing. This indicates a state corruption issue. To fix this, you can:\n"+
				"1. Find the project ID using: acloud management project list\n"+
				"2. Import the resource: terraform import arubacloud_project.test <project_id>\n"+
				"Or manually edit the terraform.tfstate file to add the ID.",
		)
		return
	}

	// Get current project details to preserve existing values
	getResponse, err := r.client.Client.FromProject().Get(ctx, projectID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current project",
			NewTransportError("read", "Project", err).Error(),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"Project Not Found",
			"Project not found or no data returned",
		)
		return
	}

	currentProject := getResponse.Data

	// Extract tags from Terraform list
	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Build the update request with current values as defaults
	updateRequest := sdktypes.ProjectRequest{
		Metadata: sdktypes.ResourceMetadataRequest{
			Name: data.Name.ValueString(),
			Tags: tags,
		},
		Properties: sdktypes.ProjectPropertiesRequest{
			Default: currentProject.Properties.Default,
		},
	}

	// Update description if provided
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		description := data.Description.ValueString()
		updateRequest.Properties.Description = &description
	} else if currentProject.Properties.Description != nil {
		updateRequest.Properties.Description = currentProject.Properties.Description
	}

	// Update the project using the SDK
	response, err := r.client.Client.FromProject().Update(ctx, projectID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating project",
			NewTransportError("update", "Project", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("update", "Project", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}

	// Ensure ID and Name are set from state (they are immutable)
	data.Id = state.Id
	data.Name = state.Name

	if response != nil && response.Data != nil {
		// Update ID from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
		// Update name from response if available (should match state)
		if response.Data.Metadata.Name != nil {
			data.Name = types.StringValue(*response.Data.Metadata.Name)
		}
		// Update description from response if available
		if response.Data.Properties.Description != nil {
			data.Description = types.StringValue(*response.Data.Properties.Description)
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
			// If the config had null tags, keep it null to avoid drift
			// If it had an empty list, set empty list
			if data.Tags.IsNull() {
				data.Tags = types.ListNull(types.StringType)
			} else {
				// Set empty list if no tags and config had a list
				emptyList, diags := types.ListValue(types.StringType, []attr.Value{})
				resp.Diagnostics.Append(diags...)
				if !resp.Diagnostics.HasError() {
					data.Tags = emptyList
				}
			}
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.Id.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"Project ID is required to delete the project",
		)
		return
	}

	// Delete the project using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	deletionChecker := func(ctx context.Context) (bool, error) {
		getResp, getErr := r.client.Client.FromProject().Get(ctx, projectID, nil)
		if getErr != nil {
			return false, NewTransportError("get", "Project", getErr)
		}
		if provErr := CheckResponse("get", "Project", getResp); provErr != nil {
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
			resp, err := r.client.Client.FromProject().Delete(ctx, projectID, nil)
			if err != nil {
				return NewTransportError("delete", "Project", err)
			}
			return CheckResponse("delete", "Project", resp)
		},
		"Project",
		projectID,
		r.client.ResourceTimeout,
		deletionChecker,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting project",
			NewTransportError("delete", "Project", err).Error(),
		)
		return
	}

	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "Project", projectID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for Project deletion", waitErr.Error())
		return
	}

	tflog.Trace(ctx, "deleted a project resource", map[string]interface{}{
		"project_id": projectID,
	})
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
