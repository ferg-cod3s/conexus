# Python Project Integration

This guide covers integrating Conexus with Python projects, including Django, Flask, FastAPI, and data science applications.

## Quick Setup

### 1. Install Conexus

```bash
# Clone Conexus repository
git clone https://github.com/ferg-cod3s/conexus.git
cd conexus

# Build binaries
./scripts/build-binaries.sh

# Or install via pip (if available)
pip install conexus-mcp
```

### 2. Configure MCP Client

**For OpenCode** (`.opencode/opencode.jsonc`):

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "conexus": {
      "type": "local",
      "command": ["bunx", "-y", "@agentic-conexus/mcp"],
      "environment": {
        "CONEXUS_DB_PATH": "./.conexus/db.sqlite",
        "CONEXUS_ROOT_PATH": "."
      },
      "enabled": true
    }
  },
  "agent": {
    "python-pro": {
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
    }
  }
}
```

**For Claude Desktop:**

```json
{
  "mcpServers": {
    "conexus": {
      "command": "bunx",
      "args": ["-y", "@agentic-conexus/mcp"],
      "env": {
        "CONEXUS_DB_PATH": "/path/to/project/.conexus/db.sqlite",
        "CONEXUS_ROOT_PATH": "/path/to/project"
      }
    }
  }
}
```

**For Claude Code** (`~/.claude/mcp.json`):

```json
{
  "mcpServers": {
    "conexus": {
      "command": "bunx",
      "args": ["-y", "@agentic-conexus/mcp"],
      "env": {
        "CONEXUS_DB_PATH": "./.conexus/db.sqlite",
        "CONEXUS_ROOT_PATH": "."
      }
    }
  }
}
```

### 3. Project Configuration

Create `.conexus/config.yml`:

```yaml
project:
  name: "my-python-app"
  description: "Python web application"

codebase:
  root: "."
  include_patterns:
    - "**/*.py"
    - "**/*.ipynb"
    - "**/requirements*.txt"
    - "**/pyproject.toml"
    - "**/setup.py"
    - "**/Pipfile"
    - "**/*.md"
  exclude_patterns:
    - "**/__pycache__/**"
    - "**/venv/**"
    - "**/env/**"
    - "**/.env/**"
    - "**/build/**"
    - "**/dist/**"
    - "**/*.pyc"
    - "**/.pytest_cache/**"
    - "**/.coverage/**"
    - "**/htmlcov/**"

indexing:
  auto_reindex: true
  reindex_interval: "45m"
  chunk_size: 400

search:
  max_results: 50
  similarity_threshold: 0.7
```

## Framework-Specific Examples

### Django Application

**Project Structure:**
```
django-project/
├── manage.py
├── myapp/
│   ├── models.py
│   ├── views.py
│   ├── urls.py
│   ├── admin.py
│   └── tests.py
├── config/
│   ├── settings.py
│   ├── urls.py
│   └── wsgi.py
├── .conexus/
└── .opencode/
```

**Configuration:**
```yaml
# .conexus/config.yml
codebase:
  include_patterns:
    - "**/*.py"
    - "**/manage.py"
    - "**/requirements*.txt"
    - "**/pyproject.toml"
  exclude_patterns:
    - "**/__pycache__/**"
    - "**/migrations/**"
    - "**/static/**"
    - "**/media/**"
```

**Recommended Agents:**
- `django-pro`
- `python-pro`
- `api-builder-enhanced`

**Useful Queries:**
- "Find all Django models and their fields"
- "Show me the URL patterns"
- "Search for view functions"
- "Locate the settings configuration"

### Flask Application

**Setup:**
```python
# app.py
from flask import Flask
app = Flask(__name__)

@app.route('/')
def hello():
    return 'Hello World'
```

**Configuration:**
```yaml
codebase:
  include_patterns:
    - "**/*.py"
    - "**/requirements.txt"
    - "**/Pipfile"
  exclude_patterns:
    - "**/__pycache__/**"
    - "**/instance/**"
    - "**/.env/**"
```

**Queries:**
- "Find all Flask routes"
- "Show me the application factory pattern"
- "Search for blueprint definitions"
- "Locate error handlers"

### FastAPI Application

**Project Structure:**
```
fastapi-app/
├── main.py
├── models/
├── routers/
├── dependencies.py
├── .conexus/
└── requirements.txt
```

**Configuration:**
```yaml
codebase:
  include_patterns:
    - "**/*.py"
    - "**/requirements.txt"
    - "**/pyproject.toml"
  exclude_patterns:
    - "**/__pycache__/**"
    - "**/.env/**"
```

**Recommended Agents:**
- `fastapi-pro`
- `python-pro`

**Queries:**
- "Find all API endpoints"
- "Show me the Pydantic models"
- "Search for dependency injection"
- "Locate the route definitions"

### Data Science / ML Projects

**Configuration:**
```yaml
# .conexus/config.yml
codebase:
  include_patterns:
    - "**/*.py"
    - "**/*.ipynb"
    - "**/requirements*.txt"
    - "**/pyproject.toml"
    - "**/*.md"
    - "**/*.yaml"
    - "**/*.yml"
  exclude_patterns:
    - "**/__pycache__/**"
    - "**/venv/**"
    - "**/.env/**"
    - "**/data/**"
    - "**/models/**"
    - "**/checkpoints/**"
```

**Recommended Agents:**
- `data-scientist`
- `python-pro`

**Queries:**
- "Find all machine learning models"
- "Show me the data preprocessing code"
- "Search for training functions"
- "Locate evaluation metrics"

## Development Workflow

### Virtual Environment Setup

```bash
# Create virtual environment
python -m venv venv
source venv/bin/activate  # Linux/macOS
# or
venv\Scripts\activate     # Windows

# Install dependencies
pip install -r requirements.txt

# Install Conexus (if available via pip)
pip install conexus-mcp
```

### Pyproject.toml Integration

```toml
# pyproject.toml
[tool.conexus]
db_path = ".conexus/db.sqlite"
root_path = "."
auto_reindex = true

[tool.conexus.indexing]
chunk_size = 400
reindex_interval = "45m"

[tool.conexus.search]
max_results = 50
similarity_threshold = 0.7
```

### Pre-commit Hooks

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: conexus-index
        name: Index codebase with Conexus
        entry: ../conexus/bin/conexus-darwin-arm64
        args: [index, --quiet]
        language: system
        pass_filenames: false

  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.4.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
```

### VS Code Integration

```json
// .vscode/settings.json
{
  "python.defaultInterpreterPath": "./venv/bin/python",
  "mcp.server.conexus": {
    "command": "npx",
    "args": ["-y", "@agentic-conexus/mcp"],
    "env": {
      "CONEXUS_DB_PATH": "${workspaceFolder}/.conexus/db.sqlite",
      "CONEXUS_ROOT_PATH": "${workspaceFolder}"
    }
  }
}
```

## Testing Integration

### Pytest Configuration

```ini
# pytest.ini
[tool:pytest]
testpaths = tests
python_files = test_*.py *_test.py
python_classes = Test*
python_functions = test_*
addopts = -v --tb=short --cov=src --cov-report=html
```

**Conexus can help with:**
- "Find all test functions"
- "Show me the test fixtures"
- "Search for mocked dependencies"
- "Locate integration tests"

### Coverage Integration

```ini
# .coveragerc
[run]
source = src
omit =
    */tests/*
    */venv/*
    */__pycache__/*

[report]
exclude_lines =
    pragma: no cover
    def __repr__
    raise AssertionError
    raise NotImplementedError
```

## Performance Optimization

### For Large Python Codebases

```yaml
# .conexus/config.yml
indexing:
  chunk_size: 300
  workers: 2
  memory_limit: "512MB"

search:
  max_results: 30
  cache_enabled: true
  cache_ttl: "2h"

codebase:
  exclude_patterns:
    - "**/large_data_files/**"
    - "**/cache/**"
    - "**/logs/**"
```

### Memory Management

```bash
# Environment variables
export CONEXUS_VECTORSTORE_MEMORY_LIMIT=512MB
export CONEXUS_INDEXING_MEMORY_LIMIT=256MB

# Python memory settings
export PYTHONPATH="${PYTHONPATH}:."
export PYTHONDONTWRITEBYTECODE=1
```

## Troubleshooting

### Common Python Issues

**Import errors:**
```bash
# Check Python path
python -c "import sys; print(sys.path)"

# Install missing dependencies
pip install -r requirements.txt

# Check virtual environment
which python
python --version
```

**Module not found:**
```bash
# Ensure you're in the right directory
pwd
ls -la

# Check if package is installed
pip list | grep package_name

# Reinstall package
pip uninstall package_name
pip install package_name
```

**Conexus indexing issues:**
```bash
# Check Python files are included
find . -name "*.py" | head -10

# Test Conexus directly
../conexus/bin/conexus-darwin-arm64 index --dry-run
```

### Framework-Specific Issues

**Django:**
- Ensure `migrations/` are excluded
- Include `manage.py` and settings files
- Add Django apps to include patterns

**Flask:**
- Include all Python files with routes
- Exclude instance directories
- Check for application factory patterns

**FastAPI:**
- Include router files and models
- Exclude generated files
- Check for async/await patterns

## Best Practices

1. **Virtual Environments:** Always use virtual environments and exclude them from indexing

2. **Dependency Management:** Include `requirements.txt`, `pyproject.toml`, and `Pipfile` for better understanding

3. **Test Coverage:** Include test files and coverage reports for comprehensive code understanding

4. **Documentation:** Include `README.md`, docstrings, and documentation files

5. **Configuration Files:** Include configuration files like `settings.py`, `.env.example`

## Integration Examples

### With Black and isort

```toml
# pyproject.toml
[tool.black]
line-length = 88
target-version = ['py38']

[tool.isort]
profile = "black"
multi_line_output = 3
```

### With MyPy

```ini
# mypy.ini
[mypy]
python_version = 3.8
warn_return_any = True
warn_unused_configs = True
disallow_untyped_defs = True
disallow_incomplete_defs = True
check_untyped_defs = True
disallow_untyped_decorators = True
no_implicit_optional = True
warn_redundant_casts = True
warn_unused_ignores = True
warn_no_return = True
warn_unreachable = True
strict_equality = True
```

### With Poetry

```toml
# pyproject.toml
[tool.poetry]
name = "my-python-app"
version = "0.1.0"
description = "Python application"

[tool.poetry.dependencies]
python = "^3.8"
fastapi = "^0.104.1"

[tool.poetry.group.dev.dependencies]
pytest = "^7.4.3"
black = "^23.11.0"
isort = "^5.12.0"
mypy = "^1.7.1"

[build-system]
requires = ["poetry-core"]
build-backend = "poetry.core.masonry.api"
```

This integration allows Conexus to understand Python codebases, frameworks, dependencies, and development patterns for enhanced AI assistance.