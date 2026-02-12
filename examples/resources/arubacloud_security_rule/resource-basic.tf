resource "arubacloud_security_rule" "basic" {
  name = "basic-security-rule"
  security_group_id = "security-group-id"
}
