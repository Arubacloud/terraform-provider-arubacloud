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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type DBaaSResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Zone          types.String `tfsdk:"zone"`
	Tags          types.List   `tfsdk:"tags"`
	ProjectID     types.String `tfsdk:"project_id"`
	EngineID      types.String `tfsdk:"engine_id"`
	Flavor        types.String `tfsdk:"flavor"`
	Storage       types.Object `tfsdk:"storage"`
	Network       types.Object `tfsdk:"network"`
	BillingPeriod types.String `tfsdk:"billing_period"`
}

type DBaaSResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &DBaaSResource{}
var _ resource.ResourceWithImportState = &DBaaSResource{}

func NewDBaaSResource() resource.Resource {
	return &DBaaSResource{}
}

func (r *DBaaSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dbaas"
}

func (r *DBaaSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an ArubaCloud DBaaS cluster — a managed database cluster with automated backups and high availability.",
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
				MarkdownDescription: "Display name for the DBaaS cluster.",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region identifier (e.g., `ITBG-Bergamo`). See the [available locations and zones](https://api.arubacloud.com/docs/metadata/#location-and-data-center).",
				Required:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Availability zone within the region where the DBaaS cluster is deployed.",
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
			"engine_id": schema.StringAttribute{
				MarkdownDescription: "Database engine type and version identifier (e.g., `mysql-8.0` for MySQL 8.0, `postgresql-15` for PostgreSQL 15). See the [available engines](https://api.arubacloud.com/docs/metadata/#dbaas-engines).",
				Required:            true,
			},
			"flavor": schema.StringAttribute{
				MarkdownDescription: "Compute flavour for the DBaaS cluster nodes. See [available flavours](https://api.arubacloud.com/docs/metadata/#dbaas-flavors). For example, `DBO2A4` means 2 vCPU and 4 GB RAM.",
				Required:            true,
			},
			"storage": schema.SingleNestedAttribute{
				MarkdownDescription: "Storage configuration for the DBaaS instance.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"size_gb": schema.Int64Attribute{
						MarkdownDescription: "Storage size in GB for the DBaaS instance.",
						Required:            true,
					},
					"autoscaling": schema.SingleNestedAttribute{
						MarkdownDescription: "Optional autoscaling configuration for the DBaaS storage.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								MarkdownDescription: "Whether storage autoscaling is enabled.",
								Required:            true,
							},
							"available_space": schema.Int64Attribute{
								MarkdownDescription: "Minimum available space threshold in GB. When the available storage falls below this value, autoscaling increases storage by the `step_size` amount.",
								Required:            true,
							},
							"step_size": schema.Int64Attribute{
								MarkdownDescription: "Amount of storage (in GB) added on each autoscaling event.",
								Required:            true,
							},
						},
					},
				},
			},
			"network": schema.SingleNestedAttribute{
				MarkdownDescription: "Network configuration for the DBaaS instance. All URI references are immutable after creation.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"vpc_uri_ref": schema.StringAttribute{
						MarkdownDescription: "URI reference to the VPC resource.",
						Required:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"subnet_uri_ref": schema.StringAttribute{
						MarkdownDescription: "URI reference to the Subnet resource.",
						Required:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"security_group_uri_ref": schema.StringAttribute{
						MarkdownDescription: "URI reference to the Security Group resource.",
						Required:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"elastic_ip_uri_ref": schema.StringAttribute{
						MarkdownDescription: "Optional URI reference to an Elastic IP resource.",
						Optional:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
				},
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing cycle. Accepted values: `Hour`, `Month`, `Year`.",
				Optional:            true,
			},
		},
	}
}

func (r *DBaaSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func dbaasRef(data *DBaaSResourceModel) aruba.Ref {
	if !data.Uri.IsNull() && data.Uri.ValueString() != "" {
		return aruba.URI(data.Uri.ValueString())
	}
	return aruba.URI("/projects/" + data.ProjectID.ValueString() + "/providers/Aruba.Database/dbaas/" + data.Id.ValueString())
}

// dbaasNetworkAttrTypes returns the attr.Type map for the network object.
func dbaasNetworkAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"vpc_uri_ref":            types.StringType,
		"subnet_uri_ref":         types.StringType,
		"security_group_uri_ref": types.StringType,
		"elastic_ip_uri_ref":     types.StringType,
	}
}

// dbaasStorageAttrTypes returns the attr.Type map for the storage object.
func dbaasStorageAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"size_gb": types.Int64Type,
		"autoscaling": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"enabled":         types.BoolType,
				"available_space": types.Int64Type,
				"step_size":       types.Int64Type,
			},
		},
	}
}

// extractNetworkFromModel extracts VPC, subnet, security group, and elastic IP URIs from the model's Network object.
func extractNetworkFromModel(_ context.Context, data *DBaaSResourceModel) (vpc, subnet, sg, eip string, ok bool) {
	if data.Network.IsNull() || data.Network.IsUnknown() {
		return
	}
	attrs := data.Network.Attributes()
	getStr := func(key string) string {
		if v, exists := attrs[key]; exists {
			if s, isStr := v.(types.String); isStr && !s.IsNull() {
				return s.ValueString()
			}
		}
		return ""
	}
	vpc = getStr("vpc_uri_ref")
	subnet = getStr("subnet_uri_ref")
	sg = getStr("security_group_uri_ref")
	eip = getStr("elastic_ip_uri_ref")
	ok = vpc != "" && subnet != "" && sg != ""
	return
}

// buildNetworkObject builds the network types.Object for state from the four URIs.
func buildNetworkObject(vpc, subnet, sg, eip string) types.Object {
	eipVal := types.StringNull()
	if eip != "" {
		eipVal = types.StringValue(eip)
	}
	obj, _ := types.ObjectValue(dbaasNetworkAttrTypes(), map[string]attr.Value{
		"vpc_uri_ref":            types.StringValue(vpc),
		"subnet_uri_ref":         types.StringValue(subnet),
		"security_group_uri_ref": types.StringValue(sg),
		"elastic_ip_uri_ref":     eipVal,
	})
	return obj
}

func (r *DBaaSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DBaaSResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract storage.
	if data.Storage.IsNull() || data.Storage.IsUnknown() {
		resp.Diagnostics.AddError("Missing Storage Configuration", "Storage configuration is required to create a DBaaS instance")
		return
	}
	storageAttrs := data.Storage.Attributes()
	sizeGBAttr, _ := storageAttrs["size_gb"].(types.Int64)
	if sizeGBAttr.IsNull() || sizeGBAttr.IsUnknown() {
		resp.Diagnostics.AddError("Missing Storage Size", "Storage size_gb is required")
		return
	}

	// Extract network.
	vpc, subnet, sg, eip, netOK := extractNetworkFromModel(ctx, &data)
	if !netOK {
		resp.Diagnostics.AddError("Missing Network Configuration", "vpc_uri_ref, subnet_uri_ref, and security_group_uri_ref are required")
		return
	}

	builder := aruba.NewDBaaS().
		Named(data.Name.ValueString()).
		InProject(aruba.URI("/projects/" + projectID)).
		InRegion(aruba.Region(data.Location.ValueString())).
		InZone(aruba.Zone(data.Zone.ValueString())).
		OfEngine(aruba.DatabaseEngine(data.EngineID.ValueString())).
		OfFlavor(aruba.DBaaSFlavor(data.Flavor.ValueString())).
		SizedGB(int(sizeGBAttr.ValueInt64())).
		WithVPC(aruba.URI(vpc)).
		WithSubnet(aruba.URI(subnet)).
		WithSecurityGroup(aruba.URI(sg)).
		Tagged(tags...)

	if eip != "" {
		builder = builder.WithElasticIP(aruba.URI(eip))
	}

	// Autoscaling.
	if autoscalingAttr, ok := storageAttrs["autoscaling"]; ok {
		if autoscalingObj, ok := autoscalingAttr.(types.Object); ok && !autoscalingObj.IsNull() {
			asAttrs := autoscalingObj.Attributes()
			enabledAttr, _ := asAttrs["enabled"].(types.Bool)
			availableAttr, _ := asAttrs["available_space"].(types.Int64)
			stepAttr, _ := asAttrs["step_size"].(types.Int64)
			if !enabledAttr.IsNull() && enabledAttr.ValueBool() {
				builder = builder.WithAutoscaling(int(availableAttr.ValueInt64()), int(stepAttr.ValueInt64()))
			}
		}
	}

	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		builder = builder.BilledBy(aruba.BillingPeriod(data.BillingPeriod.ValueString()))
	}

	dbaas, err := r.client.Client.FromDatabase().DBaaS().Create(ctx, builder)
	if provErr := CheckResponseErr("create", "DBaaS", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(dbaas.ID())
	data.Uri = strVal(dbaas.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if waitErr := dbaas.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
		ReportWaitResult(&resp.Diagnostics, waitErr, "DBaaS", data.Id.ValueString())
		return
	}

	// Refresh URI from re-read and preserve network from plan.
	fresh, freshErr := r.client.Client.FromDatabase().DBaaS().Get(ctx, dbaasRef(&data))
	if freshErr == nil {
		data.Uri = strVal(fresh.URI())
	} else {
		tflog.Warn(ctx, fmt.Sprintf("Failed to refresh DBaaS after creation: %v", freshErr))
	}

	data.Network = buildNetworkObject(vpc, subnet, sg, eip)

	tflog.Trace(ctx, "created a DBaaS resource", map[string]interface{}{
		"dbaas_id":   data.Id.ValueString(),
		"dbaas_name": data.Name.ValueString(),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DBaaSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DBaaSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	dbaas, err := r.client.Client.FromDatabase().DBaaS().Get(ctx, dbaasRef(&data))
	if provErr := CheckResponseErr("read", "DBaaS", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	st := string(dbaas.State())
	switch {
	case isFailedState(st):
		resp.Diagnostics.AddWarning("Resource in Failed State",
			fmt.Sprintf("DBaaS %q is in a terminal failure state (%s). "+
				"Run `terraform destroy` to clean it up, or `terraform apply -replace=<address>` to recreate it.", data.Id.ValueString(), st))
	case IsCreatingState(st):
		if waitErr := dbaas.WaitUntilReady(ctx, sdkWaitOptions(r.client.ResourceTimeout)...); waitErr != nil {
			ReportWaitResult(&resp.Diagnostics, waitErr, "DBaaS", data.Id.ValueString())
			return
		}
		dbaas, err = r.client.Client.FromDatabase().DBaaS().Get(ctx, dbaasRef(&data))
		if provErr := CheckResponseErr("read", "DBaaS", err); provErr != nil {
			resp.Diagnostics.AddError("API Error", provErr.Error())
			return
		}
	}

	// Preserve immutable fields from state.
	projectID := data.ProjectID
	zone := data.Zone
	engineID := data.EngineID
	flavor := data.Flavor
	network := data.Network
	storage := data.Storage

	data.Id = types.StringValue(dbaas.ID())
	data.Uri = strVal(dbaas.URI())
	data.Name = types.StringValue(dbaas.Name())
	data.Tags = TagsToListPreserveNull(dbaas.Tags(), data.Tags)
	data.ProjectID = projectID
	data.Zone = zone // zone not returned by API

	if dbaas.Region() != "" {
		data.Location = types.StringValue(string(dbaas.Region()))
	}

	// engine_id is immutable and the API may normalize the value (e.g. "mysql" → "mysql-8.0"),
	// so always preserve the state value to avoid a perpetual diff.
	data.EngineID = engineID
	if f := string(dbaas.Flavor()); f != "" {
		data.Flavor = types.StringValue(f)
	} else {
		data.Flavor = flavor
	}
	if bp := string(dbaas.BillingPeriod()); bp != "" {
		data.BillingPeriod = types.StringValue(bp)
	} else {
		data.BillingPeriod = types.StringNull()
	}

	// Storage: update size_gb from API, preserve autoscaling from state.
	storageAttrTypes := dbaasStorageAttrTypes()
	storageAttrs := map[string]attr.Value{}
	if s := dbaas.SizeGB(); s > 0 {
		storageAttrs["size_gb"] = types.Int64Value(int64(s))
	} else if !storage.IsNull() {
		if existingAttrs := storage.Attributes(); existingAttrs != nil {
			if v, ok := existingAttrs["size_gb"]; ok {
				storageAttrs["size_gb"] = v
			}
		}
	}
	if _, ok := storageAttrs["size_gb"]; !ok {
		storageAttrs["size_gb"] = types.Int64Null()
	}
	// Preserve autoscaling from state.
	autoscalingObjType, _ := storageAttrTypes["autoscaling"].(types.ObjectType)
	if !storage.IsNull() {
		existingAttrs := storage.Attributes()
		if v, ok := existingAttrs["autoscaling"]; ok {
			storageAttrs["autoscaling"] = v
		} else {
			storageAttrs["autoscaling"] = types.ObjectNull(autoscalingObjType.AttrTypes)
		}
	} else {
		storageAttrs["autoscaling"] = types.ObjectNull(autoscalingObjType.AttrTypes)
	}
	storageObj, diags := types.ObjectValue(storageAttrTypes, storageAttrs)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		data.Storage = storageObj
	}

	// Network: preserve from state (not returned by API reliably).
	data.Network = network

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DBaaSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DBaaSResourceModel
	var state DBaaSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := ListToTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	dbaas, err := r.client.Client.FromDatabase().DBaaS().Get(ctx, dbaasRef(&state))
	if provErr := CheckResponseErr("read", "DBaaS", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	dbaas.Named(data.Name.ValueString())
	if tags != nil {
		dbaas.RetaggedAs(tags...)
	} else {
		dbaas.RetaggedAs(dbaas.Tags()...)
	}

	// Update storage size if changed.
	if !data.Storage.IsNull() {
		storageAttrs := data.Storage.Attributes()
		if sizeAttr, ok := storageAttrs["size_gb"].(types.Int64); ok && !sizeAttr.IsNull() {
			dbaas.SizedGB(int(sizeAttr.ValueInt64()))
		}
		// Update autoscaling if provided.
		if asAttr, ok := storageAttrs["autoscaling"]; ok {
			if asObj, ok := asAttr.(types.Object); ok && !asObj.IsNull() {
				asAttrs := asObj.Attributes()
				enabledAttr, _ := asAttrs["enabled"].(types.Bool)
				availableAttr, _ := asAttrs["available_space"].(types.Int64)
				stepAttr, _ := asAttrs["step_size"].(types.Int64)
				if !enabledAttr.IsNull() && enabledAttr.ValueBool() {
					dbaas.WithAutoscaling(int(availableAttr.ValueInt64()), int(stepAttr.ValueInt64()))
				} else {
					dbaas.WithoutAutoscaling()
				}
			}
		}
	}

	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		dbaas.BilledBy(aruba.BillingPeriod(data.BillingPeriod.ValueString()))
	}

	updated, err := r.client.Client.FromDatabase().DBaaS().Update(ctx, dbaas)
	if provErr := CheckResponseErr("update", "DBaaS", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	// Preserve immutable fields.
	data.Id = state.Id
	data.ProjectID = state.ProjectID
	data.Uri = state.Uri
	data.Zone = state.Zone
	data.EngineID = state.EngineID
	data.Flavor = state.Flavor
	data.Network = state.Network // network is immutable

	if updated.URI() != "" {
		data.Uri = types.StringValue(updated.URI())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DBaaSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DBaaSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := dbaasRef(&data)
	dbaasID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromDatabase().DBaaS().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "DBaaS", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErrAsError("delete", "DBaaS",
			r.client.Client.FromDatabase().DBaaS().Delete(ctx, ref))
	}, "DBaaS", dbaasID, r.client.ResourceTimeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting DBaaS", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "DBaaS", dbaasID, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for DBaaS deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a DBaaS resource", map[string]interface{}{"dbaas_id": dbaasID})
}

func (r *DBaaSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
