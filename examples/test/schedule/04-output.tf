output "recurring_job_id" {
  value       = arubacloud_schedulejob.recurring.id
  description = "Recurring schedule job ID"
}

output "recurring_job_uri" {
  value       = arubacloud_schedulejob.recurring.uri
  description = "Recurring schedule job URI"
}

output "oneshot_job_id" {
  value       = arubacloud_schedulejob.oneshot.id
  description = "One-shot schedule job ID"
}

output "oneshot_job_uri" {
  value       = arubacloud_schedulejob.oneshot.uri
  description = "One-shot schedule job URI"
}
