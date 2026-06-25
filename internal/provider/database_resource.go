package provider

import (
	"context"
	"fmt"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type DatabaseResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Uri       types.String `tfsdk:"uri"`
	ProjectID types.String `tfsdk:"project_id"`
	DBaaSID   types.String `tfsdk:"dbaas_id"`
	Name      types.String `tfsdk:"name"`
}

type DatabaseResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &DatabaseResource{}
var _ resource.ResourceWithImportState = &DatabaseResource{}

func NewDatabaseResource() resource.Resource {
	return &DatabaseResource{}
}

func (r *DatabaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *DatabaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a logical database within an ArubaCloud DBaaS cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Unique identifier for the resource (same as the database name).",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Computed by the API. Full resource URI used as a reference value in other resources.",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project that owns this resource.",
				Required:            true,
			},
			"dbaas_id": schema.StringAttribute{
				MarkdownDescription: "ID of the parent DBaaS cluster this database belongs to.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the database.",
				Required:            true,
			},
		},
	}
}

func (r *DatabaseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func databaseRef(data *DatabaseResourceModel) aruba.Ref {
	return aruba.URI("/projects/" + data.ProjectID.ValueString() +
		"/providers/Aruba.Database/dbaas/" + data.DBaaSID.ValueString() +
		"/databases/" + data.Id.ValueString())
}

func (r *DatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DatabaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()
	dbaasURI := "/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID

	var db *aruba.Database
	if createErr := CreateWithTransientRetry(ctx, func() error {
		var err error
		db, err = r.client.Client.FromDatabase().Databases().Create(ctx,
			aruba.NewDatabase().
				Named(data.Name.ValueString()).
				InDBaaS(aruba.URI(dbaasURI)),
		)
		return CheckResponseErrAsError("create", "Database", err)
	}, "Database", data.Name.ValueString(), r.client.ResourceTimeout); createErr != nil {
		resp.Diagnostics.AddError("Error creating database", createErr.Error())
		return
	}

	// Database uses name as ID.
	data.Id = types.StringValue(db.Name())
	data.Uri = strVal(db.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Databases don't have a status; wait until we can successfully Get.
	checker := func(ctx context.Context) (string, error) {
		_, getErr := r.client.Client.FromDatabase().Databases().Get(ctx, databaseRef(&data))
		if provErr := CheckResponseErr("get", "Database", getErr); provErr != nil {
			return "Unknown", provErr
		}
		return "Active", nil
	}
	if err := WaitForResourceActive(ctx, checker, "Database", data.Id.ValueString(), r.client.ResourceTimeout); err != nil {
		ReportWaitResult(&resp.Diagnostics, err, "Database", data.Id.ValueString())
		return
	}

	tflog.Trace(ctx, "created a Database resource", map[string]interface{}{
		"database_id":   data.Id.ValueString(),
		"database_name": data.Name.ValueString(),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Id.IsNull() || data.Id.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	db, err := r.client.Client.FromDatabase().Databases().Get(ctx, databaseRef(&data))
	if provErr := CheckResponseErr("read", "Database", err); provErr != nil {
		if IsNotFound(provErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(db.Name())
	data.Name = types.StringValue(db.Name())
	data.Uri = strVal(db.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DatabaseResourceModel
	var state DatabaseResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	db, err := r.client.Client.FromDatabase().Databases().Get(ctx, databaseRef(&state))
	if provErr := CheckResponseErr("read", "Database", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	db.Named(data.Name.ValueString())

	updated, err := r.client.Client.FromDatabase().Databases().Update(ctx, db)
	if provErr := CheckResponseErr("update", "Database", err); provErr != nil {
		resp.Diagnostics.AddError("API Error", provErr.Error())
		return
	}

	data.Id = types.StringValue(updated.Name()) // name can change
	data.ProjectID = state.ProjectID
	data.DBaaSID = state.DBaaSID
	data.Name = types.StringValue(updated.Name())
	data.Uri = strVal(updated.URI())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref := databaseRef(&data)
	databaseName := data.Id.ValueString()

	deletionChecker := func(ctx context.Context) (bool, error) {
		_, getErr := r.client.Client.FromDatabase().Databases().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Database", getErr); provErr != nil {
			if IsNotFound(provErr) {
				return true, nil
			}
			return false, provErr
		}
		return false, nil
	}

	deleteStart := time.Now()
	err := DeleteResourceWithRetry(ctx, func() error {
		return CheckResponseErrAsError("delete", "Database",
			r.client.Client.FromDatabase().Databases().Delete(ctx, ref))
	}, "Database", databaseName, r.client.ResourceTimeout, deletionChecker)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting database", err.Error())
		return
	}
	if waitErr := WaitForResourceDeleted(ctx, deletionChecker, "Database", databaseName, remainingTimeout(deleteStart, r.client.ResourceTimeout)); waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for Database deletion", waitErr.Error())
		return
	}
	tflog.Trace(ctx, "deleted a Database resource", map[string]interface{}{"database_id": databaseName})
}

func (r *DatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
