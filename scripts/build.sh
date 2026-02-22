#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"
APP_NAME="pico-xbar"
APP_BUNDLE="$BUILD_DIR/$APP_NAME.app"

VERSION="${VERSION:-dev}"
if [ "$VERSION" = "dev" ] && git describe --tags --exact-match 2>/dev/null; then
    VERSION=$(git describe --tags --exact-match | sed 's/^v//')
fi

echo "Building $APP_NAME version $VERSION..."

rm -rf "$BUILD_DIR"
mkdir -p "$APP_BUNDLE/Contents/MacOS"
mkdir -p "$APP_BUNDLE/Contents/Resources"

echo "Creating universal binary..."
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build \
    -ldflags "-X github.com/laborin/pico-xbar/internal/version.Version=$VERSION" \
    -o "$BUILD_DIR/${APP_NAME}-amd64" \
    "$PROJECT_ROOT/cmd/pico-xbar"

CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build \
    -ldflags "-X github.com/laborin/pico-xbar/internal/version.Version=$VERSION" \
    -o "$BUILD_DIR/${APP_NAME}-arm64" \
    "$PROJECT_ROOT/cmd/pico-xbar"

lipo -create -output "$APP_BUNDLE/Contents/MacOS/$APP_NAME" \
    "$BUILD_DIR/${APP_NAME}-amd64" \
    "$BUILD_DIR/${APP_NAME}-arm64"

rm "$BUILD_DIR/${APP_NAME}-amd64" "$BUILD_DIR/${APP_NAME}-arm64"

echo "Creating Info.plist..."
sed "s/VERSION_PLACEHOLDER/$VERSION/g" "$PROJECT_ROOT/assets/Info.plist" \
    > "$APP_BUNDLE/Contents/Info.plist"

echo "Creating app icon..."
ICONSET_DIR="$BUILD_DIR/AppIcon.iconset"
mkdir -p "$ICONSET_DIR"

SOURCE_ICON="$PROJECT_ROOT/assets/appicon.png"
sips -z 16 16     "$SOURCE_ICON" --out "$ICONSET_DIR/icon_16x16.png" >/dev/null
sips -z 32 32     "$SOURCE_ICON" --out "$ICONSET_DIR/icon_16x16@2x.png" >/dev/null
sips -z 32 32     "$SOURCE_ICON" --out "$ICONSET_DIR/icon_32x32.png" >/dev/null
sips -z 64 64     "$SOURCE_ICON" --out "$ICONSET_DIR/icon_32x32@2x.png" >/dev/null
sips -z 128 128   "$SOURCE_ICON" --out "$ICONSET_DIR/icon_128x128.png" >/dev/null
sips -z 256 256   "$SOURCE_ICON" --out "$ICONSET_DIR/icon_128x128@2x.png" >/dev/null
sips -z 256 256   "$SOURCE_ICON" --out "$ICONSET_DIR/icon_256x256.png" >/dev/null
sips -z 512 512   "$SOURCE_ICON" --out "$ICONSET_DIR/icon_256x256@2x.png" >/dev/null
sips -z 512 512   "$SOURCE_ICON" --out "$ICONSET_DIR/icon_512x512.png" >/dev/null
sips -z 1024 1024 "$SOURCE_ICON" --out "$ICONSET_DIR/icon_512x512@2x.png" >/dev/null

iconutil -c icns -o "$APP_BUNDLE/Contents/Resources/AppIcon.icns" "$ICONSET_DIR"
rm -rf "$ICONSET_DIR"

echo "Copying localizations..."
for lproj in "$PROJECT_ROOT/assets"/*.lproj; do
    if [ -d "$lproj" ]; then
        cp -R "$lproj" "$APP_BUNDLE/Contents/Resources/"
    fi
done

echo "Build complete: $APP_BUNDLE"
echo "Version: $VERSION"
