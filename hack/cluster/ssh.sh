#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath | xargs dirname | xargs dirname)"
# shellcheck source=common.sh
source "${base_dir}/hack/cluster/common.sh"

# shellcheck disable=SC2068
ssh -o "StrictHostKeyChecking=no" -o "UserKnownHostsFile=/dev/null" -i "${ssh_key}" -p 2222 core@localhost ${@}
