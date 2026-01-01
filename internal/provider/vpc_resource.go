// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"reflect"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type VPCResource struct {
	client *ArubaCloudClient
}

type VPCResourceModel struct {
	Id        types.String `tfsdk:"id"`
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
		MarkdownDescription: "VPC resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VPC identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "VPC name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "VPC location",
				Required:            true,
				// Validators removed for v1.16.1 compatibility
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the VPC",
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
			fmt.Sprintf("Unable to create VPC: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create VPC"
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
		resp.Diagnostics.AddError(
			"VPC Not Active",
			fmt.Sprintf("VPC was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a VPC resource", map[string]interface{}{
		"vpc_id": data.Id.ValueString(),
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

	if projectID == "" || vpcID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and VPC ID are required to read the VPC",
		)
		return
	}

	// Get VPC details using the SDK
	response, err := r.client.Client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading VPC",
			fmt.Sprintf("Unable to read VPC: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read VPC"
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
		vpc := response.Data

		if vpc.Metadata.ID != nil {
			data.Id = types.StringValue(*vpc.Metadata.ID)
		}
		if vpc.Metadata.Name != nil {
			data.Name = types.StringValue(*vpc.Metadata.Name)
		}
		if vpc.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(vpc.Metadata.LocationResponse.Value)
		}

		// Update tags from response
		if len(vpc.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(vpc.Metadata.Tags))
			for i, tag := range vpc.Metadata.Tags {
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

	projectID := data.ProjectID.ValueString()
	vpcID := data.Id.ValueString()

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
			fmt.Sprintf("Unable to get current VPC: %s", err),
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
			fmt.Sprintf("Unable to update VPC: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to update VPC"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
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

	// Delete the VPC using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	// The response from VPCs().Delete() has StatusCode and Error fields directly accessible
	deleteFunc := func() (interface{}, error) {
		return r.client.Client.FromNetwork().VPCs().Delete(ctx, projectID, vpcID, nil)
	}

	// Extract error info - the response type has StatusCode and Error as fields
	// The response from VPCs().Delete() is the same type as in Read/Update methods
	// which access response.StatusCode and response.Error.Title directly
	extractErrorFunc := func(response interface{}) (int, *string, *string, bool) {
		// Check if response has IsError() method
		type errorResponse interface {
			IsError() bool
		}
		resp, ok := response.(errorResponse)
		if !ok {
			return 0, nil, nil, false
		}
		if !resp.IsError() {
			return 0, nil, nil, false
		}

		// The response from deleteFunc is the actual SDK response type
		// which has StatusCode and Error as direct struct fields
		// We can access them by type asserting to the actual response type
		// Since deleteFunc returns the concrete type, we can access fields directly
		// The response type matches what we see in Read/Update: response.StatusCode, response.Error.Title
		// We'll access them using the response directly where we know the type
		// For VPC, the response is the same type as used elsewhere in this file
		// We can access response.StatusCode and response.Error.Title directly
		// by using the response type from the SDK
		return extractVPCDelError(response)
	}

	err := DeleteResourceWithRetry(
		ctx,
		deleteFunc,
		extractErrorFunc,
		"VPC",
		vpcID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting VPC",
			fmt.Sprintf("Unable to delete VPC: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a VPC resource", map[string]interface{}{
		"vpc_id": vpcID,
	})
}

// extractVPCDelError extracts error from VPC delete response using reflection
// The response from VPCs().Delete() has StatusCode and Error as direct struct fields
// We use reflection to access these fields since we have interface{}
func extractVPCDelError(response interface{}) (int, *string, *string, bool) {
	// Use reflection to access StatusCode and Error fields
	v := reflect.ValueOf(response)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Access StatusCode field
	statusCodeField := v.FieldByName("StatusCode")
	if !statusCodeField.IsValid() || !statusCodeField.CanInterface() {
		return 0, nil, nil, true // Return that it's an error, but can't extract details
	}
	statusCode := int(statusCodeField.Int())

	// Access Error field
	errorField := v.FieldByName("Error")
	if !errorField.IsValid() {
		return statusCode, nil, nil, true
	}

	// Access Error.Title and Error.Detail
	var errorTitle, errorDetail *string
	if errorField.Kind() == reflect.Struct || (errorField.Kind() == reflect.Ptr && errorField.Elem().Kind() == reflect.Struct) {
		errorVal := errorField
		if errorVal.Kind() == reflect.Ptr {
			errorVal = errorVal.Elem()
		}

		titleField := errorVal.FieldByName("Title")
		if titleField.IsValid() && titleField.CanInterface() && !titleField.IsNil() {
			if titlePtr, ok := titleField.Interface().(*string); ok {
				errorTitle = titlePtr
			}
		}

		detailField := errorVal.FieldByName("Detail")
		if detailField.IsValid() && detailField.CanInterface() && !detailField.IsNil() {
			if detailPtr, ok := detailField.Interface().(*string); ok {
				errorDetail = detailPtr
			}
		}
	}

	return statusCode, errorTitle, errorDetail, true
}

func (r *VPCResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *VPCResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}
