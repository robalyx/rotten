#!/bin/bash
set -euo pipefail

# Configuration
REPO_OWNER="robalyx"
REPO_NAME="rotten"
EXPORT_DIR="exports/official"
GITHUB_TOKEN="$1"

# Read version from export config
ENGINE_VERSION=$(jq -r '.engineVersion' "$EXPORT_DIR/export_config.json")
EXPORT_VERSION=$(jq -r '.exportVersion' "$EXPORT_DIR/export_config.json")

if [ -z "$EXPORT_VERSION" ] || [ -z "$ENGINE_VERSION" ]; then
    echo "Error: Failed to read version from export config"
    exit 1
fi


# Create export zip file
EXPORT_ZIP="export-${EXPORT_VERSION}.zip"
cd "$EXPORT_DIR"
zip -r "../../$EXPORT_ZIP" ./*
cd ../..

# Calculate export checksum
EXPORT_CHECKSUM=$(sha256sum "$EXPORT_ZIP" | cut -d' ' -f1)

# Calculate executable checksums
LINUX_CHECKSUM=$(sha256sum "rotten-linux-amd64" | cut -d' ' -f1)
WINDOWS_CHECKSUM=$(sha256sum "rotten-windows-amd64.exe" | cut -d' ' -f1)
DARWIN_AMD64_CHECKSUM=$(sha256sum "rotten-darwin-amd64" | cut -d' ' -f1)
DARWIN_ARM64_CHECKSUM=$(sha256sum "rotten-darwin-arm64" | cut -d' ' -f1)

# Create release notes
RELEASE_NOTES="## Engine Version
${ENGINE_VERSION}

## Usage
1. Download the appropriate executable for your platform
2. Run the executable from the terminal/command prompt
3. You can download official exports directly from within the program when selecting directories

## Export Data
SHA256 hash for verification:
\`\`\`
${EXPORT_ZIP} SHA256: ${EXPORT_CHECKSUM}
\`\`\`

## Executables
SHA256 hashes for verification:
\`\`\`
rotten-linux-amd64 SHA256: ${LINUX_CHECKSUM}
rotten-windows-amd64.exe SHA256: ${WINDOWS_CHECKSUM}
rotten-darwin-amd64 SHA256: ${DARWIN_AMD64_CHECKSUM}
rotten-darwin-arm64 SHA256: ${DARWIN_ARM64_CHECKSUM}
\`\`\`"

# Verify release notes format
if ! echo "$RELEASE_NOTES" | grep -q "^${EXPORT_ZIP} SHA256: ${EXPORT_CHECKSUM}$"; then
  echo "Error: Release notes checksum format is incorrect"
  exit 1
fi

# Create the release with formatted notes
gh release create "$EXPORT_VERSION" \
  --title "Rotten v$EXPORT_VERSION" \
  --notes "$RELEASE_NOTES" \
  --draft \
  rotten-linux-amd64 \
  rotten-windows-amd64.exe \
  rotten-darwin-amd64 \
  rotten-darwin-arm64 \
  "$EXPORT_ZIP"

# Cleanup
rm "$EXPORT_ZIP" 