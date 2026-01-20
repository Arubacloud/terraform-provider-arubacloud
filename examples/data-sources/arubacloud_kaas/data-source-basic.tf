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
output "kaas_billing_period" {
  value = data.arubacloud_kaas.basic.billing_period
}
output "kaas_vpc_uri_ref" {
  value = data.arubacloud_kaas.basic.vpc_uri_ref
}
output "kaas_subnet_uri_ref" {
  value = data.arubacloud_kaas.basic.subnet_uri_ref
}
output "kaas_node_cidr_address" {
  value = data.arubacloud_kaas.basic.node_cidr_address
}
output "kaas_node_cidr_name" {
  value = data.arubacloud_kaas.basic.node_cidr_name
}
output "kaas_security_group_name" {
  value = data.arubacloud_kaas.basic.security_group_name
}
output "kaas_pod_cidr" {
  value = data.arubacloud_kaas.basic.pod_cidr
}
output "kaas_kubernetes_version" {
  value = data.arubacloud_kaas.basic.kubernetes_version
}
output "kaas_node_pools" {
  value = data.arubacloud_kaas.basic.node_pools
}
