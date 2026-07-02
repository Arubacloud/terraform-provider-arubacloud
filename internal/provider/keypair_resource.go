package provider

import (
	"context"
	"fmt"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
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
	Timeout   types.String `tfsdk:"timeout"`
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
		MarkdownDescription: "Manages an ArubaCloud SSH KeyPair.",
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
				MarkdownDescription: "Display name for the KeyPair. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
			"value": schema.StringAttribute{
				MarkdownDescription: "OpenSSH-format public key string (e.g., `ssh-rsa AAAA...`). The provider uploads this to ArubaCloud; the corresponding private key is never stored. Write-only — this value is sent to the API but is not returned in subsequent read responses. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Optional:            true,
			},
			"timeout": schema.StringAttribute{
				MarkdownDescription: "Per-resource timeout override (e.g. `\"15m\"`, `\"1h\"`). Overrides the provider-level `resource_timeout` for this resource's Create and Delete operations. Uses Go duration syntax.",
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
		resp.Diagnostics.AddError("Missing Project ID", "Project ID is required to create a keypair")
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	builder := aruba.NewKeyPair().
		Named(data.Name.ValueString()).
		InProject(aruba.URI("/projects/" + projectID)).
		InRegion(aruba.Region(data.Location.ValueString())).
		WithPublicKey(data.Value.ValueString()).
		Tagged(tags...)

	kp, err := r.client.Client.FromCompute().KeyPairs().Create(ctx, builder)
	if provErr := CheckResponseErr("create", "Keypair", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(kp.ID())
	if uri := kp.URI(); uri != "" {
		data.Uri = types.StringValue(uri)
	} else {
		data.Uri = types.StringNull()
	}

	keypairID := data.Id.ValueString()

	// KeyPairs may go through a short provisioning phase; wait until ready.
	if err := kp.WaitUntilReady(ctx, sdkWaitOptions(effectiveTimeout(data.Timeout, r.client.ResourceTimeout))...); err != nil {
		ReportWaitResult(&resp.Diagnostics, err, "Keypair", keypairID)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	tflog.Trace(ctx, "created a Keypair resource", map[string]interface{}{
		"keypair_id":   keypairID,
		"keypair_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// keypairRef returns the Ref to use for Get/Update/Delete.
// Uses the stored URI when available (normal flow); falls back to a constructed URI for the import flow.
func keypairRef(data *KeypairResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.URI("/projects/" + data.ProjectID.ValueString() + "/compute/keyPairs/" + data.Id.ValueString())
}

func (r *KeypairResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KeypairResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id.IsUnknown() || data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Debug(ctx, "Reading keypair", map[string]interface{}{
		"project_id":  data.ProjectID.ValueString(),
		"keypair_id":  data.Id.ValueString(),
		"keypair_uri": data.Uri.ValueString(),
	})

	kp, err := r.client.Client.FromCompute().KeyPairs().Get(ctx, keypairRef(&data))
	if provErr := CheckResponseErr("read", "Keypair", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	// Preserve write-only / non-returned fields from state.
	projectIDFromState := data.ProjectID
	valueFromState := data.Value

	data.Id = types.StringValue(kp.ID())
	data.Name = types.StringValue(kp.Name())
	if uri := kp.URI(); uri != "" {
		data.Uri = types.StringValue(uri)
	} else {
		data.Uri = types.StringNull()
	}

	raw := kp.Raw()
	if raw != nil && raw.Metadata.LocationResponse != nil {
		data.Location = types.StringValue(string(raw.Metadata.LocationResponse.Value))
	}

	data.Tags = TagsToListPreserveNull(kp.Tags(), data.Tags)

	data.ProjectID = projectIDFromState
	data.Value = valueFromState

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

	// Refresh state via Get so URI and computed fields stay current.
	// name, value, location, and project_id all have RequiresReplace — the only
	// mutable field that can trigger Update is tags.
	kp, err := r.client.Client.FromCompute().KeyPairs().Get(ctx, keypairRef(&state))
	if provErr := CheckResponseErr("read", "Keypair", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	// Preserve immutable and write-only fields from state.
	data.Id = state.Id
	data.ProjectID = state.ProjectID
	data.Location = state.Location
	data.Value = state.Value
	if uri := kp.URI(); uri != "" {
		data.Uri = types.StringValue(uri)
	} else {
		data.Uri = state.Uri
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeypairResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KeypairResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keypairID := data.Id.ValueString()
	if keypairID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Keypair ID is required to delete the keypair")
		return
	}

	ref := keypairRef(&data)

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromCompute().KeyPairs().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Keypair", getErr); provErr != nil {
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
			delErr := r.client.Client.FromCompute().KeyPairs().Delete(ctx, ref)
			return CheckResponseErrAsError("delete", "Keypair", delErr)
		},
		"Keypair",
		keypairID,
		effectiveTimeout(data.Timeout, r.client.ResourceTimeout),
		deletionChecker,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting keypair", err.Error())
		return
	}

	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "Keypair", keypairID, remainingTimeout(deleteStart, effectiveTimeout(data.Timeout, r.client.ResourceTimeout))); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for Keypair deletion", waitErr.Error())
		return
	}

	tflog.Trace(ctx, "deleted a Keypair resource", map[string]interface{}{"keypair_id": keypairID})
}

func (r *KeypairResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseImportID(req.ID, "<project_id>/<keypair_id>", "proj-abc/kp-xyz", 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
