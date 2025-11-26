resource "arubacloud_blockstorage" "example" {
  name       = "example-blockstorage"
  location   = "example-location"
  size       = 100
  project_id = "example-project"
}
