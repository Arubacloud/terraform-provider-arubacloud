data "arubacloud_containerregistry" "example" {
  id = "your-containerregistry-id"
}

output "containerregistry_name" {
  value = data.arubacloud_containerregistry.example.name
}
output "containerregistry_location" {
  value = data.arubacloud_containerregistry.example.location
}
output "containerregistry_tags" {
  value = data.arubacloud_containerregistry.example.tags
}
output "containerregistry_project_id" {
  value = data.arubacloud_containerregistry.example.project_id
}
output "containerregistry_billing_period" {
  value = data.arubacloud_containerregistry.example.billing_period
}
output "containerregistry_public_ip_uri_ref" {
  value = data.arubacloud_containerregistry.example.public_ip_uri_ref
}
output "containerregistry_vpc_uri_ref" {
  value = data.arubacloud_containerregistry.example.vpc_uri_ref
}
output "containerregistry_subnet_uri_ref" {
  value = data.arubacloud_containerregistry.example.subnet_uri_ref
}
output "containerregistry_security_group_uri_ref" {
  value = data.arubacloud_containerregistry.example.security_group_uri_ref
}
output "containerregistry_block_storage_uri_ref" {
  value = data.arubacloud_containerregistry.example.block_storage_uri_ref
}
output "containerregistry_concurrent_users_flavor" {
  value = data.arubacloud_containerregistry.example.concurrent_users_flavor
}
