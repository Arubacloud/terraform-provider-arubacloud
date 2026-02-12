data "arubacloud_dbaas" "basic" {
  id = "dbaas-id"
}

output "dbaas_name" {
  value = data.arubacloud_dbaas.basic.name
}
output "dbaas_location" {
  value = data.arubacloud_dbaas.basic.location
}
output "dbaas_project_id" {
  value = data.arubacloud_dbaas.basic.project_id
}
output "dbaas_storage_size_gb" {
  value = data.arubacloud_dbaas.basic.storage_size_gb
}
output "dbaas_autoscaling_enabled" {
  value = data.arubacloud_dbaas.basic.autoscaling_enabled
}
output "dbaas_autoscaling_available_space" {
  value = data.arubacloud_dbaas.basic.autoscaling_available_space
}
output "dbaas_autoscaling_step_size" {
  value = data.arubacloud_dbaas.basic.autoscaling_step_size
}
output "dbaas_vpc_uri_ref" {
  value = data.arubacloud_dbaas.basic.vpc_uri_ref
}
output "dbaas_subnet_uri_ref" {
  value = data.arubacloud_dbaas.basic.subnet_uri_ref
}
output "dbaas_security_group_uri_ref" {
  value = data.arubacloud_dbaas.basic.security_group_uri_ref
}
output "dbaas_elastic_ip_uri_ref" {
  value = data.arubacloud_dbaas.basic.elastic_ip_uri_ref
}
