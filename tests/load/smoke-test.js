// Smoke Test: Verify basic functionality under minimal load
// Purpose: Ensure system works before running full load tests
// Duration: 1 minute
// VUs: 1-5

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const searchDuration = new Trend('search_duration');

export const options = {
  stages: [
    { duration: '30s', target: 1 },  // Ramp up to 1 user
    { duration: '30s', target: 5 },  // Increase to 5 users
  ],
  thresholds: {
    'http_req_duration': ['p(95)<1000'], // 95% of requests should be below 1s
    'http_req_failed': ['rate<0.01'],    // Error rate should be below 1%
    'errors': ['rate<0.05'],              // Custom error rate below 5%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  // Test 1: Health check
  const healthRes = http.get(`${BASE_URL}/health`);
  const healthCheck = check(healthRes, {
    'health status is 200': (r) => r.status === 200,
    'health response is healthy': (r) => r.json('status') === 'healthy',
  });
  errorRate.add(!healthCheck);

  sleep(1);

  // Test 2: List tools (MCP)
  const listToolsPayload = JSON.stringify({
    jsonrpc: '2.0',
    id: 1,
    method: 'tools/list',
    params: {},
  });

  const listToolsRes = http.post(
    `${BASE_URL}/mcp`,
    listToolsPayload,
    {
      headers: { 'Content-Type': 'application/json' },
    }
  );

  const toolsCheck = check(listToolsRes, {
    'tools/list status is 200': (r) => r.status === 200,
    'tools/list has result': (r) => r.json('result') !== undefined,
    'tools/list has tools array': (r) => r.json('result.tools') !== undefined,
  });
  errorRate.add(!toolsCheck);

  sleep(1);

  // Test 3: Context search (if implemented)
  const searchPayload = JSON.stringify({
    jsonrpc: '2.0',
    id: 2,
    method: 'tools/call',
    params: {
      name: 'context.search',
      arguments: {
        query: 'test function',
        top_k: 10,
      },
    },
  });

  const searchStart = Date.now();
  const searchRes = http.post(
    `${BASE_URL}/mcp`,
    searchPayload,
    {
      headers: { 'Content-Type': 'application/json' },
    }
  );
  const searchTime = Date.now() - searchStart;
  searchDuration.add(searchTime);

  const searchCheck = check(searchRes, {
    'search status is 200': (r) => r.status === 200,
    'search response is valid': (r) => r.json('result') !== undefined || r.json('error') !== undefined,
  });
  errorRate.add(!searchCheck);

  sleep(2);
}

export function handleSummary(data) {
  return {
    'results/smoke-test-summary.json': JSON.stringify(data, null, 2),
    stdout: textSummary(data, { indent: ' ', enableColors: true }),
  };
}

function textSummary(data, options) {
  const indent = options.indent || '';
  const enableColors = options.enableColors || false;
  
  let summary = '\n\n';
  summary += `${indent}Smoke Test Summary\n`;
  summary += `${indent}==================\n\n`;
  
  // Requests
  const requests = data.metrics.http_reqs;
  summary += `${indent}Total Requests: ${requests.values.count}\n`;
  summary += `${indent}Request Rate: ${requests.values.rate.toFixed(2)} req/s\n\n`;
  
  // Response Times
  const duration = data.metrics.http_req_duration;
  summary += `${indent}Response Times:\n`;
  summary += `${indent}  p50: ${duration.values['p(50)'].toFixed(2)}ms\n`;
  summary += `${indent}  p95: ${duration.values['p(95)'].toFixed(2)}ms\n`;
  summary += `${indent}  p99: ${duration.values['p(99)'].toFixed(2)}ms\n`;
  summary += `${indent}  max: ${duration.values.max.toFixed(2)}ms\n\n`;
  
  // Error Rates
  const failed = data.metrics.http_req_failed;
  const errors = data.metrics.errors;
  summary += `${indent}Error Rates:\n`;
  summary += `${indent}  HTTP Failures: ${(failed.values.rate * 100).toFixed(2)}%\n`;
  summary += `${indent}  Check Failures: ${(errors.values.rate * 100).toFixed(2)}%\n\n`;
  
  // Custom Metrics
  if (data.metrics.search_duration) {
    const searchDur = data.metrics.search_duration;
    summary += `${indent}Search Performance:\n`;
    summary += `${indent}  avg: ${searchDur.values.avg.toFixed(2)}ms\n`;
    summary += `${indent}  p95: ${searchDur.values['p(95)'].toFixed(2)}ms\n\n`;
  }
  
  return summary;
}
