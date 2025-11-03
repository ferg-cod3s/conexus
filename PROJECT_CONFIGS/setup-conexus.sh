#!/bin/bash

# Conexus Universal Setup Script
# Automatically configures conexus for any project based on detected project type
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Function to detect project type
detect_project_type() {
    local project_name=""
    local project_type="generic"

    # Try to get project name from various sources
    if [ -f "package.json" ]; then
        project_name=$(jq -r '.name // empty' package.json 2>/dev/null || echo "")
        if [ -n "$project_name" ]; then
            project_type="nodejs"
        fi
    elif [ -f "go.mod" ]; then
        project_name=$(grep "^module " go.mod | head -1 | sed 's/module //' | xargs basename 2>/dev/null || echo "")
        project_type="go"
    elif [ -f "pyproject.toml" ]; then
        project_name=$(grep "^name = " pyproject.toml | head -1 | sed 's/name = //' | tr -d '"' | xargs 2>/dev/null || echo "")
        project_type="python"
    elif [ -f "Cargo.toml" ]; then
        project_name=$(grep "^name = " Cargo.toml | head -1 | sed 's/name = //' | tr -d '"' | xargs 2>/dev/null || echo "")
        project_type="rust"
    elif [ -f "composer.json" ]; then
        project_name=$(jq -r '.name // empty' composer.json 2>/dev/null || echo "")
        project_type="php"
    elif [ -f "Gemfile" ]; then
        project_name=$(basename "$(pwd)" | tr '_' '-')
        project_type="ruby"
    elif [ -f "pom.xml" ] || [ -f "build.gradle" ] || [ -f "build.gradle.kts" ]; then
        project_name=$(basename "$(pwd)")
        project_type="java"
    elif [ -f "Dockerfile" ] || [ -d "docker" ]; then
        project_name=$(basename "$(pwd)")
        project_type="docker"
    else
        project_name=$(basename "$(pwd)")
    fi

    # Clean up project name
    project_name=$(echo "$project_name" | sed 's/[^a-zA-Z0-9_-]/-/g' | sed 's/^-*//' | sed 's/-*$//')

    if [ -z "$project_name" ]; then
        project_name="unknown-project"
    fi

    echo "$project_name:$project_type"
}

# Function to get agent configuration based on project type
get_agent_config() {
    local project_type="$1"

    case "$project_type" in
        "nodejs")
            cat << 'EOF'
    "typescript-pro": {
      "tools": {
        "conexus": true
      }
    },
    "javascript-pro": {
      "tools": {
        "conexus": true
      }
    },
    "frontend-developer": {
      "tools": {
        "conexus": true
      }
    },
    "api-builder-enhanced": {
      "tools": {
        "conexus": true
      }
    },
    "astro-pro": {
      "tools": {
        "conexus": true
      }
    },
    "nextjs-pro": {
      "tools": {
        "conexus": true
      }
    },
    "react-pro": {
      "tools": {
        "conexus": true
      }
    },
    "vue-pro": {
      "tools": {
        "conexus": true
      }
    },
    "node-js-developer": {
      "tools": {
        "conexus": true
      }
    },
EOF
            ;;
        "go")
            cat << 'EOF'
    "go-expert": {
      "tools": {
        "conexus": true
      }
    },
    "golang-developer": {
      "tools": {
        "conexus": true
      }
    },
    "api-builder-enhanced": {
      "tools": {
        "conexus": true
      }
    },
    "backend-developer": {
      "tools": {
        "conexus": true
      }
    },
EOF
            ;;
        "python")
            cat << 'EOF'
    "python-pro": {
      "tools": {
        "conexus": true
      }
    },
    "python-developer": {
      "tools": {
        "conexus": true
      }
    },
    "api-builder-enhanced": {
      "tools": {
        "conexus": true
      }
    },
    "data-scientist": {
      "tools": {
        "conexus": true
      }
    },
    "backend-developer": {
      "tools": {
        "conexus": true
      }
    },
EOF
            ;;
        "rust")
            cat << 'EOF'
    "rust-pro": {
      "tools": {
        "conexus": true
      }
    },
    "rust-developer": {
      "tools": {
        "conexus": true
      }
    },
    "systems-programmer": {
      "tools": {
        "conexus": true
      }
    },
    "backend-developer": {
      "tools": {
        "conexus": true
      }
    },
EOF
            ;;
        "php")
            cat << 'EOF'
    "php-pro": {
      "tools": {
        "conexus": true
      }
    },
    "php-developer": {
      "tools": {
        "conexus": true
      }
    },
    "laravel-pro": {
      "tools": {
        "conexus": true
      }
    },
    "backend-developer": {
      "tools": {
        "conexus": true
      }
    },
EOF
            ;;
        "ruby")
            cat << 'EOF'
    "ruby-pro": {
      "tools": {
        "conexus": true
      }
    },
    "ruby-on-rails-pro": {
      "tools": {
        "conexus": true
      }
    },
    "ruby-developer": {
      "tools": {
        "conexus": true
      }
    },
    "backend-developer": {
      "tools": {
        "conexus": true
      }
    },
EOF
            ;;
        "java")
            cat << 'EOF'
    "java-pro": {
      "tools": {
        "conexus": true
      }
    },
    "spring-boot-pro": {
      "tools": {
        "conexus": true
      }
    },
    "java-developer": {
      "tools": {
        "conexus": true
      }
    },
    "backend-developer": {
      "tools": {
        "conexus": true
      }
    },
EOF
            ;;
        "docker")
            cat << 'EOF'
    "devops-engineer": {
      "tools": {
        "conexus": true
      }
    },
    "infrastructure-developer": {
      "tools": {
        "conexus": true
      }
    },
    "container-specialist": {
      "tools": {
        "conexus": true
      }
    },
EOF
            ;;
        *)
            cat << 'EOF'
    "full-stack-developer": {
      "tools": {
        "conexus": true
      }
    },
    "software-engineer": {
      "tools": {
        "conexus": true
      }
    },
EOF
            ;;
    esac
}

# Function to create opencode.jsonc configuration
create_opencode_config() {
    local project_name="$1"
    local project_type="$2"

    print_info "Creating .opencode/opencode.jsonc configuration..."

    cat > .opencode/opencode.jsonc << EOF
{
  "\$schema": "https://opencode.ai/config.json",
  "mcp": {
    "conexus": {
      "type": "local",
      "command": ["../bin/conexus-darwin-arm64"],
      "enabled": true,
      "environment": {
        "CONEXUS_PORT": "0",
        "CONEXUS_DB_PATH": "{env:CONEXUS_DB_PATH:./data/conexus.db}",
        "CONEXUS_LOG_LEVEL": "info",
        "CONEXUS_SENTRY_ENABLED": "true",
        "CONEXUS_SENTRY_DSN": "https://7e54c8bc81fb554a460d4331e5c23fe0@sentry.fergify.work/15",
        "CONEXUS_SENTRY_ENVIRONMENT": "development",
        "CONEXUS_SENTRY_SAMPLE_RATE": "1.0",
        "CONEXUS_SENTRY_RELEASE": "0.1.0-alpha"
      }
    }
  },
  "agent": {
$(get_agent_config "$project_type")
    "conexus-expert": {
      "tools": {
        "conexus": true
      }
    }
  }
}
EOF

    print_status "Configuration file created"
}

# Function to create environment file
create_env_file() {
    local project_name="$1"

    print_info "Creating .env file..."

    cat > .env << EOF
# $project_name Environment Configuration
# Conexus Database and Settings
CONEXUS_DB_PATH=./data/conexus.db
CONEXUS_PORT=0
CONEXUS_LOG_LEVEL=info

# Sentry Configuration (optional - remove if not needed)
# CONEXUS_SENTRY_ENABLED=true
# CONEXUS_SENTRY_DSN=https://7e54c8bc81fb554a460d4331e5c23fe0@sentry.fergify.work/15
# CONEXUS_SENTRY_ENVIRONMENT=development
# CONEXUS_SENTRY_SAMPLE_RATE=1.0
# CONEXUS_SENTRY_RELEASE=0.1.0-alpha

# Additional environment variables can be added here
# CONEXUS_ROOT_PATH=./
# CONEXUS_EMBEDDING_PROVIDER=mock
# CONEXUS_EMBEDDING_MODEL=mock-768
# CONEXUS_EMBEDDING_DIMENSIONS=768
EOF

    print_status "Environment file created"
}

# Function to check if conexus binary exists
check_binary() {
    local binary_path="../bin/conexus-darwin-arm64"

    if [ ! -f "$binary_path" ]; then
        print_error "Conexus binary not found at $binary_path"
        print_error "Please ensure conexus is built and available in the bin directory"
        exit 1
    fi

    if [ ! -x "$binary_path" ]; then
        print_warning "Conexus binary is not executable, making it executable..."
        chmod +x "$binary_path"
        print_status "Binary made executable"
    fi

    print_status "Conexus binary found and ready"
}

# Main setup function
main() {
    print_info "Conexus Universal Setup Script"
    print_info "=============================="

    # Detect project type
    local detection_result=$(detect_project_type)
    local project_name=$(echo "$detection_result" | cut -d: -f1)
    local project_type=$(echo "$detection_result" | cut -d: -f2)

    print_info "Detected project: $project_name"
    print_info "Project type: $project_type"

    # Check if conexus binary exists
    check_binary

    # Create .opencode directory
    print_info "Creating .opencode directory..."
    mkdir -p .opencode
    print_status "Directory created"

    # Create data directory
    print_info "Creating data directory..."
    mkdir -p data
    print_status "Directory created"

    # Create configuration file
    create_opencode_config "$project_name" "$project_type"

    # Create environment file
    create_env_file "$project_name"

    # Final instructions
    echo ""
    print_status "Conexus setup completed successfully!"
    echo ""
    print_info "Next steps:"
    echo "  1. Review and customize .opencode/opencode.jsonc if needed"
    echo "  2. Run 'source .env' to load environment variables"
    echo "  3. Start using conexus with your preferred MCP-compatible client"
    echo ""
    print_info "Configuration details:"
    echo "  - Config: .opencode/opencode.jsonc"
    echo "  - Database: ./data/conexus.db"
    echo "  - Binary: ../bin/conexus-darwin-arm64"
    echo "  - Environment: .env"
}

# Run main function
main "$@"