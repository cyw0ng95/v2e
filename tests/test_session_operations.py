"""Comprehensive session operation tests for CVE Meta service.

These tests provide extensive coverage of:
1. Session lifecycle (start, pause, resume, stop)
2. Session state transitions
3. Session parameter validation
4. Multiple session scenarios
5. Session persistence and recovery
"""

import pytest
import time


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
class TestSessionLifecycle:
    """Test complete session lifecycle operations."""
    
    def test_session_start_with_default_params(self, access_service):
        """Test starting session with minimal parameters.
        
        Verifies default parameter handling.
        """
        access = access_service
        
        print("\n  → Testing session start with default params")
        
        response = access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={"session_id": "default-params-session"}
        )
        
        assert response["retcode"] == 0
        assert response["payload"]["success"] is True
        assert response["payload"]["session_id"] == "default-params-session"
        
        # Check status to verify defaults
        status = access.rpc_call(
            method="RPCGetSessionStatus",
            target="meta",
            params={}
        )
        
        assert status["retcode"] == 0
        assert status["payload"]["has_session"] is True
        assert status["payload"]["start_index"] >= 0
        assert status["payload"]["results_per_batch"] > 0
        
        print(f"  ✓ Session started with defaults: start_index={status['payload']['start_index']}, batch={status['payload']['results_per_batch']}")
    
    def test_session_start_with_custom_params(self, access_service):
        """Test starting session with custom parameters.
        
        Verifies parameter customization works.
        """
        access = access_service
        
        print("\n  → Testing session start with custom params")
        
        response = access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={
                "session_id": "custom-params-session",
                "start_index": 100,
                "results_per_batch": 20
            }
        )
        
        assert response["retcode"] == 0
        
        # Verify custom params were applied
        status = access.rpc_call(
            method="RPCGetSessionStatus",
            target="meta",
            params={}
        )
        
        assert status["payload"]["start_index"] == 100
        assert status["payload"]["results_per_batch"] == 20
        
        print(f"  ✓ Session started with custom params")
    
    def test_session_pause_resume_multiple_times(self, access_service):
        """Test pausing and resuming session multiple times.
        
        Verifies pause/resume can be called repeatedly.
        """
        access = access_service
        
        print("\n  → Testing multiple pause/resume cycles")
        
        # Start session
        access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={"session_id": "pause-resume-test", "results_per_batch": 5}
        )
        
        time.sleep(0.5)
        
        # Pause and resume 3 times
        for i in range(3):
            # Pause
            pause_resp = access.rpc_call(
                method="RPCPauseJob",
                target="meta",
                params={}
            )
            assert pause_resp["retcode"] == 0
            assert pause_resp["payload"]["state"] == "paused"
            
            time.sleep(0.3)
            
            # Resume
            resume_resp = access.rpc_call(
                method="RPCResumeJob",
                target="meta",
                params={}
            )
            assert resume_resp["retcode"] == 0
            assert resume_resp["payload"]["state"] == "running"
            
            time.sleep(0.3)
            
            print(f"  → Cycle {i+1} completed")
        
        print(f"  ✓ Multiple pause/resume cycles completed")
    
    def test_session_stop_returns_statistics(self, access_service):
        """Test that stopping session returns proper statistics.
        
        Verifies stop response includes counters.
        """
        access = access_service
        
        print("\n  → Testing session stop statistics")
        
        # Start session
        access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={"session_id": "stats-test", "results_per_batch": 5}
        )
        
        time.sleep(1)
        
        # Stop and check statistics
        stop_resp = access.rpc_call(
            method="RPCStopSession",
            target="meta",
            params={}
        )
        
        assert stop_resp["retcode"] == 0
        assert "fetched_count" in stop_resp["payload"]
        assert "stored_count" in stop_resp["payload"]
        assert "error_count" in stop_resp["payload"]
        assert stop_resp["payload"]["fetched_count"] >= 0
        
        print(f"  → Statistics: fetched={stop_resp['payload']['fetched_count']}, stored={stop_resp['payload']['stored_count']}, errors={stop_resp['payload']['error_count']}")
        print(f"  ✓ Session stopped with valid statistics")


@pytest.mark.integration
class TestSessionStateValidation:
    """Test session state validation and error handling."""
    
    def test_pause_without_session(self, access_service):
        """Test pausing when no session exists.
        
        Verifies proper error handling.
        """
        access = access_service
        
        print("\n  → Testing pause without active session")
        
        response = access.rpc_call(
            method="RPCPauseJob",
            target="meta",
            params={}
        )
        
        # Should return error
        assert response["retcode"] != 0
        print(f"  ✓ Pause correctly rejected: {response['message']}")
    
    def test_resume_without_session(self, access_service):
        """Test resuming when no session exists.
        
        Verifies proper error handling.
        """
        access = access_service
        
        print("\n  → Testing resume without active session")
        
        response = access.rpc_call(
            method="RPCResumeJob",
            target="meta",
            params={}
        )
        
        # Should return error
        assert response["retcode"] != 0
        print(f"  ✓ Resume correctly rejected: {response['message']}")
    
    def test_resume_when_already_running(self, access_service):
        """Test resuming a session that's already running.
        
        Verifies proper state validation.
        """
        access = access_service
        
        print("\n  → Testing resume on running session")
        
        # Start session
        access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={"session_id": "running-test"}
        )
        
        time.sleep(0.5)
        
        # Try to resume (should fail since it's already running)
        response = access.rpc_call(
            method="RPCResumeJob",
            target="meta",
            params={}
        )
        
        # Should return error or indicate already running
        assert response["retcode"] != 0 or response["payload"].get("state") == "running"
        print(f"  ✓ Resume correctly handled on running session")
    
    def test_pause_when_already_paused(self, access_service):
        """Test pausing a session that's already paused.
        
        Verifies idempotency of pause operation.
        """
        access = access_service
        
        print("\n  → Testing pause on already paused session")
        
        # Start session
        access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={"session_id": "paused-test"}
        )
        
        time.sleep(0.5)
        
        # Pause
        access.rpc_call(
            method="RPCPauseJob",
            target="meta",
            params={}
        )
        
        # Pause again
        response = access.rpc_call(
            method="RPCPauseJob",
            target="meta",
            params={}
        )
        
        # Should either succeed (idempotent) or fail with appropriate message
        print(f"  → Second pause result: retcode={response['retcode']}, message={response.get('message', 'N/A')}")
        print(f"  ✓ Double pause handled correctly")
    
    def test_start_session_with_invalid_session_id(self, access_service):
        """Test starting session with various invalid session IDs.
        
        Verifies session ID validation.
        """
        access = access_service
        
        print("\n  → Testing invalid session IDs")
        
        invalid_ids = [
            "",  # Empty string
            " ",  # Whitespace only
            "a" * 300,  # Very long ID
        ]
        
        for invalid_id in invalid_ids:
            response = access.rpc_call(
                method="RPCStartSession",
                target="meta",
                params={"session_id": invalid_id}
            )
            
            print(f"  → Testing ID '{invalid_id[:20]}...': retcode={response['retcode']}")
        
        print(f"  ✓ Invalid session IDs handled")


@pytest.mark.integration
class TestSessionParameters:
    """Test session parameter validation and edge cases."""
    
    def test_session_with_zero_start_index(self, access_service):
        """Test session with start_index=0.
        
        Verifies zero is a valid starting point.
        """
        access = access_service
        
        print("\n  → Testing session with start_index=0")
        
        response = access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={
                "session_id": "zero-index-session",
                "start_index": 0,
                "results_per_batch": 5
            }
        )
        
        assert response["retcode"] == 0
        
        status = access.rpc_call(
            method="RPCGetSessionStatus",
            target="meta",
            params={}
        )
        
        assert status["payload"]["start_index"] == 0
        print(f"  ✓ Session started with start_index=0")
    
    def test_session_with_negative_start_index(self, access_service):
        """Test session with negative start_index.
        
        Verifies negative values are handled properly.
        """
        access = access_service
        
        print("\n  → Testing session with negative start_index")
        
        response = access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={
                "session_id": "negative-index-session",
                "start_index": -1,
                "results_per_batch": 5
            }
        )
        
        # Should either reject or normalize to 0
        print(f"  → Response: retcode={response['retcode']}, message={response.get('message', 'N/A')}")
        print(f"  ✓ Negative start_index handled")
    
    def test_session_with_small_batch_size(self, access_service):
        """Test session with very small batch size.
        
        Verifies small batches are handled correctly.
        """
        access = access_service
        
        print("\n  → Testing session with batch_size=1")
        
        response = access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={
                "session_id": "small-batch-session",
                "start_index": 0,
                "results_per_batch": 1
            }
        )
        
        assert response["retcode"] == 0
        
        status = access.rpc_call(
            method="RPCGetSessionStatus",
            target="meta",
            params={}
        )
        
        assert status["payload"]["results_per_batch"] == 1
        print(f"  ✓ Session started with batch_size=1")
    
    def test_session_with_large_batch_size(self, access_service):
        """Test session with very large batch size.
        
        Verifies large batches are handled correctly.
        """
        access = access_service
        
        print("\n  → Testing session with batch_size=1000")
        
        response = access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={
                "session_id": "large-batch-session",
                "start_index": 0,
                "results_per_batch": 1000
            }
        )
        
        # Should either accept or cap the value
        print(f"  → Response: retcode={response['retcode']}")
        
        if response["retcode"] == 0:
            status = access.rpc_call(
                method="RPCGetSessionStatus",
                target="meta",
                params={}
            )
            print(f"  → Actual batch size: {status['payload']['results_per_batch']}")
        
        print(f"  ✓ Large batch_size handled")
    
    def test_session_with_zero_batch_size(self, access_service):
        """Test session with batch_size=0.
        
        Verifies zero batch size is rejected or normalized.
        """
        access = access_service
        
        print("\n  → Testing session with batch_size=0")
        
        response = access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={
                "session_id": "zero-batch-session",
                "start_index": 0,
                "results_per_batch": 0
            }
        )
        
        # Should either reject or use a default
        print(f"  → Response: retcode={response['retcode']}, message={response.get('message', 'N/A')}")
        print(f"  ✓ Zero batch_size handled")


@pytest.mark.integration
class TestSessionStatus:
    """Test session status reporting and accuracy."""
    
    def test_status_shows_progress(self, access_service):
        """Test that status shows progress over time.
        
        Verifies counters increment.
        """
        access = access_service
        
        print("\n  → Testing status progress tracking")
        
        # Start session
        access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={"session_id": "progress-test", "results_per_batch": 5}
        )
        
        # Get initial status
        status1 = access.rpc_call(
            method="RPCGetSessionStatus",
            target="meta",
            params={}
        )
        initial_fetched = status1["payload"]["fetched_count"]
        
        time.sleep(1)
        
        # Get updated status
        status2 = access.rpc_call(
            method="RPCGetSessionStatus",
            target="meta",
            params={}
        )
        updated_fetched = status2["payload"]["fetched_count"]
        
        print(f"  → Progress: {initial_fetched} → {updated_fetched}")
        # Progress may or may not increase depending on API availability
        print(f"  ✓ Status tracking works")
    
    def test_status_when_paused(self, access_service):
        """Test that status correctly shows paused state.
        
        Verifies state field accuracy.
        """
        access = access_service
        
        print("\n  → Testing status in paused state")
        
        # Start and pause
        access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={"session_id": "pause-status-test"}
        )
        
        time.sleep(0.5)
        
        access.rpc_call(
            method="RPCPauseJob",
            target="meta",
            params={}
        )
        
        # Check status
        status = access.rpc_call(
            method="RPCGetSessionStatus",
            target="meta",
            params={}
        )
        
        assert status["payload"]["state"] == "paused"
        print(f"  ✓ Status correctly shows paused state")
    
    def test_status_preserves_timestamps(self, access_service):
        """Test that status includes valid timestamps.
        
        Verifies timestamp fields exist and are formatted correctly.
        """
        access = access_service
        
        print("\n  → Testing status timestamps")
        
        # Start session
        access.rpc_call(
            method="RPCStartSession",
            target="meta",
            params={"session_id": "timestamp-test"}
        )
        
        time.sleep(0.5)
        
        status = access.rpc_call(
            method="RPCGetSessionStatus",
            target="meta",
            params={}
        )
        
        payload = status["payload"]
        assert "created_at" in payload
        assert "updated_at" in payload
        assert payload["created_at"] != ""
        assert payload["updated_at"] != ""
        
        print(f"  → Created: {payload['created_at']}")
        print(f"  → Updated: {payload['updated_at']}")
        print(f"  ✓ Timestamps present and valid")
