# Password must be 8-20 characters, using at least one number, one uppercase letter, 
# one lowercase letter, and one special character. Spaces are not allowed.
# The password must be base64 encoded using the base64encode() function.
resource "arubacloud_dbaasuser" "example" {
  dbaas_id   = "example-dbaas-id"
  project_id = "example-project-id"
  username   = "example-user"
  password   = base64encode("Example123!")  # In production, use a secure password or variable
}
