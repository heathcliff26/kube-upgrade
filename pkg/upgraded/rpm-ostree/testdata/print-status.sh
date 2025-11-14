#!/bin/bash

script_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)"

cat "${script_dir}/status.json"
