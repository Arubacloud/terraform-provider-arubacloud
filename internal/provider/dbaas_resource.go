// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
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
		MarkdownDescription: "DBaaS resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "DBaaS identifier",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "DBaaS URI",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "DBaaS name",
				Required:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "DBaaS location",
				Required:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone",
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
			"engine_id": schema.StringAttribute{
				MarkdownDescription: "Database engine ID. Available engines are described in the [ArubaCloud API documentation](https://api.arubacloud.com/docs/metadata/#dbaas-engines). For example, `mysql-8.0` for MySQL version 8.0.",
				Required:            true,
			},
			"flavor": schema.StringAttribute{
				MarkdownDescription: "DBaaS flavor name. Available flavors are described in the [ArubaCloud API documentation](https://api.arubacloud.com/docs/metadata/#dbaas-flavors). For example, `DBO2A4` means 2 CPU and 4GB RAM.",
				Required:            true,
			},
			"storage": schema.SingleNestedAttribute{
				MarkdownDescription: "Storage configuration for the DBaaS instance",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"size_gb": schema.Int64Attribute{
						MarkdownDescription: "Storage size in GB for the DBaaS instance",
						Required:            true,
					},
					"autoscaling": schema.SingleNestedAttribute{
						MarkdownDescription: "Autoscaling configuration for the DBaaS instance",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								MarkdownDescription: "Enable autoscaling",
								Required:            true,
							},
							"available_space": schema.Int64Attribute{
								MarkdownDescription: "Minimum available space threshold in GB. When the available storage falls below this value, autoscaling will increase the storage by the step_size amount.",
								Required:            true,
							},
							"step_size": schema.Int64Attribute{
								MarkdownDescription: "Step size for autoscaling (in GB)",
								Required:            true,
							},
						},
					},
				},
			},
			"network": schema.SingleNestedAttribute{
				MarkdownDescription: "Network configuration for the DBaaS instance",
				Required:            true,
				Attributes: map[string]schema.Attribute{
			"vpc_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI reference to the VPC resource (e.g., `arubacloud_vpc.example.uri`)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subnet_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI reference to the Subnet resource (e.g., `arubacloud_subnet.example.uri`)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"security_group_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI reference to the Security Group resource (e.g., `arubacloud_securitygroup.example.uri`)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"elastic_ip_uri_ref": schema.StringAttribute{
				MarkdownDescription: "URI reference to the Elastic IP resource (e.g., `arubacloud_elasticip.example.uri`)",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
				},
			},
			"billing_period": schema.StringAttribute{
				MarkdownDescription: "Billing period (Hour, Month, Year)",
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

func (r *DBaaSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DBaaSResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"Project ID is required to create a DBaaS instance",
		)
		return
	}

	// Extract tags
	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	engineID := data.EngineID.ValueString()
	flavor := data.Flavor.ValueString()
	zone := data.Zone.ValueString()

	// Extract storage configuration from nested object
	if data.Storage.IsNull() || data.Storage.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing Storage Configuration",
			"Storage configuration is required to create a DBaaS instance",
		)
		return
	}

	storageObj, diags := data.Storage.ToObjectValue(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	storageAttrs := storageObj.Attributes()
	sizeGBAttr, ok := storageAttrs["size_gb"].(types.Int64)
	if !ok || sizeGBAttr.IsNull() || sizeGBAttr.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing Storage Size",
			"Storage size_gb is required to create a DBaaS instance",
		)
		return
	}
	storageSizeGB := int32(sizeGBAttr.ValueInt64())
	storageSizeGBPtr := &storageSizeGB

	// Extract autoscaling if present
	var autoscaling *sdktypes.DBaaSAutoscaling
	if autoscalingAttr, ok := storageAttrs["autoscaling"]; ok && autoscalingAttr != nil {
		if autoscalingObj, ok := autoscalingAttr.(types.Object); ok && !autoscalingObj.IsNull() && !autoscalingObj.IsUnknown() {
			autoscalingObjValue, diags := autoscalingObj.ToObjectValue(ctx)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				autoscalingAttrs := autoscalingObjValue.Attributes()
				enabledAttr, _ := autoscalingAttrs["enabled"].(types.Bool)
				availableSpaceAttr, _ := autoscalingAttrs["available_space"].(types.Int64)
				stepSizeAttr, _ := autoscalingAttrs["step_size"].(types.Int64)

				if !enabledAttr.IsNull() && !availableSpaceAttr.IsNull() && !stepSizeAttr.IsNull() {
					enabled := enabledAttr.ValueBool()
					availableSpace := int32(availableSpaceAttr.ValueInt64())
					stepSize := int32(stepSizeAttr.ValueInt64())
					autoscaling = &sdktypes.DBaaSAutoscaling{
						Enabled:        &enabled,
						AvailableSpace: &availableSpace,
						StepSize:       &stepSize,
					}
				}
			}
		}
	}

	// Extract network configuration from nested object
	if data.Network.IsNull() || data.Network.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing Network Configuration",
			"Network configuration is required to create a DBaaS instance",
		)
		return
	}

	networkObj, diags := data.Network.ToObjectValue(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	networkAttrs := networkObj.Attributes()
	vpcUriRefAttr, ok := networkAttrs["vpc_uri_ref"].(types.String)
	if !ok || vpcUriRefAttr.IsNull() || vpcUriRefAttr.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing VPC URI Reference",
			"VPC URI reference is required in network configuration",
		)
		return
	}
	vpcUriRef := vpcUriRefAttr.ValueString()

	subnetUriRefAttr, ok := networkAttrs["subnet_uri_ref"].(types.String)
	if !ok || subnetUriRefAttr.IsNull() || subnetUriRefAttr.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing Subnet URI Reference",
			"Subnet URI reference is required in network configuration",
		)
		return
	}
	subnetUriRef := subnetUriRefAttr.ValueString()

	securityGroupUriRefAttr, ok := networkAttrs["security_group_uri_ref"].(types.String)
	if !ok || securityGroupUriRefAttr.IsNull() || securityGroupUriRefAttr.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing Security Group URI Reference",
			"Security Group URI reference is required in network configuration",
		)
		return
	}
	securityGroupUriRef := securityGroupUriRefAttr.ValueString()

	// Elastic IP is optional
	var elasticIpUriRef string
	if elasticIpAttr, ok := networkAttrs["elastic_ip_uri_ref"]; ok && elasticIpAttr != nil {
		if elasticIpStr, ok := elasticIpAttr.(types.String); ok && !elasticIpStr.IsNull() && !elasticIpStr.IsUnknown() {
			elasticIpUriRef = elasticIpStr.ValueString()
		}
	}

	// Build the create request
	// Network configuration using DBaaSNetworking structure
	networking := &sdktypes.DBaaSNetworking{
		VPCURI:           &vpcUriRef,
		SubnetURI:        &subnetUriRef,
		SecurityGroupURI: &securityGroupUriRef,
	}
	
	// Add optional Elastic IP if provided
	if elasticIpUriRef != "" {
		networking.ElasticIPURI = &elasticIpUriRef
	}

	createRequest := sdktypes.DBaaSRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: data.Location.ValueString(),
			},
		},
		Properties: sdktypes.DBaaSPropertiesRequest{
			Engine: &sdktypes.DBaaSEngine{
				ID: &engineID,
			},
			Flavor: &sdktypes.DBaaSFlavor{
				Name: &flavor,
			},
			Storage: &sdktypes.DBaaSStorage{
				SizeGB: storageSizeGBPtr,
			},
			Autoscaling: autoscaling,
			Networking: networking,
			Zone:       &zone,
		},
	}

	// Add optional billing period
	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		billingPeriod := data.BillingPeriod.ValueString()
		createRequest.Properties.BillingPlan = &sdktypes.DBaaSBillingPlan{
			BillingPeriod: &billingPeriod,
		}
	}

	// Log the full request for debugging
	debugMap := map[string]interface{}{
		"project_id":        projectID,
		"name":              data.Name.ValueString(),
		"location":          data.Location.ValueString(),
		"zone":              zone,
		"engine_id":         engineID,
		"flavor":            flavor,
		"storage_size_gb":   storageSizeGB,
		"vpc_uri":           vpcUriRef,
		"subnet_uri":        subnetUriRef,
		"security_group_uri": securityGroupUriRef,
		"elastic_ip_uri":    elasticIpUriRef,
		"autoscaling":       autoscaling != nil,
	}
	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		debugMap["billing_period"] = data.BillingPeriod.ValueString()
	}
	tflog.Debug(ctx, "DBaaS create request", debugMap)

	// Create the DBaaS instance using the SDK
	response, err := r.client.Client.FromDatabase().DBaaS().Create(ctx, projectID, createRequest, nil)
	if err != nil {
		tflog.Error(ctx, "DBaaS create error", map[string]interface{}{
			"error":      err.Error(),
			"project_id": projectID,
		})
		resp.Diagnostics.AddError(
			"Error creating DBaaS instance",
			fmt.Sprintf("Unable to create DBaaS instance: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to create DBaaS instance"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		
		// Log detailed error information for debugging
		errorDetails := map[string]interface{}{
			"project_id": projectID,
			"engine_id":  engineID,
			"flavor":     flavor,
			"vpc_uri":    vpcUriRef,
			"subnet_uri": subnetUriRef,
			"sg_uri":     securityGroupUriRef,
		}
		if response.Error.Title != nil {
			errorDetails["error_title"] = *response.Error.Title
		}
		if response.Error.Detail != nil {
			errorDetails["error_detail"] = *response.Error.Detail
		}
		if response.Error.Status != nil {
			errorDetails["error_status"] = *response.Error.Status
		}
		if response.Error.Type != nil {
			errorDetails["error_type"] = *response.Error.Type
		}
		
		// Log full request and error response JSON only on errors for debugging
		if requestJSON, jsonErr := json.MarshalIndent(createRequest, "", "  "); jsonErr == nil {
			tflog.Debug(ctx, "Full DBaaS create request JSON (error case)", map[string]interface{}{
				"request_json": string(requestJSON),
			})
		}
		if errorJSON, jsonErr := json.MarshalIndent(response.Error, "", "  "); jsonErr == nil {
			tflog.Debug(ctx, "Full API error response JSON", map[string]interface{}{
				"error_json": string(errorJSON),
			})
		}
		
		tflog.Error(ctx, "DBaaS create request failed", errorDetails)
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	if response != nil && response.Data != nil {
		if response.Data.Metadata.ID != nil {
			data.Id = types.StringValue(*response.Data.Metadata.ID)
		}
		if response.Data.Metadata.URI != nil {
			data.Uri = types.StringValue(*response.Data.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"DBaaS instance created but no data returned from API",
		)
		return
	}

	// Wait for DBaaS to be active before returning
	// This ensures Terraform doesn't proceed until DBaaS is ready
	dbaasID := data.Id.ValueString()
	checker := func(ctx context.Context) (string, error) {
		getResp, err := r.client.Client.FromDatabase().DBaaS().Get(ctx, projectID, dbaasID, nil)
		if err != nil {
			return "", err
		}
		if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
			return *getResp.Data.Status.State, nil
		}
		return "Unknown", nil
	}

	// Wait for DBaaS to be active - block until ready (using configured timeout)
	if err := WaitForResourceActive(ctx, checker, "DBaaS", dbaasID, r.client.ResourceTimeout); err != nil {
		resp.Diagnostics.AddError(
			"DBaaS Not Active",
			fmt.Sprintf("DBaaS instance was created but did not become active within the timeout period: %s", err),
		)
		return
	}

	// Re-read the DBaaS instance to populate URI
	getResp, err := r.client.Client.FromDatabase().DBaaS().Get(ctx, projectID, dbaasID, nil)
	if err == nil && getResp != nil && getResp.Data != nil {
		dbaas := getResp.Data
		if dbaas.Metadata.URI != nil {
			data.Uri = types.StringValue(*dbaas.Metadata.URI)
		}

		// Build network object from extracted values
		networkAttrTypes := map[string]attr.Type{
			"vpc_uri_ref":            types.StringType,
			"subnet_uri_ref":         types.StringType,
			"security_group_uri_ref": types.StringType,
			"elastic_ip_uri_ref":     types.StringType,
		}
		networkAttrs := map[string]attr.Value{
			"vpc_uri_ref":            types.StringValue(vpcUriRef),
			"subnet_uri_ref":         types.StringValue(subnetUriRef),
			"security_group_uri_ref": types.StringValue(securityGroupUriRef),
		}
		if elasticIpUriRef != "" {
			networkAttrs["elastic_ip_uri_ref"] = types.StringValue(elasticIpUriRef)
		} else {
			networkAttrs["elastic_ip_uri_ref"] = types.StringNull()
		}
		networkObj, diags := types.ObjectValue(networkAttrTypes, networkAttrs)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Network = networkObj
		}
	}

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

	projectID := data.ProjectID.ValueString()
	dbaasID := data.Id.ValueString()

	if projectID == "" || dbaasID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and DBaaS ID are required to read the DBaaS instance",
		)
		return
	}

	// Get DBaaS instance details using the SDK
	response, err := r.client.Client.FromDatabase().DBaaS().Get(ctx, projectID, dbaasID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading DBaaS instance",
			fmt.Sprintf("Unable to read DBaaS instance: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		if response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		errorMsg := "Failed to read DBaaS instance"
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
		dbaas := response.Data

		if dbaas.Metadata.ID != nil {
			data.Id = types.StringValue(*dbaas.Metadata.ID)
		}
		if dbaas.Metadata.URI != nil {
			data.Uri = types.StringValue(*dbaas.Metadata.URI)
		} else {
			data.Uri = types.StringNull()
		}
		if dbaas.Metadata.Name != nil {
			data.Name = types.StringValue(*dbaas.Metadata.Name)
		}
		if dbaas.Metadata.LocationResponse != nil {
			data.Location = types.StringValue(dbaas.Metadata.LocationResponse.Value)
		}
		// Zone is not in the response, preserve from state
		// data.Zone is already set from state in Read function
		if dbaas.Properties.Engine != nil && dbaas.Properties.Engine.ID != nil {
			data.EngineID = types.StringValue(*dbaas.Properties.Engine.ID)
		}
		if dbaas.Properties.Flavor != nil && dbaas.Properties.Flavor.Name != nil {
			data.Flavor = types.StringValue(*dbaas.Properties.Flavor.Name)
		}
		if dbaas.Properties.BillingPlan != nil && dbaas.Properties.BillingPlan.BillingPeriod != nil {
			data.BillingPeriod = types.StringValue(*dbaas.Properties.BillingPlan.BillingPeriod)
		} else {
			data.BillingPeriod = types.StringNull()
		}

		// Build nested storage object from API response
		storageAttrs := make(map[string]attr.Value)
		storageAttrTypes := map[string]attr.Type{
			"size_gb":    types.Int64Type,
			"autoscaling": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"enabled":         types.BoolType,
					"available_space": types.Int64Type,
					"step_size":       types.Int64Type,
				},
			},
		}

		// Set storage size_gb
		if dbaas.Properties.Storage != nil && dbaas.Properties.Storage.SizeGB != nil {
			storageAttrs["size_gb"] = types.Int64Value(int64(*dbaas.Properties.Storage.SizeGB))
		} else {
			// If not in API response, preserve from state
			if !data.Storage.IsNull() && !data.Storage.IsUnknown() {
				storageObj, diags := data.Storage.ToObjectValue(ctx)
				if !diags.HasError() {
					existingAttrs := storageObj.Attributes()
					if sizeGB, ok := existingAttrs["size_gb"]; ok {
						storageAttrs["size_gb"] = sizeGB
					}
				}
			}
		}

		// Set autoscaling - preserve from state since API response structure may differ
		// DBaaSAutoscalingResponse has a different structure than DBaaSAutoscaling
		// We'll preserve autoscaling from state to maintain consistency
		if !data.Storage.IsNull() && !data.Storage.IsUnknown() {
			storageObj, diags := data.Storage.ToObjectValue(ctx)
			if !diags.HasError() {
				storageAttrsFromState := storageObj.Attributes()
				if autoscalingAttr, ok := storageAttrsFromState["autoscaling"]; ok && autoscalingAttr != nil {
					storageAttrs["autoscaling"] = autoscalingAttr
				} else {
					// If no autoscaling in state, set to null
					storageAttrs["autoscaling"] = types.ObjectNull(map[string]attr.Type{
						"enabled":         types.BoolType,
						"available_space": types.Int64Type,
						"step_size":       types.Int64Type,
					})
				}
			}
		} else {
			// If not in API response, preserve from state
			if !data.Storage.IsNull() && !data.Storage.IsUnknown() {
				storageObj, diags := data.Storage.ToObjectValue(ctx)
				if !diags.HasError() {
					existingAttrs := storageObj.Attributes()
					if autoscaling, ok := existingAttrs["autoscaling"]; ok {
						storageAttrs["autoscaling"] = autoscaling
					}
				}
			}
			// If no autoscaling in state either, set to null
			if _, ok := storageAttrs["autoscaling"]; !ok {
				storageAttrs["autoscaling"] = types.ObjectNull(map[string]attr.Type{
					"enabled":         types.BoolType,
					"available_space": types.Int64Type,
					"step_size":       types.Int64Type,
				})
			}
		}

		// Build storage object
		storageObj, diags := types.ObjectValue(storageAttrTypes, storageAttrs)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Storage = storageObj
		}

		// Build nested network object from state
		// Network configuration is preserved from state since it's not in API response
		networkAttrs := make(map[string]attr.Value)
		networkAttrTypes := map[string]attr.Type{
			"vpc_uri_ref":            types.StringType,
			"subnet_uri_ref":         types.StringType,
			"security_group_uri_ref": types.StringType,
			"elastic_ip_uri_ref":     types.StringType,
		}

		// Preserve network configuration from state
		if !data.Network.IsNull() && !data.Network.IsUnknown() {
			networkObj, diags := data.Network.ToObjectValue(ctx)
			if !diags.HasError() {
				existingNetworkAttrs := networkObj.Attributes()
				networkAttrs["vpc_uri_ref"] = existingNetworkAttrs["vpc_uri_ref"]
				networkAttrs["subnet_uri_ref"] = existingNetworkAttrs["subnet_uri_ref"]
				networkAttrs["security_group_uri_ref"] = existingNetworkAttrs["security_group_uri_ref"]
				if elasticIp, ok := existingNetworkAttrs["elastic_ip_uri_ref"]; ok {
					networkAttrs["elastic_ip_uri_ref"] = elasticIp
				} else {
					networkAttrs["elastic_ip_uri_ref"] = types.StringNull()
				}
			}
		} else {
			// If network not in state, set to null (should not happen)
			networkAttrs["vpc_uri_ref"] = types.StringNull()
			networkAttrs["subnet_uri_ref"] = types.StringNull()
			networkAttrs["security_group_uri_ref"] = types.StringNull()
			networkAttrs["elastic_ip_uri_ref"] = types.StringNull()
		}

		// Build network object
		networkObj, diags := types.ObjectValue(networkAttrTypes, networkAttrs)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Network = networkObj
		}

		// Update tags
		if len(dbaas.Metadata.Tags) > 0 {
			tagValues := make([]types.String, len(dbaas.Metadata.Tags))
			for i, tag := range dbaas.Metadata.Tags {
				tagValues[i] = types.StringValue(tag)
			}
			tagsList, diags := types.ListValueFrom(ctx, types.StringType, tagValues)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = tagsList
			}
		} else {
			emptyList, diags := types.ListValue(types.StringType, []attr.Value{})
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Tags = emptyList
			}
		}
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DBaaSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DBaaSResourceModel
	var state DBaaSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get IDs from state (not plan) - IDs are immutable and should always be in state
	projectID := state.ProjectID.ValueString()
	dbaasID := state.Id.ValueString()

	if projectID == "" || dbaasID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and DBaaS ID are required to update the DBaaS instance",
		)
		return
	}

	// Get current DBaaS instance details
	getResponse, err := r.client.Client.FromDatabase().DBaaS().Get(ctx, projectID, dbaasID, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching current DBaaS instance",
			fmt.Sprintf("Unable to get current DBaaS instance: %s", err),
		)
		return
	}

	if getResponse == nil || getResponse.Data == nil {
		resp.Diagnostics.AddError(
			"DBaaS Instance Not Found",
			"DBaaS instance not found or no data returned",
		)
		return
	}

	current := getResponse.Data

	// Get region value
	regionValue := ""
	if current.Metadata.LocationResponse != nil {
		regionValue = current.Metadata.LocationResponse.Value
	}
	if regionValue == "" {
		resp.Diagnostics.AddError(
			"Missing Region",
			"Unable to determine region value for DBaaS instance",
		)
		return
	}

	// Extract tags
	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		diags := data.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		tags = current.Metadata.Tags
	}

	// Extract storage configuration from plan's nested storage object
	var storageSizeGBPtr *int32
	var autoscaling *sdktypes.DBaaSAutoscaling

	if !data.Storage.IsNull() && !data.Storage.IsUnknown() {
		storageObj, diags := data.Storage.ToObjectValue(ctx)
		if !diags.HasError() {
			storageAttrs := storageObj.Attributes()
			
			// Extract size_gb
			if sizeGBAttr, ok := storageAttrs["size_gb"].(types.Int64); ok && !sizeGBAttr.IsNull() && !sizeGBAttr.IsUnknown() {
				storageSizeGB := int32(sizeGBAttr.ValueInt64())
				storageSizeGBPtr = &storageSizeGB
			} else if current.Properties.Storage != nil && current.Properties.Storage.SizeGB != nil {
				// Fallback to current if not in plan
				sizeGB := *current.Properties.Storage.SizeGB
				storageSizeGBPtr = &sizeGB
			}

			// Extract autoscaling if present
			if autoscalingAttr, ok := storageAttrs["autoscaling"]; ok && autoscalingAttr != nil {
				if autoscalingObj, ok := autoscalingAttr.(types.Object); ok && !autoscalingObj.IsNull() && !autoscalingObj.IsUnknown() {
					autoscalingObjValue, diags := autoscalingObj.ToObjectValue(ctx)
					if !diags.HasError() {
						autoscalingAttrs := autoscalingObjValue.Attributes()
						enabledAttr, _ := autoscalingAttrs["enabled"].(types.Bool)
						availableSpaceAttr, _ := autoscalingAttrs["available_space"].(types.Int64)
						stepSizeAttr, _ := autoscalingAttrs["step_size"].(types.Int64)

						if !enabledAttr.IsNull() && !availableSpaceAttr.IsNull() && !stepSizeAttr.IsNull() {
							enabled := enabledAttr.ValueBool()
							availableSpace := int32(availableSpaceAttr.ValueInt64())
							stepSize := int32(stepSizeAttr.ValueInt64())
							autoscaling = &sdktypes.DBaaSAutoscaling{
								Enabled:        &enabled,
								AvailableSpace: &availableSpace,
								StepSize:       &stepSize,
							}
						}
					}
				}
			} else if current.Properties.Autoscaling != nil {
				// Fallback to current if not in plan
				// Convert from DBaaSAutoscalingResponse to DBaaSAutoscaling
				// Note: Response type may have different structure, extract what we can
				availableSpace := int32(0)
				stepSize := int32(0)
				if current.Properties.Autoscaling.AvailableSpace != nil {
					availableSpace = *current.Properties.Autoscaling.AvailableSpace
				}
				if current.Properties.Autoscaling.StepSize != nil {
					stepSize = *current.Properties.Autoscaling.StepSize
				}
				// Default enabled to true if autoscaling exists (we can't determine from response)
				enabled := true
				autoscaling = &sdktypes.DBaaSAutoscaling{
					Enabled:        &enabled,
					AvailableSpace: &availableSpace,
					StepSize:       &stepSize,
				}
			}
		}
	} else {
		// If storage not in plan, use current values
		if current.Properties.Storage != nil && current.Properties.Storage.SizeGB != nil {
			sizeGB := *current.Properties.Storage.SizeGB
			storageSizeGBPtr = &sizeGB
		}
		if current.Properties.Autoscaling != nil {
			// Convert from DBaaSAutoscalingResponse to DBaaSAutoscaling
			availableSpace := int32(0)
			stepSize := int32(0)
			if current.Properties.Autoscaling.AvailableSpace != nil {
				availableSpace = *current.Properties.Autoscaling.AvailableSpace
			}
			if current.Properties.Autoscaling.StepSize != nil {
				stepSize = *current.Properties.Autoscaling.StepSize
			}
			// Default enabled to true if autoscaling exists
			enabled := true
			autoscaling = &sdktypes.DBaaSAutoscaling{
				Enabled:        &enabled,
				AvailableSpace: &availableSpace,
				StepSize:       &stepSize,
			}
		}
	}

	// Build update request - only name and tags can be updated
	updateRequest := sdktypes.DBaaSRequest{
		Metadata: sdktypes.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: sdktypes.ResourceMetadataRequest{
				Name: data.Name.ValueString(),
				Tags: tags,
			},
			Location: sdktypes.LocationRequest{
				Value: regionValue,
			},
		},
		Properties: sdktypes.DBaaSPropertiesRequest{
			// Preserve current engine if it exists
			Engine: func() *sdktypes.DBaaSEngine {
				if current.Properties.Engine != nil {
					return &sdktypes.DBaaSEngine{
						ID: current.Properties.Engine.ID,
					}
				}
				return nil
			}(),
			// Preserve current flavor if it exists
			Flavor: func() *sdktypes.DBaaSFlavor {
				if current.Properties.Flavor != nil {
					return &sdktypes.DBaaSFlavor{
						Name: current.Properties.Flavor.Name,
					}
				}
				return nil
			}(),
			// Use storage from plan or current
			Storage: func() *sdktypes.DBaaSStorage {
				if storageSizeGBPtr != nil {
					return &sdktypes.DBaaSStorage{
						SizeGB: storageSizeGBPtr,
					}
				}
				return nil
			}(),
			Autoscaling: autoscaling,
			// Preserve zone from plan or state (zone is immutable)
			Zone: func() *string {
				if !data.Zone.IsNull() && !data.Zone.IsUnknown() {
					zone := data.Zone.ValueString()
					return &zone
				}
				// Preserve from state if not in plan
				if !state.Zone.IsNull() && !state.Zone.IsUnknown() {
					zone := state.Zone.ValueString()
					return &zone
				}
				return nil
			}(),
			// Preserve networking from Terraform state (networking is immutable)
			// Extract from state since response type structure may differ
			Networking: func() *sdktypes.DBaaSNetworking {
				if !state.Network.IsNull() && !state.Network.IsUnknown() {
					networkObj, diags := state.Network.ToObjectValue(ctx)
					if !diags.HasError() {
						networkAttrs := networkObj.Attributes()
						vpcUriAttr, _ := networkAttrs["vpc_uri_ref"].(types.String)
						subnetUriAttr, _ := networkAttrs["subnet_uri_ref"].(types.String)
						securityGroupUriAttr, _ := networkAttrs["security_group_uri_ref"].(types.String)
						elasticIpUriAttr, _ := networkAttrs["elastic_ip_uri_ref"].(types.String)

						networking := &sdktypes.DBaaSNetworking{}
						if !vpcUriAttr.IsNull() && !vpcUriAttr.IsUnknown() {
							vpcUri := vpcUriAttr.ValueString()
							networking.VPCURI = &vpcUri
						}
						if !subnetUriAttr.IsNull() && !subnetUriAttr.IsUnknown() {
							subnetUri := subnetUriAttr.ValueString()
							networking.SubnetURI = &subnetUri
						}
						if !securityGroupUriAttr.IsNull() && !securityGroupUriAttr.IsUnknown() {
							securityGroupUri := securityGroupUriAttr.ValueString()
							networking.SecurityGroupURI = &securityGroupUri
						}
						if !elasticIpUriAttr.IsNull() && !elasticIpUriAttr.IsUnknown() {
							elasticIpUri := elasticIpUriAttr.ValueString()
							networking.ElasticIPURI = &elasticIpUri
						}
						return networking
					}
				}
				return nil
			}(),
		},
	}

	// Add billing period if provided, otherwise preserve from current
	if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
		billingPeriod := data.BillingPeriod.ValueString()
		updateRequest.Properties.BillingPlan = &sdktypes.DBaaSBillingPlan{
			BillingPeriod: &billingPeriod,
		}
	} else if current.Properties.BillingPlan != nil && current.Properties.BillingPlan.BillingPeriod != nil {
		updateRequest.Properties.BillingPlan = &sdktypes.DBaaSBillingPlan{
			BillingPeriod: current.Properties.BillingPlan.BillingPeriod,
		}
	}

	// Update the DBaaS instance using the SDK
	response, err := r.client.Client.FromDatabase().DBaaS().Update(ctx, projectID, dbaasID, updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating DBaaS instance",
			fmt.Sprintf("Unable to update DBaaS instance: %s", err),
		)
		return
	}

	if response != nil && response.IsError() && response.Error != nil {
		errorMsg := "Failed to update DBaaS instance"
		if response.Error.Title != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *response.Error.Title)
		}
		if response.Error.Detail != nil {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, *response.Error.Detail)
		}
		
		// Log full request and error response JSON only on errors for debugging
		if requestJSON, jsonErr := json.MarshalIndent(updateRequest, "", "  "); jsonErr == nil {
			tflog.Debug(ctx, "Full DBaaS update request JSON (error case)", map[string]interface{}{
				"request_json": string(requestJSON),
			})
		}
		if errorJSON, jsonErr := json.MarshalIndent(response.Error, "", "  "); jsonErr == nil {
			tflog.Debug(ctx, "Full API error response JSON", map[string]interface{}{
				"error_json": string(errorJSON),
			})
		}
		
		resp.Diagnostics.AddError("API Error", errorMsg)
		return
	}

	// Preserve immutable fields from state
	data.Id = state.Id
	data.ProjectID = state.ProjectID
	data.Uri = state.Uri
	// Preserve network configuration from state
	// Network configuration is immutable, so we preserve it from state
	data.Network = state.Network
	// Storage is preserved from plan (which includes autoscaling)

	// Re-read the DBaaS instance to get the latest state
	getResp, err := r.client.Client.FromDatabase().DBaaS().Get(ctx, projectID, dbaasID, nil)
	if err == nil && getResp != nil && getResp.Data != nil {
		dbaas := getResp.Data
		if dbaas.Metadata.URI != nil {
			data.Uri = types.StringValue(*dbaas.Metadata.URI)
		}
		// Note: Network URI references are preserved from state
		// They are not yet available in the SDK response structure
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DBaaSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DBaaSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	dbaasID := data.Id.ValueString()

	if projectID == "" || dbaasID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Fields",
			"Project ID and DBaaS ID are required to delete the DBaaS instance",
		)
		return
	}

	// Delete the DBaaS instance using the SDK with retry mechanism
	// Retry on any error except 404 (Resource Not Found)
	err := DeleteResourceWithRetry(
		ctx,
		func() (interface{}, error) {
			return r.client.Client.FromDatabase().DBaaS().Delete(ctx, projectID, dbaasID, nil)
		},
		ExtractSDKError,
		"DBaaS",
		dbaasID,
		r.client.ResourceTimeout,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting DBaaS instance",
			fmt.Sprintf("Unable to delete DBaaS instance: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a DBaaS resource", map[string]interface{}{
		"dbaas_id": dbaasID,
	})
}

func (r *DBaaSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
