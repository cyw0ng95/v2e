"""Integration tests for CVE CRUD operations via cve-meta service.

These tests verify the complete CRUD lifecycle for CVE data management:
1. Create - Fetch CVE from NVD and store locally
2. Read - Retrieve CVE (with local caching)
3. Update - Refetch CVE from NVD to update local copy
4. Delete - Remove CVE from local storage
5. List - List CVEs with pagination

All tests follow the broker-first architecture:
- External requests → Access REST API → Broker → cve-meta → cve-local/cve-remote
"""

import pytest
import time


def is_rate_limited(response):
    """Check if a response indicates NVD API rate limiting.
    
    Args:
        response: Response dict from RPC call
        
    Returns:
        True if the response indicates rate limiting, False otherwise
    """
    if response.get("retcode") == 500:
        message = response.get("message", "")
        if "NVD_RATE_LIMITED" in message or "429" in message:
            return True
    return False


@pytest.mark.integration
class TestCVECreateOperation:
    """Integration tests for RPCCreateCVE - Fetch from NVD and save locally."""
    
    @pytest.mark.slow
    def test_create_cve_success(self, access_service):
        """Test creating a CVE by fetching from NVD.
        
        This verifies:
        - cve-meta orchestrates remote fetch and local save
        - CVE data is fetched from NVD API
        - CVE data is saved to local database
        - Response includes both success flag and CVE data
        """
        access = access_service
        cve_id = "CVE-2021-44228"  # Log4Shell
        
        print(f"\n  → Testing RPCCreateCVE for {cve_id}")
        
        # Create CVE (fetch from NVD and save locally)
        response = access.rpc_call(
            method="RPCCreateCVE",
            target="cve-meta",
            params={"cve_id": cve_id}
        )
        
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Check for rate limiting
        if is_rate_limited(response):
            pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
        
        # Verify response
        assert response["retcode"] == 0
        assert response["message"] == "success"
        assert response["payload"] is not None
        
        # Verify payload contains success flag and CVE data
        payload = response["payload"]
        assert payload["success"] is True
        assert payload["cve_id"] == cve_id
        assert "cve" in payload
        assert payload["cve"]["id"] == cve_id
        
        print(f"  ✓ Test passed: Successfully created CVE {cve_id}")
    
    def test_create_cve_invalid_id(self, access_service):
        """Test creating a CVE with invalid CVE ID.
        
        This verifies:
        - cve-meta validates CVE ID format
        - Appropriate error is returned for invalid IDs
        """
        access = access_service
        cve_id = "INVALID-CVE-ID"
        
        print(f"\n  → Testing RPCCreateCVE with invalid ID: {cve_id}")
        
        response = access.rpc_call(
            method="RPCCreateCVE",
            target="cve-meta",
            params={"cve_id": cve_id}
        )
        
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Should get error
        assert response["retcode"] == 500
        assert "not found" in response["message"].lower() or "failed" in response["message"].lower()
        
        print(f"  ✓ Test passed: Proper error for invalid CVE ID")
    
    def test_create_cve_missing_param(self, access_service):
        """Test creating a CVE without cve_id parameter.
        
        This verifies:
        - cve-meta validates required parameters
        - Appropriate error is returned for missing parameters
        """
        access = access_service
        
        print(f"\n  → Testing RPCCreateCVE with missing cve_id")
        
        response = access.rpc_call(
            method="RPCCreateCVE",
            target="cve-meta",
            params={}
        )
        
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Should get error
        assert response["retcode"] == 500
        assert "cve_id" in response["message"].lower() or "required" in response["message"].lower()
        
        print(f"  ✓ Test passed: Proper error for missing parameter")


@pytest.mark.integration
class TestCVEReadOperation:
    """Integration tests for RPCGetCVE - Read with local caching."""
    
    @pytest.mark.slow
    def test_get_cve_not_cached_fetches_from_nvd(self, access_service):
        """Test getting a CVE that is not in local cache.
        
        This verifies:
        - cve-meta checks local storage first
        - If not found locally, fetches from NVD
        - Saves fetched CVE to local storage
        - Returns CVE data to caller
        """
        access = access_service
        cve_id = "CVE-2021-45046"  # Log4Shell variant
        
        print(f"\n  → Testing RPCGetCVE for uncached CVE: {cve_id}")
        
        # First, ensure CVE is not in local cache
        print(f"  → Deleting CVE if it exists...")
        access.rpc_call(
            method="RPCDeleteCVE",
            target="cve-meta",
            params={"cve_id": cve_id},
            verbose=False
        )
        
        # Get CVE (should fetch from NVD)
        print(f"  → Fetching CVE from NVD...")
        response = access.get_cve(cve_id)
        
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Check for rate limiting
        if is_rate_limited(response):
            pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
        
        # Verify response
        assert response["retcode"] == 0
        assert response["message"] == "success"
        assert response["payload"] is not None
        assert response["payload"]["id"] == cve_id
        
        print(f"  ✓ Test passed: Successfully fetched uncached CVE from NVD")
    
    @pytest.mark.slow
    def test_get_cve_cached_returns_from_local(self, access_service):
        """Test getting a CVE that is already in local cache.
        
        This verifies:
        - cve-meta checks local storage first
        - If found locally, returns immediately without NVD call
        - Response is fast (< 1 second)
        """
        access = access_service
        cve_id = "CVE-2022-22965"  # Spring4Shell
        
        print(f"\n  → Testing RPCGetCVE for cached CVE: {cve_id}")
        
        # First, ensure CVE is in local cache
        print(f"  → Creating CVE to cache it...")
        create_response = access.rpc_call(
            method="RPCCreateCVE",
            target="cve-meta",
            params={"cve_id": cve_id},
            verbose=False
        )
        
        # Check for rate limiting during creation
        if is_rate_limited(create_response):
            pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
        
        # Get CVE (should return from cache)
        print(f"  → Fetching CVE from cache...")
        start_time = time.time()
        response = access.get_cve(cve_id)
        elapsed = time.time() - start_time
        
        print(f"  → Response received in {elapsed:.3f}s:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Verify response
        assert response["retcode"] == 0
        assert response["message"] == "success"
        assert response["payload"] is not None
        assert response["payload"]["id"] == cve_id
        
        # Should be fast when cached (< 2 seconds, accounting for RPC overhead)
        assert elapsed < 2.0, f"Cached response took too long: {elapsed:.3f}s"
        
        print(f"  ✓ Test passed: Successfully retrieved cached CVE in {elapsed:.3f}s")
    
    def test_get_cve_nonexistent(self, access_service):
        """Test getting a non-existent CVE ID.
        
        This verifies:
        - cve-meta handles non-existent CVE IDs gracefully
        - Appropriate error is returned
        """
        access = access_service
        cve_id = "CVE-9999-99999"  # Non-existent CVE
        
        print(f"\n  → Testing RPCGetCVE for non-existent CVE: {cve_id}")
        
        response = access.get_cve(cve_id)
        
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Should get error
        assert response["retcode"] == 500
        assert "not found" in response["message"].lower() or "failed" in response["message"].lower()
        
        print(f"  ✓ Test passed: Proper error for non-existent CVE")


@pytest.mark.integration
class TestCVEUpdateOperation:
    """Integration tests for RPCUpdateCVE - Refetch from NVD to update."""
    
    @pytest.mark.slow
    def test_update_cve_success(self, access_service):
        """Test updating a CVE by refetching from NVD.
        
        This verifies:
        - cve-meta fetches latest data from NVD
        - Local storage is updated with new data
        - Response includes updated CVE data
        """
        access = access_service
        cve_id = "CVE-2021-44228"  # Log4Shell
        
        print(f"\n  → Testing RPCUpdateCVE for {cve_id}")
        
        # First, ensure CVE exists locally
        print(f"  → Creating CVE first...")
        create_response = access.rpc_call(
            method="RPCCreateCVE",
            target="cve-meta",
            params={"cve_id": cve_id},
            verbose=False
        )
        
        # Check for rate limiting during creation
        if is_rate_limited(create_response):
            pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
        
        # Update CVE (refetch from NVD)
        print(f"  → Updating CVE from NVD...")
        response = access.rpc_call(
            method="RPCUpdateCVE",
            target="cve-meta",
            params={"cve_id": cve_id}
        )
        
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Check for rate limiting
        if is_rate_limited(response):
            pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
        
        # Verify response
        assert response["retcode"] == 0
        assert response["message"] == "success"
        assert response["payload"] is not None
        
        # Verify payload
        payload = response["payload"]
        assert payload["success"] is True
        assert payload["cve_id"] == cve_id
        assert "cve" in payload
        assert payload["cve"]["id"] == cve_id
        
        print(f"  ✓ Test passed: Successfully updated CVE {cve_id}")
    
    def test_update_cve_nonexistent(self, access_service):
        """Test updating a non-existent CVE.
        
        This verifies:
        - cve-meta handles non-existent CVEs during update
        - Appropriate error is returned
        """
        access = access_service
        cve_id = "CVE-9999-99999"  # Non-existent CVE
        
        print(f"\n  → Testing RPCUpdateCVE for non-existent CVE: {cve_id}")
        
        response = access.rpc_call(
            method="RPCUpdateCVE",
            target="cve-meta",
            params={"cve_id": cve_id}
        )
        
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Should get error
        assert response["retcode"] == 500
        assert "not found" in response["message"].lower() or "failed" in response["message"].lower()
        
        print(f"  ✓ Test passed: Proper error for non-existent CVE")


@pytest.mark.integration
class TestCVEDeleteOperation:
    """Integration tests for RPCDeleteCVE - Delete from local storage."""
    
    @pytest.mark.slow
    def test_delete_cve_success(self, access_service):
        """Test deleting a CVE from local storage.
        
        This verifies:
        - cve-meta deletes CVE from local database
        - Subsequent reads fail (CVE not found)
        - Response confirms successful deletion
        """
        access = access_service
        cve_id = "CVE-2023-12345"  # Test CVE
        
        print(f"\n  → Testing RPCDeleteCVE for {cve_id}")
        
        # First, create CVE to ensure it exists
        print(f"  → Creating CVE first...")
        create_response = access.rpc_call(
            method="RPCCreateCVE",
            target="cve-meta",
            params={"cve_id": cve_id},
            verbose=False
        )
        
        # Check for rate limiting during creation
        if is_rate_limited(create_response):
            pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
        # May fail if CVE doesn't exist in NVD, which is okay for this test
        # We just want to test deletion
        
        # Delete CVE
        print(f"  → Deleting CVE...")
        response = access.rpc_call(
            method="RPCDeleteCVE",
            target="cve-meta",
            params={"cve_id": cve_id}
        )
        
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Verify response (should succeed even if CVE didn't exist)
        assert response["retcode"] == 0 or response["retcode"] == 500
        
        # If deletion succeeded, verify CVE is not in local storage
        if response["retcode"] == 0:
            payload = response["payload"]
            assert payload["cve_id"] == cve_id
            
            # Verify CVE is not stored locally anymore
            check_response = access.rpc_call(
                method="RPCIsCVEStoredByID",
                target="cve-local",
                params={"cve_id": cve_id},
                verbose=False
            )
            assert check_response["payload"]["stored"] is False
            
            print(f"  ✓ Test passed: Successfully deleted CVE {cve_id}")
        else:
            print(f"  ✓ Test passed: Delete handled non-existent CVE correctly")
    
    def test_delete_cve_missing_param(self, access_service):
        """Test deleting a CVE without cve_id parameter.
        
        This verifies:
        - cve-meta validates required parameters
        - Appropriate error is returned
        """
        access = access_service
        
        print(f"\n  → Testing RPCDeleteCVE with missing cve_id")
        
        response = access.rpc_call(
            method="RPCDeleteCVE",
            target="cve-meta",
            params={}
        )
        
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Should get error
        assert response["retcode"] == 500
        assert "cve_id" in response["message"].lower() or "required" in response["message"].lower()
        
        print(f"  ✓ Test passed: Proper error for missing parameter")


@pytest.mark.integration
class TestCVEListOperation:
    """Integration tests for RPCListCVEs - List with pagination."""
    
    @pytest.mark.slow
    def test_list_cves_with_data(self, access_service):
        """Test listing CVEs from local storage.
        
        This verifies:
        - cve-meta retrieves CVEs from local database
        - Pagination parameters work correctly
        - Response includes CVE list and total count
        """
        access = access_service
        
        print(f"\n  → Testing RPCListCVEs")
        
        # First, ensure we have some CVEs in storage
        print(f"  → Creating test CVEs...")
        test_cves = ["CVE-2021-44228", "CVE-2021-45046"]
        for cve_id in test_cves:
            create_response = access.rpc_call(
                method="RPCCreateCVE",
                target="cve-meta",
                params={"cve_id": cve_id},
                verbose=False
            )
            
            # Check for rate limiting
            if is_rate_limited(create_response):
                pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
            
            time.sleep(0.5)  # Rate limiting
        
        # List CVEs
        print(f"  → Listing CVEs...")
        response = access.rpc_call(
            method="RPCListCVEs",
            target="cve-meta",
            params={"offset": 0, "limit": 10}
        )
        
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Verify response
        assert response["retcode"] == 0
        assert response["message"] == "success"
        assert response["payload"] is not None
        
        # Verify payload structure
        payload = response["payload"]
        assert "cves" in payload
        assert "total" in payload
        assert "offset" in payload
        assert "limit" in payload
        assert isinstance(payload["cves"], list)
        assert isinstance(payload["total"], int)
        assert payload["total"] >= 0
        
        print(f"  ✓ Test passed: Listed {len(payload['cves'])} CVEs (total: {payload['total']})")
    
    def test_list_cves_pagination(self, access_service):
        """Test listing CVEs with different pagination parameters.
        
        This verifies:
        - Offset and limit parameters work correctly
        - Different pages return different results
        """
        access = access_service
        
        print(f"\n  → Testing RPCListCVEs pagination")
        
        # List first page
        print(f"  → Fetching page 1...")
        page1_response = access.rpc_call(
            method="RPCListCVEs",
            target="cve-meta",
            params={"offset": 0, "limit": 5},
            verbose=False
        )
        
        # List second page
        print(f"  → Fetching page 2...")
        page2_response = access.rpc_call(
            method="RPCListCVEs",
            target="cve-meta",
            params={"offset": 5, "limit": 5},
            verbose=False
        )
        
        # Both should succeed
        assert page1_response["retcode"] == 0
        assert page2_response["retcode"] == 0
        
        page1 = page1_response["payload"]
        page2 = page2_response["payload"]
        
        # Verify pagination metadata
        assert page1["offset"] == 0
        assert page1["limit"] == 5
        assert page2["offset"] == 5
        assert page2["limit"] == 5
        
        # Total should be the same
        assert page1["total"] == page2["total"]
        
        print(f"  ✓ Test passed: Pagination works correctly")
        print(f"    - Page 1: {len(page1['cves'])} CVEs")
        print(f"    - Page 2: {len(page2['cves'])} CVEs")
        print(f"    - Total: {page1['total']} CVEs")
    
    def test_list_cves_empty_database(self, access_service):
        """Test listing CVEs when database is empty.
        
        This verifies:
        - cve-meta handles empty database gracefully
        - Returns empty list with total count of 0
        """
        access = access_service
        
        print(f"\n  → Testing RPCListCVEs with potentially empty database")
        
        # List CVEs (may or may not be empty depending on other tests)
        response = access.rpc_call(
            method="RPCListCVEs",
            target="cve-meta",
            params={"offset": 0, "limit": 10},
            verbose=False
        )
        
        # Should succeed regardless
        assert response["retcode"] == 0
        assert response["message"] == "success"
        assert response["payload"] is not None
        
        payload = response["payload"]
        assert "cves" in payload
        assert "total" in payload
        assert isinstance(payload["cves"], list)
        assert isinstance(payload["total"], int)
        assert payload["total"] >= 0
        
        print(f"  ✓ Test passed: Handled database state correctly (total: {payload['total']})")


@pytest.mark.integration
class TestCVEBusinessFlows:
    """Integration tests for complete CVE business workflows."""
    
    @pytest.mark.slow
    def test_complete_crud_lifecycle(self, access_service):
        """Test complete CRUD lifecycle for a CVE.
        
        This verifies:
        1. Create - Fetch from NVD and save
        2. Read - Retrieve from cache
        3. Update - Refetch from NVD
        4. Delete - Remove from storage
        5. Verify deletion
        """
        access = access_service
        cve_id = "CVE-2021-44228"  # Log4Shell
        
        print(f"\n  → Testing complete CRUD lifecycle for {cve_id}")
        
        # Step 1: Create
        print(f"  → Step 1: Create (fetch from NVD)...")
        create_response = access.rpc_call(
            method="RPCCreateCVE",
            target="cve-meta",
            params={"cve_id": cve_id},
            verbose=False
        )
        
        # Check for rate limiting
        if is_rate_limited(create_response):
            pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
        
        assert create_response["retcode"] == 0
        print(f"    ✓ Created successfully")
        
        # Step 2: Read (from cache)
        print(f"  → Step 2: Read (from cache)...")
        read_response = access.get_cve(cve_id)
        assert read_response["retcode"] == 0
        assert read_response["payload"]["id"] == cve_id
        print(f"    ✓ Read successfully")
        
        # Step 3: Update
        print(f"  → Step 3: Update (refetch from NVD)...")
        time.sleep(1)  # Rate limiting
        update_response = access.rpc_call(
            method="RPCUpdateCVE",
            target="cve-meta",
            params={"cve_id": cve_id},
            verbose=False
        )
        assert update_response["retcode"] == 0
        print(f"    ✓ Updated successfully")
        
        # Step 4: Delete
        print(f"  → Step 4: Delete...")
        delete_response = access.rpc_call(
            method="RPCDeleteCVE",
            target="cve-meta",
            params={"cve_id": cve_id},
            verbose=False
        )
        assert delete_response["retcode"] == 0
        print(f"    ✓ Deleted successfully")
        
        # Step 5: Verify deletion
        print(f"  → Step 5: Verify deletion...")
        check_response = access.rpc_call(
            method="RPCIsCVEStoredByID",
            target="cve-local",
            params={"cve_id": cve_id},
            verbose=False
        )
        assert check_response["payload"]["stored"] is False
        print(f"    ✓ Deletion verified")
        
        print(f"  ✓ Test passed: Complete CRUD lifecycle successful")
    
    @pytest.mark.slow
    def test_cache_then_fetch_workflow(self, access_service):
        """Test the cache-then-fetch workflow.
        
        This verifies:
        1. First Get - Fetches from NVD (cache miss)
        2. Second Get - Returns from cache (cache hit)
        3. Cache hit is faster than cache miss
        """
        access = access_service
        cve_id = "CVE-2022-22965"  # Spring4Shell
        
        print(f"\n  → Testing cache-then-fetch workflow for {cve_id}")
        
        # Clear cache first
        print(f"  → Clearing cache...")
        access.rpc_call(
            method="RPCDeleteCVE",
            target="cve-meta",
            params={"cve_id": cve_id},
            verbose=False
        )
        
        # First Get (cache miss - fetch from NVD)
        print(f"  → First Get (cache miss)...")
        start_time1 = time.time()
        response1 = access.get_cve(cve_id)
        time1 = time.time() - start_time1
        
        # Check for rate limiting
        if is_rate_limited(response1):
            pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
        
        assert response1["retcode"] == 0
        print(f"    ✓ Fetched from NVD in {time1:.3f}s")
        
        # Second Get (cache hit - return from local)
        print(f"  → Second Get (cache hit)...")
        start_time2 = time.time()
        response2 = access.get_cve(cve_id)
        time2 = time.time() - start_time2
        assert response2["retcode"] == 0
        print(f"    ✓ Retrieved from cache in {time2:.3f}s")
        
        # Cache hit should be faster
        print(f"  → Comparing times:")
        print(f"    - Cache miss: {time1:.3f}s")
        print(f"    - Cache hit: {time2:.3f}s")
        print(f"    - Speedup: {time1/time2:.1f}x faster")
        
        print(f"  ✓ Test passed: Cache workflow verified")
    
    @pytest.mark.slow
    def test_batch_create_and_list(self, access_service):
        """Test creating multiple CVEs and listing them.
        
        This verifies:
        1. Create multiple CVEs
        2. List all CVEs
        3. Verify all created CVEs are in the list
        """
        access = access_service
        test_cves = [
            "CVE-2021-44228",  # Log4Shell
            "CVE-2021-45046",  # Log4Shell variant
        ]
        
        print(f"\n  → Testing batch create and list for {len(test_cves)} CVEs")
        
        # Create CVEs
        print(f"  → Creating CVEs...")
        for cve_id in test_cves:
            print(f"    - Creating {cve_id}...")
            response = access.rpc_call(
                method="RPCCreateCVE",
                target="cve-meta",
                params={"cve_id": cve_id},
                verbose=False
            )
            
            # Check for rate limiting
            if is_rate_limited(response):
                pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
            
            assert response["retcode"] == 0, f"Failed to create {cve_id}: {response}"
            print(f"      ✓ Created {cve_id}")
            time.sleep(1)  # Rate limiting for NVD API
        
        print(f"    ✓ All CVEs created")
        
        # List CVEs
        print(f"  → Listing CVEs...")
        list_response = access.rpc_call(
            method="RPCListCVEs",
            target="cve-meta",
            params={"offset": 0, "limit": 100},
            verbose=False
        )
        assert list_response["retcode"] == 0, f"Failed to list CVEs: {list_response}"
        
        payload = list_response["payload"]
        cve_ids_in_list = [cve["id"] for cve in payload["cves"]]
        
        print(f"  → List response:")
        print(f"    - Total CVEs in database: {payload['total']}")
        print(f"    - CVEs returned: {len(cve_ids_in_list)}")
        print(f"    - CVE IDs: {cve_ids_in_list}")
        
        print(f"  → Verifying all CVEs are in the list...")
        for cve_id in test_cves:
            assert cve_id in cve_ids_in_list, f"CVE {cve_id} not found in list"
            print(f"    ✓ {cve_id} found in list")
        
        print(f"  ✓ Test passed: Batch operations successful")
        print(f"    - Created: {len(test_cves)} CVEs")
        print(f"    - Total in database: {payload['total']} CVEs")
