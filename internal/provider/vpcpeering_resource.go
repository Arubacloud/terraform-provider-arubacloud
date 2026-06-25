package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &VpcPeeringResource{}
var _ resource.ResourceWithImportState = &VpcPeeringResource{}

func NewVpcPeeringResource() resource.Resource {
	return &VpcPeeringResource{}
}

type VpcPeeringResource struct {
	client *ArubaCloudClient
}

type VpcPeeringResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	Tags      types.List   `tfsdk:"tags"`
	ProjectId types.String `tfsdk:"project_id"`
	VpcId     types.String `tfsdk:"vpc_id"`
	PeerVpc   types.String `tfsdk:"peer_vpc"`
}

func (r *VpcPeeringResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpcpeering"
}

func (r *VpcPeeringResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an ArubaCloud VPC Peering connection between two VPCs, enabling private IP routing between them without traversing the public internet. Both VPCs must exist in the same region. Use `arubacloud_vpcpeeringroute` to configure the routes within each peered VPC.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource.",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources (e.g., as a `*_uri_ref` attribute).",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the VPC peering.",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier for the resource (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "ID of the local VPC initiating this peering connection.",
				Required:            true,
			},
			"peer_vpc": schema.StringAttribute{
				MarkdownDescription: "ID or URI of the remote peer VPC to connect to.",
				Required:            true,
			},
		},
	}
}

func (r *VpcPeeringResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func vpcPeeringRef(data *VpcPeeringResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.VPCPeeringRef(data.ProjectId.ValueString(), data.VpcId.ValueString(), data.Id.ValueString())
}

func applyVPCPeeringToModel(p *aruba.VPCPeering, data *VpcPeeringResourceModel) {
	data.Id = types.StringValue(p.ID())
	data.Uri = strVal(p.URI())
	data.Name = types.StringValue(p.Name())
	data.Tags = TagsToListPreserveNull(p.Tags(), data.Tags)
	if p.RemoteVPCURI() != "" {
		data.PeerVpc = types.StringValue(p.RemoteVPCURI())
	}
	raw := p.Raw()
	if raw != nil && raw.Metadata.LocationResponse != nil {
		data.Location = types.StringValue(string(raw.Metadata.LocationResponse.Value))
	}
}

func (r *VpcPeeringResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VpcPeeringResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	vpcID := data.VpcId.ValueString()

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Normalise peer VPC value: if bare ID, construct the full URI.
	peerVPCURI := data.PeerVpc.ValueString()
	if !strings.HasPrefix(peerVPCURI, "/") {
		peerVPCURI = fmt.Sprintf("/projects/%s/providers/Aruba.Network/vpcs/%s", projectID, peerVPCURI)
	}

	vpcURI := aruba.URI("/projects/" + projectID + "/network/vpcs/" + vpcID)
	peering, err := r.client.Client.FromNetwork().VPCPeerings().Create(ctx,
		aruba.NewVPCPeering().
			Named(data.Name.ValueString()).
			InVPC(vpcURI).
			InRegion(aruba.Region(data.Location.ValueString())).
			Tagged(tags...).
			PeeredWith(aruba.URI(peerVPCURI)),
	)
	if provErr := CheckResponseErr("create", "VPCPeering", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(peering.ID())
	data.Uri = strVal(peering.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if waitErr := peering.WaitUntilReady(ctx, aruba.WithTimeout(r.client.ResourceTimeout)); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "VPCPeering", data.Id.ValueString())
		return
	}

	fresh, freshErr := r.client.Client.FromNetwork().VPCPeerings().Get(ctx, vpcPeeringRef(&data))
	if freshErr == nil {
		projectID := data.ProjectId
		vpcID := data.VpcId
		applyVPCPeeringToModel(fresh, &data)
		data.ProjectId = projectID
		data.VpcId = vpcID
	} else {
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh VPCPeering after creation: %v", freshErr))
	}

	tflog.Trace(ctx, "created a VPC Peering resource", map[string]interface{}{
		"vpcpeering_id":   data.Id.ValueString(),
		"vpcpeering_name": data.Name.ValueString(),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpcPeeringResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VpcPeeringResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	peering, err := r.client.Client.FromNetwork().VPCPeerings().Get(ctx, vpcPeeringRef(&data))
	if provErr := CheckResponseErr("read", "VPCPeering", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	st := string(peering.State())
	switch {
	case isFailedState(st):
		resp.Diagnostics.AddWarning("Resource in Failed State",
			fmt.Sprintf("VPCPeering %q is in a terminal failure state (%s). "+
				"Run `terraform destroy` to clean it up, or `terraform apply -replace=<address>` to recreate it.", data.Id.ValueString(), st))
	case IsCreatingState(st):
		if waitErr := peering.WaitUntilReady(ctx, aruba.WithTimeout(r.client.ResourceTimeout)); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "VPCPeering", data.Id.ValueString())
			return
		}
		peering, err = r.client.Client.FromNetwork().VPCPeerings().Get(ctx, vpcPeeringRef(&data))
		if provErr := CheckResponseErr("read", "VPCPeering", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	projectID := data.ProjectId
	vpcID := data.VpcId
	applyVPCPeeringToModel(peering, &data)
	data.ProjectId = projectID
	data.VpcId = vpcID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpcPeeringResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VpcPeeringResourceModel
	var state VpcPeeringResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	peering, err := r.client.Client.FromNetwork().VPCPeerings().Get(ctx, vpcPeeringRef(&state))
	if provErr := CheckResponseErr("read", "VPCPeering", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	peering.Named(data.Name.ValueString())
	if tags != nil {
		peering.RetaggedAs(tags...)
	} else {
		peering.RetaggedAs(peering.Tags()...)
	}

	updated, err := r.client.Client.FromNetwork().VPCPeerings().Update(ctx, peering)
	if provErr := CheckResponseErr("update", "VPCPeering", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = state.Id
	data.Uri = state.Uri
	data.ProjectId = state.ProjectId
	data.VpcId = state.VpcId
	data.PeerVpc = state.PeerVpc
	data.Location = state.Location
	data.Name = types.StringValue(updated.Name())
	data.Tags = TagsToListPreserveNull(updated.Tags(), data.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpcPeeringResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VpcPeeringResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := vpcPeeringRef(&data)
	peeringID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromNetwork().VPCPeerings().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "VPCPeering", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErrAsError("delete", "VPCPeering",
			r.client.Client.FromNetwork().VPCPeerings().Delete(ctx, ref))
	}, "VPCPeering", peeringID, r.client.ResourceTimeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting VPCPeering", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "VPCPeering", peeringID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for VPCPeering deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a VPC Peering resource", map[string]interface{}{"vpcpeering_id": peeringID})
}

func (r *VpcPeeringResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
