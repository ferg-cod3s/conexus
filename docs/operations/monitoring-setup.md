# Monitoring and Error Tracking Setup

## Sentry MCP Integration

The Conexus project includes a comprehensive Sentry MCP server integration that provides powerful error tracking, performance monitoring, and debugging capabilities.

### Current Configuration

✅ **Sentry MCP Server Status**: Configured and operational
- **Host**: sentry.fergify.work
- **Organization**: unfergettable-designs
- **Projects**: conexus, tunnelforge-go, tunnelforge-web, and 13 others

### Available Tools

The Sentry MCP server provides the following tools for enhanced context and insights:

#### Core Tools (Always Available)
- `whoami` - Verify authentication and user details
- `find_organizations` - List accessible organizations
- `find_teams` - List teams in an organization
- `find_projects` - List projects in an organization
- `find_releases` - Find release information
- `get_issue_details` - Get detailed issue information
- `get_trace_details` - Get trace performance data
- `get_event_attachment` - Download event attachments
- `find_dsns` - Get Sentry DSN keys
- `search_docs` - Search Sentry documentation
- `get_doc` - Fetch full documentation pages

#### AI-Powered Tools (Require OpenAI API Key)
- `search_events` - Natural language event search with aggregations
- `search_issues` - Natural language issue search
- `use_sentry` - General Sentry AI assistant

### Enabling AI-Powered Features

To enable the powerful AI-powered search tools, add your OpenAI API key:

```bash
# Add to your environment or opencode config
export OPENAI_API_KEY="your-openai-api-key-here"
```

Or add it to your OpenCode MCP configuration:

```json
"sentry": {
  "environment": {
    "OPENAI_API_KEY": "your-openai-api-key-here"
  }
}
```

### Usage Examples

#### Basic Error Investigation
```
"Show me unresolved errors in the conexus project from the last 24 hours"
"Get details for issue PROJECT-123"
"Find all database connection errors"
```

#### Performance Analysis
```
"Show me slow API calls in the tunnelforge-web project"
"Get trace details for request abc123def456"
"Find performance bottlenecks in user authentication"
```

#### Release Tracking
```
"What was deployed in the last release?"
"Show me issues introduced in version 1.2.3"
"Find releases with high error rates"
```

### Integration Benefits

1. **Enhanced Debugging**: Direct access to error traces, stack traces, and event data
2. **Performance Insights**: Trace analysis and performance bottleneck identification
3. **Context-Aware Development**: Real-time error monitoring during development
4. **Release Quality**: Pre-deployment error analysis and post-deployment monitoring
5. **Team Collaboration**: Shared error context and investigation workflows

### Best Practices

1. **Use Specific Queries**: Include project names and time ranges for better results
2. **Leverage AI Search**: Use natural language queries for complex investigations
3. **Combine with Conexus**: Use Sentry data alongside codebase context for comprehensive debugging
4. **Monitor Key Metrics**: Set up alerts for critical errors and performance issues
5. **Document Findings**: Use Sentry insights to improve code quality and reliability

### Troubleshooting

**AI Tools Not Working**: Ensure `OPENAI_API_KEY` environment variable is set
**Authentication Issues**: Verify Sentry access token and organization permissions
**No Data Returned**: Check project names and query parameters
**Performance Issues**: Use specific time ranges and project filters

### Related Documentation

- [Security Compliance](../Security-Compliance.md)
- [Development Workflow](../contributing/contributing-guide.md)
- [Error Handling Patterns](../internal/testing/integration/error_handling.go)