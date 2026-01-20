data "arubacloud_cloudserver" "example" {
  id = "your-cloudserver-id"
}

output "cloudserver_name" {
  value = data.arubacloud_cloudserver.example.name
}
output "cloudserver_location" {
  value = data.arubacloud_cloudserver.example.location
}
output "cloudserver_project_id" {
  value = data.arubacloud_cloudserver.example.project_id
}
output "cloudserver_zone" {
  value = data.arubacloud_cloudserver.example.zone
}
output "cloudserver_tags" {
  value = data.arubacloud_cloudserver.example.tags
}
output "cloudserver_vpc_uri_ref" {
  value = data.arubacloud_cloudserver.example.vpc_uri_ref
}
output "cloudserver_elastic_ip_uri_ref" {
  value = data.arubacloud_cloudserver.example.elastic_ip_uri_ref
}
output "cloudserver_subnet_uri_refs" {
  value = data.arubacloud_cloudserver.example.subnet_uri_refs
}
output "cloudserver_securitygroup_uri_refs" {
  value = data.arubacloud_cloudserver.example.securitygroup_uri_refs
}
output "cloudserver_flavor_name" {
  value = data.arubacloud_cloudserver.example.flavor_name
}
output "cloudserver_key_pair_uri_ref" {
  value = data.arubacloud_cloudserver.example.key_pair_uri_ref
}
output "cloudserver_user_data" {
  value = data.arubacloud_cloudserver.example.user_data
}
output "cloudserver_boot_volume_uri_ref" {
  value = data.arubacloud_cloudserver.example.boot_volume_uri_ref
}
