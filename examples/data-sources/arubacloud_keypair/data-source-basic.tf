data "arubacloud_keypair" "basic" {
  id = "your-keypair-id"
}

output "keypair_name" {
  value = data.arubacloud_keypair.basic.name
}
output "keypair_location" {
  value = data.arubacloud_keypair.basic.location
}
output "keypair_project_id" {
  value = data.arubacloud_keypair.basic.project_id
}
output "keypair_value" {
  value = data.arubacloud_keypair.basic.value
}
output "keypair_tags" {
  value = data.arubacloud_keypair.basic.tags
}
