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
output "cloudserver_vpc_id" {
  value = data.arubacloud_cloudserver.example.vpc_id
}
output "cloudserver_flavor_name" {
  value = data.arubacloud_cloudserver.example.flavor_name
}
output "cloudserver_elastic_ip_id" {
  value = data.arubacloud_cloudserver.example.elastic_ip_id
}
output "cloudserver_boot_volume" {
  value = data.arubacloud_cloudserver.example.boot_volume
}
output "cloudserver_key_pair_id" {
  value = data.arubacloud_cloudserver.example.key_pair_id
}
output "cloudserver_subnets" {
  value = data.arubacloud_cloudserver.example.subnets
}
output "cloudserver_securitygroups" {
  value = data.arubacloud_cloudserver.example.securitygroups
}
