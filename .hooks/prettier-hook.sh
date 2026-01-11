#!/usr/bin/env bash
# Run prettier on files

set -euo pipefail

# Get the list of files passed by pre-commit
FILES=("$@")

# Check if prettier is installed
if ! command -v prettier &>/dev/null; then
	echo "prettier is not installed. Installing..."
	npm install -g prettier
fi

# Run prettier on the files
for file in "${FILES[@]}"; do
	prettier --write "$file"
done
