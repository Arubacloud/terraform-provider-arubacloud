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
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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

func kmsRef(data *KMSResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.URI("/projects/" + data.ProjectID.ValueString() +
		"/providers/Aruba.Security/kms/" + data.Id.ValueString())
}

func applyKMSToModel(kms *aruba.KMS, data *KMSResourceModel) {
	data.Id = types.StringValue(kms.ID())
	if uri := kms.URI(); uri != "" {
		data.Uri = types.StringValue(uri)
	} else {
		data.Uri = types.StringNull()
	}
	data.Name = types.StringValue(kms.Name())
	data.Tags = TagsToListPreserveNull(kms.Tags(), data.Tags)
	if r := string(kms.Region()); r != "" {
		data.Location = types.StringValue(r)
	}
	if bp := string(kms.BillingPeriod()); bp != "" {
		data.BillingPeriod = types.StringValue(bp)
	}
}

func (r *KMSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KMSResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Project ID is required to create a KMS")
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	builder := aruba.NewKMS().
		Named(data.Name.ValueString()).
		InProject(aruba.URI("/projects/" + projectID)).
		Tagged(tags...)

	if !data.Location.IsNull() && !data.Location.IsUnknown() {
		builder = builder.InRegion(aruba.Region(data.Location.ValueString()))
	}
	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		builder = builder.BilledBy(aruba.BillingPeriod(data.BillingPeriod.ValueString()))
	}

	kms, err := r.client.Client.FromSecurity().KMS().Create(ctx, builder)
	if provErr := CheckResponseErr("create", "KMS", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	applyKMSToModel(kms, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if waitErr := kms.WaitUntilReady(ctx, aruba.WithTimeout(r.client.ResourceTimeout)); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "KMS", data.Id.ValueString())
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

	if data.Id.IsUnknown() || data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	kms, err := r.client.Client.FromSecurity().KMS().Get(ctx, kmsRef(&data))
	if provErr := CheckResponseErr("read", "KMS", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	st := string(kms.State())
	switch {
	case isFailedState(st):
		resp.Diagnostics.AddWarning("Resource in Failed State",
			fmt.Sprintf("KMS %q is in a terminal failure state (%s). "+
				"Run `terraform destroy` to clean it up, or `terraform apply -replace=<address>` to recreate it.", data.Id.ValueString(), st))
	case IsCreatingState(st):
		if waitErr := kms.WaitUntilReady(ctx, aruba.WithTimeout(r.client.ResourceTimeout)); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "KMS", data.Id.ValueString())
			return
		}
		kms, err = r.client.Client.FromSecurity().KMS().Get(ctx, kmsRef(&data))
		if provErr := CheckResponseErr("read", "KMS", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	applyKMSToModel(kms, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KMSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KMSResourceModel
	var state KMSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	kms, err := r.client.Client.FromSecurity().KMS().Get(ctx, kmsRef(&state))
	if provErr := CheckResponseErr("read", "KMS", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	kms.Named(data.Name.ValueString())
	if tags != nil {
		kms.RetaggedAs(tags...)
	} else {
		kms.RetaggedAs(kms.Tags()...)
	}
	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		kms.BilledBy(aruba.BillingPeriod(data.BillingPeriod.ValueString()))
	}

	updated, err := r.client.Client.FromSecurity().KMS().Update(ctx, kms)
	if provErr := CheckResponseErr("update", "KMS", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = state.Id
	data.ProjectID = state.ProjectID
	applyKMSToModel(updated, &data)
	data.Id = state.Id
	data.ProjectID = state.ProjectID

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KMSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KMSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id.IsUnknown() || data.Id.IsNull() || data.Id.ValueString() == "" {
		return
	}

	ref := kmsRef(&data)
	kmsID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromSecurity().KMS().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "KMS", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErr("delete", "KMS",
			r.client.Client.FromSecurity().KMS().Delete(ctx, ref))
	}, "KMS", kmsID, r.client.ResourceTimeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting KMS", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "KMS", kmsID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for KMS deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a KMS resource", map[string]interface{}{"kms_id": kmsID})
}

func (r *KMSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
