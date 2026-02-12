data "arubacloud_elastic_ip" "example" {
  name       = "example-elastic-ip"
  project_id = "example-project"
  location   = "eu-1"
}

output "elasticip_id" {
  value = data.arubacloud_elastic_ip.example.id
}
output "elasticip_name" {
  value = data.arubacloud_elastic_ip.example.name
}
output "elasticip_location" {
  value = data.arubacloud_elastic_ip.example.location
}
output "elasticip_project_id" {
  value = data.arubacloud_elastic_ip.example.project_id
}
output "elasticip_address" {
  value = data.arubacloud_elastic_ip.example.address
}
output "elasticip_billing_period" {
  value = data.arubacloud_elastic_ip.example.billing_period
}
output "elasticip_tags" {
  value = data.arubacloud_elastic_ip.example.tags
}
