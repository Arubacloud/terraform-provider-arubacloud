output "kms_id" {
  value       = arubacloud_kms.test.id
  description = "KMS instance ID"
}

output "kms_uri" {
  value       = arubacloud_kms.test.uri
  description = "KMS instance URI"
}
