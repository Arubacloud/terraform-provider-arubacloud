# Step 2: Create Schedule Job Resources

# NOTE: Before using these examples, you need to:
# 1. Replace the resource_uri with an actual resource URI from your project
#    Format: /projects/{project_id}/providers/{provider}/{resource_type}/{resource_id}
#    Example: /projects/68398923fb2cb026400d4d31/providers/Aruba.Compute/cloudServers/69007ece4e7d691466d86223
# 2. The action_uri should be a valid action for that resource (e.g., "start", "stop", "poweroff", "restart")
#
# IMPORTANT: Steps cannot be updated in-place. If you need to modify steps,
# you must destroy the resource and recreate it (terraform destroy && terraform apply)

# Schedule Job - Recurring job example
# IMPORTANT: Update resource_uri to point to an actual resource in your account before applying
resource "arubacloud_schedulejob" "recurring" {
  name       = "recurring-schedule-job"
  project_id = arubacloud_project.test.id
  location   = "ITBG-Bergamo"  # Change to your region
  tags       = ["schedule", "recurring", "test"]
  properties = {
    enabled            = true
    schedule_job_type  = "Recurring"
    schedule_at        = "2026-02-01T10:00:00Z"  # Start date/time
    execute_until      = "2026-03-01T10:00:00Z"  # End date/time
    cron               = "0 10 * * *"             # Daily at 10:00 AM
    steps = [
      {
        name         = "Power Off Server"
        # TODO: Replace with your actual resource URI
        resource_uri = "/projects/68398923fb2cb026400d4d31/providers/Aruba.Compute/cloudServers/69007ece4e7d691466d86223"
        action_uri   = "poweroff"
        http_verb    = "POST"
        body         = null
      }
    ]
  }
}

# Schedule Job - One-shot job example
# This example uses the actual resource from your selection
resource "arubacloud_schedulejob" "oneshot" {
  name       = "oneshot-schedule-job"
  project_id = arubacloud_project.test.id
  location   = "ITBG-Bergamo"  # Change to your region
  tags       = ["schedule", "oneshot", "test"]
  properties = {
    enabled            = true
    schedule_job_type  = "OneShot"
    schedule_at        = "2026-02-15T14:30:00Z"  # Execution date/time
    steps = [
      {
        name         = "Power Off Server"
        # TODO: Replace with your actual resource URI
        resource_uri = "/projects/68398923fb2cb026400d4d31/providers/Aruba.Compute/cloudServers/69007ece4e7d691466d86223"
        action_uri   = "poweroff"
        http_verb    = "POST"
        body         = null
      }
    ]
  }
}
