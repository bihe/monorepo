#!/bin/bash

# Bundle and minify JavaScript files with UUID filename
# Usage: ./bundle-js.sh [minifier] [custom_id]
# Options: terser, esbuild, none
# Default: auto-detect available minifier
# Custom ID: optional custom identifier for output filename (defaults to git commit hash)

set -e

ASSETS_DIR="."
DIST_DIR="bundle"
TEMP_BUNDLE="temp_bundle.js"
MINIFIER="${1:-auto}"
CUSTOM_ID="${2:-}"

# Create bundle directory and clean JS files
mkdir -p "$DIST_DIR"
echo "Cleaning existing JS files from bundle directory..."
rm -f "$DIST_DIR"/*.js "$DIST_DIR"/*.js.gz

# Determine output filename
if [[ -n "$CUSTOM_ID" ]]; then
    BUNDLE_ID="$CUSTOM_ID"
    echo "Using custom ID: $BUNDLE_ID"
else
    # Get current git commit hash
    if ! command -v git &> /dev/null; then
        echo "Error: git command not found"
        exit 1
    fi

    if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
        echo "Error: not inside a git repository"
        exit 1
    fi

    BUNDLE_ID=$(git rev-parse --short HEAD)
    echo "Using git commit hash: $BUNDLE_ID"
fi

OUTPUT_FILE="$DIST_DIR/$BUNDLE_ID.js"
GZIP_FILE="$OUTPUT_FILE.gz"

# Determine minifier to use
if [[ "$MINIFIER" == "auto" ]]; then
    if command -v esbuild &> /dev/null; then
        MINIFIER="esbuild"
    elif command -v terser &> /dev/null; then
        MINIFIER="terser"
    else
        MINIFIER="none"
        echo "⚠️  No minifier found. Using concatenation only."
        echo "   Install esbuild: npm install -g esbuild"
        echo "   Install terser: npm install -g terser"
    fi
fi

# Validate minifier choice
case "$MINIFIER" in
    "esbuild")
        if ! command -v esbuild &> /dev/null; then
            echo "Error: esbuild not found. Install with: npm install -g esbuild"
            exit 1
        fi
        echo "Using esbuild for minification"
        ;;
    "terser")
        if ! command -v terser &> /dev/null; then
            echo "Error: terser not found. Install with: npm install -g terser"
            exit 1
        fi
        echo "Using terser for minification"
        ;;
    "none")
        echo "No minification - concatenation only"
        ;;
    *)
        echo "Error: Unknown minifier '$MINIFIER'. Use: terser, esbuild, or none"
        exit 1
        ;;
esac

# JavaScript files in order of dependency
JS_FILES=(
    "$ASSETS_DIR/bootstrap/js/popper.min.js"
    "$ASSETS_DIR/bootstrap/js/bootstrap.bundle.min.js"
    "$ASSETS_DIR/js/htmx.min.js"
    "$ASSETS_DIR/js/_hyperscript.min.js"
    "$ASSETS_DIR/js/Sortable.min.js"
    "$ASSETS_DIR/js/script.js"
)

echo "Bundling JavaScript files..."

# Check if all files exist
for file in "${JS_FILES[@]}"; do
    if [[ ! -f "$file" ]]; then
        echo "Error: File not found: $file"
        exit 1
    fi
done

# Concatenate files
> "$TEMP_BUNDLE"
for file in "${JS_FILES[@]}"; do
    echo "Adding: $file"
    echo "/* === $file === */" >> "$TEMP_BUNDLE"
    cat "$file" >> "$TEMP_BUNDLE"
    echo "" >> "$TEMP_BUNDLE"
done

if [[ "$MINIFIER" != "none" ]]; then
    echo "Minifying bundle with $MINIFIER..."
fi

# Minify the bundle based on chosen minifier
case "$MINIFIER" in
    "esbuild")
        esbuild "$TEMP_BUNDLE" --minify --target=es6 --outfile="$OUTPUT_FILE"
        ;;
    "terser")
        terser "$TEMP_BUNDLE" --compress --mangle --output "$OUTPUT_FILE"
        ;;
    "none")
        mv "$TEMP_BUNDLE" "$OUTPUT_FILE"
        ;;
esac

# Clean up temp file if it still exists
[[ -f "$TEMP_BUNDLE" ]] && rm "$TEMP_BUNDLE"

# Create gzip version
echo "Creating gzip compressed version..."
gzip -c "$OUTPUT_FILE" > "$GZIP_FILE"

echo "✅ Bundle created: $OUTPUT_FILE"
echo "✅ Gzip version: $GZIP_FILE"
echo "📁 Original size: $(du -h "$OUTPUT_FILE" | cut -f1)"
echo "📁 Gzip size: $(du -h "$GZIP_FILE" | cut -f1)"
echo "🔧 Minifier used: $MINIFIER"
echo "📊 Compression ratio: $(echo "scale=1; $(stat -f%z "$GZIP_FILE") * 100 / $(stat -f%z "$OUTPUT_FILE")" | bc)%"
echo "🔗 Use in HTML: <script src=\"./bundle/$BUNDLE_ID.js\"></script>"
