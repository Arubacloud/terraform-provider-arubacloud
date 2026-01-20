data "arubacloud_security_group" "example" {
  name       = "example-security-group"
  project_id = "example-project"
}

output "securitygroup_id" {
  value = data.arubacloud_security_group.example.id
}
output "securitygroup_name" {
  value = data.arubacloud_security_group.example.name
}
output "securitygroup_location" {
  value = data.arubacloud_security_group.example.location
}
output "securitygroup_project_id" {
  value = data.arubacloud_security_group.example.project_id
}
output "securitygroup_vpc_id" {
  value = data.arubacloud_security_group.example.vpc_id
}
output "securitygroup_tags" {
  value = data.arubacloud_security_group.example.tags
}
