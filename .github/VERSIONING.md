# Automatic Versioning

This repository uses automatic semantic versioning for releases. Every merge to the `main` branch triggers an automatic version bump and release.

## How it works

The repository includes two GitHub Actions workflows for automatic versioning:

### 1. Custom Auto-Tag Workflow (`auto-tag.yml`)
- Analyzes commit messages to determine version bump type
- Creates semantic version tags automatically
- Generates GitHub releases

### 2. Semantic Release Workflow (`semantic-release.yml`) - **RECOMMENDED**
- Uses the popular `mathieudutour/github-tag-action`
- Simpler and more reliable
- Better changelog generation

## Commit Message Conventions

To control version bumping, use these commit message patterns:

### Patch Version (v1.0.0 → v1.0.1)
```bash
git commit -m "fix: resolve issue with logger initialization"
git commit -m "docs: update README"
git commit -m "chore: update dependencies"
```

### Minor Version (v1.0.0 → v1.1.0)
```bash
git commit -m "feat: add new telemetry configuration option"
git commit -m "feature: implement metrics collection"
```

### Major Version (v1.0.0 → v2.0.0)
```bash
git commit -m "feat!: change telemetry API (BREAKING CHANGE)"
git commit -m "BREAKING CHANGE: remove deprecated logger methods"
```

## Skipping Releases

To skip automatic versioning for a commit:
```bash
git commit -m "docs: minor typo fix [skip ci]"
```

## Manual Tagging

If you need to create a manual tag:
```bash
git tag v1.2.3
git push origin v1.2.3
```

## Current Setup

- ✅ Auto-tagging on main branch merges
- ✅ Semantic version bumping based on commit messages  
- ✅ Automatic GitHub releases
- ✅ Changelog generation
- ✅ Support for pre-release branches (develop, beta)

Choose one of the two workflows based on your preference - the semantic-release workflow is generally recommended for its simplicity and reliability.