#!/bin/bash

# Bundle and minify CSS files with UUID filename
# Usage: ./bundle-css.sh [minifier] [custom_id]
# Options: esbuild, cssnano, clean-css, none
# Default: auto-detect available minifier
# Custom ID: optional custom identifier for output filename (defaults to git commit hash)

set -e

ASSETS_DIR="."
DIST_DIR="bundle"
TEMP_BUNDLE="temp_bundle.css"
MINIFIER="${1:-auto}"
CUSTOM_ID="${2:-}"

# Create bundle directory and clean CSS files and assets
mkdir -p "$DIST_DIR"
echo "Cleaning existing CSS files and assets from bundle directory..."
rm -f "$DIST_DIR"/*.css "$DIST_DIR"/*.css.gz
rm -rf "$DIST_DIR/fonts"

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

OUTPUT_FILE="$DIST_DIR/$BUNDLE_ID.css"
GZIP_FILE="$OUTPUT_FILE.gz"

# Determine minifier to use
if [[ "$MINIFIER" == "auto" ]]; then
    if command -v esbuild &> /dev/null; then
        MINIFIER="esbuild"
    elif command -v cssnano &> /dev/null; then
        MINIFIER="cssnano"
    elif command -v cleancss &> /dev/null; then
        MINIFIER="clean-css"
    else
        MINIFIER="none"
        echo "⚠️  No minifier found. Using concatenation only."
        echo "   Install esbuild: npm install -g esbuild"
        echo "   Install cssnano: npm install -g cssnano-cli"
        echo "   Install clean-css: npm install -g clean-css-cli"
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
    "cssnano")
        if ! command -v cssnano &> /dev/null; then
            echo "Error: cssnano not found. Install with: npm install -g cssnano-cli"
            exit 1
        fi
        echo "Using cssnano for minification"
        ;;
    "clean-css")
        if ! command -v cleancss &> /dev/null; then
            echo "Error: clean-css not found. Install with: npm install -g clean-css-cli"
            exit 1
        fi
        echo "Using clean-css for minification"
        ;;
    "none")
        echo "No minification - concatenation only"
        ;;
    *)
        echo "Error: Unknown minifier '$MINIFIER'. Use: esbuild, cssnano, clean-css, or none"
        exit 1
        ;;
esac

# CSS files in order of dependency
CSS_FILES=(
    "$ASSETS_DIR/bootstrap/css/bootstrap.min.css"
    "$ASSETS_DIR/bootstrap-icons/bootstrap-icons.min.css"
    "$ASSETS_DIR/fonts/local.css"
    "$ASSETS_DIR/css/styles.css"
)

echo "Bundling CSS files..."

# Check if all files exist
for file in "${CSS_FILES[@]}"; do
    if [[ ! -f "$file" ]]; then
        echo "Error: File not found: $file"
        exit 1
    fi
done

# Copy font assets to bundle directory
echo "Copying font assets..."
mkdir -p "$DIST_DIR/fonts/google"
mkdir -p "$DIST_DIR/fonts"

# Copy Google fonts
if [[ -d "$ASSETS_DIR/fonts/google" ]]; then
    cp "$ASSETS_DIR/fonts/google"/*.woff2 "$DIST_DIR/fonts/google/" 2>/dev/null || true
    echo "Copied Google fonts to bundle/fonts/google/"
fi

# Copy Bootstrap Icons fonts
if [[ -d "$ASSETS_DIR/bootstrap-icons/fonts" ]]; then
    cp "$ASSETS_DIR/bootstrap-icons/fonts"/*.woff* "$DIST_DIR/fonts/" 2>/dev/null || true
    echo "Copied Bootstrap Icons fonts to bundle/fonts/"
fi

# Concatenate files with path corrections
> "$TEMP_BUNDLE"
for file in "${CSS_FILES[@]}"; do
    echo "Adding: $file"
    echo "/* === $file === */" >> "$TEMP_BUNDLE"
    
    # Fix paths in CSS files during concatenation
    if [[ "$file" == *"fonts/local.css" ]]; then
        # Fix Google fonts paths from ./google/ to ./fonts/google/
        sed 's|url(\.\/google\/|url(./fonts/google/|g' "$file" >> "$TEMP_BUNDLE"
    elif [[ "$file" == *"bootstrap-icons/bootstrap-icons.min.css" ]]; then
        # Fix Bootstrap Icons fonts paths from fonts/ to ./fonts/
        sed 's|url("fonts\/|url("./fonts/|g' "$file" >> "$TEMP_BUNDLE"
    else
        cat "$file" >> "$TEMP_BUNDLE"
    fi
    
    echo "" >> "$TEMP_BUNDLE"
done

if [[ "$MINIFIER" != "none" ]]; then
    echo "Minifying bundle with $MINIFIER..."
fi

# Minify the bundle based on chosen minifier
case "$MINIFIER" in
    "esbuild")
        esbuild "$TEMP_BUNDLE" --minify --outfile="$OUTPUT_FILE"
        ;;
    "cssnano")
        cssnano "$TEMP_BUNDLE" "$OUTPUT_FILE"
        ;;
    "clean-css")
        cleancss -o "$OUTPUT_FILE" "$TEMP_BUNDLE"
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
echo "🔗 Use in HTML: <link rel=\"stylesheet\" href=\"./bundle/$BUNDLE_ID.css\">"
