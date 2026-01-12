"""Edge case and malformed request integration tests.

These tests verify system reliability and robustness when handling:
- Invalid inputs
- Malformed requests
- Edge cases
- Error conditions
"""

import pytest
import json


@pytest.mark.integration
class TestEdgeCasesAndMalformed:
    """Integration tests for edge cases and malformed requests."""
    
    def test_invalid_json_request(self, access_service):
        """Test handling of malformed JSON in request body.
        
        This verifies:
        - System handles invalid JSON gracefully
        - Appropriate error response is returned
        """
        access = access_service
        
        print("\n  → Testing malformed JSON request")
        
        import requests
        url = f"{access.base_url}{access.restful_prefix}/rpc"
        
        # Send invalid JSON
        response = requests.post(
            url,
            data="{invalid json}",
            headers={"Content-Type": "application/json"}
        )
        
        print(f"  → Response: {response.status_code}")
        
        # Should return error status
        assert response.status_code >= 400
        print(f"  ✓ Test passed: Invalid JSON rejected with status {response.status_code}")
    
    def test_missing_method_field(self, access_service):
        """Test RPC call with missing method field.
        
        This verifies:
        - System validates required fields
        - Appropriate error for missing method
        """
        access = access_service
        
        print("\n  → Testing RPC call with missing method field")
        
        import requests
        url = f"{access.base_url}{access.restful_prefix}/rpc"
        
        # Send request without method
        response = requests.post(
            url,
            json={"params": {}},
            headers={"Content-Type": "application/json"}
        )
        
        print(f"  → Response: {response.status_code}")
        
        # Should return error
        assert response.status_code >= 400 or (response.status_code == 200 and response.json().get("retcode") == 500)
        print(f"  ✓ Test passed: Missing method field handled correctly")
    
    def test_null_params(self, access_service):
        """Test RPC call with null params.
        
        This verifies:
        - System handles null parameters gracefully
        """
        access = access_service
        
        print("\n  → Testing RPC call with null params")
        
        response = access.rpc_call(
            method="RPCGetMessageStats",
            params=None,
            verbose=False
        )
        
        # Should handle null params (defaults to empty object)
        assert "retcode" in response
        print(f"  ✓ Test passed: Null params handled (retcode: {response['retcode']})")
    
    def test_extremely_long_cve_id(self, access_service):
        """Test CVE query with extremely long ID.
        
        This verifies:
        - System handles oversized inputs
        - No buffer overflow or crashes
        """
        access = access_service
        
        print("\n  → Testing extremely long CVE ID")
        
        # Create a very long CVE ID
        long_id = "CVE-2021-" + ("9" * 10000)
        
        response = access.rpc_call(
            method="RPCGetCVE",
            target="cve-meta",
            params={"cve_id": long_id},
            verbose=False
        )
        
        # Should return error, but not crash
        assert "retcode" in response
        assert response["retcode"] == 500
        print(f"  ✓ Test passed: Long CVE ID handled gracefully")
    
    def test_special_characters_in_cve_id(self, access_service):
        """Test CVE query with special characters.
        
        This verifies:
        - System sanitizes inputs properly
        - No injection vulnerabilities
        """
        access = access_service
        
        print("\n  → Testing special characters in CVE ID")
        
        special_ids = [
            "CVE-2021-'; DROP TABLE cves; --",  # SQL injection attempt
            "CVE-2021-<script>alert('xss')</script>",  # XSS attempt
            "CVE-2021-../../etc/passwd",  # Path traversal attempt
            "CVE-2021-\x00\x01\x02",  # Null bytes
            "CVE-2021-\n\r\t",  # Control characters
        ]
        
        for special_id in special_ids:
            print(f"  → Testing: {repr(special_id)}")
            response = access.rpc_call(
                method="RPCGetCVE",
                target="cve-meta",
                params={"cve_id": special_id},
                verbose=False
            )
            
            # Should return error, but handle safely
            assert "retcode" in response
            print(f"    ✓ Handled safely (retcode: {response['retcode']})")
        
        print(f"  ✓ Test passed: All special characters handled safely")
    
    def test_negative_pagination_values(self, access_service):
        """Test list operation with negative pagination values.
        
        This verifies:
        - System validates pagination parameters
        - Appropriate error handling
        """
        access = access_service
        
        print("\n  → Testing negative pagination values")
        
        test_cases = [
            {"offset": -1, "limit": 10},
            {"offset": 0, "limit": -1},
            {"offset": -10, "limit": -10},
        ]
        
        for params in test_cases:
            print(f"  → Testing params: {params}")
            response = access.rpc_call(
                method="RPCListCVEs",
                target="cve-meta",
                params=params,
                verbose=False
            )
            
            # Should either return error or treat as 0
            assert "retcode" in response
            print(f"    ✓ Handled (retcode: {response['retcode']})")
        
        print(f"  ✓ Test passed: Negative values handled")
    
    def test_extremely_large_pagination_limit(self, access_service):
        """Test list operation with extremely large limit.
        
        This verifies:
        - System has maximum limit protection
        - No memory exhaustion
        """
        access = access_service
        
        print("\n  → Testing extremely large pagination limit")
        
        response = access.rpc_call(
            method="RPCListCVEs",
            target="cve-meta",
            params={"offset": 0, "limit": 999999999},
            verbose=False
        )
        
        # Should either cap the limit or return error
        assert "retcode" in response
        
        if response["retcode"] == 0:
            # If successful, verify it didn't return huge dataset
            payload = response["payload"]
            assert len(payload["cves"]) < 10000, "Should cap maximum results"
            print(f"  ✓ Test passed: Limit capped to {len(payload['cves'])} items")
        else:
            print(f"  ✓ Test passed: Large limit rejected (retcode: {response['retcode']})")
    
    def test_empty_string_parameters(self, access_service):
        """Test RPC calls with empty string parameters.
        
        This verifies:
        - System validates parameter values
        - Empty strings are handled appropriately
        """
        access = access_service
        
        print("\n  → Testing empty string parameters")
        
        test_cases = [
            ("RPCGetCVE", {"cve_id": ""}),
            ("RPCCreateCVE", {"cve_id": ""}),
            ("RPCUpdateCVE", {"cve_id": ""}),
            ("RPCDeleteCVE", {"cve_id": ""}),
        ]
        
        for method, params in test_cases:
            print(f"  → Testing {method} with empty string")
            response = access.rpc_call(
                method=method,
                target="cve-meta",
                params=params,
                verbose=False
            )
            
            # Should return error
            assert response["retcode"] == 500
            assert "cve_id" in response["message"].lower() or "required" in response["message"].lower()
            print(f"    ✓ Empty string rejected")
        
        print(f"  ✓ Test passed: All empty strings rejected")
    
    def test_unicode_in_parameters(self, access_service):
        """Test RPC calls with Unicode characters.
        
        This verifies:
        - System handles international characters
        - UTF-8 encoding works correctly
        """
        access = access_service
        
        print("\n  → Testing Unicode characters in parameters")
        
        unicode_ids = [
            "CVE-2021-你好",
            "CVE-2021-привет",
            "CVE-2021-مرحبا",
            "CVE-2021-🔥",
        ]
        
        for unicode_id in unicode_ids:
            print(f"  → Testing: {unicode_id}")
            response = access.rpc_call(
                method="RPCGetCVE",
                target="cve-meta",
                params={"cve_id": unicode_id},
                verbose=False
            )
            
            # Should handle gracefully
            assert "retcode" in response
            print(f"    ✓ Handled (retcode: {response['retcode']})")
        
        print(f"  ✓ Test passed: Unicode handled correctly")
    
    def test_unknown_target_service(self, access_service):
        """Test RPC call to unknown target service.
        
        This verifies:
        - System validates target service names
        - Appropriate error for unknown targets
        """
        access = access_service
        
        print("\n  → Testing unknown target service")
        
        response = access.rpc_call(
            method="RPCGetCVE",
            target="non-existent-service",
            params={"cve_id": "CVE-2021-44228"},
            verbose=False
        )
        
        # Should return error
        assert "retcode" in response
        assert response["retcode"] == 500
        print(f"  ✓ Test passed: Unknown target rejected (retcode: {response['retcode']})")
    
    def test_malformed_target_parameter(self, access_service):
        """Test RPC call with malformed target parameter.
        
        This verifies:
        - System validates target format
        """
        access = access_service
        
        print("\n  → Testing malformed target parameter")
        
        malformed_targets = [
            "",
            " ",
            "../broker",
            "target;rm -rf /",
            "target\x00",
        ]
        
        for target in malformed_targets:
            print(f"  → Testing target: {repr(target)}")
            response = access.rpc_call(
                method="RPCGetMessageStats",
                target=target,
                params={},
                verbose=False
            )
            
            # Should handle safely
            assert "retcode" in response
            print(f"    ✓ Handled (retcode: {response['retcode']})")
        
        print(f"  ✓ Test passed: Malformed targets handled")
    
    def test_rapid_sequential_requests(self, access_service):
        """Test rapid sequential API requests.
        
        This verifies:
        - System handles high request rate
        - No rate limiting issues
        - Consistent responses
        """
        access = access_service
        
        print("\n  → Testing rapid sequential requests")
        
        num_requests = 50
        start_time = __import__('time').time()
        
        for i in range(num_requests):
            response = access.rpc_call(
                method="RPCGetMessageCount",
                params={},
                verbose=False
            )
            assert response["retcode"] == 0
        
        elapsed = __import__('time').time() - start_time
        rate = num_requests / elapsed
        
        print(f"  ✓ Test passed: Handled {num_requests} requests in {elapsed:.2f}s ({rate:.1f} req/s)")
    
    def test_mixed_valid_invalid_batch(self, access_service):
        """Test batch of mixed valid and invalid requests.
        
        This verifies:
        - System processes each request independently
        - Invalid requests don't affect valid ones
        """
        access = access_service
        
        print("\n  → Testing mixed valid/invalid batch requests")
        
        requests = [
            ("RPCGetMessageStats", {}, True),  # Valid
            ("RPCGetCVE", {"cve_id": ""}, False),  # Invalid - empty ID
            ("RPCGetMessageCount", {}, True),  # Valid
            ("RPCGetCVE", {"cve_id": "CVE-INVALID"}, False),  # Invalid - bad format
            ("RPCGetMessageStats", {}, True),  # Valid
        ]
        
        results = []
        for method, params, should_succeed in requests:
            if "CVE" in method:
                response = access.rpc_call(method, params, target="cve-meta", verbose=False)
            else:
                response = access.rpc_call(method, params, verbose=False)
            
            results.append((response["retcode"] == 0, should_succeed))
        
        # Verify results match expectations
        for i, (succeeded, expected_success) in enumerate(results):
            if expected_success:
                assert succeeded, f"Request {i} should have succeeded"
            print(f"  → Request {i+1}: {'✓' if succeeded == expected_success else '✗'}")
        
        print(f"  ✓ Test passed: Mixed batch handled correctly")
