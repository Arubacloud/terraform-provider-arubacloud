resource "arubacloud_backup" "basic" {
  name          = "example-backup"
  location      = "de-1"
  project_id    = "project-123"
  type          = "full"
  volume_id     = "volume-123"
  billing_period = "monthly"

  # optional
  retention_days = 30
  tags           = ["env:dev", "service:demo"]
}
