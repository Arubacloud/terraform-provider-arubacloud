package types

// DatabaseEngine identifies a DBaaS engine.
//
// The constants below mirror the official catalog at
// https://api.arubacloud.com/docs/metadata/ — values not in the catalog
// will be rejected by the API. The authoritative live catalog is at:
//
//	GET /providers/Aruba.Database/engines
type DatabaseEngine string

const (
	DatabaseEngineMySQL80             DatabaseEngine = "mysql-8.0"
	DatabaseEngineMSSQL2022Web        DatabaseEngine = "mssql-2022-web"
	DatabaseEngineMSSQL2022Standard   DatabaseEngine = "mssql-2022-standard"
	DatabaseEngineMSSQL2022Enterprise DatabaseEngine = "mssql-2022-enterprise"
)

// DBaaSFlavor identifies a DBaaS flavor SKU.
//
// Pattern: DBO<vCPU>A<RAM>. The constants below cover SKUs referenced in
// SDK fixtures. The authoritative list is available via:
//
//	GET /providers/Aruba.Database/flavors
type DBaaSFlavor string

const (
	DBaaSFlavorDBO1A2   DBaaSFlavor = "DBO1A2"
	DBaaSFlavorDBO1A4   DBaaSFlavor = "DBO1A4"
	DBaaSFlavorDBO2A4   DBaaSFlavor = "DBO2A4"
	DBaaSFlavorDBO2A8   DBaaSFlavor = "DBO2A8"
	DBaaSFlavorDBO4A8   DBaaSFlavor = "DBO4A8"
	DBaaSFlavorDBO4A16  DBaaSFlavor = "DBO4A16"
	DBaaSFlavorDBO8A16  DBaaSFlavor = "DBO8A16"
	DBaaSFlavorDBO8A32  DBaaSFlavor = "DBO8A32"
	DBaaSFlavorDBO12A24 DBaaSFlavor = "DBO12A24"
	DBaaSFlavorDBO16A32 DBaaSFlavor = "DBO16A32"
	DBaaSFlavorDBO16A64 DBaaSFlavor = "DBO16A64"
	DBaaSFlavorDBO24A48 DBaaSFlavor = "DBO24A48"
	DBaaSFlavorDBO32A64 DBaaSFlavor = "DBO32A64"
)

// DBaaSEngineRequest contains the database engine configuration
type DBaaSEngineRequest struct {
	// ID Type of DB engine to activate (nullable)
	// For more information, check the documentation.
	ID *DatabaseEngine `json:"id,omitempty"`

	// DataCenter Datacenter location (nullable)
	// For more information, check the documentation.
	DataCenter *string `json:"dataCenter,omitempty"`
}

// DBaaSEngineResponse contains the database engine response configuration
type DBaaSEngineResponse struct {
	// ID Engine identifier (nullable)
	ID *string `json:"id,omitempty"`

	// Type Engine type (nullable)
	Type *string `json:"type,omitempty"`

	// Name Engine name (nullable)
	Name *string `json:"name,omitempty"`

	// Version Engine version (nullable)
	Version *string `json:"version,omitempty"`

	// DataCenter Datacenter location (nullable)
	// For more information, check the documentation.
	DataCenter *string `json:"dataCenter,omitempty"`

	// PrivateIPAddress Private IP address (nullable)
	PrivateIPAddress *string `json:"privateIpAddress,omitempty"`
}

// DBaaSFlavorRequest contains the flavor configuration for a DBaaS request.
type DBaaSFlavorRequest struct {
	// Name Type of flavor to use (nullable)
	// For more information, check the documentation.
	Name *DBaaSFlavor `json:"name,omitempty"`
}

// DBaaSFlavorResponse contains the flavor response configuration
type DBaaSFlavorResponse struct {
	// Name Flavor name (nullable)
	Name *string `json:"name,omitempty"`

	// Category Flavor category (nullable)
	Category *string `json:"category,omitempty"`

	// CPU Number of CPUs (nullable)
	CPU *int32 `json:"cpu,omitempty"`

	// RAM Amount of RAM in MB (nullable)
	RAM *int32 `json:"ram,omitempty"`
}

// DBaaSStorageRequest contains the storage configuration
type DBaaSStorageRequest struct {
	// SizeGB Size in GB to use (nullable)
	SizeGB *int32 `json:"sizeGb,omitempty"`
}

// DBaaSStorageResponse contains the storage response configuration
type DBaaSStorageResponse struct {
	// SizeGB Size in GB (nullable)
	SizeGB *int32 `json:"sizeGb,omitempty"`
}

// DBaaSNetworkingRequest contains the network information to use when creating the new DBaaS
type DBaaSNetworkingRequest struct {
	// VPCURI The URI of the VPC resource to bind to this DBaaS instance (nullable)
	// Required when user has at least one VPC (with at least one subnet and a security group).
	VPCURI *string `json:"vpcUri,omitempty"`

	// SubnetURI The URI of the Subnet resource to bind to this DBaaS instance (nullable)
	// It must belong to the VPC defined in VPCURI
	// Required when user has at least one VPC (with at least one subnet and a security group).
	SubnetURI *string `json:"subnetUri,omitempty"`

	// SecurityGroupURI The URI of the SecurityGroup resource to bind to this DBaaS instance (nullable)
	// It must belong to the VPC defined in VPCURI
	// Required when user has at least one VPC (with at least one subnet and a security group).
	SecurityGroupURI *string `json:"securityGroupUri,omitempty"`

	// ElasticIPURI The URI of the ElasticIP resource to bind to this DBaaS instance (nullable)
	ElasticIPURI *string `json:"elasticIpUri,omitempty"`
}

// DBaaSNetworkingResponse contains the network response information
type DBaaSNetworkingResponse struct {
	// VPC VPC resource reference (nullable)
	VPC *ReferenceResourceCommon `json:"vpc,omitempty"`

	// Subnet Subnet resource reference (nullable)
	Subnet *ReferenceResourceCommon `json:"subnet,omitempty"`

	// SecurityGroup Security group resource reference (nullable)
	SecurityGroup *ReferenceResourceCommon `json:"securityGroup,omitempty"`

	// ElasticIP Elastic IP resource reference (nullable)
	ElasticIP *ReferenceResourceCommon `json:"elasticIp,omitempty"`
}

// DBaaSAutoscalingRequest contains the autoscaling configuration
type DBaaSAutoscalingRequest struct {
	// Enabled Indicates if autoscaling is enabled (nullable)
	Enabled *bool `json:"enabled,omitempty"`

	// AvailableSpace Available space threshold (nullable)
	AvailableSpace *int32 `json:"availableSpace,omitempty"`

	// StepSize Step size for autoscaling (nullable)
	StepSize *int32 `json:"stepSize,omitempty"`
}

// DBaaSAutoscalingResponse contains the autoscaling response configuration
type DBaaSAutoscalingResponse struct {
	// Status Autoscaling status (nullable)
	Status *string `json:"status,omitempty"`

	// AvailableSpace Available space threshold (nullable)
	AvailableSpace *int32 `json:"availableSpace,omitempty"`

	// StepSize Step size for autoscaling (nullable)
	StepSize *int32 `json:"stepSize,omitempty"`

	// RuleID Rule identifier (nullable)
	RuleID *string `json:"ruleId,omitempty"`
}

// DBaaSPropertiesRequest contains properties required to create a DBaaS instance
type DBaaSPropertiesRequest struct {
	// Zone where DBaaS will be created (optional).
	// If specified, the resource is zonal; otherwise, it is regional.
	Zone *Zone `json:"dataCenter,omitempty"`

	// Engine Database engine configuration
	Engine *DBaaSEngineRequest `json:"engine,omitempty"`

	// Flavor Flavor configuration
	Flavor *DBaaSFlavorRequest `json:"flavor,omitempty"`

	// Storage Storage configuration
	Storage *DBaaSStorageRequest `json:"storage,omitempty"`

	// BillingPlanCommon Billing plan (wraps billingPeriod)
	BillingPlanCommon *BillingPlanCommon `json:"billingPlan,omitempty"`

	// Networking Network information for the DBaaS instance
	Networking *DBaaSNetworkingRequest `json:"networking,omitempty"`

	// Autoscaling Autoscaling configuration
	Autoscaling *DBaaSAutoscalingRequest `json:"autoscaling,omitempty"`
}

// DBaaSPropertiesResponse contains the response properties of a DBaaS instance
type DBaaSPropertiesResponse struct {
	// LinkedResources Array of resources linked to the DBaaS instance (nullable)
	LinkedResources []LinkedResourceCommon `json:"linkedResources,omitempty"`

	// Engine Database engine response configuration
	Engine *DBaaSEngineResponse `json:"engine,omitempty"`

	// Flavor Flavor response configuration
	Flavor *DBaaSFlavorResponse `json:"flavor,omitempty"`

	// Networking Network response configuration
	Networking *DBaaSNetworkingResponse `json:"networking,omitempty"`

	// Storage Storage response configuration
	Storage *DBaaSStorageResponse `json:"storage,omitempty"`

	// BillingPlanCommon Billing plan (wraps billingPeriod)
	BillingPlanCommon *BillingPlanCommon `json:"billingPlan,omitempty"`

	// Autoscaling Autoscaling response configuration
	Autoscaling *DBaaSAutoscalingResponse `json:"autoscaling,omitempty"`
}

type DBaaSRequest struct {
	// Metadata of the DBaaS instance
	Metadata RegionalResourceMetadataRequest `json:"metadata"`

	// Spec contains the DBaaS instance specification
	Properties DBaaSPropertiesRequest `json:"properties"`
}

type DBaaSResponse struct {
	// Metadata of the DBaaS instance
	Metadata ResourceMetadataResponse `json:"metadata"`

	// Spec contains the DBaaS instance specification
	Properties DBaaSPropertiesResponse `json:"properties"`

	Status ResourceStatusResponse `json:"status,omitempty"`
}

type DBaaSListResponse struct {
	ListResponse
	Values []DBaaSResponse `json:"values"`
}
