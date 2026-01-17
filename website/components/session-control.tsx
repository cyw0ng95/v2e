/**
 * Session Control Component
 * Manages job session control (start, stop, pause, resume)
 */

'use client';

import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  useSessionStatus,
  useStartSession,
  useStopSession,
  usePauseJob,
  useResumeJob,
} from "@/lib/hooks";
import { Play, Square, Pause, RotateCw } from "lucide-react";
import { toast } from "sonner";

export function SessionControl() {
  const { data: sessionStatus } = useSessionStatus();
  const startSession = useStartSession();
  const stopSession = useStopSession();
  const pauseJob = usePauseJob();
  const resumeJob = useResumeJob();

  const [sessionId, setSessionId] = useState(`session-${Date.now()}`);
  const [startIndex, setStartIndex] = useState(0);
  const [resultsPerBatch, setResultsPerBatch] = useState(100);

  const handleStartSession = () => {
    startSession.mutate(
      {
        sessionId,
        startIndex,
        resultsPerBatch,
      },
      {
        onSuccess: () => {
          toast.success("Session started successfully");
        },
        onError: (error) => {
          toast.error(`Failed to start session: ${error.message}`);
        },
      }
    );
  };

  const handleStopSession = () => {
    stopSession.mutate(undefined, {
      onSuccess: (data) => {
        toast.success(
          `Session stopped. Fetched: ${data?.fetchedCount}, Stored: ${data?.storedCount}`
        );
      },
      onError: (error) => {
        toast.error(`Failed to stop session: ${error.message}`);
      },
    });
  };

  const handlePauseJob = () => {
    pauseJob.mutate(undefined, {
      onSuccess: () => {
        toast.success("Job paused");
      },
      onError: (error) => {
        toast.error(`Failed to pause job: ${error.message}`);
      },
    });
  };

  const handleResumeJob = () => {
    resumeJob.mutate(undefined, {
      onSuccess: () => {
        toast.success("Job resumed");
      },
      onError: (error) => {
        toast.error(`Failed to resume job: ${error.message}`);
      },
    });
  };

  const isRunning = sessionStatus?.hasSession && sessionStatus.state === 'running';
  const isPaused = sessionStatus?.hasSession && sessionStatus.state === 'paused';

  return (
    <Card>
      <CardHeader>
        <CardTitle>Session Control</CardTitle>
        <CardDescription>
          Start, stop, or manage CVE fetching job sessions
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {/* Session Configuration */}
          {!sessionStatus?.hasSession && (
            <div className="grid gap-4 md:grid-cols-3">
              <div className="space-y-2">
                <Label htmlFor="sessionId">Session ID</Label>
                <Input
                  id="sessionId"
                  value={sessionId}
                  onChange={(e) => setSessionId(e.target.value)}
                  placeholder="Enter session ID"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="startIndex">Start Index</Label>
                <Input
                  id="startIndex"
                  type="number"
                  value={startIndex}
                  onChange={(e) => setStartIndex(parseInt(e.target.value) || 0)}
                  placeholder="0"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="resultsPerBatch">Results per Batch</Label>
                <Input
                  id="resultsPerBatch"
                  type="number"
                  value={resultsPerBatch}
                  onChange={(e) => setResultsPerBatch(parseInt(e.target.value) || 100)}
                  placeholder="100"
                />
              </div>
            </div>
          )}

          {/* Control Buttons */}
          <div className="flex gap-2">
            {!sessionStatus?.hasSession && (
              <Button
                onClick={handleStartSession}
                disabled={startSession.isPending || !sessionId}
              >
                <Play className="h-4 w-4 mr-2" />
                Start Session
              </Button>
            )}

            {isRunning && (
              <>
                <Button
                  onClick={handlePauseJob}
                  disabled={pauseJob.isPending}
                  variant="outline"
                >
                  <Pause className="h-4 w-4 mr-2" />
                  Pause
                </Button>
                <Button
                  onClick={handleStopSession}
                  disabled={stopSession.isPending}
                  variant="destructive"
                >
                  <Square className="h-4 w-4 mr-2" />
                  Stop
                </Button>
              </>
            )}

            {isPaused && (
              <>
                <Button
                  onClick={handleResumeJob}
                  disabled={resumeJob.isPending}
                >
                  <RotateCw className="h-4 w-4 mr-2" />
                  Resume
                </Button>
                <Button
                  onClick={handleStopSession}
                  disabled={stopSession.isPending}
                  variant="destructive"
                >
                  <Square className="h-4 w-4 mr-2" />
                  Stop
                </Button>
              </>
            )}
          </div>

          {/* Session Info */}
          {sessionStatus?.hasSession && (
            <div className="rounded-lg border p-4 space-y-2">
              <div className="grid grid-cols-2 gap-2 text-sm">
                <div>
                  <span className="font-medium">Session ID:</span>{" "}
                  <span className="text-muted-foreground">{sessionStatus.sessionId}</span>
                </div>
                <div>
                  <span className="font-medium">State:</span>{" "}
                  <span className="text-muted-foreground">{sessionStatus.state}</span>
                </div>
                <div>
                  <span className="font-medium">Start Index:</span>{" "}
                  <span className="text-muted-foreground">{sessionStatus.startIndex}</span>
                </div>
                <div>
                  <span className="font-medium">Batch Size:</span>{" "}
                  <span className="text-muted-foreground">{sessionStatus.resultsPerBatch}</span>
                </div>
                <div>
                  <span className="font-medium">Fetched:</span>{" "}
                  <span className="text-muted-foreground">{sessionStatus.fetchedCount}</span>
                </div>
                <div>
                  <span className="font-medium">Stored:</span>{" "}
                  <span className="text-muted-foreground">{sessionStatus.storedCount}</span>
                </div>
                <div>
                  <span className="font-medium">Errors:</span>{" "}
                  <span className="text-muted-foreground">{sessionStatus.errorCount}</span>
                </div>
              </div>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
