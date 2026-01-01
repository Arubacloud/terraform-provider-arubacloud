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
		MarkdownDescription: "Project resource",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Project name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Project description",
				Optional:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the project",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Project Identifier",
				Computed:            true,
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
			fmt.Sprintf("Unable to create project: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create project"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
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
		"project_id": data.Id.ValueString(),
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
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"Project ID is required to read the project",
		)
		return
	}

	// Get project details using the SDK
	response, err := r.client.Client.FromProject().Get(ctx, projectID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading project",
			fmt.Sprintf("Unable to read project: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		// If project not found, mark as removed
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read project"
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
			// Set empty list if no tags
			emptyList, diags := types.ListValue(types.StringType, []attr.Value{})
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = emptyList
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

	projectID := data.Id.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"Project ID is required to update the project",
		)
		return
	}

	// Get current project details to preserve existing values
	getResponse, err := r.client.Client.FromProject().Get(ctx, projectID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current project",
			fmt.Sprintf("Unable to get current project: %s", err),
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
			fmt.Sprintf("Unable to update project: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to update project"
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
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromProject().Delete(ctx, projectID, nil)
		},
		ExtractSDKError,
		"Project",
		projectID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting project",
			fmt.Sprintf("Unable to delete project: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a project resource", map[string]interface{}{
		"project_id": projectID,
	})
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
