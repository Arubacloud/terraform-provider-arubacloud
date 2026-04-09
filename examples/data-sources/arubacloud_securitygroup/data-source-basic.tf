data "arubacloud_securitygroup" "example" {
  id         = "your-securitygroup-id"
  project_id = "your-project-id"
  vpc_id     = "your-vpc-id"
}

output "securitygroup_name" {
  value = data.arubacloud_securitygroup.example.name
}
output "securitygroup_location" {
  value = data.arubacloud_securitygroup.example.location
}
output "securitygroup_tags" {
  value = data.arubacloud_securitygroup.example.tags
}
