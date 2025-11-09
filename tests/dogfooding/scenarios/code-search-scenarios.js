/**
 * Code Search Scenarios for Dogfooding Tests
 * 
 * Defines realistic developer queries for testing Conexus context search
 * against standard file search approaches.
 */

export const codeSearchScenarios = [
  {
    id: 'function_implementations',
    query: 'function implementations for error handling',
    description: 'Find all error handling function implementations',
    expectedTypes: ['function', 'method'],
    expectedFiles: ['*.go', '*.js', '*.ts'],
    context: 'Looking for error handling patterns across the codebase'
  },
  {
    id: 'database_connections',
    query: 'database connection setup and configuration',
    description: 'Find database connection initialization code',
    expectedTypes: ['function', 'struct', 'config'],
    expectedFiles: ['*.go', 'config.*'],
    context: 'Understanding how database connections are established'
  },
  {
    id: 'authentication_flow',
    query: 'authentication and authorization flow',
    description: 'Find authentication logic and flow',
    expectedTypes: ['function', 'middleware', 'handler'],
    expectedFiles: ['*.go', 'security/*'],
    context: 'Reviewing auth implementation for security audit'
  },
  {
    id: 'api_endpoints',
    query: 'API endpoint definitions and handlers',
    description: 'Find all API endpoint implementations',
    expectedTypes: ['handler', 'route', 'endpoint'],
    expectedFiles: ['*.go', 'internal/mcp/*'],
    context: 'Mapping out the API surface for documentation'
  },
  {
    id: 'logging_mechanisms',
    query: 'logging setup and usage patterns',
    description: 'Find logging configuration and usage',
    expectedTypes: ['logger', 'config', 'function'],
    expectedFiles: ['*.go', 'observability/*'],
    context: 'Standardizing logging across the application'
  },
  {
    id: 'test_coverage',
    query: 'test coverage and testing patterns',
    description: 'Find test files and testing approaches',
    expectedTypes: ['test', 'suite', 'fixture'],
    expectedFiles: ['*_test.go', 'tests/*'],
    context: 'Assessing test coverage and quality'
  },
  {
    id: 'data_validation',
    query: 'data validation and sanitization',
    description: 'Find input validation logic',
    expectedTypes: ['validator', 'function', 'middleware'],
    expectedFiles: ['*.go', 'validation/*'],
    context: 'Ensuring data integrity and security'
  },
  {
    id: 'configuration_management',
    query: 'configuration management and loading',
    description: 'Find config loading and management',
    expectedTypes: ['config', 'loader', 'struct'],
    expectedFiles: ['config.*', '*.go'],
    context: 'Understanding configuration architecture'
  }
];
