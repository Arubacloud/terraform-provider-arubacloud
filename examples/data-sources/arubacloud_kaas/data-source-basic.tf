data "arubacloud_kaas" "basic" {
  id = "your-kaas-id"
}

output "kaas_name" {
  value = data.arubacloud_kaas.basic.name
}
output "kaas_location" {
  value = data.arubacloud_kaas.basic.location
}
output "kaas_tags" {
  value = data.arubacloud_kaas.basic.tags
}
output "kaas_project_id" {
  value = data.arubacloud_kaas.basic.project_id
}
output "kaas_preset" {
  value = data.arubacloud_kaas.basic.preset
}
output "kaas_vpc_id" {
  value = data.arubacloud_kaas.basic.vpc_id
}
output "kaas_subnet_id" {
  value = data.arubacloud_kaas.basic.subnet_id
}
output "kaas_node_cidr" {
  value = data.arubacloud_kaas.basic.node_cidr
}
output "kaas_security_group_name" {
  value = data.arubacloud_kaas.basic.security_group_name
}
output "kaas_version" {
  value = data.arubacloud_kaas.basic.version
}
output "kaas_node_pools" {
  value = data.arubacloud_kaas.basic.node_pools
}
output "kaas_ha" {
  value = data.arubacloud_kaas.basic.ha
}
output "kaas_billing_period" {
  value = data.arubacloud_kaas.basic.billing_period
}
