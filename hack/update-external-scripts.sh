#!/bin/bash

script_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)"

echo "Updating CRD conversion script"
curl -SL -o "${script_dir}/external/openapi2jsonschema.py" https://raw.githubusercontent.com/yannh/kubeconform/master/scripts/openapi2jsonschema.py
