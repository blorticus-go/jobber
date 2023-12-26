#!/bin/bash

export PATH=/opt/openshift/sales/client:/usr/bin:/bin

stdin_json=$(cat -)
default_namespace_name=$(echo $stdin_json | /usr/bin/jq -r '.Config.DefaultNamespace.Name // empty')

if [[ $default_namespace_name == "" ]]; then
    echo ".Config.DefaultNamespace.Name is empty or does not exist" >&2
    exit 1
fi

if err=$(oc adm policy add-scc-to-user anyuid -z default -n "$default_namespace_name" 2>&1); then
    :
else
    echo "failed to add (anyuid) scc to (default) service account in namespace ($default_namespace_name): $err" >&2
    exit 2
fi
