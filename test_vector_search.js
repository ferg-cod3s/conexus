const http = require('http');

// Test vector search with GitHub source type filter
const requestData = {
  jsonrpc: "2.0",
  id: 1,
  method: "tools/call",
  params: {
    name: "context.search",
    arguments: {
      query: "authentication",
      filters: {
        source_types: ["github_issue"]
      },
      limit: 5
    }
  }
};

const options = {
  hostname: 'localhost',
  port: 8081,
  path: '/mcp',
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Content-Length': Buffer.byteLength(JSON.stringify(requestData))
  }
};

const req = http.request(options, (res) => {
  console.log(`Status: ${res.statusCode}`);
  console.log(`Headers:`, res.headers);

  let data = '';
  res.on('data', (chunk) => {
    data += chunk;
  });

  res.on('end', () => {
    try {
      const response = JSON.parse(data);
      console.log('Response:', JSON.stringify(response, null, 2));
    } catch (e) {
      console.log('Raw response:', data);
    }
  });
});

req.on('error', (e) => {
  console.error(`Problem with request: ${e.message}`);
});

req.write(JSON.stringify(requestData));
req.end();
