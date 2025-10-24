# Global Setup Instructions

## Files to Move

Please run these commands to complete the global setup:

```bash
# Create global directories
mkdir -p ~/.config/opencode/agent

# Move global configuration
mv ./global-opencode.jsonc ~/.config/opencode/opencode.jsonc

# Move conexus expert agent
mv ./conexus-expert.md ~/.config/opencode/agent/

# Move setup script to a convenient location
mv ./setup-conexus-env.sh ~/bin/conexus-setup 2>/dev/null || mv ./setup-conexus-env.sh ~/.local/bin/conexus-setup 2>/dev/null || echo "Keep setup script in project directories"
```

## Verify Installation

```bash
# Check that files are in place
ls -la ~/.config/opencode/
ls -la ~/.config/opencode/agent/conexus-expert.md
cat ~/.config/opencode/opencode.jsonc
```

## Test Global Configuration

```bash
# Test that OpenCode can see the global config
opencode --help
```

After completing these steps, the global Conexus expert agent will be available across all projects.