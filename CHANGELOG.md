## 0.3.0 (June 29, 2026)

ENHANCEMENTS:

* **All resources with enum string fields** (`billing_period`, `type`, `protocol`, `kind`, etc.): Added `stringvalidator.OneOf()` validators to all enum attributes. Terraform now rejects invalid values at plan time with a clear diagnostic instead of forwarding them to the API and surfacing an opaque rejection error ([#170](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/170), [#171](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/171)).
* **18 data sources** (`arubacloud_backup`, `arubacloud_blockstorage`, `arubacloud_database`, `arubacloud_databasebackup`, `arubacloud_databasegrant`, `arubacloud_dbaasuser`, `arubacloud_elasticip`, `arubacloud_keypair`, `arubacloud_kms`, `arubacloud_restore`, `arubacloud_schedulejob`, `arubacloud_securitygroup`, `arubacloud_snapshot`, `arubacloud_vpc`, `arubacloud_vpcpeering`, `arubacloud_vpcpeeringroute`, `arubacloud_vpnroute`, `arubacloud_vpntunnel`): Added the missing `uri` attribute. The `uri` value can now be used directly in `*_uri_ref` attributes of other resources without requiring a separate resource lookup ([#170](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/170), [#171](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/171)).

BUG FIXES:

* **All 24 resources**: Implemented composite `ImportState` using the `<project_id>/<resource_id>` format (e.g., `terraform import arubacloud_vpc.example proj-abc/vpc-xyz`). Previously `ImportState` was either absent or only restored the resource `id`, requiring manual state editing to populate `project_id` ([#160](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/160)).
* **Multiple resources** (`arubacloud_vpnroute`, `arubacloud_vpntunnel`, `arubacloud_schedulejob`, `arubacloud_snapshot`, `arubacloud_vpcpeering`, `arubacloud_vpcpeeringroute`, and others): Added `RequiresReplace` plan modifiers to immutable fields (e.g., `location`, `type`, `billing_period`) that previously had none, replacing silent API rejections on update with a clear plan-time replacement proposal ([#161](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/161)).
* **Multiple resources**: Added `UseStateForUnknown` to stable computed fields (`id`, `uri`) to prevent spurious `(known after apply)` plans on resources that have already been created ([#162](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/162)).
* **Multiple resources** (`arubacloud_elasticip`, `arubacloud_blockstorage`, `arubacloud_backup`, `arubacloud_restore`, and others): Fixed perpetual `billing_period` drift. The API historically returned legacy lowercase variants (`"hourly"`, `"monthly"`, `"yearly"`); the provider now normalises these to canonical form (`"Hour"`, `"Month"`, `"Year"`) when reading state ([#163](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/163)).
* `arubacloud_securityrule`: Added `RequiresReplace` to all `properties.*` attributes. The API does not support in-place property updates; the provider previously attempted the call, received a rejection, and surfaced a runtime error. `RequiresReplace` now surfaces this as a destroy-and-recreate plan before `apply` ([#164](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/164)).
* `arubacloud_backup`, `arubacloud_restore`: Fixed a stale-state bug where `Create` returned the initial API response instead of re-reading the resource after it reached `Active`. Computed attributes (`uri`, `size`, `status`) are now populated from the final settled state ([#166](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/166), [#167](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/167)).
* **Resource wait logic**: Fixed a URI last-segment guard that caused a panic when the API returned a URI with no path segments; the provider now returns an empty string safely ([#168](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/168)).
* **Resource wait logic**: Fixed `remainingTimeout` going negative when the deadline was nearly exhausted — the value is now clamped to a minimum of 1 second before being passed to the SDK retry loop ([#169](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/169)).
* **Resource wait logic** (`WaitForResourceDeleted`): Added initial-check tolerance so the first poll after a delete call does not fail immediately when the resource still appears active — consistent with the behaviour of `WaitForResourceActive` ([#172](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/172)).
* `arubacloud_snapshot` data source: Fixed a nil-pointer panic when reading a snapshot whose associated `volume` field is empty ([#168](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/168)).

## 0.2.0 (June 26, 2026)

BREAKING CHANGES:

* Provider authentication renamed: `api_key` → `client_id`, `api_secret` → `client_secret` ([#134](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/134)).
* Environment variables renamed: `ARUBACLOUD_API_KEY` → `ARUBACLOUD_CLIENT_ID`, `ARUBACLOUD_API_SECRET` → `ARUBACLOUD_CLIENT_SECRET`.
* sdk-go upgraded from v0.1.24 to v1.0.4 — internal builder API migration; no resource schema changes.

FEATURES:

* All 25 resources and 25 data sources migrated to the sdk-go v1.0.4 fluent builder API, eliminating all `pkg/types` imports from production code ([#136](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/136), [#137](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/137), [#138](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/138), [#139](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/139), [#140](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/140), [#141](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/141), [#142](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/142), [#143](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/143)).

BUG FIXES:

* **All resources**: Create timeouts no longer taint the resource. Previously, when a resource did not become active within the timeout the provider returned a hard error, causing Terraform to mark the resource as tainted and propose destroy-and-recreate on the next plan. The provider now emits a warning, preserves the resource ID in state, and allows `Read()` to reconcile the actual backend state on the next `terraform apply` — avoiding unnecessary replacement of resources that are still provisioning.
* **All resources**: The provider-level `resource_timeout` default has been raised from 10 minutes to 30 minutes to accommodate long-running resources such as `arubacloud_kaas` and `arubacloud_containerregistry`. The value can still be overridden in the provider block (e.g. `resource_timeout = "45m"`).
* **All resources**: Fixed a secondary timeout bug where the SDK's internal retry counter (fixed at 60 attempts × 10 s = 10 min) would exhaust before the configured timeout, returning an opaque error that was incorrectly treated as a hard failure. Retry count is now derived from the configured timeout, ensuring the context deadline always fires first and produces a recoverable warning.
* `arubacloud_cloudserver`: Fixed a panic during `Create` and `Read` when the `network`, `settings`, or `storage` nested objects are null or unknown in state. The provider now initialises these with properly-typed null values before attempting to read them, preventing a "MISSING TYPE" framework error for list attributes (`subnet_uri_refs`, `securitygroup_uri_refs`).
* `arubacloud_securityrule`: Fixed perpetual drift on `properties.protocol` and `properties.target.kind`. The API normalises casing (e.g. `"tcp"` → `"TCP"`, `"Ip"` → `"IP"`), which caused Terraform to detect a diff on every plan. The provider now preserves the user's original casing when the API value is semantically identical, eliminating the spurious diff.
* `arubacloud_dbaas`: Fixed perpetual drift on `engine_id`. The API normalises engine identifiers (e.g. `"mysql"` → `"mysql-8.0"`), causing a diff on every plan. The provider now always preserves the state value for this immutable field.
* `arubacloud_dbaas`, `arubacloud_database`, `arubacloud_dbaasuser`, `arubacloud_databasegrant`: Added `UseStateForUnknown()` plan modifier to `id` and `uri` attributes, preventing these computed fields from appearing as `(known after apply)` on plans after the resource has already been created.

NOTES:

* The `uri` attribute on all resources may show a value change during the first `terraform refresh` after upgrading. This is expected (URI path segment casing may differ between SDK versions) and does **not** trigger resource replacement — all `uri` attributes are `Computed: true` and not `ForceNew`.
* See [upgrade guide](docs/guides/upgrade-to-v0.2.0.md) for step-by-step migration instructions.

## 0.1.7 (June 22, 2026)

BUG FIXES:

* **All stateful resources**: Resources that reach a terminal `Failed` state no longer block `terraform destroy` ([#116](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/116)). Previously, `Read()` returned a hard error for `Failed`-state resources, which prevented every subsequent Terraform operation — including `destroy` — from running and left CI pipelines stuck with no automated recovery path. `Read()` now emits a warning instead, populates state normally (the API returns valid metadata for failed resources), and allows Terraform to build a full destroy plan so that `Delete()` is invoked in the correct reverse-dependency order.

## 0.1.6 (May 27, 2026)

BUG FIXES:

* `arubacloud_elasticip`: Fix `billing_period` value sent to the API after backend breaking change — the ElasticIP API now uses `Hour`/`Month`/`Year` (same canonical form as all other resources) instead of the legacy `hourly`/`monthly`/`yearly` lowercase values. The create request was incorrectly sending the old lowercase form, causing API rejections.

## 0.0.1 (January 26, 2026)

FEATURES:

**Resources:**
* `arubacloud_project` - Manage ArubaCloud projects
* `arubacloud_cloudserver` - Manage Cloud Servers (VMs)
* `arubacloud_keypair` - Manage SSH keypairs
* `arubacloud_elasticip` - Manage Elastic IPs
* `arubacloud_blockstorage` - Manage Block Storage volumes
* `arubacloud_snapshot` - Manage Block Storage snapshots
* `arubacloud_vpc` - Manage Virtual Private Clouds
* `arubacloud_subnet` - Manage VPC subnets
* `arubacloud_securitygroup` - Manage Security Groups
* `arubacloud_securityrule` - Manage Security Group rules
* `arubacloud_vpcpeering` - Manage VPC peering connections
* `arubacloud_vpcpeeringroute` - Manage VPC peering routes
* `arubacloud_vpntunnel` - Manage VPN tunnels
* `arubacloud_vpnroute` - Manage VPN routes
* `arubacloud_kaas` - Manage Kubernetes as a Service clusters
* `arubacloud_containerregistry` - Manage Container Registry
* `arubacloud_dbaas` - Manage DBaaS instances
* `arubacloud_database` - Manage databases within DBaaS
* `arubacloud_dbaasuser` - Manage DBaaS users
* `arubacloud_databasegrant` - Manage database grants
* `arubacloud_databasebackup` - Manage database backups
* `arubacloud_backup` - Manage volume backups
* `arubacloud_restore` - Manage volume restores
* `arubacloud_schedulejob` - Manage scheduled jobs
* `arubacloud_kms` - Manage Key Management Service

**Data Sources:**
* All resources have corresponding data sources for importing existing infrastructure

NOTES:
* Initial release of the ArubaCloud Terraform Provider
* Supports all major ArubaCloud services: Compute, Storage, Networking, Kubernetes, DBaaS, and Security
* Provider uses the official ArubaCloud Go SDK
* Resources include automatic waiting for active state after creation
* Comprehensive examples and documentation included
* Known limitation: Key and KMIP resources temporarily disabled pending SDK updates
