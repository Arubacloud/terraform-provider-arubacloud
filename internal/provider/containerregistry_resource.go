package provider

import (
	"context"
	"fmt"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ContainerRegistryResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	ProjectID     types.String `tfsdk:"project_id"`
	BillingPeriod types.String `tfsdk:"billing_period"`
	Network       types.Object `tfsdk:"network"`
	Storage       types.Object `tfsdk:"storage"`
	Settings      types.Object `tfsdk:"settings"`
}

type ContainerRegistryNetworkModel struct {
	PublicIpUriRef      types.String `tfsdk:"public_ip_uri_ref"`
	VpcUriRef           types.String `tfsdk:"vpc_uri_ref"`
	SubnetUriRef        types.String `tfsdk:"subnet_uri_ref"`
	SecurityGroupUriRef types.String `tfsdk:"security_group_uri_ref"`
}

type ContainerRegistryStorageModel struct {
	BlockStorageUriRef types.String `tfsdk:"block_storage_uri_ref"`
}

type ContainerRegistrySettingsModel struct {
	AdminUser             types.String `tfsdk:"admin_user"`
	ConcurrentUsersFlavor types.String `tfsdk:"concurrent_users_flavor"`
}

type ContainerRegistryResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &ContainerRegistryResource{}
var _ resource.ResourceWithImportState = &ContainerRegistryResource{}

func NewContainerRegistryResource() resource.Resource {
	return &ContainerRegistryResource{}
}

func (r *ContainerRegistryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_containerregistry"
}

func (r *ContainerRegistryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an ArubaCloud Container Registry — a private OCI-compatible image registry.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource.",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the container registry.",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
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
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle. Accepted values: `Hour`, `Month`, `Year`.",
				Optional:            true,
			},
			"network": schema.SingleNestedAttribute{
				MarkdownDescription: "Network resources attached to the registry.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"public_ip_uri_ref": schema.StringAttribute{
						MarkdownDescription: "URI of the Elastic IP that exposes the registry endpoint (e.g., `arubacloud_elasticip.example.uri`).",
						Required:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"vpc_uri_ref": schema.StringAttribute{
						MarkdownDescription: "URI of the VPC that hosts the registry (e.g., `arubacloud_vpc.example.uri`).",
						Required:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"subnet_uri_ref": schema.StringAttribute{
						MarkdownDescription: "URI of the subnet within the VPC (e.g., `arubacloud_subnet.example.uri`).",
						Required:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"security_group_uri_ref": schema.StringAttribute{
						MarkdownDescription: "URI of the security group controlling registry traffic (e.g., `arubacloud_securitygroup.example.uri`).",
						Required:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
				},
			},
			"storage": schema.SingleNestedAttribute{
				MarkdownDescription: "Block storage volume that backs the registry image store.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"block_storage_uri_ref": schema.StringAttribute{
						MarkdownDescription: "URI of the block storage volume (e.g., `arubacloud_blockstorage.example.uri`).",
						Required:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
				},
			},
			"settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Optional registry configuration settings.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"admin_user": schema.StringAttribute{
						MarkdownDescription: "Administrator username for the registry.",
						Optional:            true,
					},
					"concurrent_users_flavor": schema.StringAttribute{
						MarkdownDescription: "Concurrency tier that determines how many simultaneous push/pull sessions are supported. Accepted values: `Small`, `Medium`, `HighPerf`.",
						Optional:            true,
					},
				},
			},
		},
	}
}

func (r *ContainerRegistryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func containerRegistryRef(data *ContainerRegistryResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.URI("/projects/" + data.ProjectID.ValueString() +
		"/providers/Aruba.Container/containerRegistries/" + data.Id.ValueString())
}

func applyContainerRegistryToModel(reg *aruba.ContainerRegistry, data *ContainerRegistryResourceModel) {
	data.Id = types.StringValue(reg.ID())
	data.Uri = strVal(reg.URI())
	data.Name = types.StringValue(reg.Name())
	data.Tags = TagsToListPreserveNull(reg.Tags(), data.Tags)
	if reg.Region() != "" {
		data.Location = types.StringValue(string(reg.Region()))
	}
	data.BillingPeriod = strVal(string(reg.BillingPeriod()))

	networkObj, _ := types.ObjectValue(
		map[string]attr.Type{
			"public_ip_uri_ref":      types.StringType,
			"vpc_uri_ref":            types.StringType,
			"subnet_uri_ref":         types.StringType,
			"security_group_uri_ref": types.StringType,
		},
		map[string]attr.Value{
			"public_ip_uri_ref":      types.StringValue(reg.ElasticIP()),
			"vpc_uri_ref":            types.StringValue(reg.VPC()),
			"subnet_uri_ref":         types.StringValue(reg.Subnet()),
			"security_group_uri_ref": types.StringValue(reg.SecurityGroup()),
		},
	)
	data.Network = networkObj

	storageObj, _ := types.ObjectValue(
		map[string]attr.Type{"block_storage_uri_ref": types.StringType},
		map[string]attr.Value{"block_storage_uri_ref": types.StringValue(reg.BlockStorage())},
	)
	data.Storage = storageObj

	settingsObj, _ := types.ObjectValue(
		map[string]attr.Type{
			"admin_user":              types.StringType,
			"concurrent_users_flavor": types.StringType,
		},
		map[string]attr.Value{
			"admin_user":              strVal(reg.AdminUsername()),
			"concurrent_users_flavor": strVal(string(reg.SizeFlavor())),
		},
	)
	data.Settings = settingsObj
}

func (r *ContainerRegistryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ContainerRegistryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var networkModel ContainerRegistryNetworkModel
	resp.Diagnostics.Append(data.Network.As(ctx, &networkModel, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	var storageModel ContainerRegistryStorageModel
	resp.Diagnostics.Append(data.Storage.As(ctx, &storageModel, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	builder := aruba.NewContainerRegistry().
		Named(data.Name.ValueString()).
		InProject(aruba.URI("/projects/" + projectID)).
		InRegion(aruba.Region(data.Location.ValueString())).
		Tagged(tags...).
		WithElasticIP(aruba.URI(networkModel.PublicIpUriRef.ValueString())).
		WithVPC(aruba.URI(networkModel.VpcUriRef.ValueString())).
		WithSubnet(aruba.URI(networkModel.SubnetUriRef.ValueString())).
		WithSecurityGroup(aruba.URI(networkModel.SecurityGroupUriRef.ValueString())).
		WithBlockStorage(aruba.URI(storageModel.BlockStorageUriRef.ValueString()))

	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		builder = builder.BilledBy(aruba.BillingPeriod(data.BillingPeriod.ValueString()))
	}

	if !data.Settings.IsNull() && !data.Settings.IsUnknown() {
		var settingsModel ContainerRegistrySettingsModel
		resp.Diagnostics.Append(data.Settings.As(ctx, &settingsModel, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !settingsModel.AdminUser.IsNull() && !settingsModel.AdminUser.IsUnknown() {
			builder = builder.WithAdminUsername(settingsModel.AdminUser.ValueString())
		}
		if !settingsModel.ConcurrentUsersFlavor.IsNull() && !settingsModel.ConcurrentUsersFlavor.IsUnknown() {
			builder = builder.OfSize(aruba.ContainerRegistrySizeFlavor(settingsModel.ConcurrentUsersFlavor.ValueString()))
		}
	}

	registry, err := r.client.Client.FromContainer().ContainerRegistry().Create(ctx, builder)
	if provErr := CheckResponseErr("create", "ContainerRegistry", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(registry.ID())
	data.Uri = strVal(registry.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// ContainerRegistry can take 20-40 minutes to converge.
	if waitErr := registry.WaitUntilReady(ctx, aruba.WithTimeout(40*time.Minute)); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "ContainerRegistry", data.Id.ValueString())
		return
	}

	fresh, freshErr := r.client.Client.FromContainer().ContainerRegistry().Get(ctx, containerRegistryRef(&data))
	if freshErr == nil {
		projectID := data.ProjectID
		applyContainerRegistryToModel(fresh, &data)
		data.ProjectID = projectID
	} else {
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh ContainerRegistry after creation: %v", freshErr))
	}

	tflog.Trace(ctx, "created a Container Registry resource", map[string]interface{}{
		"containerregistry_id":   data.Id.ValueString(),
		"containerregistry_name": data.Name.ValueString(),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContainerRegistryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ContainerRegistryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	registry, err := r.client.Client.FromContainer().ContainerRegistry().Get(ctx, containerRegistryRef(&data))
	if provErr := CheckResponseErr("read", "ContainerRegistry", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	st := string(registry.State())
	switch {
	case isFailedState(st):
		resp.Diagnostics.AddWarning("Resource in Failed State",
			fmt.Sprintf("ContainerRegistry %q is in a terminal failure state (%s). "+
				"Run `terraform destroy` to clean it up, or `terraform apply -replace=<address>` to recreate it.", data.Id.ValueString(), st))
	case IsCreatingState(st):
		if waitErr := registry.WaitUntilReady(ctx, aruba.WithTimeout(40*time.Minute)); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "ContainerRegistry", data.Id.ValueString())
			return
		}
		registry, err = r.client.Client.FromContainer().ContainerRegistry().Get(ctx, containerRegistryRef(&data))
		if provErr := CheckResponseErr("read", "ContainerRegistry", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	projectID := data.ProjectID
	applyContainerRegistryToModel(registry, &data)
	data.ProjectID = projectID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContainerRegistryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ContainerRegistryResourceModel
	var state ContainerRegistryResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	registry, err := r.client.Client.FromContainer().ContainerRegistry().Get(ctx, containerRegistryRef(&state))
	if provErr := CheckResponseErr("read", "ContainerRegistry", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	registry.Named(data.Name.ValueString())
	if tags != nil {
		registry.RetaggedAs(tags...)
	} else {
		registry.RetaggedAs(registry.Tags()...)
	}

	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		registry.BilledBy(aruba.BillingPeriod(data.BillingPeriod.ValueString()))
	}

	if !data.Network.IsNull() && !data.Network.IsUnknown() {
		var networkModel ContainerRegistryNetworkModel
		resp.Diagnostics.Append(data.Network.As(ctx, &networkModel, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		registry.WithElasticIP(aruba.URI(networkModel.PublicIpUriRef.ValueString())).
			WithVPC(aruba.URI(networkModel.VpcUriRef.ValueString())).
			WithSubnet(aruba.URI(networkModel.SubnetUriRef.ValueString())).
			WithSecurityGroup(aruba.URI(networkModel.SecurityGroupUriRef.ValueString()))
	}

	if !data.Storage.IsNull() && !data.Storage.IsUnknown() {
		var storageModel ContainerRegistryStorageModel
		resp.Diagnostics.Append(data.Storage.As(ctx, &storageModel, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		registry.WithBlockStorage(aruba.URI(storageModel.BlockStorageUriRef.ValueString()))
	}

	if !data.Settings.IsNull() && !data.Settings.IsUnknown() {
		var settingsModel ContainerRegistrySettingsModel
		resp.Diagnostics.Append(data.Settings.As(ctx, &settingsModel, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !settingsModel.AdminUser.IsNull() {
			registry.WithAdminUsername(settingsModel.AdminUser.ValueString())
		}
		if !settingsModel.ConcurrentUsersFlavor.IsNull() {
			registry.OfSize(aruba.ContainerRegistrySizeFlavor(settingsModel.ConcurrentUsersFlavor.ValueString()))
		}
	}

	updated, err := r.client.Client.FromContainer().ContainerRegistry().Update(ctx, registry)
	if provErr := CheckResponseErr("update", "ContainerRegistry", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = state.Id
	data.ProjectID = state.ProjectID
	projectID := data.ProjectID
	applyContainerRegistryToModel(updated, &data)
	data.ProjectID = projectID
	data.Id = state.Id

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContainerRegistryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ContainerRegistryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := containerRegistryRef(&data)
	registryID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromContainer().ContainerRegistry().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "ContainerRegistry", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErrAsError("delete", "ContainerRegistry",
			r.client.Client.FromContainer().ContainerRegistry().Delete(ctx, ref))
	}, "ContainerRegistry", registryID, r.client.ResourceTimeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting ContainerRegistry", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "ContainerRegistry", registryID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for ContainerRegistry deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a Container Registry resource", map[string]interface{}{"containerregistry_id": registryID})
}

func (r *ContainerRegistryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
