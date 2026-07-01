#!/usr/bin/env bash
#
# run-acceptance-tests.sh — ArubaCloud Terraform Provider acceptance test runner
#
# Credentials and optional fixture IDs are read from environment variables.
# No secrets are accepted as command-line flags to avoid shell history exposure.
#
# Required env vars
#   ARUBACLOUD_CLIENT_ID       OAuth2 client ID
#   ARUBACLOUD_CLIENT_SECRET   OAuth2 client secret
#   ARUBACLOUD_PROJECT_ID      Target project
#
# Optional env vars (prompt if missing; tests skip gracefully when absent)
#   ARUBACLOUD_ZONE                 TestAccKaasResource, TestAccKaasDataSource
#   ARUBACLOUD_KAAS_NODE_INSTANCE   TestAccKaasResource, TestAccKaasDataSource
#   ARUBACLOUD_OS_IMAGE_ID          TestAccCloudserverResource, TestAccBlockStorageResource_Bootable,
#                                   TestAccCloudserverDataSource, TestAccSchedulejobDataSource
#   ARUBACLOUD_DBAAS_ID             TestAccDatabaseResource
#   ARUBACLOUD_VPNTUNNEL_ID         TestAccVpntunnelDataSource, TestAccVpnrouteDataSource
#   ARUBACLOUD_VPNROUTE_ID          TestAccVpnrouteDataSource
#
# Options
#   -r, --run PATTERN       go -run regex filter     (default: ^TestAcc)
#   -t, --timeout DURATION  test timeout             (default: 120m)
#   -l, --log-level LEVEL   TF_LOG: OFF|WARN|INFO|DEBUG|TRACE  (default: WARN)
#   -y, --yes               skip prompts for optional vars (auto-confirm skip)
#   -h, --help              show this help
#
# Artifacts (written to ./artifacts/)
#   tests-TIMESTAMP.log        full go test -v output
#   tf-provider-TIMESTAMP.log  Terraform provider log (set by TF_LOG / TF_LOG_PATH)
#   summary-TIMESTAMP.txt      concise pass/fail/skip summary (also shown on terminal)
#
# Examples
#   # Run all acceptance tests
#   ./run-acceptance-tests.sh
#
#   # Run a single resource test
#   ./run-acceptance-tests.sh --run '^TestAccBackupResource$'
#
#   # Run the DBaaS stack tests with debug logging
#   ./run-acceptance-tests.sh --run '^TestAccDatabase' --log-level DEBUG --timeout 60m
#
#   # Via make
#   make testacc
#   make testacc-run TEST=TestAccBackupResource

set -euo pipefail

# ── defaults ──────────────────────────────────────────────────────────────────
RUN_FILTER="^TestAcc"
TIMEOUT="120m"
LOG_LEVEL="WARN"
YES=false

# ── argument parsing ───────────────────────────────────────────────────────────
usage() {
    # Print only the leading comment block (stop at first blank line after shebang)
    awk 'NR==1{next} /^$/{exit} {sub(/^# ?/,""); print}' "$0"
    exit 0
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        -r|--run)        RUN_FILTER="$2"; shift 2 ;;
        -t|--timeout)    TIMEOUT="$2";    shift 2 ;;
        -l|--log-level)  LOG_LEVEL="$2";  shift 2 ;;
        -y|--yes)        YES=true;        shift   ;;
        -h|--help)       usage ;;
        *) echo "ERROR: unknown option: $1" >&2
           echo "Run './run-acceptance-tests.sh --help' for usage." >&2
           exit 1 ;;
    esac
done

# ── validate required env vars ─────────────────────────────────────────────────
MISSING=()
for var in ARUBACLOUD_CLIENT_ID ARUBACLOUD_CLIENT_SECRET ARUBACLOUD_PROJECT_ID; do
    [[ -z "${!var:-}" ]] && MISSING+=("$var")
done
if [[ ${#MISSING[@]} -gt 0 ]]; then
    echo "ERROR: missing required environment variables:" >&2
    printf '  %s\n' "${MISSING[@]}" >&2
    echo "" >&2
    echo "Export them before running:" >&2
    echo "  export ARUBACLOUD_CLIENT_ID=<id>" >&2
    echo "  export ARUBACLOUD_CLIENT_SECRET=<secret>" >&2
    echo "  export ARUBACLOUD_PROJECT_ID=<project-id>" >&2
    exit 1
fi

# ── optional env vars ─────────────────────────────────────────────────────────
# Ordered list of optional vars and the tests they gate.
OPT_VAR_NAMES=(
    ARUBACLOUD_ZONE
    ARUBACLOUD_KAAS_NODE_INSTANCE
    ARUBACLOUD_OS_IMAGE_ID
    ARUBACLOUD_DBAAS_ID
    ARUBACLOUD_VPNTUNNEL_ID
    ARUBACLOUD_VPNROUTE_ID
)
OPT_VAR_TESTS=(
    "TestAccKaasResource, TestAccKaasDataSource"
    "TestAccKaasResource, TestAccKaasDataSource"
    "TestAccCloudserverResource, TestAccBlockStorageResource_Bootable, TestAccCloudserverDataSource, TestAccSchedulejobDataSource"
    "TestAccDatabaseResource"
    "TestAccVpntunnelDataSource, TestAccVpnrouteDataSource"
    "TestAccVpnrouteDataSource"
)

WILL_SKIP=()
for i in "${!OPT_VAR_NAMES[@]}"; do
    var="${OPT_VAR_NAMES[$i]}"
    tests="${OPT_VAR_TESTS[$i]}"
    if [[ -z "${!var:-}" ]]; then
        if [[ "$YES" == "false" ]] && [[ -t 0 ]]; then
            echo ""
            echo "  Optional var not set: $var"
            echo "  Skips tests         : $tests"
            printf "  Enter value to set it, or press Enter to skip those tests: "
            read -r val </dev/tty
            if [[ -n "$val" ]]; then
                export "$var"="$val"
            else
                WILL_SKIP+=("$var ($tests)")
            fi
        else
            WILL_SKIP+=("$var ($tests)")
        fi
    fi
done

if [[ ${#WILL_SKIP[@]} -gt 0 ]]; then
    echo ""
    echo "  NOTE: the following optional vars are unset — related tests will be skipped:"
    for entry in "${WILL_SKIP[@]}"; do
        printf '    %s\n' "$entry"
    done
    echo ""
    if [[ "$YES" == "false" ]] && [[ -t 0 ]]; then
        printf "  Continue with skipped tests? [Y/n]: "
        read -r confirm </dev/tty
        if [[ "${confirm,,}" == "n" ]]; then
            echo "Aborted." >&2
            exit 1
        fi
    fi
fi

# ── artifact setup ─────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ARTIFACTS_DIR="${SCRIPT_DIR}/artifacts"
mkdir -p "$ARTIFACTS_DIR"
TS=$(date +"%Y%m%d-%H%M%S")
LOGFILE="${ARTIFACTS_DIR}/tests-${TS}.log"
TF_LOGFILE="${ARTIFACTS_DIR}/tf-provider-${TS}.log"
SUMMARY_FILE="${ARTIFACTS_DIR}/summary-${TS}.txt"

# ── header ─────────────────────────────────────────────────────────────────────
{
    echo "========================================================"
    echo "  ArubaCloud Provider — Acceptance Tests"
    printf "  Started  : %s\n" "$(date)"
    printf "  Filter   : %s\n" "$RUN_FILTER"
    printf "  Timeout  : %s\n" "$TIMEOUT"
    printf "  TF_LOG   : %s\n" "$LOG_LEVEL"
    echo "========================================================"
    echo ""
    echo "--- toolchain ---"
    go version
    if command -v terraform &>/dev/null; then
        terraform version 2>&1 | head -1 || true
    else
        echo "terraform: not in PATH (only needed for make generate)"
    fi
    echo ""
    echo "--- environment (secrets redacted) ---"
    env | sort \
        | grep -E '^(TF_|ARUBACLOUD_)' \
        | sed \
            -e 's/\(CLIENT_ID=\).*/\1***/' \
            -e 's/\(CLIENT_SECRET=\).*/\1***/' \
        || true
    echo ""
} | tee "$LOGFILE"

# ── run ────────────────────────────────────────────────────────────────────────
export TF_ACC=1
export TF_LOG="${LOG_LEVEL}"
export TF_LOG_PATH="${TF_LOGFILE}"

echo "--- running: go test -v -count=1 -timeout=${TIMEOUT} -run '${RUN_FILTER}' ---" \
    | tee -a "$LOGFILE"
echo "" | tee -a "$LOGFILE"

START_EPOCH=$(date +%s)

set +e
go test \
    -v \
    -count=1 \
    -timeout="${TIMEOUT}" \
    ./internal/provider/... \
    -run "${RUN_FILTER}" \
    2>&1 | tee -a "$LOGFILE"
EXIT_CODE=${PIPESTATUS[0]}
set -e

END_EPOCH=$(date +%s)
ELAPSED=$(( END_EPOCH - START_EPOCH ))
ELAPSED_FMT=$(printf '%dm%02ds' $(( ELAPSED / 60 )) $(( ELAPSED % 60 )))

# ── summary ────────────────────────────────────────────────────────────────────
PASS_COUNT=$(grep -c '^--- PASS:' "$LOGFILE" || true)
FAIL_COUNT=$(grep -c '^--- FAIL:' "$LOGFILE" || true)
SKIP_COUNT=$(grep -c '^--- SKIP:' "$LOGFILE" || true)
TOTAL=$(( PASS_COUNT + FAIL_COUNT + SKIP_COUNT ))

{
    echo ""
    echo "========================================================"
    echo "  SUMMARY"
    echo "========================================================"
    printf '  Total    : %d\n' "$TOTAL"
    printf '  Passed   : %d\n' "$PASS_COUNT"
    printf '  Failed   : %d\n' "$FAIL_COUNT"
    printf '  Skipped  : %d\n' "$SKIP_COUNT"
    printf '  Duration : %s\n' "$ELAPSED_FMT"
    echo ""

    if [[ $FAIL_COUNT -gt 0 ]]; then
        echo "  --- failing tests ---"
        while IFS= read -r fail_line; do
            test_name=$(echo "$fail_line" | awk '{print $3}')
            duration=$(echo "$fail_line"  | awk '{print $4}')
            printf '  ✗  %s %s\n' "$test_name" "$duration"
            # First error/step line from this test's output block
            first_error=$(awk -v t="$test_name" '
                $0 ~ ("^=== RUN[[:space:]]+" t "$") { capture=1; next }
                capture && /^    / && /\.go:[0-9]+:/ { print; capture=0 }
                /^--- (PASS|FAIL|SKIP):/ { capture=0 }
            ' "$LOGFILE" | head -1 | sed 's/^[[:space:]]*//')
            [[ -n "$first_error" ]] && printf '     %s\n' "$first_error"
            echo ""
        done < <(grep '^--- FAIL:' "$LOGFILE")
    fi

    if [[ $SKIP_COUNT -gt 0 ]]; then
        echo "  --- skipped tests ---"
        grep '^--- SKIP:' "$LOGFILE" \
            | awk '{printf "  ⊘  %s %s\n", $3, $4}' \
            || true
        echo ""
    fi

    echo "  --- artifacts ---"
    printf '  Test log  : %s\n' "$LOGFILE"
    printf '  TF log    : %s\n' "$TF_LOGFILE"
    printf '  Summary   : %s\n' "$SUMMARY_FILE"
    echo ""
    echo "========================================================"
    printf '  Finished : %s\n' "$(date)"
    if [[ $EXIT_CODE -eq 0 ]]; then
        echo "  Result   : ✓  PASSED"
    else
        echo "  Result   : ✗  FAILED (exit ${EXIT_CODE})"
    fi
    echo "========================================================"
} | tee -a "$LOGFILE" | tee "$SUMMARY_FILE"

exit $EXIT_CODE
