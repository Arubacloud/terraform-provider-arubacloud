data "arubacloud_security_rule" "example" {
  id = "your-securityrule-id"
}

output "securityrule_name" {
  value = data.arubacloud_security_rule.example.name
}
output "securityrule_location" {
  value = data.arubacloud_security_rule.example.location
}
output "securityrule_project_id" {
  value = data.arubacloud_security_rule.example.project_id
}
output "securityrule_vpc_id" {
  value = data.arubacloud_security_rule.example.vpc_id
}
output "securityrule_security_group_id" {
  value = data.arubacloud_security_rule.example.security_group_id
}
output "securityrule_properties" {
  value = data.arubacloud_security_rule.example.properties
}
