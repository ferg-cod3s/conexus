/**
 * Relationship Discovery Scenarios for Dogfooding Tests
 * 
 * Defines queries that test the ability to discover relationships between
 * code elements, files, and concepts.
 */

export const relationshipDiscoveryScenarios = [
  {
    id: 'file_dependencies',
    query: 'dependencies and imports for the MCP handlers',
    description: 'Find all files that depend on or import MCP handler functions',
    expectedRelationships: ['imports', 'calls', 'references'],
    focusFile: 'internal/mcp/handlers.go',
    context: 'Understanding the impact of changes to MCP handlers'
  },
  {
    id: 'cross_references',
    query: 'cross-references for the Agent interface',
    description: 'Find all implementations and usages of the Agent interface',
    expectedRelationships: ['implements', 'uses', 'extends'],
    focusType: 'Agent',
    context: 'Refactoring the Agent interface safely'
  },
  {
    id: 'error_handling_chain',
    query: 'error handling chain from user input to logging',
    description: 'Trace error handling from input validation through to logging',
    expectedRelationships: ['calls', 'handles', 'logs'],
    startPoint: 'validation',
    endPoint: 'logging',
    context: 'Improving error handling consistency'
  },
  {
    id: 'data_flow',
    query: 'data flow from API request to database storage',
    description: 'Trace how data flows from API endpoints to database',
    expectedRelationships: ['receives', 'validates', 'stores'],
    startPoint: 'API handler',
    endPoint: 'database',
    context: 'Optimizing data processing pipeline'
  },
  {
    id: 'configuration_usage',
    query: 'how configuration values are used throughout the system',
    description: 'Find all places where configuration values are accessed',
    expectedRelationships: ['reads', 'uses', 'depends_on'],
    focusType: 'config',
    context: 'Ensuring configuration changes are safe'
  },
  {
    id: 'test_to_code_mapping',
    query: 'mapping between test files and implementation files',
    description: 'Find which tests cover which implementation files',
    expectedRelationships: ['tests', 'covers', 'verifies'],
    focusPattern: '*_test.go',
    context: 'Understanding test coverage gaps'
  }
];
