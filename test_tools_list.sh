#!/bin/bash

echo "Testing what tools are actually available..."

echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./bin/conexus-darwin-arm64