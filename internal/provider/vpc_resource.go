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

type VPCResource struct {
	client *ArubaCloudClient
}

type VPCResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
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

func (r *VPCResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (r *VPCResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an ArubaCloud VPC (Virtual Private Cloud) — the isolated network boundary within a region where subnets, security groups, and server instances are provisioned.",
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
				MarkdownDescription: "Display name for the VPC.",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier for the resource (e.g., `ITBG-Bergamo`). Changing this value forces a new resource.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource. Changing this value forces a new resource.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
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
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ArubaCloudClient, got: %T.", req.ProviderData))
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
	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	vpc, err := r.client.Client.FromNetwork().VPCs().Create(ctx,
		aruba.NewVPC().
			Named(data.Name.ValueString()).
			InProject(aruba.URI("/projects/"+projectID)).
			InRegion(aruba.Region(data.Location.ValueString())).
			NotDefault().
			WithoutPreset().
			Tagged(tags...),
	)
	if provErr := CheckResponseErr("create", "VPC", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(vpc.ID())
	data.Uri = strVal(vpc.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if waitErr := vpc.WaitUntilReady(ctx, aruba.WithTimeout(r.client.ResourceTimeout)); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "VPC", data.Id.ValueString())
		return
	}

	// Re-read to get server-assigned fields.
	fresh, freshErr := r.client.Client.FromNetwork().VPCs().Get(ctx, aruba.URI(vpc.URI()))
	if freshErr == nil {
		applyVPCToModel(fresh, &data)
	} else {
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh VPC after creation: %v", freshErr))
	}

	tflog.Trace(ctx, "created a VPC resource", map[string]interface{}{"vpc_id": data.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func vpcRef(data *VPCResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.URI("/projects/" + data.ProjectID.ValueString() + "/network/vpcs/" + data.Id.ValueString())
}

func applyVPCToModel(vpc *aruba.VPC, data *VPCResourceModel) {
	data.Id = types.StringValue(vpc.ID())
	data.Uri = strVal(vpc.URI())
	data.Name = types.StringValue(vpc.Name())
	data.Tags = TagsToListPreserveNull(vpc.Tags(), data.Tags)
	raw := vpc.Raw()
	if raw != nil && raw.Metadata.LocationResponse != nil {
		data.Location = types.StringValue(string(raw.Metadata.LocationResponse.Value))
	}
}

func (r *VPCResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VPCResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	vpc, err := r.client.Client.FromNetwork().VPCs().Get(ctx, vpcRef(&data))
	if provErr := CheckResponseErr("read", "VPC", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	st := string(vpc.State())
	switch {
	case isFailedState(st):
		resp.Diagnostics.AddWarning("Resource in Failed State",
			fmt.Sprintf("VPC %q is in a terminal failure state (%s).", data.Id.ValueString(), st))
	case IsCreatingState(st):
		if waitErr := vpc.WaitUntilReady(ctx, aruba.WithTimeout(r.client.ResourceTimeout)); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "VPC", data.Id.ValueString())
			return
		}
		vpc, err = r.client.Client.FromNetwork().VPCs().Get(ctx, vpcRef(&data))
		if provErr := CheckResponseErr("read", "VPC", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	projectID := data.ProjectID // preserve
	applyVPCToModel(vpc, &data)
	data.ProjectID = projectID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPCResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VPCResourceModel
	var state VPCResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	vpc, err := r.client.Client.FromNetwork().VPCs().Get(ctx, vpcRef(&state))
	if provErr := CheckResponseErr("read", "VPC", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	vpc.Named(data.Name.ValueString())
	if tags != nil {
		vpc.RetaggedAs(tags...)
	} else {
		vpc.RetaggedAs(vpc.Tags()...)
	}

	updated, err := r.client.Client.FromNetwork().VPCs().Update(ctx, vpc)
	if provErr := CheckResponseErr("update", "VPC", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = state.Id
	data.Uri = state.Uri
	data.ProjectID = state.ProjectID
	data.Name = types.StringValue(updated.Name())
	data.Tags = TagsToListPreserveNull(updated.Tags(), data.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPCResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VPCResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := vpcRef(&data)
	vpcID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromNetwork().VPCs().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "VPC", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		delErr := r.client.Client.FromNetwork().VPCs().Delete(ctx, ref)
		return CheckResponseErrAsError("delete", "VPC", delErr)
	}, "VPC", vpcID, r.client.ResourceTimeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting VPC", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "VPC", vpcID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for VPC deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a VPC resource", map[string]interface{}{"vpc_id": vpcID})
}

func (r *VPCResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// strVal converts an API string to types.StringValue (non-empty) or types.StringNull.
func strVal(s string) types.String {
	if s != "" {
		return types.StringValue(s)
	}
	return types.StringNull()
}
