data "arubacloud_security_rule" "example" {
  name       = "example-security-rule"
  project_id = "example-project"
  direction  = "inbound"
  protocol   = "tcp"
  port       = 80
}
