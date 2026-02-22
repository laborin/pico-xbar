#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"
APP_NAME="pico-xbar"
APP_BUNDLE="$BUILD_DIR/$APP_NAME.app"

if [ ! -d "$APP_BUNDLE" ]; then
    echo "Error: $APP_BUNDLE not found. Run build.sh first."
    exit 1
fi

if [ -z "$SIGNING_IDENTITY" ]; then
    echo "Error: SIGNING_IDENTITY not set"
    echo "Set it to your Developer ID Application certificate name, e.g.:"
    echo "  export SIGNING_IDENTITY='Developer ID Application: Your Name (TEAMID)'"
    exit 1
fi

if [ -z "$APPLE_ID" ] || [ -z "$APPLE_TEAM_ID" ] || [ -z "$APPLE_APP_PASSWORD" ]; then
    echo "Error: Missing Apple credentials for notarization"
    echo "Required environment variables:"
    echo "  APPLE_ID          - Your Apple ID email"
    echo "  APPLE_TEAM_ID     - Your Team ID"
    echo "  APPLE_APP_PASSWORD - App-specific password"
    exit 1
fi

VERSION=$(/usr/libexec/PlistBuddy -c "Print CFBundleShortVersionString" "$APP_BUNDLE/Contents/Info.plist")
echo "Signing and notarizing $APP_NAME version $VERSION..."

echo "Signing application..."
codesign --force --deep --options runtime \
    --sign "$SIGNING_IDENTITY" \
    --timestamp \
    "$APP_BUNDLE"

codesign --verify --verbose "$APP_BUNDLE"

echo "Creating zip for notarization..."
NOTARIZE_ZIP="$BUILD_DIR/${APP_NAME}-notarize.zip"
ditto -c -k --keepParent "$APP_BUNDLE" "$NOTARIZE_ZIP"

echo "Submitting for notarization..."
xcrun notarytool submit "$NOTARIZE_ZIP" \
    --apple-id "$APPLE_ID" \
    --team-id "$APPLE_TEAM_ID" \
    --password "$APPLE_APP_PASSWORD" \
    --wait

rm "$NOTARIZE_ZIP"

echo "Stapling notarization ticket..."
xcrun stapler staple "$APP_BUNDLE"

echo "Creating distribution zip..."
DIST_ZIP="$BUILD_DIR/${APP_NAME}-${VERSION}.zip"
ditto -c -k --keepParent "$APP_BUNDLE" "$DIST_ZIP"

echo "Verifying notarization..."
spctl --assess --verbose --type execute "$APP_BUNDLE"

echo ""
echo "Done! Distribution package: $DIST_ZIP"
echo "Upload this file to your GitHub release."
