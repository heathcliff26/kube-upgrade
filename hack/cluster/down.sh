#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath | xargs dirname | xargs dirname)"
# shellcheck source=common.sh
source "${base_dir}/hack/cluster/common.sh"

if [ ! -f "${pid_file}" ]; then
    echo "No cluster running, exiting..."
    exit 0
fi

kill "$(cat "${pid_file}")"
while [ -f "${pid_file}" ]; do
    sleep 1
done
