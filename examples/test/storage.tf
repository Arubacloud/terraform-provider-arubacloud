resource "arubacloud_blockstorage" "example" {
  name       = "example-block-storage"
  project_id = arubacloud_project.example.id
  properties = {
    size_gb        = 100
    billing_period = "Hour"
    zone           = "ITBG-Bergamo"
    type           = "Standard"
    bootable       = true
    image          = "ubuntu-22.04"
    snapshot_id    = arubacloud_snapshot.example.id
  }
}

resource "arubacloud_snapshot" "example" {
  name          = "example-snapshot"
  project_id    = arubacloud_project.example.id
  location      = "ITBG-Bergamo"
  billing_period = "Hour"
  volume_id     = arubacloud_blockstorage.example.id
}
