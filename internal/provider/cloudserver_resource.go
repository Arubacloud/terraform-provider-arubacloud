package provider

import (
	"context"
	"fmt"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type CloudServerResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	Name      types.String `tfsdk:"name"`
	Location  types.String `tfsdk:"location"`
	ProjectID types.String `tfsdk:"project_id"`
	Zone      types.String `tfsdk:"zone"`
	Tags      types.List   `tfsdk:"tags"`
	Network   types.Object `tfsdk:"network"`
	Settings  types.Object `tfsdk:"settings"`
	Storage   types.Object `tfsdk:"storage"`
	Timeout   types.String `tfsdk:"timeout"`
}

type CloudServerNetworkModel struct {
	VpcUriRef            types.String `tfsdk:"vpc_uri_ref"`
	ElasticIpUriRef      types.String `tfsdk:"elastic_ip_uri_ref"`
	SubnetUriRefs        types.List   `tfsdk:"subnet_uri_refs"`
	SecurityGroupUriRefs types.List   `tfsdk:"securitygroup_uri_refs"`
}

type CloudServerSettingsModel struct {
	FlavorName    types.String `tfsdk:"flavor_name"`
	KeyPairUriRef types.String `tfsdk:"key_pair_uri_ref"`
	UserData      types.String `tfsdk:"user_data"`
}

type CloudServerStorageModel struct {
	BootVolumeUriRef types.String `tfsdk:"boot_volume_uri_ref"`
}

type CloudServerResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &CloudServerResource{}
var _ resource.ResourceWithImportState = &CloudServerResource{}

func NewCloudServerResource() resource.Resource {
	return &CloudServerResource{}
}

func (r *CloudServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudserver"
}

func (r *CloudServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an ArubaCloud CloudServer virtual machine.",
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
				MarkdownDescription: "Display name for the CloudServer.",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier for the resource (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center). Changing this value forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource. Changing this value forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Availability zone within the region (e.g., `ITBG-1`). See [available zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center). Changing this value forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of string tags attached to the resource for filtering and organisation.",
				Optional:            true,
			},
			"network": schema.SingleNestedAttribute{
				MarkdownDescription: "Network configuration for the CloudServer.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"vpc_uri_ref": schema.StringAttribute{
						MarkdownDescription: "URI of the VPC to attach this CloudServer to. Reference the `uri` attribute of an `arubacloud_vpc` resource. Changing this value forces a new resource.",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							stringplanmodifier.RequiresReplace(),
						},
					},
					"elastic_ip_uri_ref": schema.StringAttribute{
						MarkdownDescription: "URI of an Elastic IP to associate with this CloudServer. Reference the `uri` attribute of an `arubacloud_elasticip` resource. Optional — omit to use a dynamic IP. Changing this value forces a new resource.",
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							stringplanmodifier.RequiresReplace(),
						},
					},
					"subnet_uri_refs": schema.ListAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "List of subnet URIs to attach this CloudServer to. Reference the `uri` attribute of each `arubacloud_subnet` resource. Changing this value forces a new resource.",
						Required:            true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						},
					},
					"securitygroup_uri_refs": schema.ListAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "List of security group URIs to apply to this CloudServer. Reference the `uri` attribute of each `arubacloud_securitygroup` resource. Changing this value forces a new resource.",
						Required:            true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Compute and access settings for the CloudServer.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"flavor_name": schema.StringAttribute{
						MarkdownDescription: "Compute flavour name (e.g., `CSO4A8` for 4 vCPU / 8 GB RAM). See [available flavours](https://api.arubacloud.com/docs/metadata/#cloudserver-flavors).",
						Required:            true,
					},
					"key_pair_uri_ref": schema.StringAttribute{
						MarkdownDescription: "URI of the SSH key pair to inject at boot. Reference the `uri` attribute of an `arubacloud_keypair` resource.",
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"user_data": schema.StringAttribute{
						MarkdownDescription: "Cloud-Init configuration passed verbatim to the instance at first boot (raw YAML or shell-script). Write-only — this value is sent to the API but is not returned in subsequent read responses. Changing this value forces a new resource.",
						Optional:            true,
						Sensitive:           true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"storage": schema.SingleNestedAttribute{
				MarkdownDescription: "Storage configuration for the CloudServer.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"boot_volume_uri_ref": schema.StringAttribute{
						MarkdownDescription: "URI of the bootable block storage volume. Reference the `uri` attribute of an `arubacloud_blockstorage` resource (must be bootable). Changing this value forces a new resource.",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"timeout": schema.StringAttribute{
				MarkdownDescription: "Per-resource timeout override (e.g. `\"15m\"`, `\"1h\"`). Overrides the provider-level `resource_timeout` for this resource's Create and Delete operations. Uses Go duration syntax.",
				Optional:            true,
			},
		},
	}
}

func (r *CloudServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CloudServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CloudServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError("Missing Project ID", "Project ID is required to create a cloud server")
		return
	}

	var networkModel CloudServerNetworkModel
	resp.Diagnostics.Append(data.Network.As(ctx, &networkModel, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	var settingsModel CloudServerSettingsModel
	resp.Diagnostics.Append(data.Settings.As(ctx, &settingsModel, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	var storageModel CloudServerStorageModel
	resp.Diagnostics.Append(data.Storage.As(ctx, &storageModel, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var subnetURIs []string
	if !networkModel.SubnetUriRefs.IsNull() && !networkModel.SubnetUriRefs.IsUnknown() {
		resp.Diagnostics.Append(networkModel.SubnetUriRefs.ElementsAs(ctx, &subnetURIs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	subnetRefs := make([]aruba.Ref, len(subnetURIs))
	for i, u := range subnetURIs {
		subnetRefs[i] = aruba.URI(u)
	}

	var sgURIs []string
	if !networkModel.SecurityGroupUriRefs.IsNull() && !networkModel.SecurityGroupUriRefs.IsUnknown() {
		resp.Diagnostics.Append(networkModel.SecurityGroupUriRefs.ElementsAs(ctx, &sgURIs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	sgRefs := make([]aruba.Ref, len(sgURIs))
	for i, u := range sgURIs {
		sgRefs[i] = aruba.URI(u)
	}

	builder := aruba.NewCloudServer().
		Named(data.Name.ValueString()).
		InProject(aruba.URI("/projects/" + projectID)).
		InRegion(aruba.Region(data.Location.ValueString())).
		InZone(aruba.Zone(data.Zone.ValueString())).
		OfFlavor(aruba.CloudServerFlavor(settingsModel.FlavorName.ValueString())).
		WithVPC(aruba.URI(networkModel.VpcUriRef.ValueString())).
		BootingFrom(aruba.URI(storageModel.BootVolumeUriRef.ValueString())).
		OnSubnets(subnetRefs...).
		WithSecurityGroups(sgRefs...).
		Tagged(tags...)

	if !settingsModel.KeyPairUriRef.IsNull() && settingsModel.KeyPairUriRef.ValueString() != "" {
		builder = builder.UsingKeyPair(aruba.URI(settingsModel.KeyPairUriRef.ValueString()))
	}
	if !networkModel.ElasticIpUriRef.IsNull() && networkModel.ElasticIpUriRef.ValueString() != "" {
		builder = builder.WithElasticIP(aruba.URI(networkModel.ElasticIpUriRef.ValueString()))
	}
	if !settingsModel.UserData.IsNull() && settingsModel.UserData.ValueString() != "" {
		builder = builder.WithUserData(settingsModel.UserData.ValueString())
	}

	server, err := r.client.Client.FromCompute().CloudServers().Create(ctx, builder)
	if provErr := CheckResponseErr("create", "CloudServer", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	serverID := server.ID()
	data.Id = types.StringValue(serverID)
	if uri := server.URI(); uri != "" {
		data.Uri = types.StringValue(uri)
	} else {
		data.Uri = types.StringNull()
	}

	// Save partial state so destroy can clean up on timeout.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := server.WaitUntilReady(ctx, sdkWaitOptions(effectiveTimeout(data.Timeout, r.client.ResourceTimeout))...); err != nil {
		ReportWaitResult(&resp.Diagnostics, err, "CloudServer", serverID)
		return
	}

	// Re-read to populate URI and server-assigned fields after provisioning.
	fresh, freshErr := r.client.Client.FromCompute().CloudServers().Get(ctx, aruba.URI(server.URI()))
	if freshErr == nil {
		// Pass &data as originalState so the plan's typed list values for
		// subnet_uri_refs and securitygroup_uri_refs are preserved — the API
		// does not return these in a usable form, and zero-value types.List{}
		// (when originalState is nil) causes an "MISSING TYPE" object error.
		r.applyServerToState(ctx, fresh, &data, &data, &resp.Diagnostics)
	} else {
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh CloudServer after creation: %v", freshErr))
	}

	tflog.Trace(ctx, "created a CloudServer resource", map[string]interface{}{
		"cloudserver_id":   serverID,
		"cloudserver_name": data.Name.ValueString(),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// cloudServerRef returns the Ref to use for Get/Update/Delete.
// Falls back to a constructed URI for the import flow where stored URI may be empty.
func cloudServerRef(data *CloudServerResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.URI("/projects/" + data.ProjectID.ValueString() + "/compute/cloudServers/" + data.Id.ValueString())
}

func (r *CloudServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var originalState CloudServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &originalState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := originalState.Id.ValueString()
	if originalState.Id.IsUnknown() || originalState.Id.IsNull() || serverID == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	server, err := r.client.Client.FromCompute().CloudServers().Get(ctx, cloudServerRef(&originalState))
	if provErr := CheckResponseErr("read", "CloudServer", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	// Resume provisioning wait if the resource is still in a transitional state.
	st := string(server.State())
	switch {
	case isFailedState(st):
		resp.Diagnostics.AddWarning(
			"Resource in Failed State",
			fmt.Sprintf("CloudServer %q is in a terminal failure state (%s). "+
				"Run `terraform destroy` to clean it up, or `terraform apply -replace=<address>` to recreate it.", serverID, st),
		)
	case IsCreatingState(st):
		if waitErr := server.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "CloudServer", serverID)
			return
		}
		// Re-read after wait.
		server, err = r.client.Client.FromCompute().CloudServers().Get(ctx, cloudServerRef(&originalState))
		if provErr := CheckResponseErr("read", "CloudServer", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	var data CloudServerResourceModel
	r.applyServerToState(ctx, server, &data, &originalState, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// applyServerToState populates data from the SDK wrapper, preserving fields from
// originalState that the API does not return. Pass originalState=nil on first Create.
func (r *CloudServerResource) applyServerToState(
	ctx context.Context,
	server *aruba.CloudServer,
	data *CloudServerResourceModel,
	originalState *CloudServerResourceModel,
	diags *diag.Diagnostics,
) {
	data.Id = types.StringValue(server.ID())
	if uri := server.URI(); uri != "" {
		data.Uri = types.StringValue(uri)
	} else {
		data.Uri = types.StringNull()
	}
	data.Name = types.StringValue(server.Name())
	data.Tags = TagsToListPreserveNull(server.Tags(), data.Tags)

	raw := server.Raw()
	if raw != nil {
		if raw.Metadata.LocationResponse != nil && raw.Metadata.LocationResponse.Value != "" {
			data.Location = types.StringValue(string(raw.Metadata.LocationResponse.Value))
		}
		data.Zone = resolveAPIStringRef(string(raw.Properties.Zone), firstString(originalState, func(s *CloudServerResourceModel) types.String { return s.Zone }))
	}

	if originalState != nil {
		data.ProjectID = originalState.ProjectID
	}

	// ── Network object ────────────────────────────────────────────────────────
	// Initialize with properly-typed nulls so types.ObjectValue never receives
	// a zero-value types.List{} (which has no element type).  We only overwrite
	// from state when the stored object is non-null and non-unknown.
	origNetwork := CloudServerNetworkModel{
		VpcUriRef:            types.StringNull(),
		ElasticIpUriRef:      types.StringNull(),
		SubnetUriRefs:        types.ListNull(types.StringType),
		SecurityGroupUriRefs: types.ListNull(types.StringType),
	}
	if originalState != nil && !originalState.Network.IsNull() && !originalState.Network.IsUnknown() {
		diags.Append(originalState.Network.As(ctx, &origNetwork, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return
		}
	}
	// Subnets are returned by the API via NetworkInterfaces[].Subnet — use the
	// live value so drift is detected. Fall back to state only when the API
	// returns nothing (e.g. immediately after create before the response hydrates).
	subnetUriRefs := origNetwork.SubnetUriRefs
	if apiSubnets := server.Subnets(); len(apiSubnets) > 0 {
		vals := make([]attr.Value, len(apiSubnets))
		for i, s := range apiSubnets {
			vals[i] = types.StringValue(s)
		}
		if lv, d := types.ListValue(types.StringType, vals); !d.HasError() {
			subnetUriRefs = lv
		}
	}

	// elastic_ip_uri_ref and securitygroup_uri_refs are not returned by the API;
	// preserve them from the prior state to avoid spurious diffs.
	networkAttrs := map[string]attr.Value{
		"vpc_uri_ref":            resolveAPIStringRef(server.VPC(), origNetwork.VpcUriRef),
		"elastic_ip_uri_ref":     origNetwork.ElasticIpUriRef,
		"subnet_uri_refs":        subnetUriRefs,
		"securitygroup_uri_refs": origNetwork.SecurityGroupUriRefs,
	}
	networkObj, d := types.ObjectValue(csNetworkAttrTypes(), networkAttrs)
	diags.Append(d...)
	data.Network = networkObj

	// ── Settings object ───────────────────────────────────────────────────────
	var origSettings CloudServerSettingsModel
	if originalState != nil && !originalState.Settings.IsNull() && !originalState.Settings.IsUnknown() {
		diags.Append(originalState.Settings.As(ctx, &origSettings, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return
		}
	}
	settingsAttrs := map[string]attr.Value{
		"flavor_name":      types.StringValue(string(server.Flavor())),
		"key_pair_uri_ref": resolveKeyPairUriRef(server.KeyPair(), origSettings.KeyPairUriRef),
		"user_data":        origSettings.UserData, // write-only; never returned by API
	}
	settingsObj, d := types.ObjectValue(csSettingsAttrTypes(), settingsAttrs)
	diags.Append(d...)
	data.Settings = settingsObj

	// ── Storage object ────────────────────────────────────────────────────────
	var origStorage CloudServerStorageModel
	if originalState != nil && !originalState.Storage.IsNull() && !originalState.Storage.IsUnknown() {
		diags.Append(originalState.Storage.As(ctx, &origStorage, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return
		}
	}
	storageAttrs := map[string]attr.Value{
		"boot_volume_uri_ref": resolveAPIStringRef(server.BootVolume(), origStorage.BootVolumeUriRef),
	}
	storageObj, d := types.ObjectValue(csStorageAttrTypes(), storageAttrs)
	diags.Append(d...)
	data.Storage = storageObj
}

// firstString is a tiny helper to safely read a field from an optional state pointer.
func firstString(s *CloudServerResourceModel, fn func(*CloudServerResourceModel) types.String) types.String {
	if s == nil {
		return types.StringNull()
	}
	return fn(s)
}

func csNetworkAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"vpc_uri_ref":            types.StringType,
		"elastic_ip_uri_ref":     types.StringType,
		"subnet_uri_refs":        types.ListType{ElemType: types.StringType},
		"securitygroup_uri_refs": types.ListType{ElemType: types.StringType},
	}
}

func csSettingsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"flavor_name":      types.StringType,
		"key_pair_uri_ref": types.StringType,
		"user_data":        types.StringType,
	}
}

func csStorageAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{"boot_volume_uri_ref": types.StringType}
}

func (r *CloudServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CloudServerResourceModel
	var state CloudServerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract nested plan models for state reconstruction.
	var planNetworkModel CloudServerNetworkModel
	resp.Diagnostics.Append(data.Network.As(ctx, &planNetworkModel, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	var planSettingsModel CloudServerSettingsModel
	resp.Diagnostics.Append(data.Settings.As(ctx, &planSettingsModel, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	var planStorageModel CloudServerStorageModel
	resp.Diagnostics.Append(data.Storage.As(ctx, &planStorageModel, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch-mutate-update: get current wrapper, apply name/tag mutations, submit.
	server, err := r.client.Client.FromCompute().CloudServers().Get(ctx, cloudServerRef(&state))
	if provErr := CheckResponseErr("read", "CloudServer", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	server.Named(data.Name.ValueString())
	if tags != nil {
		server.RetaggedAs(tags...)
	} else {
		server.RetaggedAs(server.Tags()...)
	}

	// Subnets and security groups are not returned by the API on Get, so the
	// server object would send null for both fields without these explicit calls,
	// causing a 400 "subnet/securityGroup cannot be null" error.
	if !planNetworkModel.SubnetUriRefs.IsNull() && !planNetworkModel.SubnetUriRefs.IsUnknown() {
		var subnetURIs []string
		resp.Diagnostics.Append(planNetworkModel.SubnetUriRefs.ElementsAs(ctx, &subnetURIs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		subnetRefs := make([]aruba.Ref, len(subnetURIs))
		for i, u := range subnetURIs {
			subnetRefs[i] = aruba.URI(u)
		}
		server.OnSubnets(subnetRefs...)
	}

	if !planNetworkModel.SecurityGroupUriRefs.IsNull() && !planNetworkModel.SecurityGroupUriRefs.IsUnknown() {
		var sgURIs []string
		resp.Diagnostics.Append(planNetworkModel.SecurityGroupUriRefs.ElementsAs(ctx, &sgURIs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		sgRefs := make([]aruba.Ref, len(sgURIs))
		for i, u := range sgURIs {
			sgRefs[i] = aruba.URI(u)
		}
		server.WithSecurityGroups(sgRefs...)
	}

	updated, err := r.client.Client.FromCompute().CloudServers().Update(ctx, server)
	if provErr := CheckResponseErr("update", "CloudServer", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	if waitErr := updated.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "CloudServer", state.Id.ValueString())
		return
	}

	if n := updated.Name(); n != "" {
		data.Name = types.StringValue(n)
	}
	data.Tags = TagsToListPreserveNull(updated.Tags(), data.Tags)

	// Preserve immutable fields from state.
	data.Id = state.Id
	data.Uri = state.Uri
	data.ProjectID = state.ProjectID
	data.Zone = state.Zone

	// Rebuild nested objects from plan values.
	networkAttrs := map[string]attr.Value{
		"vpc_uri_ref":            planNetworkModel.VpcUriRef,
		"elastic_ip_uri_ref":     planNetworkModel.ElasticIpUriRef,
		"subnet_uri_refs":        planNetworkModel.SubnetUriRefs,
		"securitygroup_uri_refs": planNetworkModel.SecurityGroupUriRefs,
	}
	networkObj, d := types.ObjectValue(csNetworkAttrTypes(), networkAttrs)
	resp.Diagnostics.Append(d...)
	data.Network = networkObj

	settingsAttrs := map[string]attr.Value{
		"flavor_name":      planSettingsModel.FlavorName,
		"key_pair_uri_ref": planSettingsModel.KeyPairUriRef,
		"user_data":        planSettingsModel.UserData,
	}
	settingsObj, d := types.ObjectValue(csSettingsAttrTypes(), settingsAttrs)
	resp.Diagnostics.Append(d...)
	data.Settings = settingsObj

	storageAttrs := map[string]attr.Value{"boot_volume_uri_ref": planStorageModel.BootVolumeUriRef}
	storageObj, d := types.ObjectValue(csStorageAttrTypes(), storageAttrs)
	resp.Diagnostics.Append(d...)
	data.Storage = storageObj

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CloudServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CloudServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := data.Id.ValueString()
	if serverID == "" {
		resp.Diagnostics.AddError("Missing Required Fields", "Server ID is required to delete the cloud server")
		return
	}

	ref := cloudServerRef(&data)

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromCompute().CloudServers().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "CloudServer", getErr); provErr != nil {
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
			delErr := r.client.Client.FromCompute().CloudServers().Delete(ctx, ref)
			return CheckResponseErrAsError("delete", "CloudServer", delErr)
		},
		"CloudServer",
		serverID,
		effectiveTimeout(data.Timeout, r.client.ResourceTimeout),
		deletionChecker,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting cloud server", err.Error())
		return
	}

	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "CloudServer", serverID, remainingTimeout(deleteStart, effectiveTimeout(data.Timeout, r.client.ResourceTimeout))); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for CloudServer deletion", waitErr.Error())
		return
	}

	tflog.Trace(ctx, "deleted a CloudServer resource", map[string]interface{}{"cloudserver_id": serverID})
}

func (r *CloudServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseImportID(req.ID, "<project_id>/<cloudserver_id>", "proj-abc/srv-xyz", 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// resolveAPIStringRef returns a StringValue from the API-provided string when
// non-empty, falling back to the value preserved from state otherwise.
func resolveAPIStringRef(apiValue string, stateValue types.String) types.String {
	if apiValue != "" {
		return types.StringValue(apiValue)
	}
	return stateValue
}

// resolveKeyPairUriRef resolves key_pair_uri_ref for Read state:
// API returned URI → use it; API returned empty but state had URI → null (detached); no URI either side → preserve state.
func resolveKeyPairUriRef(apiURI string, stateRef types.String) types.String {
	if apiURI != "" {
		return types.StringValue(apiURI)
	}
	if !stateRef.IsNull() {
		return types.StringNull()
	}
	return stateRef
}
