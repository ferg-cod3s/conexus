/**
 * Cross-Reference Scenarios for Dogfooding Tests
 * 
 * Defines queries that test linking between different domains like
 * tickets, documentation, code, and requirements.
 */

export const crossReferenceScenarios = [
  {
    id: 'ticket_to_code',
    query: 'code changes for ticket PERF-456',
    description: 'Find code that implements or relates to a specific ticket',
    ticketId: 'PERF-456',
    ticketTitle: 'Optimize search performance',
    expectedLinks: ['implements', 'fixes', 'relates_to'],
    context: 'Verifying ticket implementation'
  },
  {
    id: 'requirement_to_implementation',
    query: 'implementation of user authentication requirement',
    description: 'Find code that fulfills a specific requirement',
    requirement: 'User authentication with JWT',
    expectedLinks: ['implements', 'satisfies'],
    context: 'Requirement traceability'
  },
  {
    id: 'documentation_to_code',
    query: 'code examples from the API documentation',
    description: 'Find code that matches documentation examples',
    docReference: 'API-Specification.md',
    expectedLinks: ['documents', 'example_of'],
    context: 'Documentation accuracy check'
  },
  {
    id: 'bug_to_fix',
    query: 'fix for null pointer exception in handlers',
    description: 'Find the code fix for a reported bug',
    bugReport: 'Null pointer in MCP handlers',
    expectedLinks: ['fixes', 'resolves'],
    context: 'Bug fix verification'
  },
  {
    id: 'feature_to_tests',
    query: 'tests for the new MCP search feature',
    description: 'Find tests that cover a new feature',
    feature: 'Enhanced context search',
    expectedLinks: ['tests', 'verifies', 'covers'],
    context: 'Feature testing completeness'
  },
  {
    id: 'design_to_code',
    query: 'code following the federation design pattern',
    description: 'Find code that implements a specific design pattern',
    designPattern: 'Federation pattern',
    expectedLinks: ['implements', 'follows'],
    context: 'Design pattern adherence'
  }
];
