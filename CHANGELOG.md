## 0.4.0 (July 2, 2026)

BREAKING CHANGES:

* `arubacloud_vpcpeering`: The `location` attribute has been removed. It was not supported by the API and caused import and refresh failures. Remove it from any existing configurations ([#219](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/219), [#220](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/220)).
* `arubacloud_containerregistry`: The `admin_user` block is now `Required`. Configurations that omitted it will fail plan-time validation. Add the block with the desired admin username to any existing configurations ([#194](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/194)).
* `arubacloud_securityrule`: The `ImportState` ID format now requires `location` as the fourth segment: `<project_id>/<vpc_id>/<security_group_id>/<location>/<id>`. Re-import any security rules that were previously imported with the old three-segment format ([#262](https://github.com/Arubacloud/terraform-provider-arubacloud/pull/262)).

FEATURES:

* **All resources**: Added an optional `timeout` attribute per resource that overrides the provider-level `resource_timeout` for that specific resource. Example: `timeout = "60m"` on a slow `arubacloud_kaas` while keeping the provider default for all other resources ([#188](https://github.com/Arubacloud/terraform-provider-arubacloud/pull/188)).

BUG FIXES:

* `arubacloud_backup`: Fixed `retention_days` drift on `Update` — the current plan value was not being forwarded to the API, causing the field to silently revert to the server default on every apply ([#240](https://github.com/Arubacloud/terraform-provider-arubacloud/pull/240)).
* `arubacloud_backup`: Marked `retention_days` as `RequiresReplace`. The API does not support changing this field in-place; Terraform now proposes a destroy-and-recreate instead of attempting an unsupported update ([#269](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/269)).
* `arubacloud_backup`: Added a propagation poll before `Create` when the source volume was provisioned in the same plan. The backup API returns 404 briefly for newly created volumes, causing spurious create failures ([#206](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/206)).
* `arubacloud_blockstorage`: Marked `type` as `RequiresReplace` — the API returns 400 when attempting to change storage type in-place ([#193](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/193)).
* `arubacloud_blockstorage`: `Delete` now removes all associated snapshots before deleting the volume. The API rejects volume deletion when snapshots still exist, leaving destroy stuck without this ordering fix ([#258](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/258)).
* `arubacloud_cloudserver`: Fixed `Update` to always include the subnet list in the request body — the API returns 400 when the field is null ([#247](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/247)).
* `arubacloud_cloudserver`: Fixed `Update` to always include the security group list in the request body, matching the subnet fix above ([#241](https://github.com/Arubacloud/terraform-provider-arubacloud/pull/241)).
* `arubacloud_cloudserver`: Added a `WaitForResourceActive` call after `Update`. CloudServer updates (e.g. flavor resize) trigger a reboot, leaving the server in a transitional state; the provider now waits for the resource to settle before writing state ([#175](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/175)).
* `arubacloud_cloudserver`: `Read` now detects subnet drift — changes to the subnet list made outside of Terraform are reflected on the next `terraform refresh` ([#188](https://github.com/Arubacloud/terraform-provider-arubacloud/pull/188)).
* `arubacloud_containerregistry`: `Delete` no longer hangs indefinitely when the registry is in `Failed` state. The provider now treats `Failed` as a terminal condition and proceeds with deletion ([#228](https://github.com/Arubacloud/terraform-provider-arubacloud/pull/228)).
* `arubacloud_containerregistry`: Fixed the resource reference construction to prefer the ID-based path, avoiding SDK path mismatches that caused `Read` to fail after `ImportState` ([#239](https://github.com/Arubacloud/terraform-provider-arubacloud/pull/239)).
* `arubacloud_database`: Marked `name` as `RequiresReplace` — the API returns 405 on rename attempts; Terraform now proposes destroy-and-recreate when this field changes ([#264](https://github.com/Arubacloud/terraform-provider-arubacloud/pull/264)).
* `arubacloud_databasebackup`: Extended the database readiness poll before `Create` to 300 seconds and added a retry on propagation lag. Newly created databases can take several minutes before the database backup API recognises them ([#197](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/197), [#210](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/210)).
* `arubacloud_dbaas`: `billing_period` is no longer shown as `(known after apply)` after `Create` — it is now populated from the API response immediately after provisioning ([#238](https://github.com/Arubacloud/terraform-provider-arubacloud/pull/238)).
* `arubacloud_dbaas`: Fixed perpetual `billing_period` drift when the attribute is omitted from config. A `useStateIfConfigNull` plan modifier now preserves the API-assigned value rather than proposing a null diff on every plan ([#195](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/195), [#208](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/208)).
* `arubacloud_dbaasuser`: Added a propagation poll before `Create` to handle the window between a DBaaS instance becoming active and its user management API becoming available ([#209](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/209)).
* `arubacloud_dbaasuser`, `arubacloud_databasegrant`: Marked all attributes as `RequiresReplace` — neither resource has an update endpoint. The provider previously attempted updates, received API errors, and left resources in a broken state ([#252](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/252)).
* `arubacloud_elasticip`: Fixed URI path-segment casing — the provider stored `elasticips` but the API normalises to `elasticIPs`, causing `Read` to fail with 404 on resources created in a previous apply ([#275](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/275)).
* `arubacloud_kaas`: `Read` now preserves `management_ip` and `pod_cidr` from prior state when the API response omits them. These fields are not always returned, causing spurious diffs on subsequent plans ([#276](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/276)).
* `arubacloud_keypair`: Added `RequiresReplace` to `name`, `value`, `location`, and `project_id`. The API has no update endpoint for keypairs; changes to these fields now correctly propose a destroy-and-recreate instead of failing silently ([#278](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/278)).
* `arubacloud_schedulejob`: `Delete` no longer hangs indefinitely when the job is in `Failed` state — the provider now surfaces a warning and proceeds ([#237](https://github.com/Arubacloud/terraform-provider-arubacloud/pull/237), [#245](https://github.com/Arubacloud/terraform-provider-arubacloud/pull/245)).
* `arubacloud_schedulejob`: Fixed the delete poll to treat `Deleted` and `Deleting` as successfully removed, preventing false "resource still exists" errors at the end of destroy ([#280](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/280)).
* `arubacloud_securityrule`: Removed the case-normalising plan modifiers from `properties.protocol` and `properties.target.kind` that were introduced in v0.3.0. They caused perpetual plan diffs for certain protocol values; the API-returned casing is now used as-is ([#271](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/271)).
* `arubacloud_vpc`: `Read` now removes the resource from state when the API returns a `Deleting` status, enabling correct out-of-band drift detection for VPCs being deleted externally ([#279](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/279)).
* `arubacloud_vpcpeering`: Normalised `peer_vpc` to the short resource ID format; removed the unsupported `location` attribute that caused import and refresh failures ([#219](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/219), [#220](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/220)).
* `arubacloud_vpntunnel`: Fixed three schema errors that blocked resource creation: added missing `subnet_cidr` attribute to the local subnet block; corrected `dh_group` to accept numeric group identifiers (`2`, `5`, `14`–`21`) instead of OpenVPN-style names; corrected `pfs` to accept `enable`/`disable` instead of boolean-style values ([#159](https://github.com/Arubacloud/terraform-provider-arubacloud/issues/159)).

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
