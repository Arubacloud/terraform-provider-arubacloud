data "arubacloud_blockstorage" "example" {
  id = "your-blockstorage-id"
}

output "blockstorage_name" {
  value = data.arubacloud_blockstorage.example.name
}
output "blockstorage_project_id" {
  value = data.arubacloud_blockstorage.example.project_id
}
output "blockstorage_location" {
  value = data.arubacloud_blockstorage.example.location
}
output "blockstorage_size_gb" {
  value = data.arubacloud_blockstorage.example.size_gb
}
output "blockstorage_billing_period" {
  value = data.arubacloud_blockstorage.example.billing_period
}
output "blockstorage_zone" {
  value = data.arubacloud_blockstorage.example.zone
}
output "blockstorage_type" {
  value = data.arubacloud_blockstorage.example.type
}
output "blockstorage_tags" {
  value = data.arubacloud_blockstorage.example.tags
}
output "blockstorage_snapshot_id" {
  value = data.arubacloud_blockstorage.example.snapshot_id
}
output "blockstorage_bootable" {
  value = data.arubacloud_blockstorage.example.bootable
}
output "blockstorage_image" {
  value = data.arubacloud_blockstorage.example.image
}
