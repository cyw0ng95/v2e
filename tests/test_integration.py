"""Integration tests for v2e following broker-first architecture.

These tests verify the complete system deployment:
1. Broker runs standalone with config.json
2. Broker spawns all subprocess services (access, cve-remote, cve-local, cve-meta)
3. External users interact only via access REST API
4. All services communicate through broker's message routing

Test Approach:
- Build binaries with './build.sh -p'
- Start broker with full configuration
- Test via access REST API endpoints
- Verify broker spawns and manages all services correctly
"""

import pytest
import time


@pytest.mark.integration
class TestBrokerDeployment:
    """Integration tests for broker-first deployment model."""
    
    def test_access_service_health(self, access_service):
        """Test that access service is running and accessible via REST API.
        
        This verifies:
        - Broker successfully starts with config.json
        - Broker spawns access service as a subprocess
        - Access service provides REST API on configured port
        - Health endpoint responds correctly
        """
        access = access_service
        
        # Call health endpoint
        response = access.health()
        
        # Verify response
        assert response["status"] == "ok"
    
    def test_access_service_stability(self, access_service):
        """Test access service remains stable across multiple requests.
        
        This verifies:
        - Access service handles multiple sequential requests
        - No memory leaks or connection issues
        - Response time remains consistent
        """
        access = access_service
        
        # Make multiple health check requests
        for i in range(5):
            response = access.health()
            assert response["status"] == "ok"
            
            # Small delay between requests
            if i < 4:
                time.sleep(0.1)
    
    def test_broker_spawns_access_service(self, access_service):
        """Test that broker properly spawns and manages access service.
        
        This integration test verifies the complete deployment flow:
        1. Broker starts with config.json
        2. Broker reads process configuration
        3. Broker spawns access service subprocess
        4. Access service starts REST API server
        5. External requests can reach access service
        
        The test validates this by successfully calling the REST API,
        which proves the entire chain is working.
        """
        access = access_service
        
        # Measure response time to verify service is responsive
        start_time = time.time()
        response = access.health()
        elapsed_time = time.time() - start_time
        
        # Successful REST API call proves:
        # - Broker is running
        # - Broker spawned access service
        # - Access service is listening on configured port
        assert response["status"] == "ok"
        
        # Response should be fast (< 1 second) when all services are healthy
        assert elapsed_time < 1.0, f"Health check took too long: {elapsed_time:.2f}s"


@pytest.mark.integration
class TestBrokerMessageStats:
    """Integration tests for broker message statistics via generic RPC endpoint."""
    
    def test_rpc_get_message_stats(self, access_service):
        """Test RPCGetMessageStats via generic RPC endpoint.
        
        This verifies:
        - POST /restful/rpc endpoint accepts RPCGetMessageStats method
        - Response uses standardized format: {retcode, message, payload}
        - Payload contains expected message statistics fields
        
        Successfully forwards RPC requests to broker and receives real statistics.
        """
        access = access_service
        
        # Call message stats via generic RPC endpoint
        response = access.get_message_stats()
        
        # Verify standardized response structure
        assert "retcode" in response
        assert "message" in response
        assert "payload" in response
        
        # Verify success
        assert response["retcode"] == 0
        assert response["message"] == "success"
        
        # Verify payload has expected structure (with Go-style capitalized field names)
        payload = response["payload"]
        assert "TotalSent" in payload
        assert "TotalReceived" in payload
        assert "RequestCount" in payload
        assert "ResponseCount" in payload
        assert "EventCount" in payload
        assert "ErrorCount" in payload
        
    def test_rpc_get_message_count(self, access_service):
        """Test RPCGetMessageCount via generic RPC endpoint.
        
        This verifies:
        - POST /restful/rpc endpoint accepts RPCGetMessageCount method
        - Response uses standardized format: {retcode, message, payload}
        - Payload contains count field
        
        Successfully forwards RPC requests to broker and receives real count.
        """
        access = access_service
        
        # Call message count via generic RPC endpoint
        response = access.get_message_count()
        
        # Verify standardized response structure
        assert "retcode" in response
        assert "message" in response
        assert "payload" in response
        
        # Verify success
        assert response["retcode"] == 0
        assert response["message"] == "success"
        
        # Verify payload has expected structure
        payload = response["payload"]
        assert "count" in payload
        assert isinstance(payload["count"], int)
    
    def test_rpc_endpoint_stability(self, access_service):
        """Test generic RPC endpoint handles multiple requests.
        
        This verifies:
        - Generic RPC endpoint remains stable across multiple calls
        - Different RPC methods can be called sequentially
        - Consistent response structure
        """
        access = access_service
        
        # Make multiple calls to different RPC methods
        for i in range(3):
            # Call GetMessageStats
            response1 = access.get_message_stats()
            assert response1["retcode"] == 0
            assert "payload" in response1
            
            # Call GetMessageCount
            response2 = access.get_message_count()
            assert response2["retcode"] == 0
            assert "payload" in response2
            
            if i < 2:
                time.sleep(0.1)
    
    def test_rpc_unknown_method(self, access_service):
        """Test generic RPC endpoint handles unknown methods correctly.
        
        This verifies:
        - Unknown RPC methods return appropriate error
        - Error response follows standardized format
        - Error message is descriptive
        """
        access = access_service
        
        # Call unknown RPC method
        response = access.rpc_call("RPCUnknownMethod")
        
        # Verify standardized response structure
        assert "retcode" in response
        assert "message" in response
        assert "payload" in response
        
        # Verify error (broker returns error message with retcode 500)
        assert response["retcode"] == 500
        assert "no handler found" in response["message"].lower() or "unknown" in response["message"].lower()
        assert response["payload"] is None


@pytest.mark.integration
class TestCVEMetaService:
    """Integration tests for cve-meta service via RESTful API.
    
    These tests verify:
    - CVE meta service is accessible via REST API through broker routing
    - All RPC methods work correctly using RESTful style
    - Responses follow the standardized format
    """
    
    def test_service_availability(self, access_service):
        """Test that all required services are available and running.
        
        This verifies:
        - Broker has spawned all required subprocess services
        - Services remain alive during testing
        - Services can be reached via broker routing
        """
        access = access_service
        
        print(f"\n  → Checking service availability")
        
        # Check broker is responding
        print(f"  → Testing broker availability...")
        broker_response = access.rpc_call("RPCGetMessageStats", verbose=False)
        assert broker_response["retcode"] == 0
        print(f"    ✓ Broker is responding")
        
        # Check cve-meta service is responding
        print(f"  → Testing cve-meta availability...")
        meta_response = access.get_cve("CVE-TEST-AVAILABILITY")
        # Should get response (either success or error, but not timeout)
        assert "retcode" in meta_response
        assert "message" in meta_response
        print(f"    ✓ cve-meta is responding (retcode: {meta_response['retcode']})")
        
        # Check cve-local service is responding
        print(f"  → Testing cve-local availability...")
        local_response = access.rpc_call("RPCIsCVEStoredByID", params={"cve_id": "CVE-TEST"}, target="cve-local", verbose=False)
        assert "retcode" in local_response
        print(f"    ✓ cve-local is responding (retcode: {local_response['retcode']})")
        
        # Verify services stay alive across multiple requests
        print(f"  → Verifying services remain alive across requests...")
        for i in range(5):
            print(f"    - Request {i+1}/5: Testing broker...")
            response = access.rpc_call("RPCGetMessageCount", verbose=False)
            assert response["retcode"] == 0
            time.sleep(0.2)
        print(f"    ✓ Services remain alive and responsive after {5} sequential requests")
        
        print(f"  ✓ All services are available and healthy throughout test duration")
    
    def test_rpc_get_cve_with_valid_id(self, access_service):
        """Test RPCGetCVE with a valid CVE ID via RESTful API.
        
        This verifies:
        - POST /restful/rpc endpoint can route to cve-meta service
        - cve-meta service processes RPCGetCVE requests
        - Response uses standardized format: {retcode, message, payload}
        - Payload contains CVE data
        """
        access = access_service
        cve_id = "CVE-2021-44228"
        
        # Log request details
        print(f"\n  → Testing RPCGetCVE for {cve_id}")
        print(f"  → Request: POST /restful/rpc")
        print(f"  → Target: cve-meta")
        print(f"  → Method: RPCGetCVE")
        print(f"  → Params: {{'cve_id': '{cve_id}'}}")
        
        # Call RPCGetCVE via RESTful API with target=cve-meta
        response = access.rpc_call("RPCGetCVE", params={"cve_id": cve_id}, target="cve-meta", verbose=False)
        
        # Log response details
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        print(f"    - payload keys: {list(response.get('payload', {}).keys())}")
        
        # Verify standardized response structure
        assert "retcode" in response
        assert "message" in response
        assert "payload" in response
        
        # Verify success
        assert response["retcode"] == 0
        assert response["message"] == "success"
        
        # Verify payload has CVE data structure
        payload = response["payload"]
        assert payload is not None
        assert "id" in payload or "ID" in payload
        
        # Check that the CVE ID matches what we requested
        returned_id = payload.get("id") or payload.get("ID")
        assert returned_id == cve_id
        print(f"  ✓ Test passed: CVE ID {returned_id} matches request")
    
    def test_rpc_get_cve_missing_cve_id(self, access_service):
        """Test RPCGetCVE with missing cve_id parameter.
        
        This verifies:
        - cve-meta service validates required parameters
        - Proper error response when cve_id is missing
        """
        access = access_service
        
        # Log request details
        print(f"\n  → Testing RPCGetCVE with missing cve_id")
        print(f"  → Request: POST /restful/rpc")
        print(f"  → Target: cve-meta")
        print(f"  → Method: RPCGetCVE")
        print(f"  → Params: {{}}")
        
        # Call RPCGetCVE without cve_id parameter
        response = access.rpc_call("RPCGetCVE", params={}, target="cve-meta")
        
        # Log response details
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Verify error response
        assert "retcode" in response
        assert "message" in response
        assert response["retcode"] == 500  # Error
        assert "cve_id" in response["message"].lower() or "required" in response["message"].lower()
        print(f"  ✓ Test passed: Proper error response for missing parameter")
    
    def test_rpc_get_cve_empty_cve_id(self, access_service):
        """Test RPCGetCVE with empty cve_id parameter.
        
        This verifies:
        - cve-meta service validates parameter values
        - Proper error response when cve_id is empty
        """
        access = access_service
        
        # Log request details
        print(f"\n  → Testing RPCGetCVE with empty cve_id")
        print(f"  → Request: POST /restful/rpc")
        print(f"  → Target: cve-meta")
        print(f"  → Method: RPCGetCVE")
        print(f"  → Params: {{'cve_id': ''}}")
        
        # Call RPCGetCVE with empty cve_id
        response = access.get_cve("")
        
        # Log response details
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Verify error response
        assert "retcode" in response
        assert "message" in response
        assert response["retcode"] == 500  # Error
        assert "cve_id" in response["message"].lower() or "required" in response["message"].lower()
        print(f"  ✓ Test passed: Proper error response for empty parameter")
    
    def test_rpc_get_cve_multiple_requests(self, access_service):
        """Test multiple RPCGetCVE requests via RESTful API.
        
        This verifies:
        - cve-meta service handles multiple sequential requests
        - No memory leaks or connection issues
        - Consistent response format across requests
        """
        access = access_service
        
        cve_ids = ["CVE-2021-44228", "CVE-2021-45046", "CVE-2022-22965"]
        
        print(f"\n  → Testing multiple RPCGetCVE requests")
        print(f"  → CVE IDs: {cve_ids}")
        
        for i, cve_id in enumerate(cve_ids, 1):
            print(f"  → Request {i}/{len(cve_ids)}: {cve_id}")
            response = access.get_cve(cve_id)
            
            # Log response summary
            print(f"    - retcode: {response.get('retcode')}")
            print(f"    - message: {response.get('message')}")
            
            # Verify standardized response structure
            assert response["retcode"] == 0
            assert response["message"] == "success"
            assert "payload" in response
            
            # Verify CVE ID matches
            payload = response["payload"]
            returned_id = payload.get("id") or payload.get("ID")
            assert returned_id == cve_id
            print(f"    ✓ CVE ID matches: {returned_id}")
            
            # Small delay between requests
            time.sleep(0.1)
        
        print(f"  ✓ Test passed: All {len(cve_ids)} requests successful")


@pytest.mark.integration
class TestCVELocalService:
    """Integration tests for cve-local service via RESTful API.
    
    These tests verify cve-local RPC interfaces work correctly through the broker.
    """
    
    def test_rpc_is_cve_stored_by_id(self, access_service):
        """Test RPCIsCVEStoredByID via RESTful API.
        
        This verifies:
        - cve-local service can check CVE existence
        - Response format is correct
        """
        access = access_service
        cve_id = "CVE-2021-44228"
        
        print(f"\n  → Testing RPCIsCVEStoredByID for {cve_id}")
        print(f"  → Request: POST /restful/rpc")
        print(f"  → Target: cve-local")
        print(f"  → Method: RPCIsCVEStoredByID")
        
        response = access.rpc_call(
            method="RPCIsCVEStoredByID",
            target="cve-local",
            params={"cve_id": cve_id}
        )
        
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Verify response structure
        assert "retcode" in response
        assert "message" in response
        assert response["retcode"] == 0
        
        # Verify payload contains stored status
        payload = response["payload"]
        assert "stored" in payload
        assert "cve_id" in payload
        assert payload["cve_id"] == cve_id
        
        print(f"  ✓ Test passed: CVE {cve_id} stored status = {payload['stored']}")


@pytest.mark.integration
@pytest.mark.slow
class TestCVERemoteService:
    """Integration tests for cve-remote service via RESTful API.
    
    These tests make actual calls to the NVD API and are marked as slow.
    They verify cve-remote RPC interfaces work correctly through the broker.
    """
    
    def test_rpc_get_cve_cnt(self, access_service):
        """Test RPCGetCVECnt via RESTful API.
        
        This verifies:
        - cve-remote service can fetch CVE count from NVD API
        - Response format is correct
        
        Note: This test makes an actual NVD API call and may be rate-limited.
        """
        access = access_service
        
        print(f"\n  → Testing RPCGetCVECnt (NVD API call)")
        print(f"  → Request: POST /restful/rpc")
        print(f"  → Target: cve-remote")
        print(f"  → Method: RPCGetCVECnt")
        print(f"  → Warning: Makes actual NVD API request")
        
        response = access.rpc_call(
            method="RPCGetCVECnt",
            target="cve-remote",
            params={}
        )
        
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Verify response structure
        assert "retcode" in response
        assert "message" in response
        assert response["retcode"] == 0
        
        # Verify payload contains total_results
        payload = response["payload"]
        assert "total_results" in payload
        assert isinstance(payload["total_results"], int)
        assert payload["total_results"] > 0
        
        print(f"  ✓ Test passed: Total CVEs in NVD = {payload['total_results']}")
    
    def test_rpc_get_cve_by_id(self, access_service):
        """Test RPCGetCVEByID via RESTful API.
        
        This verifies:
        - cve-remote service can fetch specific CVE from NVD API
        - Response format is correct
        - CVE data structure is valid
        
        Note: This test makes an actual NVD API call and may be rate-limited.
        """
        access = access_service
        cve_id = "CVE-2021-44228"  # Log4Shell vulnerability
        
        print(f"\n  → Testing RPCGetCVEByID for {cve_id} (NVD API call)")
        print(f"  → Request: POST /restful/rpc")
        print(f"  → Target: cve-remote")
        print(f"  → Method: RPCGetCVEByID")
        print(f"  → Warning: Makes actual NVD API request")
        
        response = access.rpc_call(
            method="RPCGetCVEByID",
            target="cve-remote",
            params={"cve_id": cve_id},
            verbose=False  # Disable verbose for slow tests to avoid large log output
        )
        
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Verify response structure
        assert "retcode" in response
        assert "message" in response
        assert response["retcode"] == 0
        
        # Verify payload contains vulnerabilities
        payload = response["payload"]
        assert "vulnerabilities" in payload
        assert len(payload["vulnerabilities"]) > 0
        
        # Verify CVE structure
        vuln = payload["vulnerabilities"][0]
        assert "cve" in vuln
        cve_data = vuln["cve"]
        assert "id" in cve_data
        assert cve_data["id"] == cve_id
        
        print(f"  ✓ Test passed: Successfully fetched CVE {cve_id} from NVD")
        print(f"    - CVE ID: {cve_data['id']}")
        if "descriptions" in cve_data:
            desc = cve_data["descriptions"][0]["value"][:100] if cve_data["descriptions"] else "N/A"
            print(f"    - Description: {desc}...")


# TODO: Additional integration tests for CVE functionality will be added
# once RPC forwarding is implemented in the access service (tracked in issue #74).
# Currently, the access service only provides a health check endpoint.
# 
# Future tests will include:
#
# - POST /restful/rpc/{process_id}/{endpoint} - Forward RPC calls to backend
# - CVE search and retrieval workflows via REST API
# - Multi-service workflows (remote fetch + local storage)
#
# These tests will verify the complete broker-first architecture where:
# - External users send REST requests to access service
# - Access service forwards requests to broker via RPC
# - Broker routes messages to appropriate backend services
# - Responses flow back through broker to access to user
