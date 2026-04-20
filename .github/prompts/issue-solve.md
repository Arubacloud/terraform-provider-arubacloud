---
mode: agent
description: Diagnose and fix a reported issue in the ArubaCloud Terraform provider
---

Read [`ai/ARCHITECTURE.md`](../../ai/ARCHITECTURE.md) and [`ai/CONVENTIONS.md`](../../ai/CONVENTIONS.md) before starting.

---

## Issue to Solve

**Resource / data source affected:**
<!-- e.g. arubacloud_kaas, arubacloud_backup -->

**Symptom:**
<!-- Brief description: unexpected plan diff, error on apply, wrong state after import, etc. -->

**Steps to reproduce:**
1. 
2. 

**Terraform logs / error output:**
```
<!-- paste relevant terraform output here -->
```

**Expected behaviour:**
<!-- What should happen -->

**Actual behaviour:**
<!-- What actually happens -->

---

## Diagnosis Checklist

Work through these steps:

1. **Locate the affected file** — find `{resource}_resource.go` or `{resource}_data_source.go` in `internal/provider/`.
2. **Read() drift** — does `Read()` correctly map all API response fields back to state? Are any fields preserved from state when the API does not return them?
3. **Create() / Update() mismatch** — does the SDK request include all fields the user can set?
4. **Error handling** — is `CheckResponse[T]()` used? Are 404 responses handled with `resp.State.RemoveResource(ctx)`?
5. **Polling** — does `WaitForResourceActive()` receive a checker that correctly identifies the active state string?
6. **Plan modifiers** — are immutable fields annotated with `RequiresReplace()`? Are computed fields annotated with `UseStateForUnknown()`?

## Fix Guidelines

- Follow the resource lifecycle patterns in [`ai/ARCHITECTURE.md`](../../ai/ARCHITECTURE.md).
- Match the code style in [`ai/CONVENTIONS.md`](../../ai/CONVENTIONS.md).
- Run `make build` and `make test` after making changes.
- Run `make generate` if any schema attribute was added or changed.
      + id             = (known after apply)
      + location       = "ITBG-Bergamo"
      + name           = "container-registry-elastic-ip"
      + project_id     = (known after apply)
      + tags           = [
          + "public",
          + "container",
          + "registry",
          + "test",
        ]
      + uri            = (known after apply)
    }

  # arubacloud_kaas.test will be created
  + resource "arubacloud_kaas" "test" {
      + billing_period = "Hour"
      + id             = (known after apply)
      + kubeconfig     = (sensitive value)
      + location       = "ITBG-Bergamo"
      + management_ip  = (known after apply)
      + name           = "test-kaas"
      + network        = {
          + node_cidr           = {
              + address = "10.0.0.0/24"
              + name    = "kaas-node-cidr"
            }
          + pod_cidr            = "10.0.3.0/24"
          + security_group_name = "kaas-security-group"
          + subnet_uri_ref      = (known after apply)
          + vpc_uri_ref         = (known after apply)
        }
      + project_id     = (known after apply)
      + settings       = {
          + ha                 = true
          + kubernetes_version = "1.33.2"
          + node_pools         = [
              + {
                  + autoscaling = true
                  + instance    = "K2A4"
                  + max_count   = 5
                  + min_count   = 1
                  + name        = "pool-1"
                  + nodes       = 2
                  + zone        = "ITBG-1"
                },
            ]
        }
      + tags           = [
          + "test-k8s",
          + "kubernetes",
          + "test",
        ]
      + uri            = (known after apply)
    }

  # arubacloud_project.test will be created
  + resource "arubacloud_project" "test" {
      + description = "Project for testing Terraform container resources"
      + id          = (known after apply)
      + name        = "terraform-container-test-project"
      + tags        = [
          + "terraform",
          + "container",
          + "test",
        ]
    }

  # arubacloud_securitygroup.container_registry will be created
  + resource "arubacloud_securitygroup" "container_registry" {
      + id         = (known after apply)
      + location   = "ITBG-Bergamo"
      + name       = "container-registry-security-group"
      + project_id = (known after apply)
      + tags       = [
          + "security",
          + "container",
          + "registry",
          + "test",
        ]
      + uri        = (known after apply)
      + vpc_id     = (known after apply)
    }

  # arubacloud_securitygroup.kaas will be created
  + resource "arubacloud_securitygroup" "kaas" {
      + id         = (known after apply)
      + location   = "ITBG-Bergamo"
      + name       = "container-kaas-security-group"
      + project_id = (known after apply)
      + tags       = [
          + "security",
          + "kaas",
          + "kubernetes",
          + "test",
        ]
      + uri        = (known after apply)
      + vpc_id     = (known after apply)
    }

  # arubacloud_securityrule.container_registry_egress will be created
  + resource "arubacloud_securityrule" "container_registry_egress" {
      + id                = (known after apply)
      + location          = "ITBG-Bergamo"
      + name              = "container-registry-egress-rule"
      + project_id        = (known after apply)
      + properties        = {
          + direction = "Egress"
          + port      = "*"
          + protocol  = "ANY"
          + target    = {
              + kind  = "Ip"
              + value = "0.0.0.0/0"
            }
        }
      + security_group_id = (known after apply)
      + uri               = (known after apply)
      + vpc_id            = (known after apply)
    }

  # arubacloud_securityrule.container_registry_https will be created
  + resource "arubacloud_securityrule" "container_registry_https" {
      + id                = (known after apply)
      + location          = "ITBG-Bergamo"
      + name              = "container-registry-https-rule"
      + project_id        = (known after apply)
      + properties        = {
          + direction = "Ingress"
          + port      = "443"
          + protocol  = "TCP"
          + target    = {
              + kind  = "Ip"
              + value = "0.0.0.0/0"
            }
        }
      + security_group_id = (known after apply)
      + uri               = (known after apply)
      + vpc_id            = (known after apply)
    }

  # arubacloud_securityrule.kaas_egress will be created
  + resource "arubacloud_securityrule" "kaas_egress" {
      + id                = (known after apply)
      + location          = "ITBG-Bergamo"
      + name              = "kaas-egress-rule"
      + project_id        = (known after apply)
      + properties        = {
          + direction = "Egress"
          + port      = "*"
          + protocol  = "ANY"
          + target    = {
              + kind  = "Ip"
              + value = "0.0.0.0/0"
            }
        }
      + security_group_id = (known after apply)
      + uri               = (known after apply)
      + vpc_id            = (known after apply)
    }

  # arubacloud_subnet.test will be created
  + resource "arubacloud_subnet" "test" {
      + id         = (known after apply)
      + location   = "ITBG-Bergamo"
      + name       = "container-test-subnet"
      + project_id = (known after apply)
      + tags       = [
          + "network",
          + "container",
          + "test",
        ]
      + type       = "Basic"
      + uri        = (known after apply)
      + vpc_id     = (known after apply)
    }

  # arubacloud_vpc.test will be created
  + resource "arubacloud_vpc" "test" {
      + id         = (known after apply)
      + location   = "ITBG-Bergamo"
      + name       = "container-test-vpc"
      + project_id = (known after apply)
      + tags       = [
          + "network",
          + "container",
          + "test",
        ]
      + uri        = (known after apply)
    }

Plan: 11 to add, 0 to change, 0 to destroy.

Changes to Outputs:
  + container_registry_elastic_ip = (known after apply)
  + kaas_id                       = (known after apply)
  + kaas_kubeconfig               = (sensitive value)
  + kaas_management_ip            = (known after apply)
  + kaas_uri                      = (known after apply)

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

arubacloud_project.test: Creating...
arubacloud_project.test: Creation complete after 0s [id=69dcf86ea076662629261702]
arubacloud_elasticip.container_registry: Creating...
arubacloud_vpc.test: Creating...
arubacloud_blockstorage.container_registry: Creating...
arubacloud_elasticip.container_registry: Still creating... [10s elapsed]
arubacloud_blockstorage.container_registry: Still creating... [10s elapsed]
arubacloud_vpc.test: Still creating... [10s elapsed]
arubacloud_elasticip.container_registry: Still creating... [20s elapsed]
arubacloud_vpc.test: Still creating... [20s elapsed]
arubacloud_blockstorage.container_registry: Still creating... [20s elapsed]
arubacloud_elasticip.container_registry: Still creating... [30s elapsed]
arubacloud_vpc.test: Still creating... [30s elapsed]
arubacloud_blockstorage.container_registry: Still creating... [30s elapsed]
arubacloud_elasticip.container_registry: Still creating... [40s elapsed]
arubacloud_vpc.test: Still creating... [40s elapsed]
arubacloud_blockstorage.container_registry: Still creating... [40s elapsed]
arubacloud_elasticip.container_registry: Still creating... [50s elapsed]
arubacloud_blockstorage.container_registry: Still creating... [50s elapsed]
arubacloud_vpc.test: Still creating... [50s elapsed]
arubacloud_vpc.test: Creation complete after 56s [id=69dcf86efa14cbff96e2e645]
arubacloud_securitygroup.kaas: Creating...
arubacloud_securitygroup.container_registry: Creating...
arubacloud_subnet.test: Creating...
arubacloud_elasticip.container_registry: Creation complete after 56s [id=69dcf86ffa14cbff96e2e647]
arubacloud_blockstorage.container_registry: Still creating... [1m0s elapsed]
arubacloud_securitygroup.container_registry: Still creating... [10s elapsed]
arubacloud_securitygroup.kaas: Still creating... [10s elapsed]
arubacloud_subnet.test: Still creating... [10s elapsed]
arubacloud_blockstorage.container_registry: Creation complete after 1m7s [id=69dcf86f0016c4fa6548f193]
arubacloud_securitygroup.container_registry: Still creating... [20s elapsed]
arubacloud_securitygroup.kaas: Still creating... [20s elapsed]
arubacloud_subnet.test: Still creating... [20s elapsed]
arubacloud_securitygroup.kaas: Creation complete after 20s [id=69dcf8aafa14cbff96e2e64b]
arubacloud_securityrule.kaas_egress: Creating...
arubacloud_securitygroup.container_registry: Creation complete after 20s [id=69dcf8abfa14cbff96e2e64e]
arubacloud_securityrule.container_registry_egress: Creating...
arubacloud_securityrule.container_registry_https: Creating...
arubacloud_subnet.test: Still creating... [30s elapsed]
arubacloud_securityrule.kaas_egress: Still creating... [10s elapsed]
arubacloud_securityrule.container_registry_https: Still creating... [10s elapsed]
arubacloud_securityrule.container_registry_egress: Still creating... [10s elapsed]
arubacloud_subnet.test: Still creating... [40s elapsed]
arubacloud_securityrule.kaas_egress: Still creating... [20s elapsed]
arubacloud_securityrule.container_registry_egress: Still creating... [20s elapsed]
arubacloud_securityrule.container_registry_https: Still creating... [20s elapsed]
arubacloud_securityrule.container_registry_egress: Creation complete after 20s [id=69dcf8bffa14cbff96e2e655]
arubacloud_securityrule.kaas_egress: Creation complete after 20s [id=69dcf8bffa14cbff96e2e653]
arubacloud_securityrule.container_registry_https: Creation complete after 25s [id=69dcf8bffa14cbff96e2e657]
arubacloud_subnet.test: Still creating... [50s elapsed]
arubacloud_subnet.test: Creation complete after 55s [id=69dcf8abfa14cbff96e2e64d]
arubacloud_kaas.test: Creating...
arubacloud_kaas.test: Still creating... [10s elapsed]
arubacloud_kaas.test: Still creating... [20s elapsed]
arubacloud_kaas.test: Still creating... [30s elapsed]
arubacloud_kaas.test: Still creating... [40s elapsed]
arubacloud_kaas.test: Still creating... [50s elapsed]
arubacloud_kaas.test: Still creating... [1m0s elapsed]
arubacloud_kaas.test: Still creating... [1m10s elapsed]
arubacloud_kaas.test: Still creating... [1m20s elapsed]
arubacloud_kaas.test: Still creating... [1m30s elapsed]
arubacloud_kaas.test: Still creating... [1m40s elapsed]
arubacloud_kaas.test: Still creating... [1m50s elapsed]
arubacloud_kaas.test: Still creating... [2m0s elapsed]
arubacloud_kaas.test: Still creating... [2m10s elapsed]
arubacloud_kaas.test: Still creating... [2m20s elapsed]
arubacloud_kaas.test: Still creating... [2m30s elapsed]
arubacloud_kaas.test: Still creating... [2m40s elapsed]
arubacloud_kaas.test: Still creating... [2m50s elapsed]
arubacloud_kaas.test: Still creating... [3m0s elapsed]
╷
│ Warning: Resource Provisioning In Progress
│ 
│   with arubacloud_kaas.test,
│   on 05-container.tf line 34, in resource "arubacloud_kaas" "test":
│   34: resource "arubacloud_kaas" "test" {
│ 
│ KaaS "69dcf8e2534c48ab645211fc" was created but did not become active
│ within the timeout. Run terraform apply again to reconcile. (timeout
│ waiting for KaaS 69dcf8e2534c48ab645211fc to become active (timeout:
│ 3m0s))
╵
╷
│ Error: Provider returned invalid result object after apply
│ 
│ After the apply operation, the provider still indicated an unknown
│ value for arubacloud_kaas.test.management_ip. All values must be known
│ after apply, so this is always a bug in the provider and should be
│ reported in the provider's own repository. Terraform will still save
│ the other known object values in the state.


```

Key observations:

* KaaS resource is created but does not become active within timeout (3 minutes)
* Warning: "Resource Provisioning In Progress"
* Error: provider returned unknown value for `management_ip` after apply
* Terraform says this is a provider bug

---

## 📜 Second apply (logs)

```

amedeopalopoli@NBK-APALOPOLI:/mnt/c/Users/amedeo.palopoli/Desktop/Software/terraform-provider-arubacloud/examples/test/container$ terraform apply
╷
│ Warning: Provider development overrides are in effect
│ 
│ The following provider development overrides are set in the CLI
│ configuration:
│  - arubacloud/arubacloud in /mnt/c/Users/amedeo.palopoli/Desktop/Software/terraform-provider-arubacloud
│ 
│ The behavior may therefore not match any released version of the
│ provider and applying changes may cause the state to become
│ incompatible with published releases.
╵
arubacloud_project.test: Refreshing state... [id=69dcf86ea076662629261702]
arubacloud_blockstorage.container_registry: Refreshing state... [id=69dcf86f0016c4fa6548f193]
arubacloud_vpc.test: Refreshing state... [id=69dcf86efa14cbff96e2e645]
arubacloud_elasticip.container_registry: Refreshing state... [id=69dcf86ffa14cbff96e2e647]
arubacloud_subnet.test: Refreshing state... [id=69dcf8abfa14cbff96e2e64d]
arubacloud_securitygroup.container_registry: Refreshing state... [id=69dcf8abfa14cbff96e2e64e]
arubacloud_securitygroup.kaas: Refreshing state... [id=69dcf8aafa14cbff96e2e64b]
arubacloud_securityrule.kaas_egress: Refreshing state... [id=69dcf8bffa14cbff96e2e653]
arubacloud_kaas.test: Refreshing state... [id=69dcf8e2534c48ab645211fc]
arubacloud_securityrule.container_registry_egress: Refreshing state... [id=69dcf8bffa14cbff96e2e655]
arubacloud_securityrule.container_registry_https: Refreshing state... [id=69dcf8bffa14cbff96e2e657]

Terraform used the selected providers to generate the following
execution plan. Resource actions are indicated with the following
symbols:
  ~ update in-place
-/+ destroy and then create replacement

Terraform will perform the following actions:

  # arubacloud_kaas.test is tainted, so must be replaced
-/+ resource "arubacloud_kaas" "test" {
      ~ id             = "69dcf8e2534c48ab645211fc" -> (known after apply)
      + kubeconfig     = (sensitive value)
      + management_ip  = (known after apply)
        name           = "test-kaas"
        tags           = [
            "test-k8s",
            "kubernetes",
            "test",
        ]
      ~ uri            = "/projects/69dcf86ea076662629261702/providers/Aruba.Container/kaas/69dcf8e2534c48ab645211fc" -> (known after apply)
        # (5 unchanged attributes hidden)
    }

  # arubacloud_securityrule.container_registry_egress will be updated in-place
  ~ resource "arubacloud_securityrule" "container_registry_egress" {
        id                = "69dcf8bffa14cbff96e2e655"
        name              = "container-registry-egress-rule"
      - tags              = [] -> null
        # (6 unchanged attributes hidden)
    }

  # arubacloud_securityrule.container_registry_https will be updated in-place
  ~ resource "arubacloud_securityrule" "container_registry_https" {
        id                = "69dcf8bffa14cbff96e2e657"
        name              = "container-registry-https-rule"
      - tags              = [] -> null
        # (6 unchanged attributes hidden)
    }

  # arubacloud_securityrule.kaas_egress will be updated in-place
  ~ resource "arubacloud_securityrule" "kaas_egress" {
        id                = "69dcf8bffa14cbff96e2e653"
        name              = "kaas-egress-rule"
      - tags              = [] -> null
        # (6 unchanged attributes hidden)
    }

Plan: 1 to add, 3 to change, 1 to destroy.
```

Key observations:

* KaaS resource is marked as **tainted** → Terraform plans destroy + recreate
* Security rules show in-place updates (`tags: [] -> null`) even though config did not change

---

## ❓ Questions

1. Why is Terraform marking the KaaS resource as tainted?

   * Is it caused by the timeout?
   * Or because the provider returned incomplete state?

2. How should long-running resources like KaaS be handled correctly?

   * Should timeouts be increased?
   * Should the provider implement polling/wait logic differently?

3. How can I determine the **real state** of the KaaS resource?

   * Could it still be provisioning successfully?
   * Is it safe to avoid recreation?

4. Why is Terraform detecting changes in security rules (`tags: [] -> null`)?

   * Is this due to schema definition (`Optional`, `Computed`)?
   * How should null vs empty values be handled?

5. What are the correct fixes in the Terraform provider implementation?

   * Schema design
   * State handling after apply
   * Handling async provisioning resources

---

## 🎯 What I need

* Root cause analysis based on logs
* Explanation of Terraform behavior (tainting, state, diff)
* Suggested fixes in provider code (Go / Terraform Plugin SDK)
* Possible Terraform-side workarounds (timeouts, lifecycle, ignore_changes)
