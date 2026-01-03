resource "arubacloud_securityrule" "example" {
  name              = "example-security-rule"
  location          = "ITBG-Bergamo"
  project_id        = "example-project-id"
  vpc_id            = "example-vpc-id"
  security_group_id = "example-security-group-id"
  tags              = ["security", "example"]
  properties = {
    direction = "Ingress"
    protocol  = "TCP"
    port      = "80"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}
