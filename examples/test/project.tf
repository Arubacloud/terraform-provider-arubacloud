resource "arubacloud_project" "example" {
  name        = "example-project"
  description = "Example ArubaCloud project"
  tags        = ["dev", "test", "terraform"]
}