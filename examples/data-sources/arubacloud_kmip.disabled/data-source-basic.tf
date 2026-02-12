data "arubacloud_kmip" "basic" {
  id = "your-kmip-id"
}

output "kmip_name" {
  value = data.arubacloud_kmip.basic.name
}
output "kmip_description" {
  value = data.arubacloud_kmip.basic.description
}
output "kmip_endpoint" {
  value = data.arubacloud_kmip.basic.endpoint
}
