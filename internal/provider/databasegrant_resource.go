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

type DatabaseGrantResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	ProjectID types.String `tfsdk:"project_id"`
	DBaaSID   types.String `tfsdk:"dbaas_id"`
	Database  types.String `tfsdk:"database"`
	UserID    types.String `tfsdk:"user_id"`
	Role      types.String `tfsdk:"role"`
	Timeout   types.String `tfsdk:"timeout"`
}

type DatabaseGrantResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &DatabaseGrantResource{}
var _ resource.ResourceWithImportState = &DatabaseGrantResource{}

func NewDatabaseGrantResource() resource.Resource {
	return &DatabaseGrantResource{}
}

func (r *DatabaseGrantResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databasegrant"
}

func (r *DatabaseGrantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a privilege grant for an ArubaCloud DBaaS user on a specific database.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource (composite key: `project_id/dbaas_id/database/user_id`).",
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
				MarkdownDescription: "ID of the project that owns this resource. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dbaas_id": schema.StringAttribute{
				MarkdownDescription: "ID of the parent DBaaS cluster this grant belongs to. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database": schema.StringAttribute{
				MarkdownDescription: "ID of the database this grant applies to. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "Name or ID of the DBaaS user receiving the grant. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Privilege level granted. Accepted values depend on the database engine (e.g., `ALL`, `READ`, `WRITE`). The DBaaS grant API does not support in-place role changes. (Immutable — changing this value forces the resource to be destroyed and re-created.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timeout": schema.StringAttribute{
				MarkdownDescription: "Per-resource timeout override (e.g. `\"15m\"`, `\"1h\"`). Overrides the provider-level `resource_timeout` for this resource's Create and Delete operations. Uses Go duration syntax.",
				Optional:            true,
			},
		},
	}
}

func (r *DatabaseGrantResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// grantCompositeRef constructs a URI using the user ID as the grant key.
// This matches the legacy behavior where userID was used as the grant identifier.
func grantCompositeRef(projectID, dbaasID, databaseName, userID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID +
		"/providers/Aruba.Database/dbaas/" + dbaasID +
		"/databases/" + databaseName +
		"/grants/" + userID)
}

// grantRefFromModel extracts IDs from the composite stored ID (project/dbaas/db/user).
func grantRefFromModel(data *DatabaseGrantResourceModel) aruba.Ref {
	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()
	databaseName := data.Database.ValueString()
	userID := data.UserID.ValueString()
	return grantCompositeRef(projectID, dbaasID, databaseName, userID)
}

func (r *DatabaseGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DatabaseGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()
	databaseName := data.Database.ValueString()
	userID := data.UserID.ValueString()

	databaseURI := "/projects/" + projectID +
		"/providers/Aruba.Database/dbaas/" + dbaasID +
		"/databases/" + databaseName

	grant, err := r.client.Client.FromDatabase().Grants().Create(ctx,
		aruba.NewGrant().
			InDatabase(aruba.URI(databaseURI)).
			ForUser(userID).
			OfRole(data.Role.ValueString()),
	)
	if provErr := CheckResponseErr("create", "DatabaseGrant", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	// Preserve the composite ID pattern for backward compatibility.
	data.Id = types.StringValue(fmt.Sprintf("%s/%s/%s/%s", projectID, dbaasID, databaseName, userID))
	data.Uri = strVal(grant.URI())
	data.Role = types.StringValue(grant.RoleName())

	tflog.Trace(ctx, "created a Database Grant resource", map[string]interface{}{"grant_id": data.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatabaseGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	grant, err := r.client.Client.FromDatabase().Grants().Get(ctx, grantRefFromModel(&data))
	if provErr := CheckResponseErr("read", "DatabaseGrant", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Role = types.StringValue(grant.RoleName())
	data.Uri = strVal(grant.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is unreachable: all attributes carry RequiresReplace so Terraform
// will never call Update — any change triggers a destroy + create cycle.
func (r *DatabaseGrantResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"The ArubaCloud DBaaS grant API does not support in-place updates. "+
			"All attribute changes force a new resource (RequiresReplace).",
	)
}

func (r *DatabaseGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatabaseGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := grantRefFromModel(&data)
	grantID := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromDatabase().Grants().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "DatabaseGrant", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErrAsError("delete", "DatabaseGrant",
			r.client.Client.FromDatabase().Grants().Delete(ctx, ref))
	}, "DatabaseGrant", grantID, effectiveTimeout(data.Timeout, r.client.ResourceTimeout), deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting DatabaseGrant", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "DatabaseGrant", grantID, remainingTimeout(deleteStart, effectiveTimeout(data.Timeout, r.client.ResourceTimeout))); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for DatabaseGrant deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a Database Grant resource", map[string]interface{}{"grant_id": grantID})
}

func (r *DatabaseGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseImportID(req.ID, "<project_id>/<dbaas_id>/<database>/<user_id>", "proj-abc/dbaas-xyz/mydb/usr-xyz", 4)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}
	// DatabaseGrant uses the composite string as its stable id (no API-assigned UUID).
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dbaas_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("database"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), parts[3])...)
}
