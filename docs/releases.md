# Release Process

This document explains the automated release process for the Agar project.

## Overview

Agar uses GitHub Actions to automate releases of both the library and CLI. The release workflow is triggered when a pull request is merged to the `main` branch.

## Version Determination

### Method 1: Explicit Version (Release Branch)

Create a branch named `release/vX.Y.Z`:

```bash
git checkout -b release/v1.2.0
git push origin release/v1.2.0
```

When the PR is merged, version `v1.2.0` will be used for the release.

### Method 2: Auto-Increment (Any Other Branch)

For any other branch name (feature/, fix/, etc.), the workflow will:
1. Find the latest tag (e.g., `v1.1.5`)
2. Increment the patch version (→ `v1.1.6`)
3. Use the new version for the release

```bash
git checkout -b feature/new-tool
# ... make changes ...
git push origin feature/new-tool
# Merge PR → Automatically releases v1.1.6
```

## What Gets Released

### Library Release (`vX.Y.Z`)

**Tag**: `vX.Y.Z`

**Includes**:
- `tools/` package
- `tui/` package
- Root go.mod
- Documentation

**Installation**:
```bash
go get github.com/geoffjay/agar@vX.Y.Z
```

### CLI Release (`cmd/agar/vX.Y.Z`)

**Tag**: `cmd/agar/vX.Y.Z`

**Process**:
1. Remove `replace` directive from cmd/agar/go.mod
2. Update requirement: `github.com/geoffjay/agar@vX.Y.Z`
3. Build and test with GOWORK=off
4. Create tag and GitHub release
5. Restore `replace` directive for development

**Installation**:
```bash
go install github.com/geoffjay/agar/cmd/agar@vX.Y.Z
```

## Automated Steps

When a PR is merged to `main`:

1. **Determine Version**
   - Parse branch name or increment latest tag

2. **Run Tests**
   - Full test suite for library
   - Verify CLI builds

3. **Release Library**
   - Create git tag: `vX.Y.Z`
   - Push tag to GitHub
   - Create GitHub release with changelog

4. **Update CLI**
   - Remove `replace` directive
   - Update library dependency to new version
   - Commit changes
   - Verify builds without workspace

5. **Release CLI**
   - Create git tag: `cmd/agar/vX.Y.Z`
   - Build optimized binary
   - Create GitHub release with binary
   - Attach installation instructions

6. **Restore Development State**
   - Add back `replace` directive
   - Commit restoration
   - Push to main

## Manual Override

To skip automatic release, include `[skip release]` in the PR title or description.

## Version Bumping

### Patch Release (v1.2.3 → v1.2.4)
- Bug fixes
- Documentation updates
- Small improvements
- **Automatic**: Merge any non-release branch

### Minor Release (v1.2.3 → v1.3.0)
- New features
- New tools/components
- **Manual**: Use `release/v1.3.0` branch

### Major Release (v1.2.3 → v2.0.0)
- Breaking API changes
- Major restructuring
- **Manual**: Use `release/v2.0.0` branch

## Workflow File

`.github/workflows/release.yml`

**Triggers**:
- Pull request merged to main

**Environment Variables**:
- `GITHUB_TOKEN` - Automatically provided by GitHub Actions

**Jobs**:
- `release` - Handles version detection, tagging, and GitHub releases

## Testing the Workflow

### Local Testing (Dry Run)

```bash
# Simulate version detection
BRANCH_NAME="release/v1.5.0"
if [[ $BRANCH_NAME =~ ^release/v([0-9]+\.[0-9]+\.[0-9]+)$ ]]; then
  echo "Would use version: v${BASH_REMATCH[1]}"
else
  LATEST=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
  echo "Would increment from: $LATEST"
fi
```

### First Release

If no tags exist yet:
```bash
git checkout -b release/v0.1.0
git push origin release/v0.1.0
# Create and merge PR
# → Creates v0.1.0 and cmd/agar/v0.1.0
```

## Troubleshooting

**If release fails**:
1. Check GitHub Actions logs
2. Verify tests pass
3. Ensure GITHUB_TOKEN has write permissions
4. Check for conflicting tags

**If CLI doesn't build after release**:
1. Verify library tag exists
2. Check go.mod has correct version
3. Run `GOWORK=off go build` to test without workspace

**To manually fix a release**:
```bash
# Delete tags if needed
git tag -d vX.Y.Z
git push origin :refs/tags/vX.Y.Z

# Delete GitHub release
gh release delete vX.Y.Z

# Re-run workflow or create tags manually
```

## Best Practices

1. **Always create feature branches** from main
2. **Use conventional commit messages** for better changelogs
3. **Test locally** before creating PR
4. **Use `release/vX.Y.Z`** for intentional version bumps
5. **Let patch auto-increment** for routine merges
6. **Review draft releases** before publishing (if enabled)

## Future Enhancements

- Automatic changelog generation from commits
- Slack/Discord notifications on release
- Binary releases for multiple platforms
- Homebrew tap updates
- Docker image publishing
