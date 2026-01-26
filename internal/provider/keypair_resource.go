package provider

import (
	"context"
	"fmt"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type KeypairResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	ProjectID types.String `tfsdk:"project_id"`
	Value     types.String `tfsdk:"value"`
	Tags      types.List   `tfsdk:"tags"`
}

type KeypairResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &KeypairResource{}
var _ resource.ResourceWithImportState = &KeypairResource{}

func NewKeypairResource() resource.Resource {
	return &KeypairResource{}
}

func (r *KeypairResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keypair"
}

func (r *KeypairResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Keypair resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Keypair identifier (name)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Keypair URI",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Keypair name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Keypair location",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Public key value",
				Required:            true,
				Sensitive:           true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the keypair",
				Optional:            true,
			},
		},
	}
}

func (r *KeypairResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KeypairResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KeypairResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"Project ID is required to create a keypair",
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
	createRequest := sdktypes.KeyPairRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.KeyPairPropertiesRequest{
			Value: data.Value.ValueString(),
		},
	}

	// Create the keypair using the SDK
	response, err := r.client.Client.FromCompute().KeyPairs().Create(ctx, projectID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating keypair",
			fmt.Sprintf("Unable to create keypair: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		logContext := map[string]interface{}{
			"keypair_name": data.Name.ValueString(),
			"project_id":   projectID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to create keypair", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		// Get ID from Metadata.ID (like other resources)
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		} else {
			resp.Diagnostics.AddError(
				"Invalid API Response",
				"Keypair created but ID is missing from response",
			)
			return
		}

		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Keypair created but no data returned from API",
		)
		return
	}

	// Wait for Keypair to be active before returning (Keypair is referenced by CloudServer)
	// This ensures Terraform doesn't proceed to create dependent resources until Keypair is ready
	keypairID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromCompute().KeyPairs().Get(ctx, projectID, keypairID, nil)
		if err != nil {
			return "", err
		}
		// Keypairs don't have a Status field - if we can get it, it's ready
		if getResp != nil && getResp.Data != nil {
			return "Active", nil
		}
		return "Unknown", nil
	}

	// Wait for Keypair to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "Keypair", keypairID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Keypair Not Active",
			fmt.Sprintf("Keypair was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a Keypair resource", map[string]interface{}{
		"keypair_id":   data.Id.ValueString(),
		"keypair_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeypairResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KeypairResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get project ID and keypair ID from state
	projectID := data.ProjectID.ValueString()
	keypairID := data.Id.ValueString()

	// If ID is unknown or null, check if this is a new resource (no state) or existing resource (state exists but ID missing)
	// For new resources (during plan), we can return early
	// For existing resources, we need the ID to read - if it's missing, that's an error
	if data.Id.IsUnknown() || data.Id.IsNull() || keypairID == "" {
		// Check if ProjectID is also unknown - if so, this is definitely a new resource
		if data.ProjectID.IsUnknown() || data.ProjectID.IsNull() {
			tflog.Info(ctx, "Keypair ID and Project ID are unknown or null during read, skipping API call (likely new resource).")
			return // Do not error, as this is expected during plan for new resources
		}
		// If ProjectID is set but ID is unknown, still skip (new resource)
		if keypairID == "" {
			tflog.Info(ctx, "Keypair ID is unknown or null during read, skipping API call (likely new resource).")
			return // Do not error, as this is expected during plan for new resources
		}
	}

	// If ProjectID is missing, we can't read the keypair
	if projectID == "" {
		// Check if ProjectID is unknown (new resource) vs missing (error)
		if data.ProjectID.IsUnknown() || data.ProjectID.IsNull() {
			tflog.Info(ctx, "Keypair Project ID is unknown or null during read, skipping API call (likely new resource).")
			return // Do not error, as this is expected during plan for new resources
		}
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID is required to read the keypair",
		)
		return
	}

	// Get keypair details using the SDK
	// The API Get method accepts the keypair ID
	tflog.Debug(ctx, "Reading keypair", map[string]interface{}{
		"project_id":   projectID,
		"keypair_id":   keypairID,
		"keypair_name": data.Name.ValueString(),
		"keypair_uri":  data.Uri.ValueString(),
	})

	response, err := r.client.Client.FromCompute().KeyPairs().Get(ctx, projectID, keypairID, nil)
	if err != nil {
		tflog.Error(ctx, "Error calling keypair Get API", map[string]interface{}{
			"error":      err,
			"project_id": projectID,
			"keypair_id": keypairID,
		})
		resp.Diagnostics.AddError(
			"Error reading keypair",
			fmt.Sprintf("Unable to read keypair: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			tflog.Info(ctx, "Keypair not found (404), removing from state", map[string]interface{}{
				"project_id": projectID,
				"keypair_id": keypairID,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		logContext := map[string]interface{}{
			"project_id": projectID,
			"keypair_id": keypairID,
		}
		errorMsg := FormatAPIError(ctx, response.Error, "Failed to read keypair", logContext)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		keypair := response.Data

		// Preserve ProjectID and Value from state (they're not in the API response)
		projectIDFromState := data.ProjectID
		valueFromState := data.Value
		idFromState := data.Id

		// Get ID from Metadata.ID (like other resources)
		if keypair.Metadata.ID != nil {
			data.Id = types.StringValue(*keypair.Metadata.ID)
		} else {
			// If API doesn't provide ID, preserve from state
			data.Id = idFromState
		}

		if keypair.Metadata.Name != nil {
			data.Name = types.StringValue(*keypair.Metadata.Name)
		}
		if keypair.Metadata.URI != nil {
			data.Uri = types.StringValue(*keypair.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if keypair.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(keypair.Metadata.LocationResponse.Value)
		}

		// Update tags from response
		if len(keypair.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(keypair.Metadata.Tags))
			for i, tag := range keypair.Metadata.Tags {
				tagValues[i] = types.StringValue(tag)
			}
			tagsList, diags := types.ListValueFrom(ctx, types.StringType, tagValues)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = tagsList
			}
		} else {
			// Set tags to null when empty to match state (prevents false changes)
			data.Tags = types.ListNull(types.StringType)
		}

		// Restore ProjectID and Value from state (they're not returned by the API)
		data.ProjectID = projectIDFromState
		data.Value = valueFromState
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeypairResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KeypairResourceModel
	var state KeypairResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use IDs from state (they are immutable)
	projectID := state.ProjectID.ValueString()
	keypairID := state.Id.ValueString()

	if projectID == "" || keypairID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Keypair ID are required to update the keypair",
		)
		return
	}

	// Keypair update is not supported by the API
	// Check if immutable fields changed
	if !data.Name.Equal(state.Name) {
		resp.Diagnostics.AddError(
			"Keypair Name Update Not Supported",
			"Changing the keypair name is not supported by the API. Please delete and recreate the keypair with the new name.",
		)
		return
	}

	if !data.Value.Equal(state.Value) {
		resp.Diagnostics.AddError(
			"Keypair Public Key Update Not Supported",
			"Changing the public key value is not supported by the API. Please delete and recreate the keypair with the new public key.",
		)
		return
	}

	if !data.Location.Equal(state.Location) {
		resp.Diagnostics.AddError(
			"Keypair Location Update Not Supported",
			"Changing the keypair location is not supported by the API. Please delete and recreate the keypair in the new location.",
		)
		return
	}

	// Since updates aren't supported, we just read the resource to refresh state
	// This ensures URI and other computed fields are up to date
	getResponse, err := r.client.Client.FromCompute().KeyPairs().Get(ctx, projectID, keypairID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading keypair",
			fmt.Sprintf("Unable to read keypair: %s", err),
		)
		return
	}

	if getResponse == nil || getResponse.IsError() || getResponse.Data == nil {
		if getResponse != nil && getResponse.StatusCode == 404 {
			tflog.Info(ctx, "Keypair not found (404), removing from state", map[string]interface{}{
				"project_id": projectID,
				"keypair_id": keypairID,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Keypair Not Found",
			"Keypair not found or no data returned",
		)
		return
	}

	keypair := getResponse.Data

	// Ensure immutable fields are set from state before saving
	data.Id = state.Id
	data.ProjectID = state.ProjectID
	data.Location = state.Location
	data.Value = state.Value

	// Update URI from API response
	if keypair.Metadata.URI != nil {
		data.Uri = types.StringValue(*keypair.Metadata.URI)
	} else {
		data.Uri = state.Uri
	}

	// Tags cannot be updated via API, so preserve tags from state
	// (The API doesn't support tag updates, so we keep what's in state)
	data.Tags = state.Tags

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeypairResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KeypairResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	keypairID := data.Id.ValueString()

	if projectID == "" || keypairID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Keypair ID are required to delete the keypair",
		)
		return
	}

	// Delete the keypair using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromCompute().KeyPairs().Delete(ctx, projectID, keypairID, nil)
		},
		ExtractSDKError,
		"Keypair",
		keypairID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting keypair",
			fmt.Sprintf("Unable to delete keypair: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a Keypair resource", map[string]interface{}{
		"keypair_id": keypairID,
	})
}

func (r *KeypairResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
