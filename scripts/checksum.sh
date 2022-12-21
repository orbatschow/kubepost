#!/usr/bin/env bash

ROOT_DIR=$(git rev-parse --show-toplevel)

mkdir -p "$ROOT_DIR"/build

# exclude all hidden directories, bin and build
find . -type f ! -path './.*/*' ! -path './build/*' ! -path './bin/*' ! -path 'CHANGELOG.md' -exec md5sum "{}" + > build/before.chk

echo "computed pre hashes"
cat "$ROOT_DIR"/build/before.chk

(cd "$ROOT_DIR" && make generate)

# exclude all hidden directories, bin and build
find . -type f ! -path './.*/*' ! -path './build/*' ! -path './bin/*' ! -path 'CHANGELOG.md' -exec md5sum "{}" + > build/after.chk

echo "computed post hashes"
cat "$ROOT_DIR"/build/after.chk

chk1=$(cksum build/before.chk | awk -F" " '{print $1}')
chk2=$(cksum build/after.chk | awk -F" " '{print $1}')

if [ "$chk1" -eq "$chk2" ]
then
  echo "client, custom resource definitions and documentation is up to date, exiting"
  exit 0
else
  echo "client, custom resource definitions or documentation is not up to date, please execute 'make generate'"
  exit 1
fi