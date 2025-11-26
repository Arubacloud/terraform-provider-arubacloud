data "arubacloud_blockstorage" "example" {
  name       = "example-blockstorage"
  project_id = "example-project"
  properties {
    size_gb        = 100
    billing_period = "Hour"
    zone           = "eu-1"
    type           = "Standard"
    bootable       = false
  }
}
