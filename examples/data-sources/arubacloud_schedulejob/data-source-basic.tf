data "arubacloud_schedulejob" "example" {
  id         = "your-schedulejob-id"
  project_id = "your-project-id"
}

output "schedulejob_name" {
  value = data.arubacloud_schedulejob.example.name
}
output "schedulejob_description" {
  value = data.arubacloud_schedulejob.example.description
}
output "schedulejob_cron" {
  value = data.arubacloud_schedulejob.example.cron
}
