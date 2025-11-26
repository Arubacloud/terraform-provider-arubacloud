// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type DBaaSResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Location      types.String `tfsdk:"location"`
	Tags          types.List   `tfsdk:"tags"`
	ProjectID     types.String `tfsdk:"project_id"`
	Engine        types.String `tfsdk:"engine"`
	Zone          types.String `tfsdk:"zone"`
	Flavor        types.String `tfsdk:"flavor"`
	StorageSize   types.Int64  `tfsdk:"storage_size"`
	BillingPeriod types.String `tfsdk:"billing_period"`
}

type DBaaSResource struct {
	client *http.Client
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
		MarkdownDescription: "DBaaS resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "DBaaS identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "DBaaS name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "DBaaS location",
				Required:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags for the DBaaS resource",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project this DBaaS belongs to",
				Required:            true,
			},
			"engine": schema.StringAttribute{
				MarkdownDescription: "Database engine (mysql-8.0, mssql-2022-web, mssql-2022-standard, mssql-2022-enterprise)",
				Required:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone (ITBG-1, ITBG-2, ITBG-3)",
				Required:            true,
			},
			"flavor": schema.StringAttribute{
				MarkdownDescription: "Flavor type",
				Required:            true,
			},
			"storage_size": schema.Int64Attribute{
				MarkdownDescription: "Storage size",
				Required:            true,
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period",
				Required:            true,
			},
			"network": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"vpc_id": schema.StringAttribute{
						MarkdownDescription: "VPC ID",
						Required:            true,
					},
					"subnet_id": schema.StringAttribute{
						MarkdownDescription: "Subnet ID",
						Required:            true,
					},
					"security_group_id": schema.StringAttribute{
						MarkdownDescription: "Security Group ID",
						Required:            true,
					},
					"elastic_ip_id": schema.StringAttribute{
						MarkdownDescription: "Elastic IP ID",
						Required:            true,
					},
				},
				Required: true,
			},
			"autoscaling": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Autoscaling enabled",
						Required:            true,
					},
					"available_space": schema.Int64Attribute{
						MarkdownDescription: "Available space for autoscaling",
						Required:            true,
					},
					"step_size": schema.Int64Attribute{
						MarkdownDescription: "Step size for autoscaling",
						Required:            true,
					},
				},
				Required: true,
			},
		},
	}
}

func (r *DBaaSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*http.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *DBaaSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DBaaSResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Simulate API response
	data.Id = types.StringValue("dbaas-id")
	tflog.Trace(ctx, "created a DBaaS resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DBaaSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DBaaSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DBaaSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DBaaSResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DBaaSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DBaaSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DBaaSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
