/**
 * Dogfooding Performance Benchmark
 * 
 * Measures response times and throughput for Conexus context search
 * using realistic developer scenarios.
 * 
 * Usage:
 *   bun run tests/dogfooding/benchmarks/performance-benchmark.js
 */

import { codeSearchScenarios } from '../scenarios/code-search-scenarios.js';
import { relationshipDiscoveryScenarios } from '../scenarios/relationship-discovery-scenarios.js';
import { contextAwareScenarios } from '../scenarios/context-aware-scenarios.js';
import { crossReferenceScenarios } from '../scenarios/cross-reference-scenarios.js';

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';
const CONCURRENT_REQUESTS = 10;
const TOTAL_ITERATIONS = 100;

// Performance metrics
const metrics = {
  responseTimes: [],
  throughput: 0,
  errorRate: 0,
  p95: 0,
  p99: 0,
  avgResponseTime: 0
};

// Helper to create MCP JSON-RPC request
function createMCPRequest(method, params = {}) {
  return JSON.stringify({
    jsonrpc: '2.0',
    id: Math.random().toString(36).substring(7),
    method: method,
    params: params,
  });
}

// Send single MCP request
async function sendMCPRequest(method, params = {}) {
  const payload = createMCPRequest(method, params);
  const startTime = Date.now();
  
  try {
    const response = await fetch(`${BASE_URL}/mcp`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: payload,
    });
    
    const duration = Date.now() - startTime;
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    
    const body = await response.json();
    
    return {
      success: true,
      duration,
      result: body.result
    };
  } catch (error) {
    return {
      success: false,
      duration: Date.now() - startTime,
      error: error.message
    };
  }
}

// Run performance test for a scenario
async function runScenarioPerformance(scenario) {
  const results = [];
  
  // Run multiple iterations
  for (let i = 0; i < 5; i++) {
    const result = await sendMCPRequest('tools/call', {
      name: 'context.search',
      arguments: {
        query: scenario.query,
        max_results: 20,
        work_context: scenario.workContext
      }
    });
    
    results.push(result);
    
    // Small delay between requests
    await new Promise(resolve => setTimeout(resolve, 50));
  }
  
  return results;
}

// Calculate percentiles
function calculatePercentile(times, p) {
  const sorted = [...times].sort((a, b) => a - b);
  const index = Math.ceil((p / 100) * sorted.length) - 1;
  return sorted[index];
}

// Run concurrent load test
async function runConcurrentLoad(scenarios, duration = 30000) { // 30 seconds
  const startTime = Date.now();
  const results = [];
  let requestCount = 0;
  
  async function worker() {
    while (Date.now() - startTime < duration) {
      const scenario = scenarios[Math.floor(Math.random() * scenarios.length)];
      
      const result = await sendMCPRequest('tools/call', {
        name: 'context.search',
        arguments: {
          query: scenario.query,
          max_results: 10
        }
      });
      
      results.push(result);
      requestCount++;
      
      // Random delay 100-500ms
      await new Promise(resolve => setTimeout(resolve, 100 + Math.random() * 400));
    }
  }
  
  // Start concurrent workers
  const workers = [];
  for (let i = 0; i < CONCURRENT_REQUESTS; i++) {
    workers.push(worker());
  }
  
  await Promise.all(workers);
  
  const totalTime = (Date.now() - startTime) / 1000;
  const throughput = requestCount / totalTime;
  
  return { results, throughput, totalTime, requestCount };
}

// Main performance benchmark
async function runPerformanceBenchmark() {
  console.log('='.repeat(80));
  console.log('Conexus Dogfooding Performance Benchmark');
  console.log('='.repeat(80));
  console.log(`Target URL: ${BASE_URL}`);
  console.log(`Concurrent Requests: ${CONCURRENT_REQUESTS}`);
  console.log(`Test Duration: 30 seconds`);
  console.log('');
  
  // Health check
  try {
    const health = await fetch(`${BASE_URL}/health`);
    if (!health.ok) throw new Error(`Health check failed: ${health.status}`);
    console.log('✓ Server health check passed');
  } catch (error) {
    console.error('✗ Server health check failed:', error.message);
    process.exit(1);
  }
  
  console.log('');
  
  // Prepare scenarios
  const allScenarios = [
    ...codeSearchScenarios,
    ...relationshipDiscoveryScenarios,
    ...contextAwareScenarios,
    ...crossReferenceScenarios
  ];
  
  console.log('Running individual scenario performance tests...');
  
  // Test individual scenarios
  const individualResults = [];
  for (const scenario of allScenarios.slice(0, 10)) { // Test first 10 scenarios
    const results = await runScenarioPerformance(scenario);
    individualResults.push(...results);
    
    const avgTime = results.reduce((sum, r) => sum + r.duration, 0) / results.length;
    const successRate = results.filter(r => r.success).length / results.length * 100;
    
    console.log(`  ${scenario.id}: ${avgTime.toFixed(1)}ms avg, ${successRate.toFixed(1)}% success`);
  }
  
  console.log('');
  console.log('Running concurrent load test...');
  
  // Run concurrent load test
  const loadResults = await runConcurrentLoad(allScenarios);
  
  console.log(`  Completed ${loadResults.requestCount} requests in ${loadResults.totalTime.toFixed(1)}s`);
  console.log(`  Throughput: ${loadResults.throughput.toFixed(1)} req/s`);
  
  // Calculate metrics
  const allResponseTimes = [...individualResults, ...loadResults.results]
    .filter(r => r.success)
    .map(r => r.duration);
  
  const errorCount = [...individualResults, ...loadResults.results]
    .filter(r => !r.success).length;
  
  const totalRequests = individualResults.length + loadResults.results.length;
  
  metrics.responseTimes = allResponseTimes;
  metrics.throughput = loadResults.throughput;
  metrics.errorRate = errorCount / totalRequests;
  metrics.p95 = calculatePercentile(allResponseTimes, 95);
  metrics.p99 = calculatePercentile(allResponseTimes, 99);
  metrics.avgResponseTime = allResponseTimes.reduce((a, b) => a + b, 0) / allResponseTimes.length;
  
  // Save results
  const fs = await import('fs');
  await fs.promises.writeFile(
    'tests/dogfooding/results/performance-results.json',
    JSON.stringify(metrics, null, 2)
  );
  
  // Print results
  console.log('');
  console.log('='.repeat(80));
  console.log('Performance Results');
  console.log('='.repeat(80));
  console.log(`Total Requests: ${totalRequests}`);
  console.log(`Successful Requests: ${totalRequests - errorCount}`);
  console.log(`Error Rate: ${(metrics.errorRate * 100).toFixed(2)}%`);
  console.log(`Average Response Time: ${metrics.avgResponseTime.toFixed(1)}ms`);
  console.log(`P95 Response Time: ${metrics.p95.toFixed(1)}ms`);
  console.log(`P99 Response Time: ${metrics.p99.toFixed(1)}ms`);
  console.log(`Throughput: ${metrics.throughput.toFixed(1)} req/s`);
  console.log('');
  console.log('Results saved to tests/dogfooding/results/performance-results.json');
  console.log('='.repeat(80));
}

// Run the benchmark
runPerformanceBenchmark().catch(console.error);
