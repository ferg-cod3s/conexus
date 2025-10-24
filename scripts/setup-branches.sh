#!/bin/bash
set -e

echo "🚀 Setting up Conexus branch strategy..."

# Check if we're in a git repo
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Error: Not in a git repository"
    exit 1
fi

# Get current branch
CURRENT_BRANCH=$(git branch --show-current)
echo "📍 Current branch: $CURRENT_BRANCH"

# Ensure we have the latest changes
echo "📥 Fetching latest changes..."
git fetch origin

# Create dev branch if it doesn't exist
if ! git rev-parse --verify origin/dev >/dev/null 2>&1; then
    echo "🌱 Creating dev branch..."
    if [ "$CURRENT_BRANCH" = "main" ]; then
        git checkout -b dev
        git push origin dev
        echo "✅ Dev branch created from main"
    else
        echo "❌ Error: Please switch to main branch first"
        exit 1
    fi
else
    echo "✅ Dev branch already exists"
fi

# Switch back to main if we're not there
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo "🔄 Switching to main branch..."
    git checkout main
fi

# Ensure main is up to date
echo "⬆️ Updating main branch..."
git pull origin main

# Ensure dev is up to date
echo "⬆️ Updating dev branch..."
git fetch origin dev
git checkout dev
git pull origin dev

# Switch back to main
git checkout main

echo ""
echo "✨ Branch setup complete!"
echo ""
echo "📋 Next steps:"
echo "1. Go to your GitHub repository settings"
echo "2. Navigate to Branches > Branch protection rules"
echo "3. Add protection rule for 'main' with:"
echo "   - Require status checks to pass before merging"
echo "   - Require branches to be up to date before merging"
echo "   - Require pull request reviews (1 reviewer)"
echo "   - Require review from CODEOWNERS"
echo "   - Do not allow bypassing the above settings"
echo ""
echo "4. Add protection rule for 'dev' with:"
echo "   - Require status checks to pass before merging"
echo "   - Allow force pushes"
echo "   - No required reviews (for development flexibility)"
echo ""
echo "5. Enable the 'Setup Branch Protection' workflow"
echo "6. Start developing on the dev branch!"
echo ""
echo "🔄 Workflow will automatically sync dev → main daily"