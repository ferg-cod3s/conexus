// Test MCP integration with Conexus
import http from 'http';

function callMCP(method, params, id = 1) {
  return new Promise((resolve, reject) => {
    const postData = JSON.stringify({
      jsonrpc: "2.0",
      id,
      method,
      params
    });

    const options = {
      hostname: 'localhost',
      port: 8080,
      path: '/mcp',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(postData)
      }
    };

    const req = http.request(options, (res) => {
      let data = '';
      res.on('data', (chunk) => data += chunk);
      res.on('end', () => {
        try {
          resolve(JSON.parse(data));
        } catch (e) {
          reject(e);
        }
      });
    });

    req.on('error', reject);
    req.write(postData);
    req.end();
  });
}

async function testMCPIntegration() {
  console.log('🔗 Testing MCP Integration with Conexus\n');

  try {
    // Test each tool that would be used in development
    console.log('🛠️  Testing Development Workflow Tools:\n');

    // 1. Index control - check status
    console.log('📊 Index Status:');
    const status = await callMCP('tools/call', { 
      name: 'context_index_control', 
      arguments: { action: 'status' } 
    });
    console.log(`   ✅ ${status.result.message}`);
    console.log(`   📈 Documents indexed: ${status.result.details.documents_indexed}\n`);

    // 2. Connector management - list available
    console.log('🔌 Available Connectors:');
    const connectors = await callMCP('tools/call', { 
      name: 'context_connector_management', 
      arguments: { action: 'list' } 
    });
    if (connectors.result.connectors && connectors.result.connectors.length > 0) {
      connectors.result.connectors.forEach(conn => {
        console.log(`   ✅ ${conn.name} (${conn.type}) - ${conn.status}`);
      });
    } else {
      console.log('   ℹ️  No connectors configured');
    }
    console.log();

    // 3. Search capability (placeholder for now)
    console.log('🔍 Code Search:');
    const search = await callMCP('tools/call', { 
      name: 'context_search', 
      arguments: { query: 'function definitions' } 
    });
    console.log(`   ✅ Search executed (${search.result.totalCount || 0} results found)`);
    console.log();

    // 4. Related info lookup
    console.log('📋 Related Information:');
    const related = await callMCP('tools/call', { 
      name: 'context_get_related_info', 
      arguments: { file_path: 'main.go' } 
    });
    console.log(`   ✅ Related info query processed`);
    console.log();

    console.log('🎉 MCP Integration Test Complete!');
    console.log('\n💡 Ready for Claude Code / OpenCode integration');
    console.log('   Copy claude-mcp-config.json to ~/.claude/mcp.json');
    console.log('   Then use: /mcp conexus tools/call context_index_control {"action": "status"}');

  } catch (error) {
    console.error('❌ Integration test failed:', error.message);
    console.log('\n🔧 Troubleshooting:');
    console.log('   1. Check if Conexus is running: curl http://localhost:8080/health');
    console.log('   2. Verify port 8080 is not in use');
    console.log('   3. Check Conexus logs: tail -f conexus.log');
  }
}

testMCPIntegration();
