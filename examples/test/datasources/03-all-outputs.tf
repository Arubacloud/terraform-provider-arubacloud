# Outputs for All Datasources
# Only outputs will be shown for datasources where you provided an ID
# Shows ALL available fields from each datasource

# Core Resources
output "project_datasource" {
  description = "Project datasource - all fields"
  value       = var.test_project_id != "" ? {
    id          = try(data.arubacloud_project.test[0].id, "")
    name        = try(data.arubacloud_project.test[0].name, "")
    description = try(data.arubacloud_project.test[0].description, "")
    tags        = try(data.arubacloud_project.test[0].tags, [])
  } : null
}

# Compute Resources
output "cloudserver_datasource" {
  description = "Cloud Server datasource - all fields (flattened)"
  value       = var.test_cloudserver_id != "" ? {
    id                       = try(data.arubacloud_cloudserver.test[0].id, "")
    uri                      = try(data.arubacloud_cloudserver.test[0].uri, "")
    name                     = try(data.arubacloud_cloudserver.test[0].name, "")
    location                 = try(data.arubacloud_cloudserver.test[0].location, "")
    project_id               = try(data.arubacloud_cloudserver.test[0].project_id, "")
    zone                     = try(data.arubacloud_cloudserver.test[0].zone, "")
    tags                     = try(data.arubacloud_cloudserver.test[0].tags, [])
    # Network fields (flattened)
    vpc_uri_ref              = try(data.arubacloud_cloudserver.test[0].vpc_uri_ref, "")
    elastic_ip_uri_ref       = try(data.arubacloud_cloudserver.test[0].elastic_ip_uri_ref, "")
    subnet_uri_refs          = try(data.arubacloud_cloudserver.test[0].subnet_uri_refs, [])
    securitygroup_uri_refs   = try(data.arubacloud_cloudserver.test[0].securitygroup_uri_refs, [])
    # Settings fields (flattened)
    flavor_name              = try(data.arubacloud_cloudserver.test[0].flavor_name, "")
    key_pair_uri_ref         = try(data.arubacloud_cloudserver.test[0].key_pair_uri_ref, "")
    user_data                = try(data.arubacloud_cloudserver.test[0].user_data, "")
    # Storage fields (flattened)
    boot_volume_uri_ref      = try(data.arubacloud_cloudserver.test[0].boot_volume_uri_ref, "")
  } : null
}

#output "snapshot_datasource" {
#  description = "Snapshot datasource details"
#  value       = var.test_snapshot_id != "" ? {
#    id   = try(data.arubacloud_snapshot.test[0].id, "")
#    name = try(data.arubacloud_snapshot.test[0].name, "")
#  } : null
#}

#output "backup_datasource" {
#  description = "Backup datasource details"
#  value       = var.test_backup_id != "" ? {
#    id   = try(data.arubacloud_backup.test[0].id, "")
#    name = try(data.arubacloud_backup.test[0].name, "")
#  } : null
#}

#output "restore_datasource" {
#  description = "Restore datasource details"
#  value       = var.test_restore_id != "" ? {
#    id   = try(data.arubacloud_restore.test[0].id, "")
#    name = try(data.arubacloud_restore.test[0].name, "")
#  } : null
#}

# Storage Resources
output "blockstorage_datasource" {
  description = "Block Storage datasource - all fields"
  value       = var.test_blockstorage_id != "" ? {
    id             = try(data.arubacloud_blockstorage.test[0].id, "")
    name           = try(data.arubacloud_blockstorage.test[0].name, "")
    project_id     = try(data.arubacloud_blockstorage.test[0].project_id, "")
    location       = try(data.arubacloud_blockstorage.test[0].location, "")
    size_gb        = try(data.arubacloud_blockstorage.test[0].size_gb, 0)
    billing_period = try(data.arubacloud_blockstorage.test[0].billing_period, "")
    zone           = try(data.arubacloud_blockstorage.test[0].zone, "")
    type           = try(data.arubacloud_blockstorage.test[0].type, "")
    tags           = try(data.arubacloud_blockstorage.test[0].tags, [])
    snapshot_id    = try(data.arubacloud_blockstorage.test[0].snapshot_id, "")
    bootable       = try(data.arubacloud_blockstorage.test[0].bootable, false)
    image          = try(data.arubacloud_blockstorage.test[0].image, "")
  } : null
}

# Network Resources
output "vpc_datasource" {
  description = "VPC datasource - all fields"
  value       = var.test_vpc_id != "" ? {
    id         = try(data.arubacloud_vpc.test[0].id, "")
    name       = try(data.arubacloud_vpc.test[0].name, "")
    location   = try(data.arubacloud_vpc.test[0].location, "")
    project_id = try(data.arubacloud_vpc.test[0].project_id, "")
    tags       = try(data.arubacloud_vpc.test[0].tags, [])
  } : null
}

output "subnet_datasource" {
  description = "Subnet datasource - all fields (flattened)"
  value       = var.test_subnet_id != "" ? {
    id           = try(data.arubacloud_subnet.test[0].id, "")
    uri          = try(data.arubacloud_subnet.test[0].uri, "")
    name         = try(data.arubacloud_subnet.test[0].name, "")
    location     = try(data.arubacloud_subnet.test[0].location, "")
    tags         = try(data.arubacloud_subnet.test[0].tags, [])
    project_id   = try(data.arubacloud_subnet.test[0].project_id, "")
    vpc_id       = try(data.arubacloud_subnet.test[0].vpc_id, "")
    type         = try(data.arubacloud_subnet.test[0].type, "")
    # Network fields (flattened)
    address      = try(data.arubacloud_subnet.test[0].address, "")
    # DHCP fields (flattened)
    dhcp_enabled = try(data.arubacloud_subnet.test[0].dhcp_enabled, false)
    dhcp_routes  = try(data.arubacloud_subnet.test[0].dhcp_routes, [])
  } : null
}

output "elasticip_datasource" {
  description = "Elastic IP datasource - all fields"
  value       = var.test_elasticip_id != "" ? {
    id             = try(data.arubacloud_elasticip.test[0].id, "")
    name           = try(data.arubacloud_elasticip.test[0].name, "")
    location       = try(data.arubacloud_elasticip.test[0].location, "")
    project_id     = try(data.arubacloud_elasticip.test[0].project_id, "")
    address        = try(data.arubacloud_elasticip.test[0].address, "")
    billing_period = try(data.arubacloud_elasticip.test[0].billing_period, "")
    tags           = try(data.arubacloud_elasticip.test[0].tags, [])
  } : null
}

output "securitygroup_datasource" {
  description = "Security Group datasource - all fields"
  value       = var.test_securitygroup_id != "" ? {
    id         = try(data.arubacloud_securitygroup.test[0].id, "")
    name       = try(data.arubacloud_securitygroup.test[0].name, "")
    location   = try(data.arubacloud_securitygroup.test[0].location, "")
    tags       = try(data.arubacloud_securitygroup.test[0].tags, [])
    project_id = try(data.arubacloud_securitygroup.test[0].project_id, "")
    vpc_id     = try(data.arubacloud_securitygroup.test[0].vpc_id, "")
  } : null
}

output "securityrule_datasource" {
  description = "Security Rule datasource - all fields (flattened)"
  value       = var.test_securityrule_id != "" ? {
    id                = try(data.arubacloud_securityrule.test[0].id, "")
    uri               = try(data.arubacloud_securityrule.test[0].uri, "")
    name              = try(data.arubacloud_securityrule.test[0].name, "")
    location          = try(data.arubacloud_securityrule.test[0].location, "")
    tags              = try(data.arubacloud_securityrule.test[0].tags, [])
    project_id        = try(data.arubacloud_securityrule.test[0].project_id, "")
    vpc_id            = try(data.arubacloud_securityrule.test[0].vpc_id, "")
    security_group_id = try(data.arubacloud_securityrule.test[0].security_group_id, "")
    # Properties fields (flattened)
    direction         = try(data.arubacloud_securityrule.test[0].direction, "")
    protocol          = try(data.arubacloud_securityrule.test[0].protocol, "")
    port              = try(data.arubacloud_securityrule.test[0].port, "")
    target_kind       = try(data.arubacloud_securityrule.test[0].target_kind, "")
    target_value      = try(data.arubacloud_securityrule.test[0].target_value, "")
  } : null
}

# VPN Resources
#output "vpntunnel_datasource" {
#  description = "VPN Tunnel datasource details"
#  value       = var.test_vpntunnel_id != "" ? {
#    id   = try(data.arubacloud_vpntunnel.test[0].id, "")
#    name = try(data.arubacloud_vpntunnel.test[0].name, "")
#  } : null
#}

#output "vpnroute_datasource" {
#  description = "VPN Route datasource details"
#  value       = var.test_vpnroute_id != "" ? {
#    id   = try(data.arubacloud_vpnroute.test[0].id, "")
#    name = try(data.arubacloud_vpnroute.test[0].name, "")
#  } : null
#}

# VPC Peering Resources
#output "vpcpeering_datasource" {
#  description = "VPC Peering datasource details"
#  value       = var.test_vpcpeering_id != "" ? {
#    id   = try(data.arubacloud_vpcpeering.test[0].id, "")
#    name = try(data.arubacloud_vpcpeering.test[0].name, "")
#  } : null
#}

#output "vpcpeeringroute_datasource" {
#  description = "VPC Peering Route datasource details"
#  value       = var.test_vpcpeeringroute_id != "" ? {
#    id   = try(data.arubacloud_vpcpeeringroute.test[0].id, "")
#    name = try(data.arubacloud_vpcpeeringroute.test[0].name, "")
#  } : null
#}

# Container Resources
output "containerregistry_datasource" {
  description = "Container Registry datasource - all fields (flattened)"
  value       = var.test_containerregistry_id != "" ? {
    id                          = try(data.arubacloud_containerregistry.test[0].id, "")
    uri                         = try(data.arubacloud_containerregistry.test[0].uri, "")
    name                        = try(data.arubacloud_containerregistry.test[0].name, "")
    location                    = try(data.arubacloud_containerregistry.test[0].location, "")
    tags                        = try(data.arubacloud_containerregistry.test[0].tags, [])
    project_id                  = try(data.arubacloud_containerregistry.test[0].project_id, "")
    billing_period              = try(data.arubacloud_containerregistry.test[0].billing_period, "")
    # Network fields (flattened)
    public_ip_uri_ref           = try(data.arubacloud_containerregistry.test[0].public_ip_uri_ref, "")
    vpc_uri_ref                 = try(data.arubacloud_containerregistry.test[0].vpc_uri_ref, "")
    subnet_uri_ref              = try(data.arubacloud_containerregistry.test[0].subnet_uri_ref, "")
    security_group_uri_ref      = try(data.arubacloud_containerregistry.test[0].security_group_uri_ref, "")
    # Storage fields (flattened)
    block_storage_uri_ref       = try(data.arubacloud_containerregistry.test[0].block_storage_uri_ref, "")
    # Settings fields (flattened)
    concurrent_users_flavor     = try(data.arubacloud_containerregistry.test[0].concurrent_users_flavor, "")
  } : null
}

output "kaas_datasource" {
  description = "Kubernetes as a Service datasource - all fields (flattened)"
  value       = var.test_kaas_id != "" ? {
    id                    = try(data.arubacloud_kaas.test[0].id, "")
    uri                   = try(data.arubacloud_kaas.test[0].uri, "")
    name                  = try(data.arubacloud_kaas.test[0].name, "")
    location              = try(data.arubacloud_kaas.test[0].location, "")
    tags                  = try(data.arubacloud_kaas.test[0].tags, [])
    project_id            = try(data.arubacloud_kaas.test[0].project_id, "")
    billing_period        = try(data.arubacloud_kaas.test[0].billing_period, "")
    management_ip         = try(data.arubacloud_kaas.test[0].management_ip, "")
    # Network fields (flattened)
    vpc_uri_ref           = try(data.arubacloud_kaas.test[0].vpc_uri_ref, "")
    subnet_uri_ref        = try(data.arubacloud_kaas.test[0].subnet_uri_ref, "")
    node_cidr_address     = try(data.arubacloud_kaas.test[0].node_cidr_address, "")
    node_cidr_name        = try(data.arubacloud_kaas.test[0].node_cidr_name, "")
    security_group_name   = try(data.arubacloud_kaas.test[0].security_group_name, "")
    # Settings fields (flattened)
    pod_cidr              = try(data.arubacloud_kaas.test[0].pod_cidr, "")
    kubernetes_version    = try(data.arubacloud_kaas.test[0].kubernetes_version, "")
    node_pools            = try(data.arubacloud_kaas.test[0].node_pools, [])
  } : null
}

# Database Resources
output "database_datasource" {
  description = "Database datasource - all fields"
  value       = var.test_database_id != "" ? {
    id         = try(data.arubacloud_database.test[0].id, "")
    project_id = try(data.arubacloud_database.test[0].project_id, "")
    dbaas_id   = try(data.arubacloud_database.test[0].dbaas_id, "")
    name       = try(data.arubacloud_database.test[0].name, "")
  } : null
}

#output "databasebackup_datasource" {
#  description = "Database Backup datasource details"
#  value       = var.test_databasebackup_id != "" ? {
#    id   = try(data.arubacloud_databasebackup.test[0].id, "")
#    name = try(data.arubacloud_databasebackup.test[0].name, "")
#  } : null
#}

#output "databasegrant_datasource" {
#  description = "Database Grant datasource details"
#  value       = var.test_databasegrant_id != "" ? {
#    id   = try(data.arubacloud_databasegrant.test[0].id, "")
#    name = try(data.arubacloud_databasegrant.test[0].name, "")
#  } : null
#}

output "dbaas_datasource" {
  description = "DBaaS datasource - all fields (flattened)"
  value       = var.test_dbaas_id != "" ? {
    id                          = try(data.arubacloud_dbaas.test[0].id, "")
    uri                         = try(data.arubacloud_dbaas.test[0].uri, "")
    name                        = try(data.arubacloud_dbaas.test[0].name, "")
    location                    = try(data.arubacloud_dbaas.test[0].location, "")
    zone                        = try(data.arubacloud_dbaas.test[0].zone, "")
    tags                        = try(data.arubacloud_dbaas.test[0].tags, [])
    project_id                  = try(data.arubacloud_dbaas.test[0].project_id, "")
    engine_id                   = try(data.arubacloud_dbaas.test[0].engine_id, "")
    flavor                      = try(data.arubacloud_dbaas.test[0].flavor, "")
    billing_period              = try(data.arubacloud_dbaas.test[0].billing_period, "")
    # Storage fields (flattened)
    storage_size_gb             = try(data.arubacloud_dbaas.test[0].storage_size_gb, 0)
    autoscaling_enabled         = try(data.arubacloud_dbaas.test[0].autoscaling_enabled, false)
    autoscaling_available_space = try(data.arubacloud_dbaas.test[0].autoscaling_available_space, 0)
    autoscaling_step_size       = try(data.arubacloud_dbaas.test[0].autoscaling_step_size, 0)
    # Network fields (flattened)
    vpc_uri_ref                 = try(data.arubacloud_dbaas.test[0].vpc_uri_ref, "")
    subnet_uri_ref              = try(data.arubacloud_dbaas.test[0].subnet_uri_ref, "")
    security_group_uri_ref      = try(data.arubacloud_dbaas.test[0].security_group_uri_ref, "")
    elastic_ip_uri_ref          = try(data.arubacloud_dbaas.test[0].elastic_ip_uri_ref, "")
  } : null
}

#output "dbaasuser_datasource" {
#  description = "DBaaS User datasource details"
#  value       = var.test_dbaasuser_id != "" ? {
#    id   = try(data.arubacloud_dbaasuser.test[0].id, "")
#    name = try(data.arubacloud_dbaasuser.test[0].name, "")
#  } : null
#}

# Security Resources
#output "kms_datasource" {
#  description = "KMS datasource details"
#  value       = var.test_kms_id != "" ? {
#    id   = try(data.arubacloud_kms.test[0].id, "")
#    name = try(data.arubacloud_kms.test[0].name, "")
#  } : null
#}

output "keypair_datasource" {
  description = "Key Pair datasource - all fields"
  value       = var.test_keypair_id != "" ? {
    id         = try(data.arubacloud_keypair.test[0].id, "")
    name       = try(data.arubacloud_keypair.test[0].name, "")
    location   = try(data.arubacloud_keypair.test[0].location, "")
    project_id = try(data.arubacloud_keypair.test[0].project_id, "")
    value      = try(data.arubacloud_keypair.test[0].value, "")
    tags       = try(data.arubacloud_keypair.test[0].tags, [])
  } : null
}

# Automation Resources
#output "schedulejob_datasource" {
#  description = "Schedule Job datasource details"
#  value       = var.test_schedulejob_id != "" ? {
#    id   = try(data.arubacloud_schedulejob.test[0].id, "")
#    name = try(data.arubacloud_schedulejob.test[0].name, "")
#  } : null
#}
