resource "arubacloud_snapshot" "example" {
  name       = "example-snapshot"
  project_id = "your-project-id"
  location   = "ITBG-Bergamo"
}
