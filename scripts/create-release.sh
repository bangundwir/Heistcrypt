#!/bin/bash

# HadesCrypt Release Script
# Usage: ./scripts/create-release.sh <version>
# Example: ./scripts/create-release.sh 2.0.1

set -e

VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 2.0.1"
    exit 1
fi

# Validate version format (semantic versioning)
if ! [[ $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must be in format X.Y.Z (e.g., 2.0.1)"
    exit 1
fi

echo "🔱 Creating release for HadesCrypt v$VERSION"

# Update VERSION file
echo "$VERSION" > VERSION
echo "✅ Updated VERSION file to $VERSION"

# Update CHANGELOG.md if it exists
if [ -f "CHANGELOG.md" ]; then
    # Add new version entry to changelog
    TEMP_FILE=$(mktemp)
    echo "# Changelog" > "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
    echo "## [v$VERSION] - $(date +%Y-%m-%d)" >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
    echo "### Added" >> "$TEMP_FILE"
    echo "- " >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
    echo "### Changed" >> "$TEMP_FILE"
    echo "- " >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
    echo "### Fixed" >> "$TEMP_FILE"
    echo "- " >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
    cat CHANGELOG.md >> "$TEMP_FILE"
    mv "$TEMP_FILE" CHANGELOG.md
    echo "✅ Updated CHANGELOG.md"
fi

# Create git tag
git add VERSION CHANGELOG.md
git commit -m "Release v$VERSION"
git tag -a "v$VERSION" -m "Release v$VERSION"

echo "✅ Created git tag v$VERSION"

# Push to remote
echo "🚀 Pushing to remote repository..."
git push origin main
git push origin "v$VERSION"

echo ""
echo "🎉 Release v$VERSION created successfully!"
echo ""
echo "Next steps:"
echo "1. GitHub Actions will automatically build and create a release"
echo "2. Check the Actions tab in your GitHub repository"
echo "3. The release will be available at: https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^.]*\).*/\1/')/releases"
echo ""
echo "To create a manual build, you can also run:"
echo "  go build -ldflags \"-s -w -H windowsgui -X main.version=$VERSION\" -o dist/windows/HadesCrypt-v$VERSION-Windows-x64.exe ."
