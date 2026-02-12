data "arubacloud_schedule_job" "example" {
  name       = "example-schedule-job"
  project_id = "example-project"
  cron       = "0 0 * * *"
}
