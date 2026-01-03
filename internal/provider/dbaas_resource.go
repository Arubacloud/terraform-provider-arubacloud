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

type DBaaSResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	Tags      types.List   `tfsdk:"tags"`
	ProjectID types.String `tfsdk:"project_id"`
	EngineID  types.String `tfsdk:"engine_id"`
	Flavor    types.String `tfsdk:"flavor"`
}

type DBaaSResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &DBaaSResource{}
var _ resource.ResourceWithImportState = &DBaaSResource{}

func NewDBaaSResource() resource.Resource {
	return &DBaaSResource{}
}

func (r *DBaaSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dbaas"
}

func (r *DBaaSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "DBaaS resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "DBaaS identifier",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "DBaaS URI",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "DBaaS name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "DBaaS location",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the DBaaS resource",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this DBaaS belongs to",
				Required:            true,
			},
			"engine_id": schema.StringAttribute{
				MarkdownDescription: "Database engine ID",
				Required:            true,
			},
			"flavor": schema.StringAttribute{
				MarkdownDescription: "DBaaS flavor name",
				Required:            true,
			},
		},
	}
}

func (r *DBaaSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DBaaSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DBaaSResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"Project ID is required to create a DBaaS instance",
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

	engineID := data.EngineID.ValueString()
	flavor := data.Flavor.ValueString()

	// Build the create request
	createRequest := sdktypes.DBaaSRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.DBaaSPropertiesRequest{
			Engine: &sdktypes.DBaaSEngine{
				ID: &engineID,
			},
			Flavor: &sdktypes.DBaaSFlavor{
				Name: &flavor,
			},
		},
	}

	// Create the DBaaS instance using the SDK
	response, err := r.client.Client.FromDatabase().DBaaS().Create(ctx, projectID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating DBaaS instance",
			fmt.Sprintf("Unable to create DBaaS instance: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create DBaaS instance"
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
			"DBaaS instance created but no data returned from API",
		)
		return
	}

	// Wait for DBaaS to be active before returning
	// This ensures Terraform doesn't proceed until DBaaS is ready
	dbaasID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromDatabase().DBaaS().Get(ctx, projectID, dbaasID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for DBaaS to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "DBaaS", dbaasID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"DBaaS Not Active",
			fmt.Sprintf("DBaaS instance was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a DBaaS resource", map[string]interface{}{
		"dbaas_id": data.Id.ValueString(),
		"dbaas_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DBaaSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DBaaSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.Id.ValueString()

	if projectID == "" || dbaasID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and DBaaS ID are required to read the DBaaS instance",
		)
		return
	}

	// Get DBaaS instance details using the SDK
	response, err := r.client.Client.FromDatabase().DBaaS().Get(ctx, projectID, dbaasID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading DBaaS instance",
			fmt.Sprintf("Unable to read DBaaS instance: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read DBaaS instance"
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
		dbaas := response.Data

		if dbaas.Metadata.ID != nil {
			data.Id = types.StringValue(*dbaas.Metadata.ID)
		}
		if dbaas.Metadata.URI != nil {
			data.Uri = types.StringValue(*dbaas.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if dbaas.Metadata.Name != nil {
			data.Name = types.StringValue(*dbaas.Metadata.Name)
		}
		if dbaas.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(dbaas.Metadata.LocationResponse.Value)
		}
		if dbaas.Properties.Engine != nil && dbaas.Properties.Engine.ID != nil {
			data.EngineID = types.StringValue(*dbaas.Properties.Engine.ID)
		}
		if dbaas.Properties.Flavor != nil && dbaas.Properties.Flavor.Name != nil {
			data.Flavor = types.StringValue(*dbaas.Properties.Flavor.Name)
		}

		// Update tags
		if len(dbaas.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(dbaas.Metadata.Tags))
			for i, tag := range dbaas.Metadata.Tags {
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

func (r *DBaaSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DBaaSResourceModel
	var state DBaaSResourceModel

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
	dbaasID := state.Id.ValueString()

	if projectID == "" || dbaasID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and DBaaS ID are required to update the DBaaS instance",
		)
		return
	}

	// Get current DBaaS instance details
	getResponse, err := r.client.Client.FromDatabase().DBaaS().Get(ctx, projectID, dbaasID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current DBaaS instance",
			fmt.Sprintf("Unable to get current DBaaS instance: %s", err),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"DBaaS Instance Not Found",
			"DBaaS instance not found or no data returned",
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
			"Unable to determine region value for DBaaS instance",
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

	// Build update request - only name and tags can be updated
	updateRequest := sdktypes.DBaaSRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.DBaaSPropertiesRequest{
			// Preserve current engine if it exists
			Engine: func() *sdktypes.DBaaSEngine {
				if current.Properties.Engine != nil {
					return &sdktypes.DBaaSEngine{
						ID: current.Properties.Engine.ID,
					}
				}
				return nil
			}(),
			// Preserve current flavor if it exists
			Flavor: func() *sdktypes.DBaaSFlavor {
				if current.Properties.Flavor != nil {
					return &sdktypes.DBaaSFlavor{
						Name: current.Properties.Flavor.Name,
					}
				}
				return nil
			}(),
		},
	}

	// Update the DBaaS instance using the SDK
	response, err := r.client.Client.FromDatabase().DBaaS().Update(ctx, projectID, dbaasID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating DBaaS instance",
			fmt.Sprintf("Unable to update DBaaS instance: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to update DBaaS instance"
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
	data.ProjectID = state.ProjectID

	if response != nil && response.Data != nil {
		// Update from response if available (should match state)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DBaaSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DBaaSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.Id.ValueString()

	if projectID == "" || dbaasID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and DBaaS ID are required to delete the DBaaS instance",
		)
		return
	}

	// Delete the DBaaS instance using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromDatabase().DBaaS().Delete(ctx, projectID, dbaasID, nil)
		},
		ExtractSDKError,
		"DBaaS",
		dbaasID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting DBaaS instance",
			fmt.Sprintf("Unable to delete DBaaS instance: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a DBaaS resource", map[string]interface{}{
		"dbaas_id": dbaasID,
	})
}

func (r *DBaaSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
