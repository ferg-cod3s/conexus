/**
 * Context-Aware Scenarios for Dogfooding Tests
 * 
 * Defines queries that leverage work context like active files,
 * git branches, open issues, etc.
 */

export const contextAwareScenarios = [
  {
    id: 'active_file_context',
    query: 'related code for the current file being edited',
    description: 'Find related functions and types when working on handlers.go',
    workContext: {
      activeFiles: ['internal/mcp/handlers.go'],
      recentFiles: ['internal/mcp/schema.go', 'pkg/schema/agent_output_v1.go']
    },
    expectedContext: 'functions called by handlers, related types',
    context: 'Working on MCP handler improvements'
  },
  {
    id: 'branch_context',
    query: 'changes made in the current feature branch',
    description: 'Find code changes related to the current git branch',
    workContext: {
      gitBranch: 'feature/mcp-improvements',
      recentCommits: ['Add new MCP tool', 'Fix handler validation']
    },
    expectedContext: 'modified functions, new endpoints',
    context: 'Reviewing changes before merge'
  },
  {
    id: 'issue_context',
    query: 'code related to open issue about error handling',
    description: 'Find code relevant to a specific open issue',
    workContext: {
      openIssues: ['ERROR-123: Improve error handling in MCP'],
      issueLabels: ['bug', 'error-handling']
    },
    expectedContext: 'error handling functions, validation logic',
    context: 'Working on issue ERROR-123'
  },
  {
    id: 'recent_changes',
    query: 'code affected by recent changes to the indexer',
    description: 'Find code impacted by recent indexer modifications',
    workContext: {
      recentChanges: ['internal/indexer/*.go'],
      timeRange: 'last 3 days'
    },
    expectedContext: 'dependent modules, integration points',
    context: 'Testing indexer changes'
  },
  {
    id: 'team_focus',
    query: 'code owned by the backend team',
    description: 'Find code in areas owned by specific team',
    workContext: {
      teamOwnership: 'backend',
      directories: ['internal/', 'cmd/']
    },
    expectedContext: 'backend services, APIs',
    context: 'Backend team code review'
  },
  {
    id: 'performance_context',
    query: 'performance-critical code paths',
    description: 'Find code that needs performance optimization',
    workContext: {
      performanceFocus: true,
      hotspots: ['search operations', 'database queries']
    },
    expectedContext: 'slow functions, bottlenecks',
    context: 'Performance optimization sprint'
  }
];
