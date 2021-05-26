#!/usr/bin/env bash

set -eu

declare -a current_checksums
declare -a computed_checksums

# Iterate over files in maifests/crd folder and build checksum
for file in "$(pwd)"/manifests/crd/*.yaml; do
	# calculate md5sum
	checksum=$(md5sum "$file" | awk '{ print $1 }')

	current_checksums=("${current_checksums[@]}" "$checksum")
done

# recreate all manifests
make manifests

# Iterate over new manifests in maifests/crd folder and build the new checksums
for file in "$(pwd)"/manifests/crd/*.yaml; do
	# calculate md5sum
	checksum=$(md5sum "$file" | awk '{ print $1 }')

	computed_checksums=("${computed_checksums[@]}" "$checksum")
done

declare -a difference

difference=($(echo "${current_checksums[@]}" "${computed_checksums[@]}" | tr ' ' '\n' | sort | uniq -u))

if [ ${#difference[@]} -eq 0 ]; then
	echo "checksum computation passed successfully"
	exit 0
else
	echo "checksum computation failed, execute 'make manifests' before pushing code with changed CRD specifications"
	exit 1
fi
