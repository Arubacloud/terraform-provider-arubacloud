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

type KMSResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	ProjectID     types.String `tfsdk:"project_id"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	BillingPeriod types.String `tfsdk:"billing_period"`
}

type KMSResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &KMSResource{}
var _ resource.ResourceWithImportState = &KMSResource{}

func NewKMSResource() resource.Resource {
	return &KMSResource{}
}

func (r *KMSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kms"
}

func (r *KMSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an ArubaCloud KMS (Key Management Service) instance for storing and managing encryption keys.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the KMS instance.",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Optional:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle. Accepted values: `Hour`, `Month`, `Year`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *KMSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KMSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KMSResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()

	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID is required to create a KMS",
		)
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the create request
	location := "ITBG-Bergamo" // Default location
	if !data.Location.IsNull() && !data.Location.IsUnknown() {
		location = data.Location.ValueString()
	}

	// Use default or user-provided billing period
	billingPeriod := "Hour"
	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		billingPeriod = data.BillingPeriod.ValueString()
	}

	createRequest := sdktypes.KmsRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: location,
			},
		},
		Properties: sdktypes.KmsPropertiesRequest{
			BillingPeriod: billingPeriod,
		},
	}

	// Create the KMS using the SDK
	response, err := r.client.Client.FromSecurity().KMS().Create(ctx, projectID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating KMS",
			NewTransportError("create", "Kms", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("create", "Kms", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}

	if response != nil && response.Data != nil {
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		} else {
			resp.Diagnostics.AddError(
				"Invalid API Response",
				"KMS created but no ID returned from API",
			)
			return
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		// Set other fields from create response
		if response.Data.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(response.Data.Metadata.LocationResponse.Value)
		}
		if response.Data.Properties.BillingPeriod != "" {
			data.BillingPeriod = types.StringValue(response.Data.Properties.BillingPeriod)
		} else if data.BillingPeriod.IsNull() || data.BillingPeriod.IsUnknown() {
			data.BillingPeriod = types.StringValue("Hour")
		}
		// Tags are preserved from plan - no need to set from response in Create
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"KMS created but no data returned from API",
		)
		return
	}

	// Wait for KMS to be active before returning
	// This ensures Terraform doesn't proceed until KMS is ready
	kmsID := data.Id.ValueString()
	if kmsID == "" {
		resp.Diagnostics.AddError(
			"Missing KMS ID",
			"KMS ID is required but was not set",
		)
		return
	}
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromSecurity().KMS().Get(ctx, projectID, kmsID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for KMS to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "KMS", kmsID, r.client.ResourceTimeout); err != nil {
		ReportWaitResult(&resp.Diagnostics, err, "KMS", kmsID)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	tflog.Trace(ctx, "created a KMS resource", map[string]interface{}{
		"kms_id":   data.Id.ValueString(),
		"kms_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KMSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KMSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	kmsID := data.Id.ValueString()

	if data.Id.IsUnknown() || data.Id.IsNull() || kmsID == "" {
		tflog.Debug(ctx, "KMS ID is empty, removing resource from state", map[string]interface{}{"kms_id": kmsID})
		resp.State.RemoveResource(ctx)
		return
	}

	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID is required to read the KMS",
		)
		return
	}

	// Get KMS details using the SDK
	response, err := r.client.Client.FromSecurity().KMS().Get(ctx, projectID, kmsID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading KMS",
			NewTransportError("read", "Kms", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("read", "Kms", response); apiErr != nil {
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
				fmt.Sprintf("KMS %q reached a terminal failure state (%s) and will not recover on its own. "+
					"Use `terraform apply -replace=<address>` to recreate it.", kmsID, st),
			)
			return
		case IsCreatingState(st):
			checker := func(ctx context.Context) (string, error) {
				getResp, err := r.client.Client.FromSecurity().KMS().Get(ctx, projectID, kmsID, nil)
				if err != nil {
					return "", err
				}
				if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
					return *getResp.Data.Status.State, nil
				}
				return "Unknown", nil
			}
			if err := WaitForResourceActive(ctx, checker, "KMS", kmsID, r.client.ResourceTimeout); err != nil {
				ReportWaitResult(&resp.Diagnostics, err, "KMS", kmsID)
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
			// Re-read to get the final active state.
			response, err = r.client.Client.FromSecurity().KMS().Get(ctx, projectID, kmsID, nil)
			if err != nil {
				resp.Diagnostics.AddError("Error reading KMS after provisioning wait",
					NewTransportError("read", "Kms", err).Error())
				return
			}
			if apiErr := CheckResponse("read", "Kms", response); apiErr != nil {
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
		kms := response.Data
		if kms.Metadata.ID != nil {
			data.Id = types.StringValue(*kms.Metadata.ID)
		}
		if kms.Metadata.URI != nil {
			data.Uri = types.StringValue(*kms.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if kms.Metadata.Name != nil {
			data.Name = types.StringValue(*kms.Metadata.Name)
		}
		if kms.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(kms.Metadata.LocationResponse.Value)
		}
		data.Tags = TagsToListPreserveNull(kms.Metadata.Tags, data.Tags)
		if kms.Properties.BillingPeriod != "" {
			data.BillingPeriod = types.StringValue(kms.Properties.BillingPeriod)
		}
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KMSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KMSResourceModel
	var state KMSResourceModel

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
	kmsID := state.Id.ValueString()

	if projectID == "" || kmsID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and KMS ID are required to update the KMS",
		)
		return
	}

	// Get current KMS to preserve fields
	getResp, err := r.client.Client.FromSecurity().KMS().Get(ctx, projectID, kmsID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting KMS",
			NewTransportError("read", "Kms", err).Error(),
		)
		return
	}

	if getResp == nil || getResp.Data == nil {
		resp.Diagnostics.AddError(
			"KMS Not Found",
			"KMS not found",
		)
		return
	}

	current := getResp.Data
	regionValue := ""
	if current.Metadata.LocationResponse != nil {
		regionValue = current.Metadata.LocationResponse.Value
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if tags == nil {
		tags = current.Metadata.Tags
	}

	// Build update request
	updateRequest := sdktypes.KmsRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.KmsPropertiesRequest{
			BillingPeriod: data.BillingPeriod.ValueString(),
		},
	}

	// Update the KMS using the SDK
	response, err := r.client.Client.FromSecurity().KMS().Update(ctx, projectID, kmsID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating KMS",
			NewTransportError("update", "Kms", err).Error(),
		)
		return
	}

	if apiErr := CheckResponse("update", "Kms", response); apiErr != nil {
		resp.Diagnostics.AddError("API Error", apiErr.Error())
		return
	}

	if response != nil && response.Data != nil {
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
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

func (r *KMSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KMSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	kmsID := data.Id.ValueString()

	// If ID is unknown or empty, the resource doesn't exist (e.g., during plan or if never created)
	// Return early without error - this is expected behavior
	if data.Id.IsUnknown() || data.Id.IsNull() || kmsID == "" {
		tflog.Debug(ctx, "KMS ID is unknown or empty, skipping delete", map[string]interface{}{
			"kms_id": kmsID,
		})
		return
	}

	// Project ID should always be set, but check to be safe
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID is required to delete the KMS",
		)
		return
	}

	// Delete the KMS using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	deletionChecker := func(ctx context.Context) (bool, error) {
		getResp, getErr := r.client.Client.FromSecurity().KMS().Get(ctx, projectID, kmsID, nil)
		if getErr != nil {
			return false, NewTransportError("get", "KMS", getErr)
		}
		if provErr := CheckResponse("get", "KMS", getResp); provErr != nil {
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
			resp, err := r.client.Client.FromSecurity().KMS().Delete(ctx, projectID, kmsID, nil)
			if err != nil {
				return NewTransportError("delete", "KMS", err)
			}
			return CheckResponse("delete", "KMS", resp)
		},
		"KMS",
		kmsID,
		r.client.ResourceTimeout,
		deletionChecker,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting KMS",
			NewTransportError("delete", "Kms", err).Error(),
		)
		return
	}

	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "KMS", kmsID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for KMS deletion", waitErr.Error())
		return
	}

	tflog.Trace(ctx, "deleted a KMS resource", map[string]interface{}{
		"kms_id": kmsID,
	})
}

func (r *KMSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
