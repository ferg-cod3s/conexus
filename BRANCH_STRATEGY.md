# Branch Strategy and Workflow

## Branch Structure

### Main Branch (`main`)
- **Purpose**: Production-ready code
- **Protection**: Fully protected with required reviews and status checks
- **Merges**: Only from `dev` branch or feature branches via PR
- **Releases**: Auto-tagged and released on merge

### Development Branch (`dev`)
- **Purpose**: Integration branch for development
- **Protection**: Semi-protected with status checks but no required reviews
- **Merges**: Direct pushes allowed, PRs from feature branches
- **Sync**: Automatically kept up-to-date with main via daily sync

### Feature Branches (`feature/*`, `bugfix/*`, `hotfix/*`)
- **Purpose**: Individual features and fixes
- **Creation**: Branch from `dev`
- **Merges**: PR to `dev` (or directly to `main` for hotfixes)

## Workflow

### Development Flow
1. Create feature branch from `dev`
   ```bash
   git checkout dev
   git pull origin dev
   git checkout -b feature/amazing-feature
   ```

2. Make changes and commit
   ```bash
   git add .
   git commit -m "feat: add amazing feature"
   git push origin feature/amazing-feature
   ```

3. Create PR to `dev` branch
   - CI/CD runs tests, linting, security scans
   - No approval required for `dev` branch
   - Merge when all checks pass

### Release Flow
1. `dev` branch is automatically synced to `main` daily
2. On merge to `main`:
   - All tests and checks run
   - Auto-tagging creates new version
   - Release is created with binaries
   - GitHub release is published

### Hotfix Flow
For critical fixes:
1. Create `hotfix/*` branch from `main`
2. Make the fix
3. Create PR directly to `main`
4. After merge, cherry-pick to `dev`

## Branch Protection Rules

### Main Branch
- ✅ Required status checks: Test, Lint, Security Scan, Build
- ✅ Required approving reviews: 1
- ✅ Require code owner reviews
- ✅ Require review thread resolution
- ✅ Enforce admins
- ❌ No force pushes
- ❌ No deletions

### Dev Branch
- ✅ Required status checks: Test, Lint, Security Scan, Build
- ❌ No required reviews (development-friendly)
- ✅ Allow force pushes (for rebase/force-push)
- ❌ No deletions

## Automated Workflows

### CI/CD Pipeline
- **Triggers**: Push to `main`/`dev`, PRs
- **Tests**: Unit tests, integration tests, race detection
- **Quality**: Linting, security scanning, coverage checks
- **Build**: Cross-platform binary compilation

### Release Automation
- **Auto-tagging**: Increments patch version on main merge
- **Release Creation**: Automatic GitHub release with binaries
- **Artifact Storage**: Binaries uploaded as release assets

### Dev Sync
- **Daily Sync**: Automatically merges `dev` into `main` if ahead
- **Conflict Handling**: Creates PR if merge conflicts occur
- **Manual Trigger**: Can be triggered via workflow dispatch

## Setup Instructions

### Initial Setup
1. Create `dev` branch if it doesn't exist:
   ```bash
   git checkout -b dev
   git push origin dev
   ```

2. Enable branch protection:
   - Go to Settings > Branches
   - Add protection rules for `main` and `dev`
   - Or run the `Setup Branch Protection` workflow

3. Configure secrets (if needed):
   - `GITHUB_TOKEN`: Automatically provided by GitHub Actions

### Required Permissions
- Repository admin access for branch protection setup
- Write access for workflows to run
- Maintainer or admin role for releases

## Best Practices

### Development
- Always branch from `dev` for new features
- Keep PRs focused and small
- Write descriptive commit messages following conventional commits
- Ensure all tests pass before pushing

### Releases
- Let automation handle versioning
- Monitor release workflow runs
- Test release artifacts locally when possible

### Branch Management
- Don't force push to `main`
- Keep `dev` reasonably up-to-date with `main`
- Delete merged feature branches
- Use hotfix branch for emergency fixes to `main`

## Troubleshooting

### Sync Conflicts
If `dev` cannot be auto-merged to `main`:
1. Workflow creates a PR labeled `sync-conflict`
2. Resolve conflicts manually
3. Merge the PR
4. Delete the sync branch

### Failed Checks
- Check workflow logs for details
- Fix issues locally
- Push fixes to the same branch
- Workflow will automatically re-run

### Release Issues
- Verify version tagging logic
- Check binary build process
- Ensure release permissions are correct
- Manual release can be created if needed