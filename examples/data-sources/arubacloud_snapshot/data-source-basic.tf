data "arubacloud_snapshot" "basic" {
  id         = "snapshot-id"
  project_id = "your-project-id"
}

output "snapshot_name" {
  value = data.arubacloud_snapshot.basic.name
}
output "snapshot_location" {
  value = data.arubacloud_snapshot.basic.location
}
output "snapshot_volume_id" {
  value = data.arubacloud_snapshot.basic.volume_id
}
