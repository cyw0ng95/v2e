/**
 * Session Control Component
 * Manages job session control (start, stop, pause, resume)
 */

'use client';

import React, { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import {
  useSessionStatus,
  useStartTypedSession,
  useStopSession,
  usePauseJob,
  useResumeJob,
} from "@/lib/hooks";
import { PlayIcon as Play, SquareIcon as Square, PauseIcon as Pause, RotateCwIcon as RotateCw } from '@/components/icons';
import { toast } from "sonner";

// Helper function to capitalize data type for display
function formatDataType(dataType: string): string {
  switch (dataType.toLowerCase()) {
    case 'cve':
      return 'CVE';
    case 'cwe':
      return 'CWE';
    case 'capec':
      return 'CAPEC';
    case 'attack':
      return 'ATT&CK';
    default:
      return dataType.toUpperCase();
  }
}

export function SessionControl() {
  const { data: sessionStatus } = useSessionStatus();
  const startTypedSession = useStartTypedSession();
  const stopSession = useStopSession();
  const pauseJob = usePauseJob();
  const resumeJob = useResumeJob();

  const [sessionId, setSessionId] = useState(`session-${Date.now()}`);
  const [startIndex, setStartIndex] = useState(0);
  const [resultsPerBatch, setResultsPerBatch] = useState(100);
  const [dataType, setDataType] = useState('cve'); // Default to CVE
  // Track if we're currently handling a stop operation
  const [isStopping, setIsStopping] = useState(false);

  const handleStartSession = () => {
    startTypedSession.mutate(
      {
        sessionId,
        dataType,
        startIndex,
        resultsPerBatch,
      },
      {
        onSuccess: () => {
          toast.success(`Session started successfully for ${formatDataType(dataType)}`);
          // Reset stop session error tracking when starting a new session
          stopSession.reset();
          setIsStopping(false);
        },
        onError: (error) => {
          toast.error(`Failed to start session: ${error.message}`);
        },
      }
    );
  };

  const handleStopSession = () => {
    // Prevent multiple simultaneous stop requests
    if (stopSession.isPending || isStopping) {
      console.warn('Stop session request already pending or in progress');
      return;
    }
    
    setIsStopping(true);
    
    stopSession.mutate(undefined, {
      onSuccess: (data) => {
        toast.success(
          `Session stopped. Fetched: ${data?.fetchedCount}, Stored: ${data?.storedCount}`
        );
        setIsStopping(false);
      },
      onError: (error) => {
        // Don't show toast for "run not active" errors as they're expected
        if (!error.message.includes('run not active')) {
          toast.error(`Failed to stop session: ${error.message}`);
        }
        setIsStopping(false);
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

  // Reset stop session state when session becomes inactive
  React.useEffect(() => {
    if (sessionStatus && !sessionStatus.hasSession) {
      setIsStopping(false);
      stopSession.reset();
    }
  }, [sessionStatus, stopSession]);

  // Cleanup function to reset state when component unmounts
  React.useEffect(() => {
    return () => {
      setIsStopping(false);
      stopSession.reset();
    };
  }, [stopSession]);

  const isRunning = sessionStatus?.hasSession && sessionStatus.state === 'running';
  const isPaused = sessionStatus?.hasSession && sessionStatus.state === 'paused';

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Job Session Control</CardTitle>
        <CardDescription>
          Start, stop, or manage data fetching job sessions (CVE, CWE, CAPEC, ATT&CK)
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
                <Label htmlFor="dataType">Data Type</Label>
                <Select value={dataType} onValueChange={setDataType}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="cve">CVE</SelectItem>
                    <SelectItem value="cwe">CWE</SelectItem>
                    <SelectItem value="capec">CAPEC</SelectItem>
                    <SelectItem value="attack">ATT&CK</SelectItem>
                  </SelectContent>
                </Select>
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
              <div className="space-y-2">
                <Label htmlFor="actions">Actions</Label>
                <div className="flex gap-2">
                  <Button
                    onClick={handleStartSession}
                    disabled={startTypedSession.isPending || !sessionId}
                    className="w-full"
                  >
                    <Play className="h-4 w-4 mr-2" />
                    Start Session
                  </Button>
                </div>
              </div>
            </div>
          )}

          {/* Control Buttons */}
          <div className="flex gap-2">
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
                  disabled={stopSession.isPending || isStopping}
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
                  disabled={stopSession.isPending || isStopping}
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
                  <span className="font-medium">Data Type:</span>{" "}
                  <span className="text-muted-foreground">{formatDataType(sessionStatus.dataType || 'cve')}</span>
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
                {sessionStatus.errorMessage && (
                  <div className="col-span-2">
                    <span className="font-medium">Error:</span>{" "}
                    <span className="text-destructive">{sessionStatus.errorMessage}</span>
                  </div>
                )}
                {sessionStatus.progress && Object.keys(sessionStatus.progress).length > 0 && (
                  <div className="col-span-2 pt-2 mt-2 border-t">
                    <div className="font-medium mb-2">Progress Details:</div>
                    {Object.entries(sessionStatus.progress).map(([dt, progress]) => {
                      const typedProgress = progress as {
                        totalCount: number;
                        processedCount: number;
                        errorCount: number;
                        startTime: string;
                        lastUpdate: string;
                        errorMessage?: string;
                      };
                      return (
                        <div key={dt} className="mb-1 text-xs">
                          <span className="font-medium">{formatDataType(dt)}:</span>{" "}
                          <span>Processed: {typedProgress.processedCount}/{typedProgress.totalCount} ({Math.round((typedProgress.processedCount / Math.max(typedProgress.totalCount, 1)) * 100)}%)</span>
                          {typedProgress.errorCount > 0 && (
                            <span className="ml-2 text-destructive">Errors: {typedProgress.errorCount}</span>
                          )}
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
