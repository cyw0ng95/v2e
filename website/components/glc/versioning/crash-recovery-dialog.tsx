'use client';

/**
 * GLC Crash Recovery Dialog
 *
 * Dialog for recovering unsaved changes after a crash.
 */

import React, { useState, useEffect, useCallback } from 'react';
import {
  AlertTriangle,
  RefreshCw,
  Trash2,
  Clock,
  FileJson,
  AlertCircle,
} from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  checkCrashRecovery,
  recoverFromCrash,
  clearCrashRecovery,
  formatRecoveryAge,
} from '@/lib/glc/versioning';
import type { Graph } from '@/lib/glc/types';

interface CrashRecoveryDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  graphId: string;
  currentVersion: number;
  onRecover: (graph: Graph) => void;
  onDismiss: () => void;
}

/**
 * Crash Recovery Dialog Component
 */
export function CrashRecoveryDialog({
  open,
  onOpenChange,
  graphId,
  currentVersion,
  onRecover,
  onDismiss,
}: CrashRecoveryDialogProps) {
  const [recoveryData, setRecoveryData] = useState<{
    hasRecovery: boolean;
    age: number;
    formattedAge: string;
    version: number;
  } | null>(null);
  const [isRecovering, setIsRecovering] = useState(false);
  const [isDismissing, setIsDismissing] = useState(false);

  // Check for crash recovery on mount
  useEffect(() => {
    if (open && graphId) {
      const result = checkCrashRecovery(graphId);
      if (result.hasRecovery && result.data) {
        setRecoveryData({
          hasRecovery: true,
          age: result.age,
          formattedAge: formatRecoveryAge(result.age),
          version: result.data.version,
        });
      } else {
        setRecoveryData({ hasRecovery: false, age: 0, formattedAge: '', version: 0 });
      }
    }
  }, [open, graphId]);

  // Handle recovery
  const handleRecover = async () => {
    setIsRecovering(true);

    try {
      const graph = recoverFromCrash(graphId);
      if (graph) {
        onRecover(graph);
        clearCrashRecovery(graphId);
        onOpenChange(false);
      }
    } catch (error) {
      console.error('Failed to recover:', error);
    } finally {
      setIsRecovering(false);
    }
  };

  // Handle dismiss
  const handleDismiss = async () => {
    setIsDismissing(true);

    try {
      clearCrashRecovery(graphId);
      onDismiss();
      onOpenChange(false);
    } catch (error) {
      console.error('Failed to dismiss:', error);
    } finally {
      setIsDismissing(false);
    }
  };

  // No recovery data yet
  if (!recoveryData) {
    return null;
  }

  // No crash recovery found
  if (!recoveryData.hasRecovery) {
    return null;
  }

  const isNewerThanCurrent = recoveryData.version > currentVersion;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[450px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-amber-500" />
            Unsaved Changes Found
          </DialogTitle>
          <DialogDescription>
            We found unsaved changes from your previous session. Would you like to
            recover them?
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Recovery info */}
          <div className="rounded-lg border border-amber-200 dark:border-amber-800 bg-amber-50/50 dark:bg-amber-900/20 p-4">
            <div className="flex items-start gap-3">
              <FileJson className="h-5 w-5 text-amber-600 dark:text-amber-400 flex-shrink-0 mt-0.5" />
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="font-medium text-amber-700 dark:text-amber-300">
                    Version {recoveryData.version}
                  </span>
                  {isNewerThanCurrent && (
                    <Badge variant="outline" className="bg-green-100 dark:bg-green-900/50 text-green-700 dark:text-green-300 border-green-300 dark:border-green-700">
                      Newer than current
                    </Badge>
                  )}
                </div>
                <div className="flex items-center gap-1 text-sm text-amber-600 dark:text-amber-400 mt-1">
                  <Clock className="h-3.5 w-3.5" />
                  Saved {recoveryData.formattedAge}
                </div>
              </div>
            </div>
          </div>

          {/* Warning */}
          <div className="flex items-start gap-3 text-sm text-gray-600 dark:text-gray-400">
            <AlertCircle className="h-4 w-4 text-gray-400 flex-shrink-0 mt-0.5" />
            <p>
              Recovering will replace your current graph with the recovered version.
              You can always undo this action.
            </p>
          </div>
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button
            variant="outline"
            onClick={handleDismiss}
            disabled={isDismissing || isRecovering}
            className="text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-900/20"
          >
            {isDismissing ? (
              <>
                <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                Dismissing...
              </>
            ) : (
              <>
                <Trash2 className="h-4 w-4 mr-2" />
                Discard Changes
              </>
            )}
          </Button>
          <Button
            onClick={handleRecover}
            disabled={isRecovering || isDismissing}
            className="bg-blue-600 hover:bg-blue-700"
          >
            {isRecovering ? (
              <>
                <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                Recovering...
              </>
            ) : (
              <>
                <RefreshCw className="h-4 w-4 mr-2" />
                Recover Changes
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

/**
 * Hook to check and manage crash recovery
 */
export function useCrashRecoveryDialog(graphId: string | null) {
  const [showDialog, setShowDialog] = useState(false);

  // Check for crash recovery - use callback to avoid calling setState in effect
  const checkRecovery = useCallback(() => {
    if (graphId) {
      const result = checkCrashRecovery(graphId);
      return result.hasRecovery;
    }
    return false;
  }, [graphId]);

  // Check recovery on mount and when graphId changes
  useEffect(() => {
    const hasRecovery = checkRecovery();
    // Use requestAnimationFrame to defer setState outside of effect
    const rafId = requestAnimationFrame(() => {
      setShowDialog(hasRecovery);
    });
    return () => cancelAnimationFrame(rafId);
  }, [checkRecovery]);

  const handleRecover = (onRecoverCallback: (graph: Graph) => void) => {
    return (graph: Graph) => {
      onRecoverCallback(graph);
      setShowDialog(false);
    };
  };

  const handleDismiss = () => {
    if (graphId) {
      clearCrashRecovery(graphId);
    }
    setShowDialog(false);
  };

  return {
    showDialog,
    setShowDialog,
    handleRecover,
    handleDismiss,
    hasRecovery: showDialog,
  };
}

export default CrashRecoveryDialog;
