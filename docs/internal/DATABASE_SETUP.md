# Database Setup for Conexus

## Overview

Conexus uses SQLite as its primary database for storing vector embeddings and connector configurations. The database is designed to be **project-specific** and automatically generated on first installation.

## Database Files

- **Location**: `./data/conexus.db` (configurable via `CONEXUS_DB_PATH` environment variable)
- **Format**: SQLite with FTS5 full-text search support
- **Auto-creation**: Database and tables are created automatically on first run
- **Git Ignored**: Database files are excluded from version control

## Automatic Initialization

When Conexus starts, it automatically:

1. **Creates the data directory** if it doesn't exist
2. **Initializes the SQLite database** with required schema
3. **Sets up FTS5 virtual tables** for hybrid search
4. **Creates indexes** for optimal performance

### Schema Components

- **Documents table**: Stores text chunks with metadata
- **FTS5 virtual table**: Enables full-text search (BM25)
- **Vector embeddings**: Stored as BLOB for similarity search
- **Connector configurations**: External data source settings

## Configuration

### Default Configuration
```yaml
database:
  path: "./data/conexus.db"
```

### Environment Variables
- `CONEXUS_DB_PATH`: Override database file location
- `CONEXUS_CONFIG_FILE`: Path to configuration file

## Project-Specific Benefits

1. **Isolation**: Each project has its own database
2. **No Conflicts**: Different projects can't interfere with each other
3. **Portability**: Database travels with the project
4. **Privacy**: Sensitive project data stays local
5. **Performance**: Optimized for specific project size

## Installation Process

1. **Clone or install Conexus** in your project directory
2. **Run Conexus** - it automatically creates `./data/conexus.db`
3. **Start indexing** your project files
4. **Database grows** with your project data

## File Structure

```
your-project/
├── .gitignore          # Automatically ignores *.db files
├── conexus             # Conexus binary
├── data/               # Auto-created on first run
│   └── conexus.db      # Project database
└── config.yml          # Optional configuration
```

## Best Practices

1. **Never commit database files** - they're automatically ignored
2. **Use relative paths** for portability across machines
3. **Backup the data directory** if you need to preserve the database
4. **Consider memory mode** (`:memory:`) for temporary/testing usage

## Database Management

### Reset Database
```bash
rm ./data/conexus.db
# Conexus will recreate it on next run
```

### Backup Database
```bash
cp ./data/conexus.db ./data/conexus.db.backup
```

### Move Database Location
```bash
export CONEXUS_DB_PATH="./custom/location/my.db"
```

## Technical Details

### Connection Management
- **Max Connections**: 1 for file-based databases (prevents locking issues)
- **Connection Pooling**: Disabled for consistency
- **Transaction Safety**: ACID compliant

### Performance Features
- **FTS5**: Full-text search with BM25 scoring
- **Vector Storage**: Efficient BLOB storage for embeddings
- **Hybrid Search**: Combines semantic and keyword search
- **Incremental Updates**: Only reindex changed files

### Security
- **Path Validation**: Prevents directory traversal attacks
- **File Permissions**: Database created with 0755 directory permissions
- **Local Storage**: No external database dependencies

## Troubleshooting

### Database Locked
```bash
# Ensure only one Conexus instance is running
pkill conexus
```

### Permission Denied
```bash
# Check directory permissions
ls -la ./data/
chmod 755 ./data/
```

### Corrupted Database
```bash
# Remove and recreate
rm ./data/conexus.db
# Conexus will recreate on next run
```

## Integration with Development Workflow

The project-specific database design integrates seamlessly with typical development workflows:

1. **Git Operations**: Database files are ignored, no conflicts
2. **CI/CD**: Fresh database created in each environment
3. **Docker**: Volume mount the `./data` directory
4. **Team Collaboration**: Each developer has their own database

This approach ensures that Conexus provides a zero-configuration experience while maintaining data isolation and portability across different projects and environments.