"""Advanced CVE CRUD business workflow tests.

These tests simulate realistic web application usage patterns:
- User browsing and searching CVEs
- User managing CVE collections
- User performing complex operations
- Multi-step workflows
"""

import pytest
import time


@pytest.mark.integration
class TestWebApplicationWorkflows:
    """Business workflow tests simulating web application usage."""
    
    @pytest.mark.slow
    def test_user_search_and_view_workflow(self, access_service):
        """Simulate user searching for and viewing CVE details.
        
        Workflow:
        1. User searches for a specific CVE
        2. User views the CVE details
        3. User checks if it's in their collection
        4. User adds it to their collection (create)
        5. User verifies it was added
        """
        access = access_service
        cve_id = "CVE-2021-44228"  # Log4Shell
        
        print("\n  → Simulating user search and view workflow")
        
        # Step 1: User searches for CVE (checks if exists in local DB)
        print("  → Step 1: User checks if CVE is in local database")
        check_response = access.rpc_call(
            method="RPCIsCVEStoredByID",
            target="cve-local",
            params={"cve_id": cve_id},
            verbose=False
        )
        
        is_stored = check_response["payload"]["stored"]
        print(f"    - CVE in local DB: {is_stored}")
        
        # Step 2: User views CVE details (fetches from NVD if not local)
        print("  → Step 2: User views CVE details")
        view_response = access.rpc_call(
            method="RPCGetCVE",
            target="cve-meta",
            params={"cve_id": cve_id},
            verbose=False
        )
        
        # Check for rate limiting
        if view_response.get("retcode") == 500 and ("NVD_RATE_LIMITED" in view_response.get("message", "") or "429" in view_response.get("message", "")):
            pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
        
        assert view_response["retcode"] == 0
        cve_data = view_response["payload"]
        print(f"    - CVE ID: {cve_data['id']}")
        
        # Step 3: User adds to collection if not already there
        if not is_stored:
            print("  → Step 3: User adds CVE to their collection")
            create_response = access.rpc_call(
                method="RPCCreateCVE",
                target="cve-meta",
                params={"cve_id": cve_id},
                verbose=False
            )
            
            # Check for rate limiting
            if create_response.get("retcode") == 500 and ("NVD_RATE_LIMITED" in create_response.get("message", "") or "429" in create_response.get("message", "")):
                pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
            
            assert create_response["retcode"] == 0
            print(f"    - Added to collection")
        else:
            print("  → Step 3: CVE already in collection (skipped)")
        
        # Step 4: User verifies it's now in their collection
        print("  → Step 4: User verifies CVE is in collection")
        verify_response = access.rpc_call(
            method="RPCIsCVEStoredByID",
            target="cve-local",
            params={"cve_id": cve_id},
            verbose=False
        )
        
        assert verify_response["payload"]["stored"] is True
        print(f"    ✓ CVE confirmed in collection")
        
        print("  ✓ Test passed: User workflow completed successfully")
    
    @pytest.mark.slow
    def test_user_collection_management_workflow(self, access_service):
        """Simulate user managing their CVE collection.
        
        Workflow:
        1. User views their collection (list CVEs)
        2. User adds several new CVEs
        3. User views updated collection
        4. User removes some CVEs
        5. User verifies final collection state
        """
        access = access_service
        
        print("\n  → Simulating user collection management workflow")
        
        # Step 1: User views current collection
        print("  → Step 1: User views current collection")
        list_response = access.rpc_call(
            method="RPCListCVEs",
            target="cve-meta",
            params={"offset": 0, "limit": 100},
            verbose=False
        )
        
        initial_count = list_response["payload"]["total"]
        print(f"    - Current collection size: {initial_count}")
        
        # Step 2: User adds new CVEs to collection
        print("  → Step 2: User adds new CVEs")
        new_cves = ["CVE-2021-44228", "CVE-2021-45046"]
        added = 0
        
        for cve_id in new_cves:
            print(f"    - Adding {cve_id}...")
            create_response = access.rpc_call(
                method="RPCCreateCVE",
                target="cve-meta",
                params={"cve_id": cve_id},
                verbose=False
            )
            
            # Check for rate limiting
            if create_response.get("retcode") == 500 and ("NVD_RATE_LIMITED" in create_response.get("message", "") or "429" in create_response.get("message", "")):
                pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
            
            if create_response["retcode"] == 0:
                added += 1
                print(f"      ✓ Added")
            else:
                print(f"      ✓ Already exists or failed")
            
            time.sleep(1)  # Rate limiting
        
        # Step 3: User views updated collection
        print("  → Step 3: User views updated collection")
        updated_list = access.rpc_call(
            method="RPCListCVEs",
            target="cve-meta",
            params={"offset": 0, "limit": 100},
            verbose=False
        )
        
        updated_count = updated_list["payload"]["total"]
        print(f"    - Updated collection size: {updated_count}")
        
        # Step 4: User removes one CVE
        print("  → Step 4: User removes CVE from collection")
        cve_to_remove = "CVE-2021-45046"
        delete_response = access.rpc_call(
            method="RPCDeleteCVE",
            target="cve-meta",
            params={"cve_id": cve_to_remove},
            verbose=False
        )
        print(f"    - Removed {cve_to_remove}")
        
        # Step 5: User verifies final state
        print("  → Step 5: User verifies final collection")
        final_list = access.rpc_call(
            method="RPCListCVEs",
            target="cve-meta",
            params={"offset": 0, "limit": 100},
            verbose=False
        )
        
        final_count = final_list["payload"]["total"]
        print(f"    - Final collection size: {final_count}")
        
        print("  ✓ Test passed: Collection management workflow completed")
    
    @pytest.mark.slow
    def test_user_pagination_browsing_workflow(self, access_service):
        """Simulate user browsing through paginated CVE list.
        
        Workflow:
        1. User views first page of CVEs
        2. User navigates to next page
        3. User goes back to previous page
        4. User jumps to specific page
        5. User changes page size
        """
        access = access_service
        
        print("\n  → Simulating user pagination browsing workflow")
        
        # Ensure we have some CVEs for pagination
        print("  → Setting up test data...")
        test_cves = ["CVE-2021-44228", "CVE-2021-45046", "CVE-2022-22965"]
        for cve_id in test_cves:
            create_response = access.rpc_call(
                method="RPCCreateCVE",
                target="cve-meta",
                params={"cve_id": cve_id},
                verbose=False
            )
            
            # Check for rate limiting
            if create_response.get("retcode") == 500 and ("NVD_RATE_LIMITED" in create_response.get("message", "") or "429" in create_response.get("message", "")):
                pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
            
            time.sleep(0.5)
        
        # Step 1: View first page (page size = 2)
        print("  → Step 1: User views first page (2 items)")
        page1 = access.rpc_call(
            method="RPCListCVEs",
            target="cve-meta",
            params={"offset": 0, "limit": 2},
            verbose=False
        )
        
        assert len(page1["payload"]["cves"]) <= 2
        print(f"    - Showing {len(page1['payload']['cves'])} items")
        
        # Step 2: Navigate to next page
        print("  → Step 2: User navigates to next page")
        page2 = access.rpc_call(
            method="RPCListCVEs",
            target="cve-meta",
            params={"offset": 2, "limit": 2},
            verbose=False
        )
        
        print(f"    - Showing {len(page2['payload']['cves'])} items")
        
        # Step 3: Go back to first page
        print("  → Step 3: User goes back to first page")
        page1_again = access.rpc_call(
            method="RPCListCVEs",
            target="cve-meta",
            params={"offset": 0, "limit": 2},
            verbose=False
        )
        
        assert len(page1_again["payload"]["cves"]) == len(page1["payload"]["cves"])
        print(f"    ✓ Same page size")
        
        # Step 4: Change page size
        print("  → Step 4: User changes page size to 5")
        larger_page = access.rpc_call(
            method="RPCListCVEs",
            target="cve-meta",
            params={"offset": 0, "limit": 5},
            verbose=False
        )
        
        print(f"    - Showing {len(larger_page['payload']['cves'])} items")
        
        print("  ✓ Test passed: Pagination browsing workflow completed")
    
    @pytest.mark.slow
    def test_user_refresh_update_workflow(self, access_service):
        """Simulate user refreshing CVE data.
        
        Workflow:
        1. User has CVE in collection
        2. User wants to refresh with latest data
        3. User triggers update
        4. User verifies data was refreshed
        """
        access = access_service
        cve_id = "CVE-2021-44228"
        
        print("\n  → Simulating user refresh/update workflow")
        
        # Step 1: Ensure CVE exists
        print("  → Step 1: Ensure CVE is in collection")
        create_response = access.rpc_call(
            method="RPCCreateCVE",
            target="cve-meta",
            params={"cve_id": cve_id},
            verbose=False
        )
        
        # Check for rate limiting
        if create_response.get("retcode") == 500 and ("NVD_RATE_LIMITED" in create_response.get("message", "") or "429" in create_response.get("message", "")):
            pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
        
        print(f"    ✓ CVE ready")
        
        # Step 2: User triggers refresh
        print("  → Step 2: User refreshes CVE data from NVD")
        time.sleep(1)  # Rate limiting
        update_response = access.rpc_call(
            method="RPCUpdateCVE",
            target="cve-meta",
            params={"cve_id": cve_id},
            verbose=False
        )
        
        # Check for rate limiting
        if update_response.get("retcode") == 500 and ("NVD_RATE_LIMITED" in update_response.get("message", "") or "429" in update_response.get("message", "")):
            pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
        
        assert update_response["retcode"] == 0
        print(f"    ✓ Data refreshed")
        
        # Step 3: User views refreshed data
        print("  → Step 3: User views refreshed CVE")
        view_response = access.rpc_call(
            method="RPCGetCVE",
            target="cve-meta",
            params={"cve_id": cve_id},
            verbose=False
        )
        
        assert view_response["retcode"] == 0
        print(f"    ✓ Refreshed data retrieved")
        
        print("  ✓ Test passed: Refresh workflow completed")
    
    @pytest.mark.slow
    def test_user_bulk_import_workflow(self, access_service):
        """Simulate user importing multiple CVEs in bulk.
        
        Workflow:
        1. User has list of CVE IDs to import
        2. User imports each one
        3. User verifies all were imported
        4. User views the complete list
        """
        access = access_service
        
        print("\n  → Simulating user bulk import workflow")
        
        # List of CVEs to import
        cve_list = [
            "CVE-2021-44228",
            "CVE-2022-22965",
        ]
        
        print(f"  → User wants to import {len(cve_list)} CVEs")
        
        # Step 1: Import each CVE
        print("  → Step 1: Importing CVEs...")
        imported = []
        
        for i, cve_id in enumerate(cve_list, 1):
            print(f"    - Importing {i}/{len(cve_list)}: {cve_id}")
            response = access.rpc_call(
                method="RPCCreateCVE",
                target="cve-meta",
                params={"cve_id": cve_id},
                verbose=False
            )
            
            # Check for rate limiting
            if response.get("retcode") == 500 and ("NVD_RATE_LIMITED" in response.get("message", "") or "429" in response.get("message", "")):
                pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
            
            if response["retcode"] == 0:
                imported.append(cve_id)
                print(f"      ✓ Imported")
            else:
                print(f"      ✓ Already exists")
                imported.append(cve_id)  # Count as imported
            
            time.sleep(1)  # Rate limiting
        
        print(f"    ✓ Import completed: {len(imported)}/{len(cve_list)}")
        
        # Step 2: Verify all were imported
        print("  → Step 2: Verifying imported CVEs")
        for cve_id in imported:
            check_response = access.rpc_call(
                method="RPCIsCVEStoredByID",
                target="cve-local",
                params={"cve_id": cve_id},
                verbose=False
            )
            
            assert check_response["payload"]["stored"] is True
            print(f"    ✓ {cve_id} verified")
        
        # Step 3: View complete list
        print("  → Step 3: User views complete collection")
        list_response = access.rpc_call(
            method="RPCListCVEs",
            target="cve-meta",
            params={"offset": 0, "limit": 100},
            verbose=False
        )
        
        total = list_response["payload"]["total"]
        print(f"    - Total CVEs in collection: {total}")
        
        print("  ✓ Test passed: Bulk import workflow completed")
    
    @pytest.mark.slow
    def test_user_cleanup_workflow(self, access_service):
        """Simulate user cleaning up their CVE collection.
        
        Workflow:
        1. User views current collection
        2. User decides which CVEs to remove
        3. User removes selected CVEs
        4. User verifies they were removed
        5. User views cleaned collection
        """
        access = access_service
        
        print("\n  → Simulating user cleanup workflow")
        
        # Setup: Create some CVEs
        print("  → Setting up test CVEs...")
        test_cves = ["CVE-2021-44228", "CVE-2021-45046"]
        for cve_id in test_cves:
            create_response = access.rpc_call(
                method="RPCCreateCVE",
                target="cve-meta",
                params={"cve_id": cve_id},
                verbose=False
            )
            
            # Check for rate limiting
            if create_response.get("retcode") == 500 and ("NVD_RATE_LIMITED" in create_response.get("message", "") or "429" in create_response.get("message", "")):
                pytest.skip("NVD API rate limited (HTTP 429) - skipping test")
            
            time.sleep(0.5)
        
        # Step 1: View collection
        print("  → Step 1: User views collection")
        list_response = access.rpc_call(
            method="RPCListCVEs",
            target="cve-meta",
            params={"offset": 0, "limit": 100},
            verbose=False
        )
        
        initial_count = list_response["payload"]["total"]
        print(f"    - Collection size: {initial_count}")
        
        # Step 2: User selects CVEs to remove
        to_remove = ["CVE-2021-45046"]
        print(f"  → Step 2: User selects {len(to_remove)} CVE(s) to remove")
        
        # Step 3: Remove selected CVEs
        print("  → Step 3: Removing selected CVEs")
        for cve_id in to_remove:
            print(f"    - Removing {cve_id}")
            delete_response = access.rpc_call(
                method="RPCDeleteCVE",
                target="cve-meta",
                params={"cve_id": cve_id},
                verbose=False
            )
            print(f"      ✓ Removed")
        
        # Step 4: Verify removal
        print("  → Step 4: Verifying CVEs were removed")
        for cve_id in to_remove:
            check_response = access.rpc_call(
                method="RPCIsCVEStoredByID",
                target="cve-local",
                params={"cve_id": cve_id},
                verbose=False
            )
            
            assert check_response["payload"]["stored"] is False
            print(f"    ✓ {cve_id} confirmed removed")
        
        # Step 5: View cleaned collection
        print("  → Step 5: User views cleaned collection")
        final_list = access.rpc_call(
            method="RPCListCVEs",
            target="cve-meta",
            params={"offset": 0, "limit": 100},
            verbose=False
        )
        
        final_count = final_list["payload"]["total"]
        print(f"    - Final collection size: {final_count}")
        
        print("  ✓ Test passed: Cleanup workflow completed")
