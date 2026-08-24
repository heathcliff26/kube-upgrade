#!/bin/bash

if [[ "${1:-}" == "version" ]]; then
    echo "v1.30.4"
    exit 0
fi

exit 1
