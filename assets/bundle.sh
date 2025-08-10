#!/bin/bash

# Master bundler script - builds both JavaScript and CSS bundles
# Usage: ./bundle.sh [minifier] [custom_id]
# Options: esbuild, terser, cssnano, clean-css, none, auto
# Default: auto-detect available minifier
# Custom ID: optional custom identifier for output filename (defaults to git commit hash)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MINIFIER="${1:-auto}"
CUSTOM_ID="${2:-}"

echo "🚀 Starting asset bundling..."
echo "📦 Minifier: $MINIFIER"
if [[ -n "$CUSTOM_ID" ]]; then
    echo "🏷️  Custom ID: $CUSTOM_ID"
else
    echo "🏷️  Using git commit hash for filename"
fi
echo ""

# Check if bundle scripts exist
if [[ ! -f "$SCRIPT_DIR/bundle-js.sh" ]]; then
    echo "❌ Error: bundle-js.sh not found in $SCRIPT_DIR"
    exit 1
fi

if [[ ! -f "$SCRIPT_DIR/bundle-css.sh" ]]; then
    echo "❌ Error: bundle-css.sh not found in $SCRIPT_DIR"
    exit 1
fi

# Make sure scripts are executable
chmod +x "$SCRIPT_DIR/bundle-js.sh"
chmod +x "$SCRIPT_DIR/bundle-css.sh"

echo "📝 Bundling JavaScript files..."
echo "================================"
"$SCRIPT_DIR/bundle-js.sh" "$MINIFIER" "$CUSTOM_ID"

echo ""
echo "🎨 Bundling CSS files..."
echo "========================"
"$SCRIPT_DIR/bundle-css.sh" "$MINIFIER" "$CUSTOM_ID"

echo ""
echo "✅ All bundles created successfully!"
echo "📁 Check the bundle/ directory for your bundled assets"