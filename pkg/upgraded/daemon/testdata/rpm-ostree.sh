#!/bin/bash

set -u

script_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)"
log_file="${RPM_OSTREE_LOG_FILE:-/tmp/rpm-ostree-calls.log}"

echo "$@" >>"${log_file}"

case "${1:-}" in
upgrade)
    if [[ "${2:-}" == "--check" ]]; then
        exit "${RPM_OSTREE_CHECK_EXIT_CODE:-77}"
    fi
    exit "${RPM_OSTREE_UPGRADE_EXIT_CODE:-0}"
;;
rebase)
    exit "${RPM_OSTREE_REBASE_EXIT_CODE:-0}"
;;
deploy)
    exit "${RPM_OSTREE_DEPLOY_EXIT_CODE:-0}"
;;
status)
    cat "${RPM_OSTREE_STATUS_FILE:-${script_dir}/status.json}"
    exit "${RPM_OSTREE_STATUS_EXIT_CODE:-0}"
;;
*)
    exit "${RPM_OSTREE_UNKNOWN_EXIT_CODE:-1}"
esac
