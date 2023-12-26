#!/bin/bash

export PATH=/opt/openshift/sales/client:/usr/bin:/bin

stdin_json=$(cat -)
default_namespace_name=$(echo $stdin_json | /usr/bin/jq -r '.Runtime.DefaultNamespace.Name // empty')
assets_root_path=$(echo $stdin_json | /usr/bin/jq -r '.Runtime.Context.CurrentCase.RetrievedAssetsDirectoryPath // empty')

if [[ $default_namespace_name == "" ]]; then
    echo ".Config.DefaultNamespace.GeneratedName is empty or does not exist" >&2
    exit 1
fi

if [[ $assets_root_path == "" ]]; then
    echo ".Config.Context.TestCaseRetrievedAssetsDirectoryPath is empty or does not exist" >&2
    exit 2
fi

if [[ $(oc get pods extractor -n "$default_namespace_name" | grep Running) == "" ]]; then
    echo "No Pod named (extractor) appears to be in Running state in Namespace ($default_namespace_name)" >&2
    exit 5
fi

if [[ ! -d "$assets_root_path" ]]; then
    echo "Declared directory .Runtime.Context.CurrentCase.RetrievedAssetsDirectoryPath ($assets_root_path) does not exist" >&2
    exit 6
fi

if err=$(oc exec extractor -n "$default_namespace_name" -- /usr/bin/tar cf - -C /opt/test_results . | tar xf - -C "$assets_root_path/$test_unit_name/$test_case_name" 2>&1); then
    :
else
    echo "Failed to extract assets files from /opt/test_results on Pod (extractor) in Namespace ($default_namespace_name) to local directory ($assets_root_path): $err" >&2
    exit 7
fi
