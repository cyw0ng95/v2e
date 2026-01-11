#!/bin/bash
# Example: Demonstrating cross-service RPC calls via broker

echo "=== Broker Message Passing Demo ==="
echo ""
echo "This demo shows how the broker routes RPC messages between services."
echo ""

# Step 1: Start broker with cve-remote and cve-local services
echo "Step 1: Building binaries..."
./build.sh -p > /dev/null 2>&1

if [ $? -ne 0 ]; then
    echo "Failed to build binaries. Please run './build.sh -p' manually."
    exit 1
fi

echo "✓ Binaries built"
echo ""

# Create a temporary config for the demo
cat > /tmp/demo-config.json <<EOF
{
  "server": {
    "address": "0.0.0.0:8080"
  },
  "broker": {
    "logs_dir": "./logs",
    "processes": [
      {
        "id": "cve-remote",
        "command": "$(pwd)/.build/package/cve-remote",
        "args": [],
        "rpc": true,
        "restart": true,
        "max_restarts": -1
      },
      {
        "id": "cve-local",
        "command": "$(pwd)/.build/package/cve-local",
        "args": [],
        "rpc": true,
        "restart": true,
        "max_restarts": -1
      }
    ]
  },
  "logging": {
    "level": "info",
    "dir": "./logs"
  }
}
EOF

echo "Step 2: Starting broker with services..."
./.build/package/broker /tmp/demo-config.json > /tmp/broker.log 2>&1 &
BROKER_PID=$!

# Wait for services to start
sleep 3

if ! kill -0 $BROKER_PID 2>/dev/null; then
    echo "Failed to start broker. Check /tmp/broker.log for details."
    exit 1
fi

echo "✓ Broker started (PID: $BROKER_PID)"
echo "✓ Services: cve-remote, cve-local"
echo ""

# Step 3: List processes
echo "Step 3: Listing managed processes..."
echo '{"type":"request","id":"RPCListProcesses","payload":{}}' | \
    nc -q 1 localhost 8080 2>/dev/null || \
    echo "Note: Direct nc connection to broker not available in this demo"
echo ""

# Step 4: Demonstrate cross-service RPC call
echo "Step 4: Cross-service RPC call example"
echo ""
echo "Command:"
echo '  echo '"'"'{"type":"request","id":"RPCInvoke","payload":{"target":"cve-local","method":"RPCIsCVEStoredByID","payload":{"cve_id":"CVE-2021-44228"}}}'"'"' | broker'
echo ""
echo "This demonstrates:"
echo "  1. Client sends RPCInvoke request to broker"
echo "  2. Broker routes the message to cve-local service"
echo "  3. cve-local processes the request and responds"
echo "  4. Broker forwards the response back to client"
echo ""

# Cleanup
echo "Cleaning up..."
kill $BROKER_PID 2>/dev/null
wait $BROKER_PID 2>/dev/null
rm -f /tmp/demo-config.json

echo ""
echo "=== Demo Complete ==="
echo ""
echo "For more information, see the README.md section on 'Cross-Service RPC Calls'"
