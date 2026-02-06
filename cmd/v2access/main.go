/*
Package main implements the access RPC service.

RPC API Specification:

Access Service
====================

Service Type: REST (HTTP/JSON)
Description: RESTful API gateway service that provides external access to the v2e system.

	Forwards RPC requests to backend services through the broker.

Available REST Endpoints:
-------------------------

 1. GET /restful/health
    Description: Health check endpoint to verify service is running
    Request Parameters: None
    Response:
    - status (string): "ok" if service is healthy
    Errors: None
    Example:
    Request:  GET /restful/health
    Response: {"status": "ok"}

 2. POST /restful/rpc
    Description: Generic RPC forwarding endpoint that routes requests to backend services
    Request Parameters:
    - method (string, required): RPC method name (e.g., "RPCGetCVE")
    - params (object, optional): Parameters to pass to the RPC method
    - target (string, optional): Target process ID (default: "broker")
    Response:
    - retcode (int): 0 for success, non-zero for errors
    - message (string): Success message or error description
    - payload (object): Response data from the backend service
    Errors:
    - Invalid JSON: retcode=400, missing or malformed request body
    - RPC timeout: retcode=500, backend service did not respond in time
    - Backend error: retcode=500, backend service returned an error
    Example:
    Request:  {"method": "RPCGetCVE", "target": "meta", "params": {"cve_id": "CVE-2021-44228"}}
    Response: {"retcode": 0, "message": "success", "payload": {"id": "CVE-2021-44228", ...}}

Notes:
------
- All RPC requests are forwarded through the broker for security and routing
- Default RPC timeout is 30 seconds
- Service runs as a subprocess managed by the broker
- Uses stdin/stdout for RPC communication with broker
- External clients access via HTTP on configured address (default: 0.0.0.0:8080)
*/
package main

func main() {
	runAccess()
}
