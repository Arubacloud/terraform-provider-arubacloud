# ArubaCloud Terraform Provider - Schedule Example

This example demonstrates schedule job management with recurring and one-time scheduled tasks.

## What This Example Does

This Terraform configuration:
1. Creates a project as the foundation
2. Sets up a recurring schedule job
3. Sets up a one-shot schedule job

**Important Note**: Steps cannot be updated in-place. If you need to modify the steps in a schedule job, you must destroy and recreate the resource.

## Features

### Schedule Jobs

#### Recurring Schedule Job
- Executes repeatedly between start and end dates
- Configured with `schedule_job_type = "Recurring"`
- Requires both `schedule_at` (start time) and `execute_until` (end time)
- Can be enabled or disabled via the `enabled` property

#### One-Shot Schedule Job
- Executes once at a specific date/time
- Configured with `schedule_job_type = "OneShot"`
- Requires `schedule_at` (execution time) and `steps` array
- Automatically disabled after execution

## Prerequisites

### 1. Credentials
Create `terraform.tfvars` with your ArubaCloud credentials:
```hcl
arubacloud_api_key    = "your-api-key"
arubacloud_api_secret = "your-api-secret"
```

### 2. Provider Binary
Build the provider if not already built:
```bash
cd /path/to/terraform-provider-arubacloud
go build -o terraform-provider-arubacloud
```

## Quick Start

### 1. Initialize Terraform
```bash
terraform init
```

### 2. Review the Execution Plan
```bash
terraform plan
```

### 3. Apply the Configuration
```bash
terraform apply
```

When prompted, type `yes` to confirm.

### 4. View Outputs
After successful deployment:
```bash
terraform output
```

## Configuration Details

### Location and Region
The example uses `ITBG-Bergamo` as the default location. Change this to match your target region:
```hcl
location = "ITBG-Bergamo"  # Change to your region
```

Common locations:
- `ITBG-Bergamo` - Italy, Bergamo
- `CZTX-Prague` - Czech Republic, Prague
- `PLWZ-Warsaw` - Poland, Warsaw

### Schedule Job Properties

#### Recurring Job Properties
```hcl
properties = {
  enabled            = true                      # Enable/disable the job
  schedule_job_type  = "Recurring"              # Job type (OneShot or Recurring)
  schedule_at        = "2026-02-01T10:00:00Z"   # Start date/time (ISO 8601)
  execute_until      = "2026-03-01T10:00:00Z"   # End date/time (ISO 8601)
  cron               = "0 10 * * *"             # CRON expression (optional)
  steps = [                                      # Array of steps to execute
    {
      name         = "Power Off Server"        # Optional descriptive name
      resource_uri = "/projects/{project_id}/providers/Aruba.Compute/cloudServers/{server_id}"
      action_uri   = "poweroff"                # Action to execute
      http_verb    = "POST"                    # HTTP method (GET, POST, PUT, DELETE)
      body         = null                      # Optional request body (JSON string)
    }
  ]
}
```

#### One-Shot Job Properties
```hcl
properties = {
  enabled            = true                      # Enable/disable the job
  schedule_job_type  = "OneShot"                # Job type (OneShot or Recurring)
  schedule_at        = "2026-02-15T14:30:00Z"   # Execution date/time (ISO 8601)
  steps = [                                      # Array of steps to execute
    {
      name         = "Restart Server"          # Optional descriptive name
      resource_uri = "/projects/{project_id}/providers/Aruba.Compute/cloudServers/{server_id}"
      action_uri   = "restart"                 # Action to execute
      http_verb    = "POST"                    # HTTP method (GET, POST, PUT, DELETE)
      body         = null                      # Optional request body (JSON string)
    }
  ]
}
```

### Date/Time Format
All timestamps must be in ISO 8601 format with UTC timezone:
- Format: `YYYY-MM-DDTHH:MM:SSZ`
- Example: `2026-02-01T10:00:00Z`
- Timezone: Always use `Z` for UTC

### Job Steps Structure
Each step in the `steps` array defines an action to be executed:

- **name** (optional): Descriptive name for the step
- **resource_uri** (required): URI of the resource on which the action will be performed
  - Format: `/projects/{project_id}/providers/{provider}/{resource_type}/{resource_id}`
  - Example: `/projects/68398923fb2cb026400d4d31/providers/Aruba.Compute/cloudServers/69007ece4e7d691466d86223`
- **action_uri** (required): Action to execute on the resource (e.g., "start", "stop", "poweroff", "restart")
- **http_verb** (required): HTTP method to use (GET, POST, PUT, DELETE)
- **body** (optional): JSON string containing the request body for the action

Example step to power off a server:
```hcl
{
  name         = "Power Off Server"
  resource_uri = "/projects/68398923fb2cb026400d4d31/providers/Aruba.Compute/cloudServers/69007ece4e7d691466d86223"
  action_uri   = "poweroff"
  http_verb    = "POST"
  body         = null
}
```

### How to Get Resource URIs

To get valid resource URIs for your schedule jobs:

1. **Using Terraform resources**: If you create resources with Terraform, you can reference their URIs:
   ```hcl
   resource "arubacloud_cloudserver" "myserver" {
     # ... server configuration
   }
   
   resource "arubacloud_schedulejob" "shutdown" {
     # ...
     steps = [{
       resource_uri = arubacloud_cloudserver.myserver.uri
       action_uri   = "poweroff"
       http_verb    = "POST"
     }]
   }
   ```

2. **Using Data Sources**: Query existing resources:
   ```hcl
   data "arubacloud_cloudserver" "existing" {
     id = "your-server-id"
     project_id = "your-project-id"
   }
   
   resource "arubacloud_schedulejob" "shutdown" {
     # ...
     steps = [{
       resource_uri = data.arubacloud_cloudserver.existing.uri
       action_uri   = "poweroff"
       http_verb    = "POST"
     }]
   }
   ```

3. **Via ArubaCloud API/Console**: Get the full URI from your existing resources in the ArubaCloud console or API

## Resource Dependencies

The example follows this simple dependency chain:
```
Project
  └── Schedule Jobs
```

## Outputs

The configuration provides outputs for:
- Recurring schedule job ID and URI
- One-shot schedule job ID and URI

These outputs can be used as inputs for other Terraform modules or for reference.

## Testing

### Verify Schedule Jobs
```bash
terraform output recurring_job_id
terraform output recurring_job_uri
terraform output oneshot_job_id
terraform output oneshot_job_uri
```

## Updating Schedule Jobs

### Editable Properties
You can update the following properties without recreating the resource:
- `name` - Schedule job name
- `tags` - Resource tags
- `enabled` - Enable/disable the job

### Non-Editable Properties (Require Recreation)
The following properties **cannot be updated** and require destroying and recreating the resource:
- `steps` - Job steps configuration
- `schedule_job_type` - Job type (OneShot or Recurring)
- `schedule_at` - Start/execution time
- `execute_until` - End time (for recurring jobs)
- `cron` - CRON expression
- `location` - Resource location
- `project_id` - Project identifier

### How to Update Non-Editable Properties

If you need to modify steps or other non-editable properties:

1. **Option 1: Destroy and Recreate**
   ```bash
   terraform destroy -target=arubacloud_schedulejob.oneshot
   terraform apply
   ```

2. **Option 2: Use Terraform Taint (Deprecated but still works)**
   ```bash
   terraform taint arubacloud_schedulejob.oneshot
   terraform apply
   ```

3. **Option 3: Use Terraform Replace**
   ```bash
   terraform apply -replace=arubacloud_schedulejob.oneshot
   ```

## Cleanup

To destroy all resources:
```bash
terraform destroy
```

When prompted, type `yes` to confirm deletion.

## Common Issues

### Issue: Cannot modify steps
**Solution**: Steps and other job properties (schedule_at, cron, etc.) cannot be updated in-place. You must destroy and recreate the resource:
```bash
terraform destroy -target=arubacloud_schedulejob.your_job
terraform apply
```
Or use the replace command:
```bash
terraform apply -replace=arubacloud_schedulejob.your_job
```

### Issue: Schedule job creation fails with date validation error
**Solution**: Ensure all timestamps are in ISO 8601 format with UTC timezone (ending with `Z`). For recurring jobs, ensure `execute_until` is after `schedule_at`.

### Issue: Schedule job doesn't execute
**Solution**: Verify that:
1. The `enabled` property is set to `true`
2. The `schedule_at` time is in the future
3. For recurring jobs, the current time is between `schedule_at` and `execute_until`

## Additional Resources

- [ArubaCloud API Documentation](https://api.arubacloud.com/docs/)
- [Schedule Job Documentation](https://api.arubacloud.com/docs/resources/#schedule-jobs)
- [Terraform Provider Documentation](../../docs/)

## Notes

- Schedule jobs require proper project setup and permissions
- Times are always in UTC timezone
- Job types are: `OneShot` (single execution) or `Recurring` (multiple executions)
- Recurring jobs support CRON expressions for scheduling patterns
- One-shot jobs are automatically disabled after execution
- The `steps` array defines the actions to be executed by the job
- **Important**: Steps and scheduling properties cannot be updated in-place - changes require resource recreation
- You can update the `enabled` property to pause/resume jobs without recreating
- Only `name`, `tags`, and `enabled` properties support in-place updates
- Ensure schedule times are set appropriately for your use case
