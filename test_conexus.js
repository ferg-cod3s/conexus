// Simple MCP client test script
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

// Test all working tools
async function testConexus() {
  console.log('🧪 Testing Conexus MCP Server\n');

  try {
    // 1. List tools
    console.log('📋 Available tools:');
    const tools = await callMCP('tools/list');
    tools.result.tools.forEach(tool => {
      console.log(`  • ${tool.name}: ${tool.description.split('.')[0]}`);
    });
    console.log();

    // 2. Test index status
    console.log('📊 Index Status:');
    const status = await callMCP('tools/call', { name: 'context_index_control', arguments: { action: 'status' } });
    console.log(`  ${status.result.message}`);
    console.log(`  Details: ${JSON.stringify(status.result.details)}\n`);

    // 3. Test connector management
    console.log('🔌 Available Connectors:');
    const connectors = await callMCP('tools/call', { name: 'context_connector_management', arguments: { action: 'list' } });
    connectors.result.connectors.forEach(conn => {
      console.log(`  • ${conn.name} (${conn.type}) - ${conn.status}`);
    });
    console.log();

    // 4. Test search (returns empty results since no data indexed)
    console.log('🔍 Search Test:');
    const search = await callMCP('tools/call', { name: 'context_search', arguments: { query: 'test query' } });
    console.log(`  Found ${search.result.totalCount} results in ${search.result.queryTime}ms\n`);

    console.log('✅ All tests completed successfully!');
    console.log('\n💡 Next steps:');
    console.log('  1. Index a codebase: Implement file indexing');
    console.log('  2. Add MCP client: Use with Claude Desktop or VS Code');
    console.log('  3. Test real queries: Search through actual code');

  } catch (error) {
    console.error('❌ Test failed:', error.message);
  }
}

testConexus();
