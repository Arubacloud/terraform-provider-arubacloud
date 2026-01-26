# All Available Datasources for Testing
# Uncomment and provide the required ID in 00-input-variables.tf to test each datasource

# Core Resources
data "arubacloud_project" "test" {
  count = var.test_project_id != "" ? 1 : 0
  id    = var.test_project_id
}

# Compute Resources
data "arubacloud_cloudserver" "test" {
  count = var.test_cloudserver_id != "" ? 1 : 0
  id    = var.test_cloudserver_id
}

#data "arubacloud_snapshot" "test" {
#  count = var.test_snapshot_id != "" ? 1 : 0
#  id    = var.test_snapshot_id
#}

#data "arubacloud_backup" "test" {
#  count = var.test_backup_id != "" ? 1 : 0
#  id    = var.test_backup_id
#}

#data "arubacloud_restore" "test" {
#  count = var.test_restore_id != "" ? 1 : 0
#  id    = var.test_restore_id
#}

# Storage Resources
data "arubacloud_blockstorage" "test" {
  count = var.test_blockstorage_id != "" ? 1 : 0
  id    = var.test_blockstorage_id
}

# Network Resources
data "arubacloud_vpc" "test" {
  count = var.test_vpc_id != "" ? 1 : 0
  id    = var.test_vpc_id
}

data "arubacloud_subnet" "test" {
  count = var.test_subnet_id != "" ? 1 : 0
  id    = var.test_subnet_id
}

data "arubacloud_elasticip" "test" {
  count = var.test_elasticip_id != "" ? 1 : 0
  id    = var.test_elasticip_id
}

data "arubacloud_securitygroup" "test" {
  count = var.test_securitygroup_id != "" ? 1 : 0
  id    = var.test_securitygroup_id
}

data "arubacloud_securityrule" "test" {
  count = var.test_securityrule_id != "" ? 1 : 0
  id    = var.test_securityrule_id
}

# VPN Resources
#data "arubacloud_vpntunnel" "test" {
#  count = var.test_vpntunnel_id != "" ? 1 : 0
#  id    = var.test_vpntunnel_id
#}

#data "arubacloud_vpnroute" "test" {
#  count = var.test_vpnroute_id != "" ? 1 : 0
#  id    = var.test_vpnroute_id
#}

# VPC Peering Resources
#data "arubacloud_vpcpeering" "test" {
#  count = var.test_vpcpeering_id != "" ? 1 : 0
#  id    = var.test_vpcpeering_id
#}

#data "arubacloud_vpcpeeringroute" "test" {
#  count = var.test_vpcpeeringroute_id != "" ? 1 : 0
#  id    = var.test_vpcpeeringroute_id
#}

# Container Resources
data "arubacloud_containerregistry" "test" {
  count = var.test_containerregistry_id != "" ? 1 : 0
  id    = var.test_containerregistry_id
}

data "arubacloud_kaas" "test" {
  count = var.test_kaas_id != "" ? 1 : 0
  id    = var.test_kaas_id
}

# Database Resources
data "arubacloud_database" "test" {
  count = var.test_database_id != "" ? 1 : 0
  id    = var.test_database_id
}

#data "arubacloud_databasebackup" "test" {
#  count = var.test_databasebackup_id != "" ? 1 : 0
#  id    = var.test_databasebackup_id
#}

#data "arubacloud_databasegrant" "test" {
#  count = var.test_databasegrant_id != "" ? 1 : 0
#  id    = var.test_databasegrant_id
#}

data "arubacloud_dbaas" "test" {
  count = var.test_dbaas_id != "" ? 1 : 0
  id    = var.test_dbaas_id
}

#data "arubacloud_dbaasuser" "test" {
#  count = var.test_dbaasuser_id != "" ? 1 : 0
#  id    = var.test_dbaasuser_id
#}

# Security Resources
#data "arubacloud_kms" "test" {
#  count = var.test_kms_id != "" ? 1 : 0
#  id    = var.test_kms_id
#}

data "arubacloud_keypair" "test" {
  count = var.test_keypair_id != "" ? 1 : 0
  id    = var.test_keypair_id
}

# Automation Resources
#data "arubacloud_schedulejob" "test" {
#  count = var.test_schedulejob_id != "" ? 1 : 0
#  id    = var.test_schedulejob_id
#}
