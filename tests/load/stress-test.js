/**
 * Conexus MCP Server - Stress Test
 * 
 * Progressively increases load to find system breaking points.
 * Monitors when response times degrade and errors start occurring.
 * 
 * Usage:
 *   k6 run tests/load/stress-test.js
 *   k6 run --out json=results/stress-test.json tests/load/stress-test.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend, Rate } from 'k6/metrics';

// Custom metrics
const requestDuration = new Trend('request_duration', true);
const errorRate = new Rate('error_rate');
const timeoutErrors = new Counter('timeout_errors');
const serverErrors = new Counter('server_errors');

// Test configuration - progressive load increase
export const options = {
  stages: [
    // Phase 1: Baseline (known working load)
    { duration: '2m', target: 50 },    // Ramp to 50 VUs
    { duration: '2m', target: 50 },    // Hold 50 VUs
    
    // Phase 2: Target load
    { duration: '1m', target: 100 },   // Ramp to 100 VUs
    { duration: '2m', target: 100 },   // Hold 100 VUs
    
    // Phase 3: Stress begins
    { duration: '1m', target: 200 },   // Ramp to 200 VUs
    { duration: '2m', target: 200 },   // Hold 200 VUs
    
    // Phase 4: High stress
    { duration: '1m', target: 300 },   // Ramp to 300 VUs
    { duration: '2m', target: 300 },   // Hold 300 VUs
    
    // Phase 5: Breaking point search
    { duration: '1m', target: 400 },   // Ramp to 400 VUs
    { duration: '2m', target: 400 },   // Hold 400 VUs
    
    // Phase 6: Extreme stress (optional - may fail)
    { duration: '1m', target: 500 },   // Ramp to 500 VUs
    { duration: '2m', target: 500 },   // Hold 500 VUs
    
    // Recovery: Rapid cool-down to observe recovery behavior
    { duration: '30s', target: 100 },  // Drop to 100 VUs
    { duration: '1m', target: 100 },   // Hold to verify recovery
    { duration: '30s', target: 0 },    // Ramp down
  ],
  
  // More lenient thresholds - we expect to exceed limits
  thresholds: {
    'http_req_duration': ['p(95)<5000'],  // Allow up to 5s at peak
    'error_rate': ['rate<0.10'],           // Allow up to 10% errors
    'http_req_failed': ['rate<0.15'],      // Allow up to 15% failures
  },
  
  // Extended timeouts for stress conditions
  httpDebug: 'full',
  noConnectionReuse: false,
  userAgent: 'k6-stress-test/1.0',
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Helper to create MCP JSON-RPC request
function createMCPRequest(method, params = {}) {
  return JSON.stringify({
    jsonrpc: '2.0',
    id: Math.random().toString(36).substring(7),
    method: method,
    params: params,
  });
}

// Test scenario: Mix of all operations
function runMixedWorkload() {
  const operations = [
    // Context search - most expensive operation
     {
       name: 'context.search',
       weight: 0.50,
       params: {
         name: 'context.search',
         arguments: {
           query: 'function implementation',
           max_results: 10,
         },
       },
     },
    // Get related info - expensive operation
     {
       name: 'context.get_related_info',
       weight: 0.30,
       params: {
         name: 'context.get_related_info',
         arguments: {
           query: 'error handling',
           max_depth: 2,
         },
       },
     },
    // Index control - lightweight operation
     {
       name: 'context.index_control',
       weight: 0.15,
       params: {
         name: 'context.index_control',
         arguments: {
           action: 'status',
         },
       },
     },
    // Connector management - lightweight operation
     {
       name: 'context.connector_management',
       weight: 0.05,
       params: {
         name: 'context.connector_management',
         arguments: {
           action: 'list',
         },
       },
     },
  ];
  
  // Weighted random selection
  const rand = Math.random();
  let cumulative = 0;
  let selectedOp = operations[0];
  
  for (const op of operations) {
    cumulative += op.weight;
    if (rand < cumulative) {
      selectedOp = op;
      break;
    }
  }
  
  // Execute request
  const startTime = Date.now();
  const payload = createMCPRequest('tools/call', selectedOp.params);
  
  const response = http.post(
    `${BASE_URL}/mcp`,
    payload,
    {
      headers: { 'Content-Type': 'application/json' },
      tags: { 
        name: selectedOp.name,
        operation: 'tools/call',
      },
      timeout: '30s',  // Extended timeout for stress conditions
    }
  );
  
  const duration = Date.now() - startTime;
  requestDuration.add(duration);
  
  // Check response
  const isSuccess = check(response, {
    'status is 200': (r) => r.status === 200,
    'response has body': (r) => r.body && r.body.length > 0,
    'valid JSON': (r) => {
      try {
        JSON.parse(r.body);
        return true;
      } catch {
        return false;
      }
    },
    'has JSON-RPC response': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.jsonrpc === '2.0';
      } catch {
        return false;
      }
    },
  });
  
  // Track error types
  if (!isSuccess) {
    errorRate.add(1);
    
    if (response.status === 0) {
      timeoutErrors.add(1);
    } else if (response.status >= 500) {
      serverErrors.add(1);
    }
  } else {
    errorRate.add(0);
  }
  
  return response;
}

// Main test function
export default function () {
  runMixedWorkload();
  
  // Shorter think time under stress - more aggressive load
  sleep(Math.random() * 2 + 0.5);  // 0.5-2.5 seconds
}

// Setup function
export function setup() {
  console.log('='.repeat(80));
  console.log('Conexus Stress Test Starting');
  console.log('='.repeat(80));
  console.log(`Target URL: ${BASE_URL}`);
  console.log(`Total Duration: ~20 minutes`);
  console.log(`Load Progression: 50 → 100 → 200 → 300 → 400 → 500 VUs`);
  console.log(`Goal: Find breaking point and observe degradation`);
  console.log('='.repeat(80));
  
  // Verify server is accessible
  const healthCheck = http.get(`${BASE_URL}/health`);
  if (healthCheck.status !== 200) {
    throw new Error(`Server health check failed: ${healthCheck.status}`);
  }
  
  console.log('✓ Server health check passed');
  console.log('');
  console.log('MONITORING TIPS:');
  console.log('  - Watch for response time degradation');
  console.log('  - Monitor error rate increases');
  console.log('  - Check system resources (CPU, memory, connections)');
  console.log('  - Observe Prometheus metrics at :9090/metrics');
  console.log('='.repeat(80));
  console.log('');
  
  return { startTime: Date.now() };
}

// Teardown function
export function teardown(data) {
  const duration = (Date.now() - data.startTime) / 1000;
  
  console.log('');
  console.log('='.repeat(80));
  console.log('Stress Test Complete');
  console.log('='.repeat(80));
  console.log(`Total Duration: ${duration.toFixed(1)}s`);
  console.log('');
  console.log('ANALYSIS CHECKLIST:');
  console.log('  □ At what VU level did p95 exceed 1s?');
  console.log('  □ At what VU level did error rate exceed 1%?');
  console.log('  □ Did the system recover when load decreased?');
  console.log('  □ What was the maximum sustainable load?');
  console.log('  □ What resources were exhausted (CPU/mem/connections)?');
  console.log('='.repeat(80));
}
