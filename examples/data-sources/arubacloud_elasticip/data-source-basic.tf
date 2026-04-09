data "arubacloud_elasticip" "example" {
  id         = "your-elasticip-id"
  project_id = "your-project-id"
}

output "elasticip_name" {
  value = data.arubacloud_elasticip.example.name
}
output "elasticip_location" {
  value = data.arubacloud_elasticip.example.location
}
output "elasticip_address" {
  value = data.arubacloud_elasticip.example.address
}
output "elasticip_billing_period" {
  value = data.arubacloud_elasticip.example.billing_period
}
output "elasticip_tags" {
  value = data.arubacloud_elasticip.example.tags
}
