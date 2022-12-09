#!/usr/bin/env bash

ROOT_DIR=$(git rev-parse --show-toplevel)

PRE_HASH=$(tar -cf - "$ROOT_DIR" | md5sum | awk '{print $1}')

echo "computed pre hash: $PRE_HASH"

(cd "$ROOT_DIR" && make generate)

POST_HASH=$(tar -cf - "$ROOT_DIR" | md5sum | awk '{print $1}')

echo "computed post hash: $PRE_HASH"

if [[ "$PRE_HASH" == "$POST_HASH" ]]; then
  echo "client and custom resource definitions are up to date, exiting"
  exit 0
else
  echo "client and custom resource definitions are not up to date, please execute 'make generate'"
  exit 1
fi
