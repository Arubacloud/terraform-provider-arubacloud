data "arubacloud_blockstorage" "example" {
  id = "your-blockstorage-id"
}

output "blockstorage_name" {
  value = data.arubacloud_blockstorage.example.name
}
output "blockstorage_project_id" {
  value = data.arubacloud_blockstorage.example.project_id
}
output "blockstorage_size_gb" {
  value = data.arubacloud_blockstorage.example.properties.size_gb
}
output "blockstorage_billing_period" {
  value = data.arubacloud_blockstorage.example.properties.billing_period
}
output "blockstorage_zone" {
  value = data.arubacloud_blockstorage.example.properties.zone
}
output "blockstorage_type" {
  value = data.arubacloud_blockstorage.example.properties.type
}
output "blockstorage_snapshot_id" {
  value = data.arubacloud_blockstorage.example.properties.snapshot_id
}
output "blockstorage_bootable" {
  value = data.arubacloud_blockstorage.example.properties.bootable
}
output "blockstorage_image" {
  value = data.arubacloud_blockstorage.example.properties.image
}
