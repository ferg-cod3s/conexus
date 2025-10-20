#!/bin/bash
set -e

# Build binaries for all supported platforms
echo "Building conexus binaries for multiple platforms..."

# Output directory
OUTDIR="bin"
mkdir -p "$OUTDIR"

# Build flags for smaller binaries
LDFLAGS="-s -w"

# Platforms to build for
PLATFORMS=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
)

for platform in "${PLATFORMS[@]}"; do
  IFS="/" read -r -a array <<< "$platform"
  GOOS="${array[0]}"
  GOARCH="${array[1]}"
  
  output_name="$OUTDIR/conexus-$GOOS-$GOARCH"
  
  if [ "$GOOS" = "windows" ]; then
    output_name+='.exe'
  fi
  
  echo "Building for $GOOS/$GOARCH..."
  env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="$LDFLAGS" -trimpath -o "$output_name" ./cmd/conexus
  
  # Make executable on Unix
  if [ "$GOOS" != "windows" ]; then
    chmod +x "$output_name"
  fi
done

echo "✓ All binaries built successfully!"
echo "Binaries are in the $OUTDIR/ directory"
