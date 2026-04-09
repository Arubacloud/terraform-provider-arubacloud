data "arubacloud_kms" "basic" {
  id         = "kms-id"
  project_id = "your-project-id"
}

output "kms_name" {
  value = data.arubacloud_kms.basic.name
}
output "kms_description" {
  value = data.arubacloud_kms.basic.description
}
output "kms_endpoint" {
  value = data.arubacloud_kms.basic.endpoint
}
