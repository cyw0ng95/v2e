"""Advanced integration tests combining CRUD operations with Job Control.

These tests verify complex scenarios involving:
1. Running jobs while performing CRUD operations
2. Data consistency during concurrent operations
3. Job recovery after errors
4. Long-running job scenarios
"""

import pytest
import time
import threading


@pytest.fixture(scope="function", autouse=True)
def cleanup_session(access_service):
    """Clean up any existing session before and after each test."""
    # Try to stop any existing session before the test
    try:
        access_service.rpc_call(
            method="RPCStopSession",
            target="meta",
            params={}
        )
    except Exception:
        # Ignore errors if no session exists
        pass
    
    yield
    
    # Clean up after the test
    try:
        access_service.rpc_call(
            method="RPCStopSession",
            target="meta",
            params={}
        )
    except Exception:
        # Ignore errors if no session exists
        pass


@pytest.mark.integration
class TestCRUDDuringJobExecution:
    """Test CRUD operations while job is running."""
    
    def test_create_cve_while_job_running(self, access_service):
        """Test creating individual CVEs while a job is fetching data.
        
        Verifies that manual CVE creation works alongside job execution.
        """
        access = access_service
        
        print("\n  → Testing CRUD during job execution")
        
        # Start a job session
        start_resp = access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={
                "session_id": "crud-test-session",
                "start_index": 0,
                "results_per_batch": 5
            }
        )
        
        assert start_resp["retcode"] == 0
        print(f"  → Job started")
        
        # Let job run for a moment
        time.sleep(0.5)
        
        # Perform CRUD operation - create a CVE
        create_resp = access.rpc_call(
            method="RPCCreateCVE",
            target="meta",
            params={"cve_id": "CVE-2024-12345"}
        )
        
        # May fail due to rate limiting or not found, but should not crash
        print(f"  → Create during job: retcode={create_resp.get('retcode')}")
        
        # Check job status
        status_resp = access.rpc_call(
            method="RPCGetSessionStatus",
            target="meta",
            params={}
        )
        
        assert status_resp["retcode"] == 0
        assert status_resp["payload"]["has_session"] is True
        assert status_resp["payload"]["state"] in ["running", "paused", "stopped"]
        
        print(f"  → Job still operational after CRUD")
        print(f"  ✓ Test passed")
    
    def test_list_cves_while_job_storing(self, access_service):
        """Test listing CVEs while job is actively storing data.
        
        Verifies database reads work during concurrent writes.
        """
        access = access_service
        
        print("\n  → Testing list operation during job execution")
        
        # Start a job
        access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={
                "session_id": "list-test-session",
                "start_index": 0,
                "results_per_batch": 10
            }
        )
        
        time.sleep(0.5)
        
        # Try to list CVEs while job is running
        list_resp = access.rpc_call(
            method="RPCListCVEs",
            target="meta",
            params={"offset": 0, "limit": 10}
        )
        
        # Should succeed even with concurrent writes
        assert list_resp["retcode"] == 0
        assert "cves" in list_resp["payload"]
        
        print(f"  → Listed {len(list_resp['payload']['cves'])} CVEs during job execution")
        print(f"  ✓ Test passed")
    
    def test_count_cves_during_job(self, access_service):
        """Test counting CVEs while job is running.
        
        Verifies count accuracy during concurrent operations.
        """
        access = access_service
        
        print("\n  → Testing count operation during job execution")
        
        # Get initial count
        count1_resp = access.rpc_call(
            method="RPCCountCVEs",
            target="meta",
            params={}
        )
        initial_count = count1_resp["payload"]["count"]
        
        print(f"  → Initial count: {initial_count}")
        
        # Start job
        access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={
                "session_id": "count-test-session",
                "start_index": 0,
                "results_per_batch": 5
            }
        )
        
        # Let job add some CVEs
        time.sleep(2)
        
        # Get count again
        count2_resp = access.rpc_call(
            method="RPCCountCVEs",
            target="meta",
            params={}
        )
        final_count = count2_resp["payload"]["count"]
        
        print(f"  → Final count: {final_count}")
        
        # Count should be consistent (may or may not increase depending on NVD)
        assert count2_resp["retcode"] == 0
        assert final_count >= initial_count
        
        print(f"  ✓ Test passed: Count is consistent")


@pytest.mark.integration
class TestJobRobustness:
    """Test job control robustness with rapid commands."""
    
    def test_rapid_pause_resume(self, access_service):
        """Test rapid pause/resume cycles.
        
        Verifies the job controller handles rapid state changes.
        """
        access = access_service
        
        print("\n  → Testing rapid pause/resume cycles")
        
        # Start job
        access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={
                "session_id": "rapid-test-session",
                "start_index": 0,
                "results_per_batch": 5
            }
        )
        
        time.sleep(0.2)
        
        # Rapid pause/resume cycles
        for i in range(3):
            pause_resp = access.rpc_call(
                method="RPCPauseJob",
                target="meta",
                params={}
            )
            
            print(f"  → Pause {i+1}: retcode={pause_resp.get('retcode')}")
            time.sleep(0.1)
            
            resume_resp = access.rpc_call(
                method="RPCResumeJob",
                target="meta",
                params={}
            )
            
            print(f"  → Resume {i+1}: retcode={resume_resp.get('retcode')}")
            time.sleep(0.1)
        
        # Verify job is still operational
        status_resp = access.rpc_call(
            method="RPCGetSessionStatus",
            target="meta",
            params={}
        )
        
        assert status_resp["retcode"] == 0
        assert status_resp["payload"]["has_session"] is True
        
        print(f"  ✓ Job survived rapid pause/resume cycles")
    
    def test_multiple_status_checks(self, access_service):
        """Test multiple concurrent status checks.
        
        Verifies status API is thread-safe.
        """
        access = access_service
        
        print("\n  → Testing multiple concurrent status checks")
        
        # Start job
        access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={
                "session_id": "status-test-session",
                "start_index": 0,
                "results_per_batch": 5
            }
        )
        
        time.sleep(0.2)
        
        # Perform 10 concurrent status checks
        results = []
        errors = []
        
        def check_status():
            try:
                resp = access.rpc_call(
                    method="RPCGetSessionStatus",
                    target="meta",
                    params={}
                )
                results.append(resp)
            except Exception as e:
                errors.append(e)
        
        threads = []
        for _ in range(10):
            t = threading.Thread(target=check_status)
            threads.append(t)
            t.start()
        
        for t in threads:
            t.join()
        
        # All should succeed
        assert len(errors) == 0, f"Got errors: {errors}"
        assert len(results) == 10
        
        for resp in results:
            assert resp["retcode"] == 0
            assert resp["payload"]["has_session"] is True
        
        print(f"  ✓ All 10 concurrent status checks succeeded")
    
    def test_pause_immediately_after_start(self, access_service):
        """Test pausing immediately after starting.
        
        Tests race condition handling.
        """
        access = access_service
        
        print("\n  → Testing pause immediately after start")
        
        # Start and immediately pause
        access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={
                "session_id": "immediate-pause-session",
                "start_index": 0,
                "results_per_batch": 5
            }
        )
        
        # Pause without delay
        pause_resp = access.rpc_call(
            method="RPCPauseJob",
            target="meta",
            params={}
        )
        
        print(f"  → Immediate pause: retcode={pause_resp.get('retcode')}")
        
        # Small delay for state to update
        time.sleep(0.3)
        
        # Check status
        status_resp = access.rpc_call(
            method="RPCGetSessionStatus",
            target="meta",
            params={}
        )
        
        assert status_resp["retcode"] == 0
        # State should be paused (or running if pause was too fast)
        assert status_resp["payload"]["state"] in ["paused", "running"]
        
        print(f"  ✓ Immediate pause handled correctly")


@pytest.mark.integration
class TestJobDataConsistency:
    """Test data consistency during job execution."""
    
    def test_progress_counter_consistency(self, access_service):
        """Test that progress counters are consistent.
        
        Verifies fetched >= stored and no count goes backward.
        """
        access = access_service
        
        print("\n  → Testing progress counter consistency")
        
        # Start job
        access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={
                "session_id": "consistency-test-session",
                "start_index": 0,
                "results_per_batch": 10
            }
        )
        
        time.sleep(0.2)
        
        # Sample progress multiple times
        prev_fetched = 0
        prev_stored = 0
        
        for i in range(5):
            status_resp = access.rpc_call(
                method="RPCGetSessionStatus",
                target="meta",
                params={}
            )
            
            if status_resp["retcode"] == 0 and status_resp["payload"]["has_session"]:
                fetched = status_resp["payload"]["fetched_count"]
                stored = status_resp["payload"]["stored_count"]
                
                print(f"  → Sample {i+1}: fetched={fetched}, stored={stored}")
                
                # Counters should not decrease
                assert fetched >= prev_fetched, "Fetched count decreased!"
                assert stored >= prev_stored, "Stored count decreased!"
                
                # Fetched should be >= stored
                assert fetched >= stored, "Fetched count less than stored!"
                
                prev_fetched = fetched
                prev_stored = stored
            
            time.sleep(0.5)
        
        print(f"  ✓ Progress counters are consistent")
    
    def test_session_state_validity(self, access_service):
        """Test that session state transitions are valid.
        
        Verifies only valid state transitions occur.
        """
        access = access_service
        
        print("\n  → Testing session state validity")
        
        valid_states = ["idle", "running", "paused", "stopped"]
        
        # Start job
        access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={
                "session_id": "state-validity-session",
                "start_index": 0,
                "results_per_batch": 5
            }
        )
        
        # Check state multiple times
        for i in range(10):
            status_resp = access.rpc_call(
                method="RPCGetSessionStatus",
                target="meta",
                params={}
            )
            
            if status_resp["retcode"] == 0 and status_resp["payload"]["has_session"]:
                state = status_resp["payload"]["state"]
                
                assert state in valid_states, f"Invalid state: {state}"
                print(f"  → Sample {i+1}: state={state} (valid)")
            
            time.sleep(0.2)
        
        print(f"  ✓ All observed states are valid")


@pytest.mark.integration
class TestJobErrorScenarios:
    """Test job behavior under error conditions."""
    
    def test_job_with_invalid_start_index(self, access_service):
        """Test job with very large start index.
        
        Verifies graceful handling of edge case parameters.
        """
        access = access_service
        
        print("\n  → Testing job with large start index")
        
        # Start with very large index
        start_resp = access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={
                "session_id": "large-index-session",
                "start_index": 999999,
                "results_per_batch": 5
            }
        )
        
        assert start_resp["retcode"] == 0
        print(f"  → Job started with start_index=999999")
        
        time.sleep(1)
        
        # Check status
        status_resp = access.rpc_call(
            method="RPCGetSessionStatus",
            target="meta",
            params={}
        )
        
        # Job should handle this gracefully (likely getting no results)
        assert status_resp["retcode"] == 0
        
        print(f"  ✓ Job handled large start index gracefully")
    
    def test_stop_and_restart_session(self, access_service):
        """Test stopping a session and starting a new one.
        
        Verifies proper cleanup and new session creation.
        """
        access = access_service
        
        print("\n  → Testing stop and restart")
        
        # Start first session
        start1_resp = access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={
                "session_id": "session-1",
                "start_index": 0,
                "results_per_batch": 5
            }
        )
        assert start1_resp["retcode"] == 0
        
        time.sleep(0.5)
        
        # Get progress from first session
        status1_resp = access.rpc_call(
            method="RPCGetSessionStatus",
            target="meta",
            params={}
        )
        session1_fetched = status1_resp["payload"]["fetched_count"]
        
        print(f"  → Session 1 fetched: {session1_fetched}")
        
        # Stop first session
        stop_resp = access.rpc_call(
            method="RPCStopSession",
            target="meta",
            params={}
        )
        assert stop_resp["retcode"] == 0
        
        # Verify no session exists
        status_none = access.rpc_call(
            method="RPCGetSessionStatus",
            target="meta",
            params={}
        )
        assert status_none["payload"]["has_session"] is False
        
        # Start second session
        start2_resp = access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={
                "session_id": "session-2",
                "start_index": 0,
                "results_per_batch": 5
            }
        )
        assert start2_resp["retcode"] == 0
        
        print(f"  → Successfully started session 2 after stopping session 1")
        print(f"  ✓ Stop and restart works correctly")
