# Input Variables for Testing All Datasources
# Fill these with actual resource IDs from your account to test datasources

# Provider Credentials
variable "arubacloud_api_key" {
  description = "ArubaCloud API Key"
  type        = string
  sensitive   = true
}

variable "arubacloud_api_secret" {
  description = "ArubaCloud API Secret"
  type        = string
  sensitive   = true
}

# Core Resources
variable "test_project_id" {
  description = "Project ID to test project datasource"
  type        = string
  default     = "68398923fb2cb026400d4d31"
}

# Compute Resources
variable "test_cloudserver_id" {
  description = "Cloud Server ID to test cloudserver datasource"
  type        = string
  default     = "69007ece4e7d691466d86223"
}

#variable "test_snapshot_id" {
#  description = "Snapshot ID to test snapshot datasource"
#  type        = string
#  default     = ""
#}

#variable "test_backup_id" {
#  description = "Backup ID to test backup datasource"
#  type        = string
#  default     = ""
#}

#variable "test_restore_id" {
#  description = "Restore ID to test restore datasource"
#  type        = string
#  default     = ""
#}

# Storage Resources
variable "test_blockstorage_id" {
  description = "Block Storage ID to test blockstorage datasource"
  type        = string
  default     = "69087e12cc4c7793b9e4d2eb"
}

# Network Resources
variable "test_vpc_id" {
  description = "VPC ID to test vpc datasource"
  type        = string
  default     = "69495ef64d0cdc87949b71ec"
}

variable "test_subnet_id" {
  description = "Subnet ID to test subnet datasource"
  type        = string
  default     = "694ba1737712ac0032dbe50a"
}

variable "test_elasticip_id" {
  description = "Elastic IP ID to test elasticip datasource"
  type        = string
  default     = ""
}

variable "test_securitygroup_id" {
  description = "Security Group ID to test securitygroup datasource"
  type        = string
  default     = "694bb9817712ac0032dbe648"
}

variable "test_securityrule_id" {
  description = "Security Rule ID to test securityrule datasource"
  type        = string
  default     = "694b06564d0cdc87949b7608"
}

# VPN Resources
#variable "test_vpntunnel_id" {
#  description = "VPN Tunnel ID to test vpntunnel datasource"
#  type        = string
#  default     = ""
#}

#variable "test_vpnroute_id" {
#  description = "VPN Route ID to test vpnroute datasource"
#  type        = string
#  default     = ""
#}

# VPC Peering Resources
#variable "test_vpcpeering_id" {
#  description = "VPC Peering ID to test vpcpeering datasource"
#  type        = string
#  default     = ""
#}

#variable "test_vpcpeeringroute_id" {
#  description = "VPC Peering Route ID to test vpcpeeringroute datasource"
#  type        = string
#  default     = ""
#}

# Container Resources
variable "test_containerregistry_id" {
  description = "Container Registry ID to test containerregistry datasource"
  type        = string
  default     = "69087e7256594a088913e09f"
}

variable "test_kaas_id" {
  description = "Kubernetes as a Service ID to test kaas datasource"
  type        = string
  default     = "694ff33bc2682f8c02f4956e"
}

# Database Resources
variable "test_database_id" {
  description = "Database ID to test database datasource"
  type        = string
  default     = "wordpress"
}

#variable "test_databasebackup_id" {
#  description = "Database Backup ID to test databasebackup datasource"
#  type        = string
#  default     = ""
#}

#variable "test_databasegrant_id" {
#  description = "Database Grant ID to test databasegrant datasource"
#  type        = string
#  default     = ""
#}

variable "test_dbaas_id" {
  description = "DBaaS ID to test dbaas datasource"
  type        = string
  default     = "68ff8d8fc1445aeb83f79438"
}

#variable "test_dbaasuser_id" {
#  description = "DBaaS User ID to test dbaasuser datasource"
#  type        = string
#  default     = "wordpress"
#}

# Security Resources
#variable "test_kms_id" {
#  description = "KMS ID to test kms datasource"
#  type        = string
#  default     = ""
#}

variable "test_keypair_id" {
  description = "Key Pair ID to test keypair datasource"
  type        = string
  default     = "6676d769ba4e44cdfd373f9f"
}

# Automation Resources
#variable "test_schedulejob_id" {
#  description = "Schedule Job ID to test schedulejob datasource"
#  type        = string
#  default     = ""
#}
