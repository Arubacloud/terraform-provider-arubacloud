package provider

import (
	"context"
	"encoding/base64"
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

type DBaaSUserResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	ProjectID types.String `tfsdk:"project_id"`
	DBaaSID   types.String `tfsdk:"dbaas_id"`
	Username  types.String `tfsdk:"username"`
	Password  types.String `tfsdk:"password"`
	Timeout   types.String `tfsdk:"timeout"`
}

type DBaaSUserResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &DBaaSUserResource{}
var _ resource.ResourceWithImportState = &DBaaSUserResource{}

func NewDBaaSUserResource() resource.Resource {
	return &DBaaSUserResource{}
}

func (r *DBaaSUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dbaasuser"
}

func (r *DBaaSUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a database user within an ArubaCloud DBaaS cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource (same as the username).",
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
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"dbaas_id": schema.StringAttribute{
				MarkdownDescription: "ID of the parent DBaaS cluster this user belongs to.",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Display name for the DBaaS user.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for the DBaaS user. Write-only — this value is sent to the API but is not returned in subsequent read responses.",
				Required:            true,
				Sensitive:           true,
			},
			"timeout": schema.StringAttribute{
				MarkdownDescription: "Per-resource timeout override (e.g. `\"15m\"`, `\"1h\"`). Overrides the provider-level `resource_timeout` for this resource's Create and Delete operations. Uses Go duration syntax.",
				Optional:            true,
			},
		},
	}
}

func (r *DBaaSUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func dbaasUserRef(data *DBaaSUserResourceModel) aruba.Ref {
	return aruba.URI("/projects/" + data.ProjectID.ValueString() +
		"/providers/Aruba.Database/dbaas/" + data.DBaaSID.ValueString() +
		"/users/" + data.Id.ValueString())
}

func (r *DBaaSUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DBaaSUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()
	username := data.Username.ValueString()
	passwordBase64 := base64.StdEncoding.EncodeToString([]byte(data.Password.ValueString()))

	dbaasURI := "/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID

	var user *aruba.User
	if createErr := CreateWithTransientRetry(ctx, func() error {
		var err error
		user, err = r.client.Client.FromDatabase().Users().Create(ctx,
			aruba.NewUser().
				WithUsername(username).
				WithPassword(passwordBase64).
				InDBaaS(aruba.URI(dbaasURI)),
		)
		return CheckResponseErrAsError("create", "DBaaSUser", err)
	}, "DBaaSUser", username, effectiveTimeout(data.Timeout, r.client.ResourceTimeout)); createErr != nil {
		resp.Diagnostics.AddError("Error creating DBaaS user", createErr.Error())
		return
	}

	data.Id = types.StringValue(user.Username())
	data.Uri = strVal(user.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Users don't have a status field; wait until we can successfully Get the user.
	checker := func(ctx context.Context) (string, error) {
		_, getErr := r.client.Client.FromDatabase().Users().Get(ctx, dbaasUserRef(&data))
		if provErr := CheckResponseErr("get", "DBaaSUser", getErr); provErr != nil {
			return "Unknown", provErr
		}
		return "Active", nil
	}
	if err := WaitForResourceActive(ctx, checker, "DBaaSUser", username, effectiveTimeout(data.Timeout, r.client.ResourceTimeout)); err != nil {
		ReportWaitResult(&resp.Diagnostics, err, "DBaaSUser", username)
		return
	}

	tflog.Trace(ctx, "created a DBaaS User resource", map[string]interface{}{
		"user_id":  data.Id.ValueString(),
		"username": data.Username.ValueString(),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DBaaSUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DBaaSUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	user, err := r.client.Client.FromDatabase().Users().Get(ctx, dbaasUserRef(&data))
	if provErr := CheckResponseErr("read", "DBaaSUser", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(user.Username())
	data.Uri = strVal(user.URI())
	data.Username = types.StringValue(user.Username())
	// Password is not returned from API; preserve existing state value.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DBaaSUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DBaaSUserResourceModel
	var state DBaaSUserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	passwordBase64 := base64.StdEncoding.EncodeToString([]byte(data.Password.ValueString()))

	user, err := r.client.Client.FromDatabase().Users().Get(ctx, dbaasUserRef(&state))
	if provErr := CheckResponseErr("read", "DBaaSUser", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	user.WithPassword(passwordBase64)

	updated, err := r.client.Client.FromDatabase().Users().Update(ctx, user)
	if provErr := CheckResponseErr("update", "DBaaSUser", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = state.Id
	data.ProjectID = state.ProjectID
	data.DBaaSID = state.DBaaSID
	data.Username = types.StringValue(updated.Username())
	data.Uri = strVal(updated.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DBaaSUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DBaaSUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := dbaasUserRef(&data)
	username := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromDatabase().Users().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "DBaaSUser", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErrAsError("delete", "DBaaSUser",
			r.client.Client.FromDatabase().Users().Delete(ctx, ref))
	}, "DBaaSUser", username, effectiveTimeout(data.Timeout, r.client.ResourceTimeout), deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting DBaaS user", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "DBaaSUser", username, remainingTimeout(deleteStart, effectiveTimeout(data.Timeout, r.client.ResourceTimeout))); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for DBaaSUser deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a DBaaS User resource", map[string]interface{}{"user_id": username})
}

func (r *DBaaSUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseImportID(req.ID, "<project_id>/<dbaas_id>/<user_id>", "proj-abc/dbaas-xyz/usr-xyz", 3)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dbaas_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[2])...)
}
