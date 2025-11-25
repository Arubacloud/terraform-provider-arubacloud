resource "arubacloud_schedulejob" "example" {
  name        = "example-job"
  project_id  = "example-project-id"
  location    = "eu-central-1"
  tags        = ["nightly", "backup"]

  properties = {
    enabled            = true
    schedule_job_type  = "OneShot" # or "Recurring"
    schedule_at        = "2025-12-01T02:00:00Z" # Only for OneShot
    execute_until      = null # Only for Recurring
    cron               = null # Only for Recurring
    steps = [
      {
        name         = "Backup Step"
        resource_uri = "" ##TO BE IMPROVED AND DYNAMIC##
        action_uri   = "" ##TO BE IMPROVED AND DYNAMIC##
        http_verb    = "POST"
        body         = "{\"example\":\"example\"}"
      }
    ]
  }
}
