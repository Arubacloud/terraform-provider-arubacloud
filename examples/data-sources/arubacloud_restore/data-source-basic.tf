data "arubacloud_restore" "basic" {
  id         = "restore-id"
  project_id = "your-project-id"
  backup_id  = "your-backup-id"
}

output "restore_name" {
  value = data.arubacloud_restore.basic.name
}
output "restore_location" {
  value = data.arubacloud_restore.basic.location
}
output "restore_volume_id" {
  value = data.arubacloud_restore.basic.volume_id
}
output "restore_tags" {
  value = data.arubacloud_restore.basic.tags
}
