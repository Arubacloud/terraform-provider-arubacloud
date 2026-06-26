---
page_title: "Upgrading to v0.2.0 - ArubaCloud Provider"
subcategory: "Guides"
description: |-
  How to upgrade from ArubaCloud provider v0.1.x to v0.2.0, including authentication rename and expected state behaviour.
---

# Upgrading to v0.2.0

This guide covers the breaking changes introduced in v0.2.0 and the steps required to upgrade.

## Breaking Changes

### Authentication Attributes Renamed

The provider authentication attributes have been renamed from the old `api_key`/`api_secret` pattern to the OAuth2-aligned `client_id`/`client_secret` naming:

| v0.1.x | v0.2.0 |
|--------|--------|
| `api_key` | `client_id` |
| `api_secret` | `client_secret` |

#### HCL Example

**Before (v0.1.x)**:

```hcl
provider "arubacloud" {
  api_key    = "YOUR_API_KEY"
  api_secret = "YOUR_API_SECRET"
}
```

**After (v0.2.0)**:

```hcl
provider "arubacloud" {
  client_id     = "YOUR_CLIENT_ID"
  client_secret = "YOUR_CLIENT_SECRET"
}
```

### Environment Variables Renamed

| v0.1.x | v0.2.0 |
|--------|--------|
| `ARUBACLOUD_API_KEY` | `ARUBACLOUD_CLIENT_ID` |
| `ARUBACLOUD_API_SECRET` | `ARUBACLOUD_CLIENT_SECRET` |

**Before (v0.1.x)**:

```bash
export ARUBACLOUD_API_KEY="your-api-key"
export ARUBACLOUD_API_SECRET="your-api-secret"
```

**After (v0.2.0)**:

```bash
export ARUBACLOUD_CLIENT_ID="your-client-id"
export ARUBACLOUD_CLIENT_SECRET="your-client-secret"
```

## Upgrade Procedure

1. **Update the version constraint** in your Terraform configuration:

   ```hcl
   terraform {
     required_providers {
       arubacloud = {
         source  = "arubacloud/arubacloud"
         version = ">= 0.2.0"
       }
     }
   }
   ```

2. **Run `terraform init -upgrade`** to download the new provider version.

3. **Update the provider block** — rename `api_key` → `client_id` and `api_secret` → `client_secret` in every provider block in your configuration:

   ```hcl
   provider "arubacloud" {
     client_id     = var.client_id
     client_secret = var.client_secret
   }
   ```

4. **Update environment variables** — rename `ARUBACLOUD_API_KEY` → `ARUBACLOUD_CLIENT_ID` and `ARUBACLOUD_API_SECRET` → `ARUBACLOUD_CLIENT_SECRET` in any scripts, CI/CD secrets, or shell profiles that set these.

5. **Run `terraform refresh`** — this syncs state with the current API. You may see changes to the `uri` attribute on some resources; this is expected (see below).

6. **Run `terraform plan`** — the plan should be empty. If any `+/-` (replace) operations appear, this is unexpected; open an issue.

## URI Attribute Refresh Note

After upgrading and running `terraform refresh`, the `uri` attribute on resources may show an updated value. This is caused by minor URI path-segment casing differences between SDK versions (for example, `securityGroups` vs `securitygroups`). This is **not a problem** — the `uri` attribute is `Computed: true` and is not `ForceNew` in any resource schema, so no resource will be destroyed and re-created as a result.

A subsequent `terraform plan` will show an empty plan.

## CHANGELOG

See [CHANGELOG.md](https://github.com/Arubacloud/terraform-provider-arubacloud/blob/main/CHANGELOG.md#020-unreleased) for the full list of changes in this release.
