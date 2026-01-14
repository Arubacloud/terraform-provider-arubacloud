# Step 4: Create Storage Resources

# Data Disk - Additional storage volume for data
resource "arubacloud_blockstorage" "data_disk" {
  name          = "test-data-disk"
  project_id    = arubacloud_project.test.id
  location      = "ITBG-Bergamo"  # Change to your region
  size_gb       = 50
  billing_period = "Hour"
  zone          = "ITBG-1"  # Change to your zone
  type          = "Standard"
  tags          = ["storage", "data", "test"]
}

# Boot Disk - Block storage that can be used as a boot volume
resource "arubacloud_blockstorage" "boot_disk" {
  name          = "test-boot-disk"
  project_id    = arubacloud_project.test.id
  location      = "ITBG-Bergamo"  # Change to your region
  size_gb       = 100
  billing_period = "Hour"
  zone          = "ITBG-1"  # Change to your zone
  type          = "Performance"  # Performance type for boot disk
  bootable      = true
  image         = "LU22-001"  # Ubuntu 22.04 - see https://api.arubacloud.com/docs/metadata/#cloud-server-bootvolume for available images
  tags          = ["storage", "boot", "test"]
}

# Snapshot of the boot disk
#resource "arubacloud_snapshot" "data_disk_snapshot" {
#  name          = "test-data-disk-snapshot"
#  project_id    = arubacloud_project.test.id
#  location      = "ITBG-Bergamo"
#  billing_period = "Hour"
#  volume_uri    = arubacloud_blockstorage.data_disk.uri
#  tags          = ["snapshot", "data", "test", "updated"]
#}

# Backup of the boot disk
#resource "arubacloud_backup" "boot_disk_backup" {
#  name          = "test-boot-disk-backup"
#  location      = "ITBG-Bergamo"
#  project_id    = arubacloud_project.test.id
#  type          = "Full"
#  volume_id     = arubacloud_blockstorage.boot_disk.id
#  retention_days = 7
#  billing_period = "Hour"
#  tags          = ["backup", "boot", "test"]
#}

# Restore from the backup to a new volume
# Note: This creates a restore operation from the backup
#resource "arubacloud_restore" "boot_disk_restore" {
#  name       = "test-boot-disk-restore"
#  location   = "ITBG-Bergamo"
#  project_id = arubacloud_project.test.id
#  backup_id  = arubacloud_backup.boot_disk_backup.id
#  volume_id  = arubacloud_blockstorage.boot_disk.id
#  tags       = ["restore", "boot", "test"]
#}

