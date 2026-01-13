"""Integration tests for CVE job control functionality via cve-meta service.

These tests verify the job control features:
1. Start a session to begin continuous CVE fetching and storing
2. Get session status
3. Pause and resume jobs
4. Stop a session
5. Single session enforcement
"""

import pytest
import time


@pytest.fixture(scope="function", autouse=True)
def cleanup_session(access_service):
    """Clean up any existing session before each test."""
    # Try to stop any existing session before the test
    try:
        access_service.rpc_call(
            method="RPCStopSession",
            target="cve-meta",
            params={}
        )
    except:
        pass
    
    yield
    
    # Clean up after the test
    try:
        access_service.rpc_call(
            method="RPCStopSession",
            target="cve-meta",
            params={}
        )
    except:
        pass


@pytest.mark.integration
class TestJobControl:
    """Integration tests for job control via cve-meta service."""
    
    def test_start_session(self, access_service):
        """Test starting a new job session.
        
        This verifies:
        - cve-meta can create a new session
        - Job starts running
        - Session state is properly initialized
        """
        access = access_service
        
        print("\n  → Testing RPCStartSession")
        
        # Start a new session
        response = access.rpc_call(
            method="RPCStartSession",
            target="cve-meta",
            params={
                "session_id": "test-session-1",
                "start_index": 0,
                "results_per_batch": 10
            }
        )
        
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Verify response
        assert response["retcode"] == 0
        assert response["message"] == "success"
        assert response["payload"] is not None
        
        # Verify payload
        payload = response["payload"]
        assert payload["success"] is True
        assert payload["session_id"] == "test-session-1"
        assert payload["state"] == "running"
        assert "created_at" in payload
        
        print(f"  ✓ Test passed: Successfully started session")
    
    def test_get_session_status_no_session(self, access_service):
        """Test getting session status when no session exists.
        
        This verifies:
        - cve-meta correctly reports no session
        """
        access = access_service
        
        print("\n  → Testing RPCGetSessionStatus (no session)")
        
        response = access.rpc_call(
            method="RPCGetSessionStatus",
            target="cve-meta",
            params={}
        )
        
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Verify response
        assert response["retcode"] == 0
        assert response["message"] == "success"
        assert response["payload"] is not None
        
        # Verify no session
        payload = response["payload"]
        assert payload["has_session"] is False
        
        print(f"  ✓ Test passed: No session reported correctly")
    
    def test_get_session_status_with_session(self, access_service):
        """Test getting session status when session exists.
        
        This verifies:
        - cve-meta returns session information
        - Session state is accessible
        """
        access = access_service
        
        print("\n  → Testing RPCGetSessionStatus (with session)")
        
        # Start a session first
        access.rpc_call(
            method="RPCStartSession",
            target="cve-meta",
            params={
                "session_id": "test-session-2",
                "start_index": 0,
                "results_per_batch": 10
            }
        )
        
        # Get session status
        response = access.rpc_call(
            method="RPCGetSessionStatus",
            target="cve-meta",
            params={}
        )
        
        print(f"  → Response received:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Verify response
        assert response["retcode"] == 0
        assert response["message"] == "success"
        assert response["payload"] is not None
        
        # Verify session information
        payload = response["payload"]
        assert payload["has_session"] is True
        assert payload["session_id"] == "test-session-2"
        assert payload["state"] == "running"
        assert payload["start_index"] == 0
        assert payload["results_per_batch"] == 10
        assert "created_at" in payload
        assert "updated_at" in payload
        assert "fetched_count" in payload
        assert "stored_count" in payload
        assert "error_count" in payload
        
        print(f"  ✓ Test passed: Session status retrieved successfully")
        
    
    def test_single_session_enforcement(self, access_service):
        """Test that only one session can exist at a time.
        
        This verifies:
        - cve-meta enforces single session rule
        - Trying to start second session fails
        """
        access = access_service
        
        print("\n  → Testing single session enforcement")
        
        # Start first session
        response1 = access.rpc_call(
            method="RPCStartSession",
            target="cve-meta",
            params={
                "session_id": "test-session-3",
                "start_index": 0,
                "results_per_batch": 10
            }
        )
        
        assert response1["retcode"] == 0
        print(f"  → First session started successfully")
        
        # Try to start second session (should fail)
        response2 = access.rpc_call(
            method="RPCStartSession",
            target="cve-meta",
            params={
                "session_id": "test-session-4",
                "start_index": 0,
                "results_per_batch": 10
            }
        )
        
        print(f"  → Second session attempt:")
        print(f"    - retcode: {response2.get('retcode')}")
        print(f"    - message: {response2.get('message')}")
        
        # Should get error
        assert response2["retcode"] == 500
        assert "already exists" in response2["message"].lower() or "session exists" in response2["message"].lower()
        
        print(f"  ✓ Test passed: Single session enforcement works")
        
    
    def test_pause_and_resume_job(self, access_service):
        """Test pausing and resuming a job.
        
        This verifies:
        - cve-meta can pause a running job
        - cve-meta can resume a paused job
        - Session state changes appropriately
        """
        access = access_service
        
        print("\n  → Testing RPCPauseJob and RPCResumeJob")
        
        # Start a session
        access.rpc_call(
            method="RPCStartSession",
            target="cve-meta",
            params={
                "session_id": "test-session-5",
                "start_index": 0,
                "results_per_batch": 10
            }
        )
        
        # Give the job a moment to actually start
        time.sleep(0.2)
        
        # Pause the job
        pause_response = access.rpc_call(
            method="RPCPauseJob",
            target="cve-meta",
            params={}
        )
        
        print(f"  → Pause response:")
        print(f"    - retcode: {pause_response.get('retcode')}")
        print(f"    - message: {pause_response.get('message')}")
        
        assert pause_response["retcode"] == 0
        assert pause_response["payload"]["success"] is True
        assert pause_response["payload"]["state"] == "paused"
        
        # Give a moment for state to fully propagate
        time.sleep(0.5)
        
        # Verify session is paused
        status = access.rpc_call(
            method="RPCGetSessionStatus",
            target="cve-meta",
            params={}
        )
        assert status["payload"]["state"] == "paused"
        
        print(f"  ✓ Job paused successfully")
        
        # Resume the job
        resume_response = access.rpc_call(
            method="RPCResumeJob",
            target="cve-meta",
            params={}
        )
        
        print(f"  → Resume response:")
        print(f"    - retcode: {resume_response.get('retcode')}")
        print(f"    - message: {resume_response.get('message')}")
        
        assert resume_response["retcode"] == 0
        assert resume_response["payload"]["success"] is True
        assert resume_response["payload"]["state"] == "running"
        
        # Verify session is running
        status = access.rpc_call(
            method="RPCGetSessionStatus",
            target="cve-meta",
            params={}
        )
        assert status["payload"]["state"] == "running"
        
        print(f"  ✓ Job resumed successfully")
        
    
    def test_stop_session(self, access_service):
        """Test stopping a session.
        
        This verifies:
        - cve-meta can stop a running session
        - Session is deleted after stop
        - Progress counters are returned
        """
        access = access_service
        
        print("\n  → Testing RPCStopSession")
        
        # Start a session
        access.rpc_call(
            method="RPCStartSession",
            target="cve-meta",
            params={
                "session_id": "test-session-6",
                "start_index": 0,
                "results_per_batch": 10
            }
        )
        
        # Let it run for a moment
        time.sleep(1)
        
        # Stop the session
        stop_response = access.rpc_call(
            method="RPCStopSession",
            target="cve-meta",
            params={}
        )
        
        print(f"  → Stop response:")
        print(f"    - retcode: {stop_response.get('retcode')}")
        print(f"    - message: {stop_response.get('message')}")
        
        assert stop_response["retcode"] == 0
        assert stop_response["payload"]["success"] is True
        assert stop_response["payload"]["session_id"] == "test-session-6"
        assert "fetched_count" in stop_response["payload"]
        assert "stored_count" in stop_response["payload"]
        assert "error_count" in stop_response["payload"]
        
        print(f"  ✓ Session stopped successfully")
        print(f"    - Fetched: {stop_response['payload']['fetched_count']}")
        print(f"    - Stored: {stop_response['payload']['stored_count']}")
        print(f"    - Errors: {stop_response['payload']['error_count']}")
        
        # Verify session is deleted
        status = access.rpc_call(
            method="RPCGetSessionStatus",
            target="cve-meta",
            params={}
        )
        assert status["payload"]["has_session"] is False
        
        print(f"  ✓ Session deleted after stop")
    
    def test_pause_job_not_running(self, access_service):
        """Test pausing when no job is running.
        
        This verifies:
        - cve-meta handles pause error correctly
        """
        access = access_service
        
        print("\n  → Testing RPCPauseJob (no job running)")
        
        response = access.rpc_call(
            method="RPCPauseJob",
            target="cve-meta",
            params={}
        )
        
        print(f"  → Response:")
        print(f"    - retcode: {response.get('retcode')}")
        print(f"    - message: {response.get('message')}")
        
        # Should get error
        assert response["retcode"] == 500
        assert "not running" in response["message"].lower()
        
        print(f"  ✓ Test passed: Proper error when pausing non-running job")
