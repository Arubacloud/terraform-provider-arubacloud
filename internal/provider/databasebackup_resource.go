// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type DatabaseBackupResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Uri           types.String `tfsdk:"uri"`
	ProjectID     types.String `tfsdk:"project_id"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	Zone          types.String `tfsdk:"zone"`
	DBaaSID       types.String `tfsdk:"dbaas_id"`
	Database      types.String `tfsdk:"database"`
	BillingPeriod types.String `tfsdk:"billing_period"`
}

type DatabaseBackupResource struct {
	client *ArubaCloudClient
}

var _ resource.Resource = &DatabaseBackupResource{}
var _ resource.ResourceWithImportState = &DatabaseBackupResource{}

func NewDatabaseBackupResource() resource.Resource {
	return &DatabaseBackupResource{}
}

func (r *DatabaseBackupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databasebackup"
}

func (r *DatabaseBackupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Database Backup resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Database Backup identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Database Backup URI",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this backup belongs to",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Database Backup name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Database Backup location",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the Database Backup resource",
				Optional:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone for the Database Backup",
				Required:            true,
			},
			"dbaas_id": schema.StringAttribute{
				MarkdownDescription: "DBaaS ID this backup belongs to",
				Required:            true,
			},
			"database": schema.StringAttribute{
				MarkdownDescription: "Database name to backup",
				Required:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period",
				Required:            true,
			},
		},
	}
}

func (r *DatabaseBackupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DatabaseBackupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DatabaseBackupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.DBaaSID.ValueString()
	databaseName := data.Database.ValueString()

	if projectID == "" || dbaasID == "" || databaseName == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID, DBaaS ID, and Database name are required to create a database backup",
		)
		return
	}

	// Get DBaaS instance to get its URI
	dbaasResp, err := r.client.Client.FromDatabase().DBaaS().Get(ctx, projectID, dbaasID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting DBaaS instance",
			fmt.Sprintf("Unable to get DBaaS instance: %s", err),
		)
		return
	}

	if dbaasResp == nil || dbaasResp.Data == nil || dbaasResp.Data.Metadata.URI == nil {
		resp.Diagnostics.AddError(
			"DBaaS Instance Not Found",
			"DBaaS instance not found or missing URI",
		)
		return
	}

	dbaasURI := *dbaasResp.Data.Metadata.URI
	// Construct database URI
	databaseURI := fmt.Sprintf("%s/databases/%s", dbaasURI, databaseName)

	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Build the create request
	createRequest := sdktypes.BackupRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.BackupPropertiesRequest{
			Zone: data.Zone.ValueString(),
			DBaaS: sdktypes.ReferenceResource{
				URI: dbaasURI,
			},
			Database: sdktypes.ReferenceResource{
				URI: databaseURI,
			},
			BillingPlan: sdktypes.BillingPeriodResource{
				BillingPeriod: data.BillingPeriod.ValueString(),
			},
		},
	}

	// Create the backup using the SDK
	response, err := r.client.Client.FromDatabase().Backups().Create(ctx, projectID, createRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating database backup",
			fmt.Sprintf("Unable to create database backup: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create database backup"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil && response.Data.Metadata.ID != nil {
		data.Id = types.StringValue(*response.Data.Metadata.ID)
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"Database backup created but no ID returned from API",
		)
		return
	}

	// Wait for Database Backup to be active before returning
	// This ensures Terraform doesn't proceed until Backup is ready
	backupID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromDatabase().Backups().Get(ctx, projectID, backupID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for Database Backup to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "DatabaseBackup", backupID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Database Backup Not Active",
			fmt.Sprintf("Database backup was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a Database Backup resource", map[string]interface{}{
		"backup_id":   data.Id.ValueString(),
		"backup_name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseBackupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatabaseBackupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	backupID := data.Id.ValueString()

	if projectID == "" || backupID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Backup ID are required to read the database backup",
		)
		return
	}

	// Get backup details using the SDK
	response, err := r.client.Client.FromDatabase().Backups().Get(ctx, projectID, backupID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading database backup",
			fmt.Sprintf("Unable to read database backup: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read database backup"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		backup := response.Data
		if backup.Metadata.ID != nil {
			data.Id = types.StringValue(*backup.Metadata.ID)
		}
		if backup.Metadata.URI != nil {
			data.Uri = types.StringValue(*backup.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if backup.Metadata.Name != nil {
			data.Name = types.StringValue(*backup.Metadata.Name)
		}
		if backup.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(backup.Metadata.LocationResponse.Value)
		}
		if backup.Metadata.Tags != nil {
			tagValues := make([]attr.Value, len(backup.Metadata.Tags))
			for i, tag := range backup.Metadata.Tags {
				tagValues[i] = types.StringValue(tag)
			}
			tagsList, diags := types.ListValue(types.StringType, tagValues)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = tagsList
			}
		} else {
			data.Tags = types.ListNull(types.StringType)
		}
		if backup.Properties.Zone != "" {
			data.Zone = types.StringValue(backup.Properties.Zone)
		}
		if backup.Properties.BillingPlan.BillingPeriod != "" {
			data.BillingPeriod = types.StringValue(backup.Properties.BillingPlan.BillingPeriod)
		}
		// Extract DBaaS ID and Database name from URIs if needed
		// Note: These may need to be stored separately or extracted from URIs
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseBackupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DatabaseBackupResourceModel
	var state DatabaseBackupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Database backups typically don't support updates
	// If they do, implement update logic here
	resp.Diagnostics.AddWarning(
		"Update Not Supported",
		"Database backups do not support updates. Changes will be ignored.",
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseBackupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatabaseBackupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	backupID := data.Id.ValueString()

	if projectID == "" || backupID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and Backup ID are required to delete the database backup",
		)
		return
	}

	// Delete the backup using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromDatabase().Backups().Delete(ctx, projectID, backupID, nil)
		},
		ExtractSDKError,
		"DatabaseBackup",
		backupID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting database backup",
			fmt.Sprintf("Unable to delete database backup: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a Database Backup resource", map[string]interface{}{
		"backup_id": backupID,
	})
}

func (r *DatabaseBackupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
