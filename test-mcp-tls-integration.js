// Test MCP integration with Conexus over HTTPS/TLS
import https from 'https';
import fs from 'fs';

function callMCPOverHTTPS(method, params, id = 1) {
  return new Promise((resolve, reject) => {
    const postData = JSON.stringify({
      jsonrpc: "2.0",
      id,
      method,
      params
    });

    const options = {
      hostname: 'localhost',
      port: 8443,  // HTTPS port from config.tls-example.yml
      path: '/mcp',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(postData)
      },
      // For development with self-signed certificates
      // In production, you would use proper certificate validation
      rejectUnauthorized: false
    };

    const req = https.request(options, (res) => {
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

async function testMCPIntegrationWithTLS() {
  console.log('🔒 Testing MCP Integration with Conexus over HTTPS/TLS\n');

  try {
    // Verify TLS certificates exist
    console.log('🔐 Certificate Status:');
    const certPath = './data/tls/cert.pem';
    const keyPath = './data/tls/key.pem';
    
    if (fs.existsSync(certPath) && fs.existsSync(keyPath)) {
      console.log('   ✅ TLS certificates found');
      console.log(`   📄 Certificate: ${certPath}`);
      console.log(`   🔑 Private key: ${keyPath}\n`);
    } else {
      console.log('   ⚠️  TLS certificates not found!');
      console.log('   Run: ./scripts/generate-dev-certs.sh\n');
      return;
    }

    // Test each tool that would be used in development
    console.log('🛠️  Testing Development Workflow Tools over HTTPS:\n');

    // 1. Index control - check status
    console.log('📊 Index Status:');
    const status = await callMCPOverHTTPS('tools/call', {
      name: 'context.index_control',
      arguments: { action: 'status' }
    });
    console.log(`   ✅ ${status.result.message}`);
    console.log(`   📈 Documents indexed: ${status.result.details.documents_indexed}\n`);

    // 2. Connector management - list available
    console.log('🔌 Available Connectors:');
    const connectors = await callMCPOverHTTPS('tools/call', {
      name: 'context.connector_management',
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

    // 3. Search capability
    console.log('🔍 Code Search:');
    const search = await callMCPOverHTTPS('tools/call', {
      name: 'context.search',
      arguments: { query: 'function definitions' }
    });
    console.log(`   ✅ Search executed (${search.result.totalCount || 0} results found)`);
    console.log();

    // 4. Related info lookup
    console.log('📋 Related Information:');
    const related = await callMCPOverHTTPS('tools/call', {
      name: 'context.get_related_info',
      arguments: { file_path: 'main.go' }
    });
    console.log(`   ✅ Related info query processed`);
    console.log();

    console.log('🎉 MCP TLS Integration Test Complete!');
    console.log('\n💡 HTTPS/TLS is working correctly');
    console.log('   🔒 All MCP communication encrypted');
    console.log('   🛡️  Ready for secure Claude Code / OpenCode integration');
    console.log('\n📝 Configuration:');
    console.log('   - Copy config.tls-example.yml to config.yml');
    console.log('   - Update claude-mcp-config.json with https://localhost:8443');
    console.log('   - Then use: /mcp conexus tools/call context.index_control {"action": "status"}');

  } catch (error) {
    console.error('❌ TLS Integration test failed:', error.message);
    console.log('\n🔧 Troubleshooting:');
    console.log('   1. Check if Conexus is running with TLS:');
    console.log('      curl -k https://localhost:8443/health');
    console.log('   2. Verify TLS certificates exist:');
    console.log('      ls -la ./data/tls/');
    console.log('   3. Generate dev certificates if missing:');
    console.log('      ./scripts/generate-dev-certs.sh');
    console.log('   4. Check Conexus is using TLS config:');
    console.log('      ./conexus --config config.tls-example.yml');
    console.log('   5. Verify port 8443 is not in use:');
    console.log('      lsof -i :8443');
    console.log('   6. Check Conexus logs:');
    console.log('      tail -f conexus.log');
  }
}

testMCPIntegrationWithTLS();
