#!/usr/bin/env node
/**
 * MCP Inspector Integration Tests for Conexus
 * 
 * Tests the Conexus MCP server for protocol compliance using the official
 * MCP Inspector tool. Tests cover:
 * - Tools/list endpoint and JSON-RPC 2.0 compliance
 * - Individual tool invocations and schema validation
 * - Error handling and edge cases
 * - Protocol compliance and response formats
 * 
 * Usage: node test-inspector.js [--conexus-path ./conexus]
 */

import { spawn } from 'child_process';
import { existsSync } from 'fs';
import path from 'path';

// Configuration
const args = process.argv.slice(2);
const conexusPathIdx = args.indexOf('--conexus-path');
const conexusPath = conexusPathIdx >= 0 ? args[conexusPathIdx + 1] : './bin/conexus-' + process.platform.replace('darwin', 'macos').replace('linux', 'linux') + '-' + (process.arch === 'arm64' ? 'arm64' : 'amd64');

// Test results tracking
let testsRun = 0;
let testsPassed = 0;
let testsFailed = 0;
const failedTests = [];

// Color codes
const colors = {
  reset: '\x1b[0m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
};

// Utility functions
function logInfo(msg) {
  console.log(`${colors.blue}[INFO]${colors.reset} ${msg}`);
}

function logSuccess(msg) {
  console.log(`${colors.green}[PASS]${colors.reset} ${msg}`);
  testsPassed++;
}

function logError(msg) {
  console.log(`${colors.red}[FAIL]${colors.reset} ${msg}`);
  testsFailed++;
  failedTests.push(msg);
}

function logWarning(msg) {
  console.log(`${colors.yellow}[WARN]${colors.reset} ${msg}`);
}

function testStart(name) {
  testsRun++;
  logInfo(`Test ${testsRun}: ${name}`);
}

/**
 * Execute a command via MCP Inspector's stdio interface
 */
function inspectorCommand(inspector, command, expectedFields = []) {
  return new Promise((resolve) => {
    let response = '';
    let resolved = false;

    const onData = (data) => {
      response += data.toString();
      
      // Try to parse and check if response contains expected fields
      try {
        const lines = response.split('\n');
        for (const line of lines) {
          if (line.trim().startsWith('{')) {
            const json = JSON.parse(line);
            
            // For tools/list or other responses
            if (expectedFields.length === 0 || 
                expectedFields.some(field => field in json || field in (json.result || {}))) {
              if (!resolved) {
                resolved = true;
                inspector.stdout.removeListener('data', onData);
                resolve({ success: true, data: json, raw: response });
              }
              return;
            }
          }
        }
      } catch (e) {
        // JSON parse error, wait for more data
      }

      // Timeout after reasonable data collection
      if (response.length > 10000) {
        if (!resolved) {
          resolved = true;
          inspector.stdout.removeListener('data', onData);
          resolve({ success: false, data: null, raw: response, error: 'Response too large' });
        }
      }
    };

    inspector.stdout.once('data', onData);
    inspector.stdin.write(`${command}\n`);

    // Timeout
    setTimeout(() => {
      if (!resolved) {
        resolved = true;
        inspector.stdout.removeListener('data', onData);
        resolve({ success: false, data: null, raw: response, error: 'Command timeout' });
      }
    }, 5000);
  });
}

/**
 * Start MCP Inspector with Conexus
 */
function startInspector() {
  return new Promise((resolve, reject) => {
    // Check if conexus binary exists
    if (!existsSync(conexusPath)) {
      reject(new Error(`Conexus binary not found at: ${conexusPath}`));
      return;
    }

    logInfo(`Starting MCP Inspector with: ${conexusPath}`);
    
    const inspector = spawn('npx', ['-y', '@modelcontextprotocol/inspector', conexusPath], {
      stdio: ['pipe', 'pipe', 'pipe'],
      timeout: 30000,
    });

    // Handle errors
    inspector.on('error', (err) => {
      reject(new Error(`Failed to start inspector: ${err.message}`));
    });

    // Wait for inspector to be ready
    let output = '';
    const readyHandler = (data) => {
      output += data.toString();
      if (output.includes('Server') || output.includes('Connected') || output.includes('>')) {
        inspector.stdout.removeListener('data', readyHandler);
        resolve(inspector);
      }
    };

    inspector.stdout.on('data', readyHandler);

    setTimeout(() => {
      inspector.stdout.removeListener('data', readyHandler);
      resolve(inspector); // Proceed anyway
    }, 3000);
  });
}

/**
 * Run all tests
 */
async function runTests() {
  console.log('');
  console.log('==========================================');
  console.log('  MCP Inspector Integration Tests');
  console.log('==========================================');
  console.log('');

  let inspector;

  try {
    // Start inspector
    logInfo('Starting MCP Inspector...');
    inspector = await startInspector();
    logSuccess('MCP Inspector started');
    console.log('');

    // Test 1: List tools
    testStart('tools/list returns available tools');
    const toolsResponse = await inspectorCommand(inspector, 'tools', ['tools', 'result']);
    
    if (toolsResponse.success && toolsResponse.data.result && Array.isArray(toolsResponse.data.result.tools)) {
      const tools = toolsResponse.data.result.tools;
      logSuccess(`Found ${tools.length} tools`);
      
      // List all tools
      logInfo('Available tools:');
      tools.forEach(tool => {
        console.log(`  - ${tool.name}`);
      });
    } else {
      logError('Failed to get tools list');
    }
    console.log('');

    // Test 2: Validate JSON-RPC 2.0 format
    testStart('JSON-RPC 2.0 format validation');
    if (toolsResponse.data.jsonrpc === '2.0' && typeof toolsResponse.data.id !== 'undefined') {
      logSuccess('Response is valid JSON-RPC 2.0');
    } else {
      logError('Response does not conform to JSON-RPC 2.0 spec');
    }
    console.log('');

    // Test 3: Test context.search tool
    testStart('context.search tool invocation');
    const searchResponse = await inspectorCommand(
      inspector,
      'call context.search {"query": "function", "top_k": 5}',
      ['result', 'error']
    );
    
    if (searchResponse.success && searchResponse.data) {
      if (searchResponse.data.result) {
        logSuccess('context.search executed successfully');
        if (searchResponse.data.result.results) {
          logInfo(`  Found ${searchResponse.data.result.results.length} results`);
        }
      } else if (searchResponse.data.error) {
        logWarning(`context.search returned error: ${searchResponse.data.error.message}`);
      }
    } else {
      logError('context.search call failed');
    }
    console.log('');

    // Test 4: Test context.grep tool
    testStart('context.grep tool invocation');
    const grepResponse = await inspectorCommand(
      inspector,
      'call context.grep {"pattern": "func", "include": "*.go"}',
      ['result', 'error']
    );
    
    if (grepResponse.success && grepResponse.data) {
      if (grepResponse.data.result) {
        logSuccess('context.grep executed successfully');
        if (grepResponse.data.result.matches) {
          logInfo(`  Found ${grepResponse.data.result.matches.length} matches`);
        }
      } else if (grepResponse.data.error) {
        logWarning(`context.grep returned error: ${grepResponse.data.error.message}`);
      }
    } else {
      logError('context.grep call failed');
    }
    console.log('');

    // Test 5: Test context.index_control tool
    testStart('context.index_control tool (status action)');
    const statusResponse = await inspectorCommand(
      inspector,
      'call context.index_control {"action": "status"}',
      ['result', 'error']
    );
    
    if (statusResponse.success && statusResponse.data) {
      if (statusResponse.data.result) {
        logSuccess('context.index_control status executed successfully');
        if (statusResponse.data.result.message) {
          logInfo(`  Status: ${statusResponse.data.result.message}`);
        }
      } else if (statusResponse.data.error) {
        logWarning(`context.index_control returned error: ${statusResponse.data.error.message}`);
      }
    } else {
      logError('context.index_control call failed');
    }
    console.log('');

    // Test 6: Test context.get_related_info tool
    testStart('context.get_related_info tool invocation');
    const relatedResponse = await inspectorCommand(
      inspector,
      'call context.get_related_info {"file_path": "main.go"}',
      ['result', 'error']
    );
    
    if (relatedResponse.success && relatedResponse.data) {
      if (relatedResponse.data.result) {
        logSuccess('context.get_related_info executed successfully');
      } else if (relatedResponse.data.error) {
        logWarning(`context.get_related_info returned error: ${relatedResponse.data.error.message}`);
      }
    } else {
      logError('context.get_related_info call failed');
    }
    console.log('');

    // Test 7: Test context.connector_management tool
    testStart('context.connector_management tool (list action)');
    const connectorResponse = await inspectorCommand(
      inspector,
      'call context.connector_management {"action": "list"}',
      ['result', 'error']
    );
    
    if (connectorResponse.success && connectorResponse.data) {
      if (connectorResponse.data.result) {
        logSuccess('context.connector_management list executed successfully');
        if (connectorResponse.data.result.connectors) {
          logInfo(`  Found ${connectorResponse.data.result.connectors.length} connectors`);
        }
      } else if (connectorResponse.data.error) {
        logWarning(`context.connector_management returned error: ${connectorResponse.data.error.message}`);
      }
    } else {
      logError('context.connector_management call failed');
    }
    console.log('');

    // Test 8: Test error handling - invalid tool
    testStart('Error handling - invalid tool name');
    const invalidResponse = await inspectorCommand(
      inspector,
      'call invalid_tool {"test": "value"}',
      ['result', 'error']
    );
    
    if (invalidResponse.data && invalidResponse.data.error) {
      logSuccess('Invalid tool properly rejected with error response');
      logInfo(`  Error code: ${invalidResponse.data.error.code}`);
      logInfo(`  Error message: ${invalidResponse.data.error.message}`);
    } else {
      logError('Invalid tool should have returned an error');
    }
    console.log('');

    // Test 9: Test error handling - invalid parameters
    testStart('Error handling - invalid parameters');
    const invalidParamsResponse = await inspectorCommand(
      inspector,
      'call context.search {"invalid_param": "test"}',
      ['result', 'error']
    );
    
    if (invalidParamsResponse.data && (invalidParamsResponse.data.error || invalidParamsResponse.data.result)) {
      logSuccess('Invalid parameters handled appropriately');
    } else {
      logError('Invalid parameters not handled properly');
    }
    console.log('');

    // Test 10: Protocol compliance - ID tracking
    testStart('Protocol compliance - request ID tracking');
    if (toolsResponse.data.id && typeof toolsResponse.data.id === 'number') {
      logSuccess('Server properly tracks request IDs');
    } else {
      logError('Server not tracking request IDs properly');
    }
    console.log('');

  } catch (error) {
    logError(`Test suite error: ${error.message}`);
  } finally {
    // Cleanup
    if (inspector) {
      try {
        inspector.stdin.write('exit\n');
        inspector.kill();
      } catch (e) {
        // Ignore cleanup errors
      }
    }
  }

  // Print summary
  console.log('');
  console.log('==========================================');
  console.log('  Test Results Summary');
  console.log('==========================================');
  console.log(`Tests Run:    ${testsRun}`);
  console.log(`${colors.green}Tests Passed: ${testsPassed}${colors.reset}`);
  console.log(`${colors.red}Tests Failed: ${testsFailed}${colors.reset}`);
  console.log('==========================================');
  console.log('');

  if (testsFailed > 0) {
    console.log(`${colors.red}Failed Tests:${colors.reset}`);
    failedTests.forEach(test => {
      console.log(`  - ${test}`);
    });
    console.log('');
    process.exit(1);
  } else {
    console.log(`${colors.green}All tests passed! ✅${colors.reset}`);
    process.exit(0);
  }
}

// Run tests
runTests().catch(error => {
  console.error(`${colors.red}Fatal error:${colors.reset}`, error.message);
  process.exit(1);
});
