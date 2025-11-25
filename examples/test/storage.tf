# Block Storage Resource Example
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

#Snapshot Resource Example
resource "arubacloud_snapshot" "example" {
  name          = "example-snapshot"
  project_id    = arubacloud_project.example.id
  location      = "ITBG-Bergamo"
  billing_period = "Hour"
  volume_id     = arubacloud_blockstorage.example.id
}

#Backup Resource Example
resource "arubacloud_backup" "example" {
  name           = "example-backup"
  location       = "ITBG-Bergamo"
  tags           = ["backup", "test"]
  project_id     = arubacloud_project.example.id
  type           = "Full"
  volume_id      = arubacloud_blockstorage.example.id
  retention_days = 30
  billing_period = "Hour"
}

#Restore Resource Example
resource "arubacloud_restore" "example" {
  name        = "example-restore"
  location    = "ITBG-Bergamo"
  tags        = ["restore", "test"]
  project_id  = arubacloud_project.example.id
  volume_id   = arubacloud_blockstorage.example.id
}
