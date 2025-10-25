/**
 * Conexus MCP Server - Load Test
 * 
 * Tests system under realistic load with 100+ concurrent users.
 * Validates performance targets: p95 < 1s, p99 < 2s, error rate < 1%
 * 
 * Usage:
 *   k6 run tests/load/load-test.js
 *   k6 run --out json=results/load-test.json tests/load/load-test.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend, Rate } from 'k6/metrics';

// Custom metrics
const searchDuration = new Trend('search_duration', true);
const relatedInfoDuration = new Trend('related_info_duration', true);
const indexControlDuration = new Trend('index_control_duration', true);
const errorRate = new Rate('error_rate');
const toolCallErrors = new Counter('tool_call_errors');

// Test configuration
export const options = {
  stages: [
    // Warm-up: Gradually ramp to target load
    { duration: '2m', target: 50 },   // Ramp to 50 VUs over 2 minutes
    { duration: '1m', target: 100 },  // Ramp to 100 VUs over 1 minute
    
    // Sustained load: Maintain target for observation
    { duration: '5m', target: 100 },  // Hold 100 VUs for 5 minutes
    
    // Peak load: Test higher concurrency
    { duration: '1m', target: 150 },  // Ramp to 150 VUs
    { duration: '3m', target: 150 },  // Hold 150 VUs for 3 minutes
    
    // Cool-down: Gradual ramp-down
    { duration: '1m', target: 50 },   // Ramp down to 50 VUs
    { duration: '1m', target: 0 },    // Ramp down to 0
  ],
  
  thresholds: {
    // HTTP request duration thresholds
    'http_req_duration': [
      'p(95)<1000',    // 95% of requests under 1s
      'p(99)<2000',    // 99% of requests under 2s
    ],
    
    // Tool-specific duration thresholds
    'search_duration': ['p(95)<800'],           // Search should be fast
    'related_info_duration': ['p(95)<1000'],    // Related info can be slower
    'index_control_duration': ['p(95)<500'],    // Control ops should be quick
    
    // Error rate thresholds
    'error_rate': ['rate<0.01'],                // Less than 1% errors
    'http_req_failed': ['rate<0.01'],           // Less than 1% HTTP failures
    
    // Check failure threshold
    'checks': ['rate>0.95'],                    // More than 95% checks pass
  },
  
  // Summary output
  summaryTrendStats: ['min', 'avg', 'med', 'p(95)', 'p(99)', 'max'],
};

// Test data - realistic queries
const searchQueries = [
  'function implementation',
  'error handling',
  'database connection',
  'authentication flow',
  'API endpoint',
  'configuration setup',
  'logging mechanism',
  'test coverage',
  'data validation',
  'security check',
];

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const RESULTS_DIR = __ENV.RESULTS_DIR || 'tests/load/results';

// Helper to create MCP JSON-RPC request
function createMCPRequest(method, params = {}) {
  return JSON.stringify({
    jsonrpc: '2.0',
    id: Math.random().toString(36).substring(7),
    method: method,
    params: params,
  });
}

// Helper to send MCP request
function sendMCPRequest(method, params = {}) {
  const payload = createMCPRequest(method, params);
  const response = http.post(
    `${BASE_URL}/mcp`,
    payload,
    {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: method },
    }
  );
  return response;
}

// Test scenario: Context Search (60% of traffic)
function testContextSearch() {
  const query = searchQueries[Math.floor(Math.random() * searchQueries.length)];
  const startTime = Date.now();
  
  const response = sendMCPRequest('tools/call', {
    name: 'context_search',
    arguments: {
      query: query,
      max_results: 10,
    },
  });
  
  const duration = Date.now() - startTime;
  searchDuration.add(duration);
  
  const success = check(response, {
    'search: status 200': (r) => r.status === 200,
    'search: has result': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.result !== undefined;
      } catch {
        return false;
      }
    },
    'search: valid JSON-RPC': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.jsonrpc === '2.0' && body.id !== undefined;
      } catch {
        return false;
      }
    },
  });
  
  if (!success) {
    errorRate.add(1);
    toolCallErrors.add(1);
  } else {
    errorRate.add(0);
  }
  
  return response;
}

// Test scenario: Get Related Info (20% of traffic)
function testGetRelatedInfo() {
  const query = searchQueries[Math.floor(Math.random() * searchQueries.length)];
  const startTime = Date.now();
  
  const response = sendMCPRequest('tools/call', {
    name: 'context_get_related_info',
    arguments: {
      query: query,
      max_depth: 2,
    },
  });
  
  const duration = Date.now() - startTime;
  relatedInfoDuration.add(duration);
  
  const success = check(response, {
    'related_info: status 200': (r) => r.status === 200,
    'related_info: has result': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.result !== undefined;
      } catch {
        return false;
      }
    },
  });
  
  if (!success) {
    errorRate.add(1);
    toolCallErrors.add(1);
  } else {
    errorRate.add(0);
  }
  
  return response;
}

// Test scenario: Index Control Status (15% of traffic)
function testIndexControlStatus() {
  const startTime = Date.now();
  
  const response = sendMCPRequest('tools/call', {
    name: 'context_index_control',
    arguments: {
      action: 'status',
    },
  });
  
  const duration = Date.now() - startTime;
  indexControlDuration.add(duration);
  
  const success = check(response, {
    'index_control: status 200': (r) => r.status === 200,
    'index_control: has result': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.result !== undefined;
      } catch {
        return false;
      }
    },
  });
  
  if (!success) {
    errorRate.add(1);
    toolCallErrors.add(1);
  } else {
    errorRate.add(0);
  }
  
  return response;
}

// Test scenario: Connector Management List (5% of traffic)
function testConnectorManagementList() {
  const response = sendMCPRequest('tools/call', {
    name: 'context_connector_management',
    arguments: {
      action: 'list',
    },
  });
  
  const success = check(response, {
    'connector_mgmt: status 200': (r) => r.status === 200,
    'connector_mgmt: has result': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.result !== undefined;
      } catch {
        return false;
      }
    },
  });
  
  if (!success) {
    errorRate.add(1);
    toolCallErrors.add(1);
  } else {
    errorRate.add(0);
  }
  
  return response;
}

// Main test function - executes for each VU iteration
export default function () {
  // Weighted random selection of operations (realistic usage pattern)
  const rand = Math.random();
  
  if (rand < 0.60) {
    // 60% - Context Search (most common operation)
    testContextSearch();
  } else if (rand < 0.80) {
    // 20% - Get Related Info
    testGetRelatedInfo();
  } else if (rand < 0.95) {
    // 15% - Index Control Status
    testIndexControlStatus();
  } else {
    // 5% - Connector Management List
    testConnectorManagementList();
  }
  
  // Realistic think time between requests (1-5 seconds)
  // Users don't make requests instantly; they read results, think, then act
  sleep(Math.random() * 4 + 1);
}

// Setup function - runs once before test starts
export function setup() {
  console.log('='.repeat(80));
  console.log('Conexus Load Test Starting');
  console.log('='.repeat(80));
  console.log(`Target URL: ${BASE_URL}`);
  console.log(`Total Duration: ~14 minutes`);
  console.log(`Peak VUs: 150`);
  console.log(`Expected Requests: ~8,000-12,000`);
  console.log('='.repeat(80));
  
  // Verify server is accessible
  const healthCheck = http.get(`${BASE_URL}/health`);
  if (healthCheck.status !== 200) {
    throw new Error(`Server health check failed: ${healthCheck.status}`);
  }
  
  console.log('✓ Server health check passed');
  console.log('');
  
  return { startTime: Date.now() };
}

// Teardown function - runs once after test completes
export function teardown(data) {
  const duration = (Date.now() - data.startTime) / 1000;
  
  console.log('');
  console.log('='.repeat(80));
  console.log('Load Test Complete');
  console.log('='.repeat(80));
  console.log(`Total Duration: ${duration.toFixed(1)}s`);
  console.log('');
  console.log('Results saved to k6 output');
  console.log(`To analyze: k6 run --out json=${RESULTS_DIR}/load-test.json tests/load/load-test.js`);
  console.log('='.repeat(80));
}
