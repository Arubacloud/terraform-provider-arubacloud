data "arubacloud_containerregistry" "example" {
  name       = "example-container-registry"
  project_id = "example-project"
  location   = "eu-1"
  type       = "Standard"
}
