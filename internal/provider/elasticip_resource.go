package provider

import (
	"context"
	"fmt"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &ElasticIPResource{}
var _ resource.ResourceWithImportState = &ElasticIPResource{}

func NewElasticIPResource() resource.Resource {
	return &ElasticIPResource{}
}

type ElasticIPResource struct {
	client *ArubaCloudClient
}

type ElasticIPResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	Address       types.String `tfsdk:"address"`
	ProjectId     types.String `tfsdk:"project_id"`
}

func (r *ElasticIPResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticip"
}

func (r *ElasticIPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an ArubaCloud Elastic IP — a static public IPv4 address.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the Elastic IP.",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier (e.g., `ITBG-Bergamo`). Changing this value forces a new resource.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"tags": schema.ListAttribute{
				ElementType: types.StringType, MarkdownDescription: "List of string tags.", Optional: true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle. Accepted values: `Hour`, `Month`, `Year`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Validators:          []validator.String{stringvalidator.OneOf("Hour", "Month", "Year")},
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. The assigned public IP address.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource. Changing this value forces a new resource.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *ElasticIPResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ArubaCloudClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *ArubaCloudClient, got: %T.", req.ProviderData))
		return
	}
	r.client = client
}

func (r *ElasticIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ElasticIPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	billingPeriod := "Hour"
	if !data.BillingPeriod.IsNull() && data.BillingPeriod.ValueString() != "" {
		billingPeriod = data.BillingPeriod.ValueString()
	}

	eip, err := r.client.Client.FromNetwork().ElasticIPs().Create(ctx,
		aruba.NewElasticIP().
			Named(data.Name.ValueString()).
			InProject(aruba.URI("/projects/"+data.ProjectId.ValueString())).
			InRegion(aruba.Region(data.Location.ValueString())).
			BilledBy(aruba.BillingPeriod(billingPeriod)).
			Tagged(tags...),
	)
	if provErr := CheckResponseErr("create", "ElasticIP", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(eip.ID())
	data.Uri = strVal(eip.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if waitErr := eip.WaitUntilReady(ctx, aruba.WithTimeout(r.client.ResourceTimeout)); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "ElasticIP", data.Id.ValueString())
		return
	}

	fresh, freshErr := r.client.Client.FromNetwork().ElasticIPs().Get(ctx, aruba.URI(eip.URI()))
	if freshErr == nil {
		applyEIPToModel(fresh, &data)
	}

	tflog.Trace(ctx, "created an ElasticIP resource", map[string]interface{}{"eip_id": data.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func eipRef(data *ElasticIPResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.URI("/projects/" + data.ProjectId.ValueString() + "/network/elasticIPs/" + data.Id.ValueString())
}

func applyEIPToModel(eip *aruba.ElasticIP, data *ElasticIPResourceModel) {
	data.Id = types.StringValue(eip.ID())
	data.Uri = strVal(eip.URI())
	data.Name = types.StringValue(eip.Name())
	data.Tags = TagsToListPreserveNull(eip.Tags(), data.Tags)
	data.BillingPeriod = strVal(string(eip.BillingPeriod()))
	data.Address = strVal(eip.Address())
	raw := eip.Raw()
	if raw != nil && raw.Metadata.LocationResponse != nil {
		data.Location = types.StringValue(string(raw.Metadata.LocationResponse.Value))
	}
}

func (r *ElasticIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ElasticIPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	eip, err := r.client.Client.FromNetwork().ElasticIPs().Get(ctx, eipRef(&data))
	if provErr := CheckResponseErr("read", "ElasticIP", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	st := string(eip.State())
	if isFailedState(st) {
		resp.Diagnostics.AddWarning("Resource in Failed State",
			fmt.Sprintf("ElasticIP %q is in a terminal failure state (%s).", data.Id.ValueString(), st))
	} else if IsCreatingState(st) {
		if waitErr := eip.WaitUntilReady(ctx, aruba.WithTimeout(r.client.ResourceTimeout)); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "ElasticIP", data.Id.ValueString())
			return
		}
		eip, err = r.client.Client.FromNetwork().ElasticIPs().Get(ctx, eipRef(&data))
		if provErr := CheckResponseErr("read", "ElasticIP", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	projectID := data.ProjectId
	applyEIPToModel(eip, &data)
	data.ProjectId = projectID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ElasticIPResourceModel
	var state ElasticIPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	eip, err := r.client.Client.FromNetwork().ElasticIPs().Get(ctx, eipRef(&state))
	if provErr := CheckResponseErr("read", "ElasticIP", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	eip.Named(data.Name.ValueString())
	if tags != nil {
		eip.RetaggedAs(tags...)
	} else {
		eip.RetaggedAs(eip.Tags()...)
	}
	if !data.BillingPeriod.IsNull() && data.BillingPeriod.ValueString() != "" {
		eip.BilledBy(aruba.BillingPeriod(data.BillingPeriod.ValueString()))
	}

	updated, err := r.client.Client.FromNetwork().ElasticIPs().Update(ctx, eip)
	if provErr := CheckResponseErr("update", "ElasticIP", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = state.Id
	data.Uri = state.Uri
	data.ProjectId = state.ProjectId
	data.Location = state.Location
	data.Address = state.Address
	data.Name = types.StringValue(updated.Name())
	data.Tags = TagsToListPreserveNull(updated.Tags(), data.Tags)
	data.BillingPeriod = strVal(string(updated.BillingPeriod()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ElasticIPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := eipRef(&data)
	eipID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromNetwork().ElasticIPs().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "ElasticIP", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		delErr := r.client.Client.FromNetwork().ElasticIPs().Delete(ctx, ref)
		return CheckResponseErrAsError("delete", "ElasticIP", delErr)
	}, "ElasticIP", eipID, r.client.ResourceTimeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting ElasticIP", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "ElasticIP", eipID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for ElasticIP deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted an ElasticIP resource", map[string]interface{}{"eip_id": eipID})
}

func (r *ElasticIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
