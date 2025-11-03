/**
 * Conexus MCP Server - Soak Test
 * 
 * Tests system stability under sustained moderate load over extended period.
 * Detects memory leaks, connection leaks, and performance degradation.
 * 
 * Usage:
 *   k6 run tests/load/soak-test.js
 *   k6 run --duration 30m tests/load/soak-test.js  # Override duration
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend, Rate, Gauge } from 'k6/metrics';

// Custom metrics for tracking trends over time
const requestDuration = new Trend('request_duration', true);
const searchDuration = new Trend('search_duration', true);
const errorRate = new Rate('error_rate');
const activeConnections = new Gauge('active_connections');
const memoryWarnings = new Counter('memory_warnings');

// Test configuration - sustained moderate load
export const options = {
  stages: [
    // Warm-up: Gradual ramp to target load
    { duration: '3m', target: 75 },    // Ramp to 75 VUs over 3 minutes
    
    // Sustained load: Hold steady for observation
    { duration: '20m', target: 75 },   // Hold 75 VUs for 20 minutes
    
    // Cool-down: Observe shutdown behavior
    { duration: '2m', target: 0 },     // Ramp down to 0
  ],
  
  thresholds: {
    // Response time should remain consistent throughout test
    'http_req_duration': [
      'p(95)<1000',    // 95% under 1s
      'p(99)<2000',    // 99% under 2s
    ],
    
    // Error rate should stay low throughout
    'error_rate': ['rate<0.01'],       // Less than 1% errors
    'http_req_failed': ['rate<0.01'],  // Less than 1% HTTP failures
    
    // Checks should consistently pass
    'checks': ['rate>0.95'],           // More than 95% checks pass
    
    // Search performance should not degrade
    'search_duration': ['p(95)<800'],  // Searches under 800ms
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const SOAK_DURATION = __ENV.SOAK_DURATION || '20m';

// Diverse test queries to simulate real usage patterns
const testQueries = [
  'function implementation',
  'error handling patterns',
  'database connection pool',
  'authentication middleware',
  'API rate limiting',
  'logging configuration',
  'test helper functions',
  'data validation rules',
  'security best practices',
  'performance optimization',
  'concurrent request handling',
  'cache invalidation',
  'session management',
  'configuration loading',
  'dependency injection',
];

// Helper to create MCP JSON-RPC request
function createMCPRequest(method, params = {}) {
  return JSON.stringify({
    jsonrpc: '2.0',
    id: Math.random().toString(36).substring(7),
    method: method,
    params: params,
  });
}

// Test scenario: Context Search
function testContextSearch() {
  const query = testQueries[Math.floor(Math.random() * testQueries.length)];
  const startTime = Date.now();

  const payload = createMCPRequest('tools/call', {
    name: 'context.search',
    arguments: {
      query: query,
      max_results: 10,
    },
  });
  
  const response = http.post(
    `${BASE_URL}/mcp`,
    payload,
    {
      headers: { 'Content-Type': 'application/json' },
       tags: { name: 'context.search' },
    }
  );
  
  const duration = Date.now() - startTime;
  requestDuration.add(duration);
  searchDuration.add(duration);
  
  const success = check(response, {
    'search: status 200': (r) => r.status === 200,
    'search: valid response': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.jsonrpc === '2.0' && body.result !== undefined;
      } catch {
        return false;
      }
    },
    'search: reasonable duration': () => duration < 2000,
  });
  
  errorRate.add(success ? 0 : 1);
  
  return response;
}

// Test scenario: Get Related Info
function testGetRelatedInfo() {
  const query = testQueries[Math.floor(Math.random() * testQueries.length)];
  const startTime = Date.now();

  const payload = createMCPRequest('tools/call', {
    name: 'context.get_related_info',
    arguments: {
      query: query,
      max_depth: 2,
    },
  });

  const response = http.post(
    `${BASE_URL}/mcp`,
    payload,
    {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'context.get_related_info' },
    }
  );
  
  const duration = Date.now() - startTime;
  requestDuration.add(duration);
  
  const success = check(response, {
    'related_info: status 200': (r) => r.status === 200,
    'related_info: valid response': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.jsonrpc === '2.0' && body.result !== undefined;
      } catch {
        return false;
      }
    },
  });
  
  errorRate.add(success ? 0 : 1);
  
  return response;
}

// Test scenario: Index Control Status
function testIndexControlStatus() {
   const payload = createMCPRequest('tools/call', {
     name: 'context.index_control',
     arguments: {
       action: 'status',
     },
   });
  
  const response = http.post(
    `${BASE_URL}/mcp`,
    payload,
    {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'context.index_control' },
    }
  );
  
  const success = check(response, {
    'index_control: status 200': (r) => r.status === 200,
    'index_control: valid response': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.jsonrpc === '2.0' && body.result !== undefined;
      } catch {
        return false;
      }
    },
  });
  
  errorRate.add(success ? 0 : 1);
  
  return response;
}

// Periodically check system health
function checkSystemHealth() {
  const response = http.get(`${BASE_URL}/health`);
  
  check(response, {
    'health: status 200': (r) => r.status === 200,
    'health: responsive': (r) => r.timings.duration < 500,
  });
  
  // Check metrics endpoint for memory issues
  const metrics = http.get(`${BASE_URL}:9090/metrics`);
  if (metrics.status === 200) {
    // Look for signs of memory issues (this is a simple heuristic)
    if (metrics.body.includes('go_memstats_alloc_bytes') && 
        metrics.body.length > 1000000) {  // Arbitrary threshold
      memoryWarnings.add(1);
    }
  }
}

// Main test function
export default function () {
  const rand = Math.random();
  
  // Realistic usage pattern
  if (rand < 0.60) {
    // 60% - Context Search
    testContextSearch();
  } else if (rand < 0.85) {
    // 25% - Get Related Info
    testGetRelatedInfo();
  } else {
    // 15% - Index Control Status
    testIndexControlStatus();
  }
  
  // Periodically check system health (every ~20 requests)
  if (Math.random() < 0.05) {
    checkSystemHealth();
  }
  
  // Realistic think time
  sleep(Math.random() * 4 + 1);  // 1-5 seconds
}

// Setup function
export function setup() {
  console.log('='.repeat(80));
  console.log('Conexus Soak Test Starting');
  console.log('='.repeat(80));
  console.log(`Target URL: ${BASE_URL}`);
  console.log(`Duration: ${SOAK_DURATION}`);
  console.log(`Sustained VUs: 75`);
  console.log(`Goal: Detect memory leaks, connection leaks, degradation`);
  console.log('='.repeat(80));
  
  // Verify server is accessible
  const healthCheck = http.get(`${BASE_URL}/health`);
  if (healthCheck.status !== 200) {
    throw new Error(`Server health check failed: ${healthCheck.status}`);
  }
  
  console.log('✓ Server health check passed');
  console.log('');
  console.log('MONITORING TIPS:');
  console.log('  - Watch for gradual response time increases');
  console.log('  - Monitor memory usage over time');
  console.log('  - Check for connection pool exhaustion');
  console.log('  - Look for goroutine leaks in Go runtime metrics');
  console.log('  - Observe Grafana dashboards for trends');
  console.log('='.repeat(80));
  console.log('');
  
  return { startTime: Date.now() };
}

// Teardown function
export function teardown(data) {
  const duration = (Date.now() - data.startTime) / 1000;
  const minutes = Math.floor(duration / 60);
  const seconds = Math.floor(duration % 60);
  
  console.log('');
  console.log('='.repeat(80));
  console.log('Soak Test Complete');
  console.log('='.repeat(80));
  console.log(`Total Duration: ${minutes}m ${seconds}s`);
  console.log('');
  console.log('ANALYSIS CHECKLIST:');
  console.log('  □ Did response times remain consistent?');
  console.log('  □ Did memory usage stabilize or grow continuously?');
  console.log('  □ Were there any connection leaks?');
  console.log('  □ Did error rates remain stable?');
  console.log('  □ Are there any patterns in metrics over time?');
  console.log('  □ Compare first 5min vs last 5min metrics');
  console.log('='.repeat(80));
}
