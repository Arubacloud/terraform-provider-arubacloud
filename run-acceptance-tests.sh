#!/usr/bin/env bash
# run-acceptance-tests.sh — run ArubaCloud provider acceptance tests
# Usage: ./run-acceptance-tests.sh --api-key KEY --api-secret SECRET --project-id ID [options]
set -euo pipefail

TIMEOUT="${TIMEOUT:-120m}"
PATTERN="${PATTERN:-^TestAcc}"

print_usage() {
  cat <<'EOF'
Usage: run-acceptance-tests.sh [options]

Required:
  --api-key      <key>     ArubaCloud API key     (or env ARUBACLOUD_API_KEY)
  --api-secret   <secret>  ArubaCloud API secret  (or env ARUBACLOUD_API_SECRET)
  --project-id   <id>      Project ID             (or env ARUBACLOUD_PROJECT_ID)

Optional fixture IDs (only needed for unconverted data source tests):
  --dbaas-id              ARUBACLOUD_DBAAS_ID
  --vpc-id                ARUBACLOUD_VPC_ID
  --cloudserver-id        ARUBACLOUD_CLOUDSERVER_ID
  --containerregistry-id  ARUBACLOUD_CONTAINERREGISTRY_ID
  --kaas-id               ARUBACLOUD_KAAS_ID
  --backup-id             ARUBACLOUD_BACKUP_ID
  --restore-id            ARUBACLOUD_RESTORE_ID
  --vpcpeering-id         ARUBACLOUD_VPCPEERING_ID
  --vpcpeeringroute-id    ARUBACLOUD_VPCPEERINGROUTE_ID
  --database-backup-id    ARUBACLOUD_DATABASE_BACKUP_ID

Test selection:
  -t, --test     <pattern>  Go -run regex (default: ^TestAcc)
      --timeout  <dur>      go test timeout (default: 120m)
  -h, --help                Show this help

Examples:
  # Smoke test a single data source
  ./run-acceptance-tests.sh --api-key KEY --api-secret SECRET --project-id PID \
    -t '^TestAccVpcDataSource$'

  # Run all converted data source tests
  ./run-acceptance-tests.sh --api-key KEY --api-secret SECRET --project-id PID \
    -t '^TestAcc(Vpc|Subnet|Keypair|Securitygroup|Securityrule|Blockstorage|Snapshot|Backup|Elasticip|Kms|Schedulejob|Vpntunnel|Vpnroute)DataSource$'

  # Include DBaaS-dependent tests
  ./run-acceptance-tests.sh --api-key KEY --api-secret SECRET --project-id PID \
    --dbaas-id DBAAS_ID \
    -t '^TestAcc(Database|Databasegrant|Dbaasuser)DataSource$'
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --api-key)              export ARUBACLOUD_API_KEY="$2";              shift 2 ;;
    --api-secret)           export ARUBACLOUD_API_SECRET="$2";           shift 2 ;;
    --project-id)           export ARUBACLOUD_PROJECT_ID="$2";           shift 2 ;;
    --dbaas-id)             export ARUBACLOUD_DBAAS_ID="$2";             shift 2 ;;
    --vpc-id)               export ARUBACLOUD_VPC_ID="$2";               shift 2 ;;
    --cloudserver-id)       export ARUBACLOUD_CLOUDSERVER_ID="$2";       shift 2 ;;
    --containerregistry-id) export ARUBACLOUD_CONTAINERREGISTRY_ID="$2"; shift 2 ;;
    --kaas-id)              export ARUBACLOUD_KAAS_ID="$2";              shift 2 ;;
    --backup-id)            export ARUBACLOUD_BACKUP_ID="$2";            shift 2 ;;
    --restore-id)           export ARUBACLOUD_RESTORE_ID="$2";           shift 2 ;;
    --vpcpeering-id)        export ARUBACLOUD_VPCPEERING_ID="$2";        shift 2 ;;
    --vpcpeeringroute-id)   export ARUBACLOUD_VPCPEERINGROUTE_ID="$2";   shift 2 ;;
    --database-backup-id)   export ARUBACLOUD_DATABASE_BACKUP_ID="$2";   shift 2 ;;
    -t|--test)              PATTERN="$2";                                 shift 2 ;;
    --timeout)              TIMEOUT="$2";                                 shift 2 ;;
    -h|--help)              print_usage; exit 0 ;;
    *) echo "error: unknown flag: $1" >&2; echo "Run with --help for usage." >&2; exit 2 ;;
  esac
done

: "${ARUBACLOUD_API_KEY:?Required. Pass --api-key or set ARUBACLOUD_API_KEY.}"
: "${ARUBACLOUD_API_SECRET:?Required. Pass --api-secret or set ARUBACLOUD_API_SECRET.}"
: "${ARUBACLOUD_PROJECT_ID:?Required. Pass --project-id or set ARUBACLOUD_PROJECT_ID.}"

export TF_ACC=1

echo "Running: go test -v -timeout=$TIMEOUT ./internal/provider/... -run \"$PATTERN\""
exec go test -v -timeout="$TIMEOUT" ./internal/provider/... -run "$PATTERN"
