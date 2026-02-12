resource "arubacloud_schedulejob" "example" {
  name       = "example-schedule-job"
  project_id = "example-project"
  location   = "example-location"
  tags       = ["tag1", "tag2"]
  properties = {
    enabled            = true
    schedule_job_type  = "Recurring"
    schedule_at        = "2025-11-26T10:00:00Z"
    execute_until      = "2025-12-01T10:00:00Z"
  }
}
