/**
 * Conexus MCP Server - Spike Test
 * 
 * Tests system behavior under sudden traffic spikes.
 * Validates graceful degradation and recovery capabilities.
 * 
 * Usage:
 *   k6 run tests/load/spike-test.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend, Rate } from 'k6/metrics';

// Custom metrics
const requestDuration = new Trend('request_duration', true);
const errorRate = new Rate('error_rate');
const spikeErrors = new Counter('spike_errors');
const recoveryTime = new Trend('recovery_time', true);

// Test configuration - sudden spikes
export const options = {
  stages: [
    // Baseline: Normal load
    { duration: '1m', target: 50 },     // Ramp to baseline
    { duration: '2m', target: 50 },     // Hold baseline
    
    // Spike 1: Moderate sudden increase
    { duration: '10s', target: 200 },   // Sudden spike to 200 VUs
    { duration: '1m', target: 200 },    // Hold spike
    { duration: '30s', target: 50 },    // Recover to baseline
    { duration: '1m', target: 50 },     // Observe recovery
    
    // Spike 2: Large sudden increase
    { duration: '10s', target: 400 },   // Sudden spike to 400 VUs
    { duration: '1m', target: 400 },    // Hold spike
    { duration: '30s', target: 50 },    // Recover to baseline
    { duration: '1m', target: 50 },     // Observe recovery
    
    // Spike 3: Extreme sudden increase
    { duration: '5s', target: 600 },    // Very sudden spike to 600 VUs
    { duration: '30s', target: 600 },   // Brief hold
    { duration: '30s', target: 50 },    // Recover to baseline
    { duration: '2m', target: 50 },     // Extended recovery observation
    
    // Cool-down
    { duration: '30s', target: 0 },     // Ramp down
  ],
  
  thresholds: {
    // More lenient thresholds - spikes may cause temporary degradation
    'http_req_duration': ['p(95)<3000'],  // Allow up to 3s during spikes
    'error_rate': ['rate<0.05'],           // Allow up to 5% errors overall
    'http_req_failed': ['rate<0.10'],      // Allow up to 10% failures
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Test queries
const queries = [
  'error handling',
  'database query',
  'API endpoint',
  'configuration',
  'authentication',
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

// Main test operation - fast context search
function testFastOperation() {
  const query = queries[Math.floor(Math.random() * queries.length)];
  const startTime = Date.now();
  
  const payload = createMCPRequest('tools/call', {
    name: 'context.search',
    arguments: {
      query: query,
      max_results: 5,  // Smaller result set for faster responses
    },
  });
  
  const response = http.post(
    `${BASE_URL}/mcp`,
    payload,
    {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'context.search' },
      timeout: '10s',
    }
  );
  
  const duration = Date.now() - startTime;
  requestDuration.add(duration);
  
  const success = check(response, {
    'status 200 or 503': (r) => r.status === 200 || r.status === 503,
    'has response body': (r) => r.body && r.body.length > 0,
    'valid JSON': (r) => {
      try {
        JSON.parse(r.body);
        return true;
      } catch {
        return false;
      }
    },
  });
  
  // Track spike-specific errors
  if (response.status === 503 || response.status === 0) {
    spikeErrors.add(1);
  }
  
  errorRate.add(success ? 0 : 1);
  
  return response;
}

// Lightweight health check
function testHealthCheck() {
  const response = http.get(`${BASE_URL}/health`, {
    timeout: '5s',
  });
  
  check(response, {
    'health: status 200': (r) => r.status === 200,
    'health: fast response': (r) => r.timings.duration < 500,
  });
  
  return response;
}

// Main test function
export default function () {
  const rand = Math.random();
  
  if (rand < 0.90) {
    // 90% - Main search operations
    testFastOperation();
    // No think time during spikes - simulate burst traffic
    sleep(0.1);
  } else {
    // 10% - Health checks
    testHealthCheck();
    sleep(0.1);
  }
}

// Setup function
export function setup() {
  console.log('='.repeat(80));
  console.log('Conexus Spike Test Starting');
  console.log('='.repeat(80));
  console.log(`Target URL: ${BASE_URL}`);
  console.log(`Total Duration: ~13 minutes`);
  console.log(`Spike Pattern: 50 → 200 → 50 → 400 → 50 → 600 → 50`);
  console.log(`Goal: Test resilience to sudden traffic bursts`);
  console.log('='.repeat(80));
  
  // Verify server is accessible
  const healthCheck = http.get(`${BASE_URL}/health`);
  if (healthCheck.status !== 200) {
    throw new Error(`Server health check failed: ${healthCheck.status}`);
  }
  
  console.log('✓ Server health check passed');
  console.log('');
  console.log('MONITORING TIPS:');
  console.log('  - Watch for graceful degradation (503 responses)');
  console.log('  - Monitor recovery time after spikes');
  console.log('  - Check for connection pool saturation');
  console.log('  - Observe queue depths and request queueing');
  console.log('  - Look for circuit breaker activation');
  console.log('='.repeat(80));
  console.log('');
  
  return { 
    startTime: Date.now(),
    spikeTimestamps: [],
  };
}

// Teardown function
export function teardown(data) {
  const duration = (Date.now() - data.startTime) / 1000;
  
  console.log('');
  console.log('='.repeat(80));
  console.log('Spike Test Complete');
  console.log('='.repeat(80));
  console.log(`Total Duration: ${duration.toFixed(1)}s`);
  console.log('');
  console.log('ANALYSIS CHECKLIST:');
  console.log('  □ Did system handle spikes gracefully (503s vs crashes)?');
  console.log('  □ How quickly did system recover after each spike?');
  console.log('  □ Were baseline metrics restored after recovery?');
  console.log('  □ At what VU count did spikes cause significant errors?');
  console.log('  □ Did multiple spikes cause cumulative degradation?');
  console.log('  □ Were there any resource leaks after spikes?');
  console.log('='.repeat(80));
}
