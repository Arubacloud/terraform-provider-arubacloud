data "arubacloud_databasebackup" "example" {
  id         = "your-backup-id"
  project_id = "your-project-id"
}

output "databasebackup_name" {
  value = data.arubacloud_databasebackup.example.name
}
output "databasebackup_location" {
  value = data.arubacloud_databasebackup.example.location
}
output "databasebackup_billing_period" {
  value = data.arubacloud_databasebackup.example.billing_period
}
output "databasebackup_tags" {
  value = data.arubacloud_databasebackup.example.tags
}
