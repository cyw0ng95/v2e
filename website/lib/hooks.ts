import { useState, useEffect, useMemo, useCallback, useRef } from 'react';
import { rpcClient } from './rpc-client';
import { createLogger } from './logger';
import { AttackListResponse, AttackTechnique } from './types';

// Create logger for hooks
const logger = createLogger('hooks');

/**
 * Common options for abortable data fetching
 */
interface AbortableFetchOptions {
  signal?: AbortSignal;
}

/**
 * Hook cleanup tracker - ensures we only update state if component is still mounted
 * and prevents race conditions during rapid unmount/remount
 */
function useMountedState() {
  const isMountedRef = useRef<boolean>(true);

  useEffect(() => {
    isMountedRef.current = true;

    return () => {
      isMountedRef.current = false;
    };
  }, []);

  return {
    isMounted: () => isMountedRef.current,
    setStateIfMounted: <T>(setter: React.Dispatch<React.SetStateAction<T>>) => {
      if (isMountedRef.current) {
        setter(undefined as T);
      }
    },
  };
}

/**
 * Generic hook factory for abortable data fetching
 * Prevents race conditions during rapid unmount by:
 * 1. Using AbortController to cancel pending requests
 * 2. Tracking mount state to prevent state updates after unmount
 * 3. Checking signal state before processing responses
 */
function createAbortableDataHook<TData, TParams>(
  fetchFn: (params: TParams, signal: AbortSignal) => Promise<{ retcode: number; message: string; payload: TData }>,
  errorMessage: string
) {
  return (params: TParams) => {
    const [data, setData] = useState<TData | null>(null);
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const [error, setError] = useState<Error | null>(null);
    const { isMounted, setStateIfMounted } = useMountedState();

    useEffect(() => {
      const abortController = new AbortController();
      const { signal } = abortController;

      const fetchData = async () => {
        if (signal.aborted) return;

        try {
          setIsLoading(true);
          const response = await fetchFn(params, signal);

          if (signal.aborted || !isMounted()) return;

          if (response.retcode !== 0) {
            throw new Error(response.message || errorMessage);
          }

          setData(response.payload);
          setError(null);
        } catch (err: unknown) {
          if (signal.aborted) return;

          const errorObj = err instanceof Error ? err : new Error(String(err));
          setStateIfMounted<Error | null>(() => setError(errorObj));
          logger.error(errorMessage, err, { params });
        } finally {
          if (!signal.aborted) {
            setIsLoading(false);
          }
        }
      };

      fetchData();

      return () => {
        abortController.abort();
      };
    }, [params, fetchFn, errorMessage, isMounted, setStateIfMounted]);

    return { data, isLoading, error };
  };
}

interface AttackQueryParams {
  offset?: number;
  limit?: number;
  search?: string;
}

export function useAttackTechniques(params: AttackQueryParams = {}) {
  const [data, setData] = useState<AttackListResponse | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  const { offset = 0, limit = 100, search = '' } = params;
  const { isMounted, setStateIfMounted } = useMountedState();

  useEffect(() => {
    // Use AbortController to cancel pending requests on unmount
    const abortController = new AbortController();
    const { signal } = abortController;

    const fetchData = async () => {
      // Don't fetch if already aborted
      if (signal.aborted) return;

      try {
        setIsLoading(true);
        const response = await rpcClient.listAttackTechniques(offset, limit);

        // Check if component was unmounted or request was aborted
        if (signal.aborted || !isMounted()) return;

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch ATT&CK techniques');
        }

        setData(response.payload);
        setError(null);
      } catch (err: unknown) {
        // Don't set error if request was aborted
        if (signal.aborted) return;

        const error = err instanceof Error ? err : new Error(String(err));
        setStateIfMounted<Error | null>(() => setError(error));
        logger.error('Error fetching ATT&CK techniques', err, { offset, limit });
      } finally {
        // Only update loading state if not aborted
        if (!signal.aborted) {
          setIsLoading(false);
        }
      }
    };

    fetchData();

    // Cleanup function - abort request and prevent state updates
    return () => {
      abortController.abort();
    };
  }, [offset, limit, search, isMounted, setStateIfMounted]);

  return { data, isLoading, error };
}

export function useAttackTactics(params: AttackQueryParams = {}) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  const { offset = 0, limit = 100, search = '' } = params;
  const { isMounted, setStateIfMounted } = useMountedState();

  useEffect(() => {
    const abortController = new AbortController();
    const { signal } = abortController;

    const fetchData = async () => {
      if (signal.aborted) return;

      try {
        setIsLoading(true);
        const response = await rpcClient.listAttackTactics(offset, limit);

        if (signal.aborted || !isMounted()) return;

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch ATT&CK tactics');
        }

        setData(response.payload);
        setError(null);
      } catch (err: unknown) {
        if (signal.aborted) return;

        const errorObj = err instanceof Error ? err : new Error(String(err));
        setStateIfMounted<Error | null>(() => setError(errorObj));
        logger.error('Error fetching ATT&CK tactics', err, { offset, limit });
      } finally {
        if (!signal.aborted) {
          setIsLoading(false);
        }
      }
    };

    fetchData();

    return () => {
      abortController.abort();
    };
  }, [offset, limit, search, isMounted, setStateIfMounted]);

  return { data, isLoading, error };
}

export function useAttackMitigations(params: AttackQueryParams = {}) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  const { offset = 0, limit = 100, search = '' } = params;
  const { isMounted, setStateIfMounted } = useMountedState();

  useEffect(() => {
    const abortController = new AbortController();
    const { signal } = abortController;

    const fetchData = async () => {
      if (signal.aborted) return;

      try {
        setIsLoading(true);
        const response = await rpcClient.listAttackMitigations(offset, limit);

        if (signal.aborted || !isMounted()) return;

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch ATT&CK mitigations');
        }

        setData(response.payload);
        setError(null);
      } catch (err: unknown) {
        if (signal.aborted) return;

        const errorObj = err instanceof Error ? err : new Error(String(err));
        setStateIfMounted<Error | null>(() => setError(errorObj));
        logger.error('Error fetching ATT&CK mitigations', err, { offset, limit });
      } finally {
        if (!signal.aborted) {
          setIsLoading(false);
        }
      }
    };

    fetchData();

    return () => {
      abortController.abort();
    };
  }, [offset, limit, search, isMounted, setStateIfMounted]);

  return { data, isLoading, error };
}

// CAPEC Hooks
export function useCAPEC(capecId?: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!capecId) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getCAPEC(capecId);
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch CAPEC');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching CAPEC', err, { capecId });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    // Cleanup function
    return () => {};
  }, [capecId]);

  return { data, isLoading, error };
}

export function useCAPECList(offset: number = 0, limit: number = 100) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);
  const { isMounted, setStateIfMounted } = useMountedState();

  useEffect(() => {
    const abortController = new AbortController();
    const { signal } = abortController;

    const fetchData = async () => {
      if (signal.aborted) return;

      try {
        setIsLoading(true);
        const response = await rpcClient.listCAPECs(offset, limit);

        if (signal.aborted || !isMounted()) return;

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch CAPEC list');
        }

        setData(response.payload);
        setError(null);
      } catch (err: unknown) {
        if (signal.aborted) return;

        const errorObj = err instanceof Error ? err : new Error(String(err));
        setStateIfMounted<Error | null>(() => setError(errorObj));
        logger.error('Error fetching CAPEC list', err, { offset, limit });
      } finally {
        if (!signal.aborted) {
          setIsLoading(false);
        }
      }
    };

    fetchData();

    return () => {
      abortController.abort();
    };
  }, [offset, limit, isMounted, setStateIfMounted]);

  return { data, isLoading, error };
}

// Session Management Hooks
export function useSessionStatus() {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);
  const { isMounted, setStateIfMounted } = useMountedState();

  useEffect(() => {
    const abortController = new AbortController();
    const { signal } = abortController;

    const fetchData = async () => {
      if (signal.aborted) return;

      try {
        setIsLoading(true);
        const response = await rpcClient.getSessionStatus();

        if (signal.aborted || !isMounted()) return;

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch session status');
        }

        setData(response.payload);
        setError(null);
      } catch (err: unknown) {
        if (signal.aborted) return;

        const errorObj = err instanceof Error ? err : new Error(String(err));
        setStateIfMounted<Error | null>(() => setError(errorObj));
        logger.error('Error fetching session status', err);
      } finally {
        if (!signal.aborted) {
          setIsLoading(false);
        }
      }
    };

    fetchData();

    return () => {
      abortController.abort();
    };
  }, [isMounted, setStateIfMounted]);

  return { data, isLoading, error };
}

export function useStartSession() {
  const [isPending, setIsPending] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);

  const mutate = async (params: { sessionId: string; startIndex?: number; resultsPerBatch?: number }, options?: { onSuccess?: (data: any) => void; onError?: (error: Error) => void }) => {
    // Destructure outside try block so variables are in scope for error logging
    const { sessionId, startIndex, resultsPerBatch } = params;

    try {
      setIsPending(true);
      setError(null);

      const response = await rpcClient.startSession(sessionId, startIndex, resultsPerBatch);

      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to start session');
      }

      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      logger.error('Error starting session', err, { sessionId, startIndex, resultsPerBatch });
      if (options?.onError) {
        options.onError(err);
      }
    } finally {
      setIsPending(false);
    }
  };

  return { mutate, isPending, error };
}

export function useStartTypedSession() {
  const [isPending, setIsPending] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);

  const mutate = async (params: { sessionId: string; dataType: string; startIndex?: number; resultsPerBatch?: number; params?: Record<string, unknown> }, options?: { onSuccess?: (data: any) => void; onError?: (error: Error) => void }) => {
    // Destructure outside try block so variables are in scope for error logging
    const { sessionId, dataType, startIndex, resultsPerBatch, params: extraParams } = params;

    try {
      setIsPending(true);
      setError(null);

      const response = await rpcClient.startTypedSession(sessionId, dataType, startIndex, resultsPerBatch, extraParams);

      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to start typed session');
      }

      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      logger.error('Error starting typed session', err, { sessionId, dataType, startIndex, resultsPerBatch });
      if (options?.onError) {
        options.onError(err);
      }
    } finally {
      setIsPending(false);
    }
  };

  return { mutate, isPending, error };
}

export function useStartCWEImport() {
  const [isPending, setIsPending] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);

  const mutate = async (params?: Record<string, unknown>, options?: { onSuccess?: (data: any) => void; onError?: (error: Error) => void }) => {
    try {
      setIsPending(true);
      setError(null);

      // TODO: Implement CWE import RPC method
      throw new Error('CWE import not yet implemented');

    } catch (err: any) {
      setError(err);
      logger.error('Error starting CWE import', err);
      if (options?.onError) {
        options.onError(err);
      }
    } finally {
      setIsPending(false);
    }
  };

  return { mutate, isPending, error };
}

export function useStartCAPECImport() {
  const [isPending, setIsPending] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);

  const mutate = async (params?: Record<string, unknown>, options?: { onSuccess?: (data: any) => void; onError?: (error: Error) => void }) => {
    try {
      setIsPending(true);
      setError(null);

      // TODO: Implement CAPEC import RPC method
      throw new Error('CAPEC import not yet implemented');

    } catch (err: any) {
      setError(err);
      logger.error('Error starting CAPEC import', err);
      if (options?.onError) {
        options.onError(err);
      }
    } finally {
      setIsPending(false);
    }
  };

  return { mutate, isPending, error };
}

export function useStartATTACKImport() {
  const [isPending, setIsPending] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);

  const mutate = async (params?: Record<string, unknown>, options?: { onSuccess?: (data: any) => void; onError?: (error: Error) => void }) => {
    try {
      setIsPending(true);
      setError(null);

      // TODO: Implement ATT&CK import RPC method
      throw new Error('ATT&CK import not yet implemented');

    } catch (err: any) {
      setError(err);
      logger.error('Error starting ATT&CK import', err);
      if (options?.onError) {
        options.onError(err);
      }
    } finally {
      setIsPending(false);
    }
  };

  return { mutate, isPending, error };
}

export function useStopSession() {
  const [isPending, setIsPending] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);
  // Track if we've already seen a "run not active" error to prevent retries
  const [seenInactiveError, setSeenInactiveError] = useState<boolean>(false);

  const mutate = async (_: undefined, options?: { onSuccess?: (data: any) => void; onError?: (error: Error) => void }) => {
    try {
      // Prevent retries if we've already seen the "run not active" error
      if (seenInactiveError) {
        logger.warn('Skipping stop session request - already saw "run not active" error');
        return;
      }
      
      setIsPending(true);
      setError(null);
      
      const response = await rpcClient.stopSession();
      
      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to stop session');
      }
      
      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      logger.error('Error stopping session', err);
      
      // Prevent infinite retry loop for "run not active" errors
      if (err.message && err.message.includes('run not active')) {
        logger.warn('Session already inactive, preventing future retries');
        setSeenInactiveError(true);
        // Don't call onError for this specific error to prevent retry loops
        return;
      }
      
      if (options?.onError) {
        options.onError(err);
      }
    } finally {
      setIsPending(false);
    }
  };

  // Reset the seen error state when component unmounts or when we want to allow retries again
  const reset = () => {
    setSeenInactiveError(false);
    setError(null);
  };

  return { mutate, isPending, error, reset };
}

export function usePauseJob() {
  const [isPending, setIsPending] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);

  const mutate = async (_: undefined, options?: { onSuccess?: (data: any) => void; onError?: (error: Error) => void }) => {
    try {
      setIsPending(true);
      setError(null);
      
      const response = await rpcClient.pauseJob();
      
      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to pause job');
      }
      
      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      logger.error('Error pausing job', err);
      if (options?.onError) {
        options.onError(err);
      }
    } finally {
      setIsPending(false);
    }
  };

  return { mutate, isPending, error };
}

export function useResumeJob() {
  const [isPending, setIsPending] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);

  const mutate = async (_: undefined, options?: { onSuccess?: (data: any) => void; onError?: (error: Error) => void }) => {
    try {
      setIsPending(true);
      setError(null);
      
      const response = await rpcClient.resumeJob();
      
      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to resume job');
      }
      
      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      logger.error('Error resuming job', err);
      if (options?.onError) {
        options.onError(err);
      }
    } finally {
      setIsPending(false);
    }
  };

  return { mutate, isPending, error };
}

// CWE Views Hooks
export function useCWEViews(offset: number = 0, limit: number = 100) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);
  const { isMounted, setStateIfMounted } = useMountedState();

  useEffect(() => {
    const abortController = new AbortController();
    const { signal } = abortController;

    const fetchData = async () => {
      if (signal.aborted) return;

      try {
        setIsLoading(true);
        const response = await rpcClient.listCWEViews(offset, limit);

        if (signal.aborted || !isMounted()) return;

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch CWE views');
        }

        setData(response.payload);
        setError(null);
      } catch (err: unknown) {
        if (signal.aborted) return;

        const errorObj = err instanceof Error ? err : new Error(String(err));
        setStateIfMounted<Error | null>(() => setError(errorObj));
        logger.error('Error fetching CWE views', err, { offset, limit });
      } finally {
        if (!signal.aborted) {
          setIsLoading(false);
        }
      }
    };

    fetchData();

    return () => {
      abortController.abort();
    };
  }, [offset, limit, isMounted, setStateIfMounted]);

  // Track refetch abort controller separately
  const abortControllerRef = useRef<AbortController | null>(null);

  const refetch = useCallback(async () => {
    // Cancel any in-progress refetch
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    abortControllerRef.current = new AbortController();
    const { signal } = abortControllerRef.current;

    try {
      setIsLoading(true);
      const response = await rpcClient.listCWEViews(offset, limit);

      if (signal.aborted || !isMounted()) return;

      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to fetch CWE views');
      }

      setData(response.payload);
      setError(null);
    } catch (err: unknown) {
      if (signal.aborted) return;

      const errorObj = err instanceof Error ? err : new Error(String(err));
      setStateIfMounted<Error | null>(() => setError(errorObj));
      console.error('Error fetching CWE views:', err);
    } finally {
      if (!signal.aborted) {
        setIsLoading(false);
      }
    }
  }, [offset, limit, isMounted, setStateIfMounted]);

  return { data, isLoading, error, refetch };
}

export function useCWEJobStatus() {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        // We'll use the session status for now as a placeholder for job status
        const response = await rpcClient.getSessionStatus();
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch CWE job status');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching CWE job status', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    // Cleanup function
    return () => {};
  }, []);

  return { data, isLoading, error };
}

export function useStartCWEViewJob() {
  const [isPending, setIsPending] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);

  const mutate = async (params: { sessionId?: string; startIndex?: number; resultsPerBatch?: number } = {}, options?: { onSuccess?: (data: any) => void; onError?: (error: Error) => void }) => {
    // Destructure outside try block so variables are in scope for error logging
    const { sessionId, startIndex, resultsPerBatch } = params;

    try {
      setIsPending(true);
      setError(null);

      const response = await rpcClient.startCWEViewJob(sessionId, startIndex, resultsPerBatch);

      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to start CWE view job');
      }

      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      logger.error('Error starting CWE view job', err, { sessionId, startIndex, resultsPerBatch });
      if (options?.onError) {
        options.onError(err);
      }
    } finally {
      setIsPending(false);
    }
  };

  return { mutate, isPending, error };
}

export function useStopCWEViewJob() {
  const [isPending, setIsPending] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);

  const mutate = async (params: { sessionId?: string } = {}, options?: { onSuccess?: (data: any) => void; onError?: (error: Error) => void }) => {
    // Destructure outside try block so variables are in scope for error logging
    const { sessionId } = params;

    try {
      setIsPending(true);
      setError(null);

      const response = await rpcClient.stopCWEViewJob(sessionId);

      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to stop CWE view job');
      }

      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      logger.error('Error stopping CWE view job', err, { sessionId });
      if (options?.onError) {
        options.onError(err);
      }
    } finally {
      setIsPending(false);
    }
  };

  return { mutate, isPending, error };
}

// System Metrics Hook
export function useSysMetrics() {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);
  const { isMounted, setStateIfMounted } = useMountedState();

  useEffect(() => {
    const abortController = new AbortController();
    const { signal } = abortController;

    const fetchData = async () => {
      if (signal.aborted) return;

      try {
        setIsLoading(true);
        const response = await rpcClient.getSysMetrics();

        if (signal.aborted || !isMounted()) return;

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch system metrics');
        }

        setData(response.payload);
        setError(null);
      } catch (err: unknown) {
        if (signal.aborted) return;

        const errorObj = err instanceof Error ? err : new Error(String(err));
        setStateIfMounted<Error | null>(() => setError(errorObj));
        logger.error('Error fetching system metrics', err);
      } finally {
        if (!signal.aborted) {
          setIsLoading(false);
        }
      }
    };

    fetchData();

    return () => {
      abortController.abort();
    };
  }, [isMounted, setStateIfMounted]);

  return { data, isLoading, error };
}

// CVE Hooks
export function useCVEList(offset: number = 0, limit: number = 100) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);
  const { isMounted, setStateIfMounted } = useMountedState();

  useEffect(() => {
    const abortController = new AbortController();
    const { signal } = abortController;

    const fetchData = async () => {
      if (signal.aborted) return;

      try {
        setIsLoading(true);
        const response = await rpcClient.listCVEs(offset, limit);

        if (signal.aborted || !isMounted()) return;

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch CVE list');
        }

        setData(response.payload);
        setError(null);
      } catch (err: unknown) {
        if (signal.aborted) return;

        const errorObj = err instanceof Error ? err : new Error(String(err));
        setStateIfMounted<Error | null>(() => setError(errorObj));
        logger.error('Error fetching CVE list', err, { offset, limit });
      } finally {
        if (!signal.aborted) {
          setIsLoading(false);
        }
      }
    };

    fetchData();

    return () => {
      abortController.abort();
    };
  }, [offset, limit, isMounted, setStateIfMounted]);

  // Derive derived state to prevent unnecessary re-renders
  const derivedState = useMemo(() => ({
    data,
    isLoading,
    error,
    hasData: !!data?.cves?.length,
    total: data?.total || 0,
  }), [data, isLoading, error]);

  return derivedState;
}

export function useCVECount() {
  const [data, setData] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.countCVEs();
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch CVE count');
        }
        
        setData(response.payload?.count || 0);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching CVE count', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    // Cleanup function
    return () => {};
  }, []);

  return { data, isLoading, error };
}

// ATT&CK Software Hook
export function useAttackSoftware(params: AttackQueryParams = {}) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  const { offset = 0, limit = 100, search = '' } = params;

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listAttackSoftware(offset, limit);
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch ATT&CK software');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching ATT&CK software', err, { offset, limit });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    // Cleanup function
    return () => {};
  }, [offset, limit, search]);

  return { data, isLoading, error };
}

// ATT&CK Groups Hook
export function useAttackGroups(params: AttackQueryParams = {}) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  const { offset = 0, limit = 100, search = '' } = params;

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listAttackGroups(offset, limit);
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch ATT&CK groups');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Failed to fetch ATT&CK groups', err, { offset, limit });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    // Cleanup function
    return () => {};
  }, [offset, limit, search]);

  return { data, isLoading, error };
}

// CWE Hooks
export function useCWEList(params: { offset?: number; limit?: number; search?: string } = {}) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  const { offset = 0, limit = 100, search = '' } = params;
  const { isMounted, setStateIfMounted } = useMountedState();

  useEffect(() => {
    const abortController = new AbortController();
    const { signal } = abortController;

    const fetchData = async () => {
      if (signal.aborted) return;

      try {
        setIsLoading(true);
        const response = await rpcClient.listCWEs({ offset, limit, search });

        if (signal.aborted || !isMounted()) return;

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch CWE list');
        }

        setData(response.payload);
        setError(null);
      } catch (err: unknown) {
        if (signal.aborted) return;

        const errorObj = err instanceof Error ? err : new Error(String(err));
        setStateIfMounted<Error | null>(() => setError(errorObj));
        logger.error('Error fetching CWE list', err, { offset, limit, search });
      } finally {
        if (!signal.aborted) {
          setIsLoading(false);
        }
      }
    };

    fetchData();

    return () => {
      abortController.abort();
    };
  }, [offset, limit, search, isMounted, setStateIfMounted]);

  return { data, isLoading, error };
}

// ============================================================================
// ASVS Hooks
// ============================================================================

export function useASVSList(params: { offset?: number; limit?: number; chapter?: string; level?: number } = {}) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  const { offset = 0, limit = 100, chapter = '', level = 0 } = params;

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listASVS({ offset, limit, chapter, level });
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch ASVS list');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        console.error('Error fetching ASVS list:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    // Cleanup function
    return () => {};
  }, [offset, limit, chapter, level]);

  return { data, isLoading, error };
}

export function useASVS(requirementId: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!requirementId) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getASVS(requirementId);
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch ASVS requirement');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        console.error('Error fetching ASVS requirement:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    // Cleanup function
    return () => {};
  }, [requirementId]);

  return { data, isLoading, error };
}

export function useImportASVS() {
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);

  const importASVS = async (url: string) => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await rpcClient.importASVS(url);
      
      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to import ASVS requirements');
      }
      
      return response.payload;
    } catch (err: any) {
      setError(err);
      console.error('Error importing ASVS:', err);
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  return { importASVS, isLoading, error };
}

// Individual ATT&CK item hooks
export function useAttackTechnique(id: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!id) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getAttackTechnique(id);
        
        if (response.retcode !== 0) {
          throw new Error(response.message || `Failed to fetch attack technique ${id}`);
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching attack technique', err, { id });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    // Cleanup function
    return () => {};
  }, [id]);

  return { data, isLoading, error };
}

export function useAttackTactic(id: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!id) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getAttackTactic(id);
        
        if (response.retcode !== 0) {
          throw new Error(response.message || `Failed to fetch attack tactic ${id}`);
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching attack tactic', err, { id });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    // Cleanup function
    return () => {};
  }, [id]);

  return { data, isLoading, error };
}

export function useAttackMitigation(id: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!id) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getAttackMitigation(id);
        
        if (response.retcode !== 0) {
          throw new Error(response.message || `Failed to fetch attack mitigation ${id}`);
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching attack mitigation', err, { id });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    // Cleanup function
    return () => {};
  }, [id]);

  return { data, isLoading, error };
}

export function useAttackSoftwareById(id: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!id) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getAttackSoftware(id);
        
        if (response.retcode !== 0) {
          throw new Error(response.message || `Failed to fetch attack software ${id}`);
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching attack software', err, { id });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    // Cleanup function
    return () => {};
  }, [id]);

  return { data, isLoading, error };
}

export function useAttackGroupById(id: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!id) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getAttackGroup(id);

        if (response.retcode !== 0) {
          throw new Error(response.message || `Failed to fetch attack group ${id}`);
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching attack group', err, { id });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    // Cleanup function
    return () => {};
  }, [id]);

  return { data, isLoading, error };
}

// ============================================================================
// SSG (SCAP Security Guide) Hooks
// ============================================================================

// SSG Import Job Hooks
export function useSSGImportStatus() {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getSSGImportStatus();

        if (response.retcode !== 0) {
          // Not an error if no active job
          setData(null);
          setError(null);
          return;
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        // Not an error if no active job
        setData(null);
        setError(null);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, []);

  const refetch = async () => {
    try {
      setIsLoading(true);
      const response = await rpcClient.getSSGImportStatus();

      if (response.retcode !== 0) {
        setData(null);
        setError(null);
        return;
      }

      setData(response.payload);
      setError(null);
    } catch (err: any) {
      setData(null);
      setError(null);
    } finally {
      setIsLoading(false);
    }
  };

  return { data, isLoading, error, refetch };
}

export function useStartSSGImportJob() {
  const [isPending, setIsPending] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);

  const mutate = async (params: { runId?: string } = {}, options?: { onSuccess?: (data: any) => void; onError?: (error: Error) => void }) => {
    const { runId } = params;

    try {
      setIsPending(true);
      setError(null);

      const response = await rpcClient.startSSGImportJob(runId);

      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to start SSG import job');
      }

      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      logger.error('Error starting SSG import job', err, { runId });
      if (options?.onError) {
        options.onError(err);
      }
    } finally {
      setIsPending(false);
    }
  };

  return { mutate, isPending, error };
}

export function useStopSSGImportJob() {
  const [isPending, setIsPending] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);

  const mutate = async (_: undefined, options?: { onSuccess?: (data: any) => void; onError?: (error: Error) => void }) => {
    try {
      setIsPending(true);
      setError(null);

      const response = await rpcClient.stopSSGImportJob();

      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to stop SSG import job');
      }

      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      logger.error('Error stopping SSG import job', err);
      if (options?.onError) {
        options.onError(err);
      }
    } finally {
      setIsPending(false);
    }
  };

  return { mutate, isPending, error };
}

export function usePauseSSGImportJob() {
  const [isPending, setIsPending] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);

  const mutate = async (_: undefined, options?: { onSuccess?: (data: any) => void; onError?: (error: Error) => void }) => {
    try {
      setIsPending(true);
      setError(null);

      const response = await rpcClient.pauseSSGImportJob();

      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to pause SSG import job');
      }

      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      logger.error('Error pausing SSG import job', err);
      if (options?.onError) {
        options.onError(err);
      }
    } finally {
      setIsPending(false);
    }
  };

  return { mutate, isPending, error };
}

export function useResumeSSGImportJob() {
  const [isPending, setIsPending] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);

  const mutate = async (params: { runId: string }, options?: { onSuccess?: (data: any) => void; onError?: (error: Error) => void }) => {
    const { runId } = params;

    try {
      setIsPending(true);
      setError(null);

      const response = await rpcClient.resumeSSGImportJob(runId);

      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to resume SSG import job');
      }

      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      logger.error('Error resuming SSG import job', err, { runId });
      if (options?.onError) {
        options.onError(err);
      }
    } finally {
      setIsPending(false);
    }
  };

  return { mutate, isPending, error };
}

// SSG Guide Hooks
export function useSSGGuides(product?: string, profileId?: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listSSGGuides(product, profileId);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch SSG guides');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching SSG guides', err, { product, profileId });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [product, profileId]);

  const refetch = async () => {
    try {
      setIsLoading(true);
      const response = await rpcClient.listSSGGuides(product, profileId);

      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to fetch SSG guides');
      }

      setData(response.payload);
      setError(null);
    } catch (err: any) {
      setError(err);
      logger.error('Error fetching SSG guides', err);
    } finally {
      setIsLoading(false);
    }
  };

  return { data, isLoading, error, refetch };
}

export function useSSGTree(guideId: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!guideId) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getSSGTree(guideId);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch SSG tree');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching SSG tree', err, { guideId });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [guideId]);

  return { data, isLoading, error };
}

export function useSSGGroup(id: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!id) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getSSGGroup(id);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch SSG group');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching SSG group', err, { id });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [id]);

  return { data, isLoading, error };
}

export function useSSGRule(id: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!id) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getSSGRule(id);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch SSG rule');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching SSG rule', err, { id });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [id]);

  return { data, isLoading, error };
}

// SSG Table Hooks

export function useSSGTables(product?: string, tableType?: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listSSGTables(product, tableType);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch SSG tables');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching SSG tables', err, { product, tableType });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [product, tableType]);

  const refetch = useCallback(async () => {
    try {
      setIsLoading(true);
      const response = await rpcClient.listSSGTables(product, tableType);

      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to fetch SSG tables');
      }

      setData(response.payload);
      setError(null);
    } catch (err: any) {
      setError(err);
      logger.error('Error refetching SSG tables', err, { product, tableType });
    } finally {
      setIsLoading(false);
    }
  }, [product, tableType]);

  return { data, isLoading, error, refetch };
}

export function useSSGTable(id: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!id) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getSSGTable(id);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch SSG table');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching SSG table', err, { id });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [id]);

  return { data, isLoading, error };
}

export function useSSGTableEntries(tableId: string, offset?: number, limit?: number) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!tableId) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getSSGTableEntries(tableId, offset, limit);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch SSG table entries');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching SSG table entries', err, { tableId, offset, limit });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [tableId, offset, limit]);

  return { data, isLoading, error };
}
// ============================================================================
// SSG Manifest Hooks
// ============================================================================

export function useSSGManifests(product?: string, limit?: number, offset?: number) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listSSGManifests(product, limit, offset);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch SSG manifests');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching SSG manifests', err, { product, limit, offset });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [product, limit, offset]);

  return { data, isLoading, error };
}

export function useSSGManifest(manifestId: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!manifestId) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getSSGManifest(manifestId);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch SSG manifest');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching SSG manifest', err, { manifestId });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [manifestId]);

  return { data, isLoading, error };
}

export function useSSGProfiles(product?: string, profileId?: string, limit?: number, offset?: number) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listSSGProfiles(product, profileId, limit, offset);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch SSG profiles');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching SSG profiles', err, { product, profileId, limit, offset });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [product, profileId, limit, offset]);

  return { data, isLoading, error };
}

export function useSSGProfile(profileId: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!profileId) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getSSGProfile(profileId);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch SSG profile');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching SSG profile', err, { profileId });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [profileId]);

  return { data, isLoading, error };
}

export function useSSGProfileRules(profileId: string, limit?: number, offset?: number) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!profileId) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getSSGProfileRules(profileId, limit, offset);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch SSG profile rules');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching SSG profile rules', err, { profileId, limit, offset });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [profileId, limit, offset]);

  return { data, isLoading, error };
}

// ============================================================================
// SSG Data Stream Hooks
// ============================================================================

export function useSSGDataStreams(product?: string, limit?: number, offset?: number) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listSSGDataStreams(product, limit, offset);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch SSG data streams');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching SSG data streams', err, { product, limit, offset });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [product, limit, offset]);

  const refetch = async () => {
    try {
      setIsLoading(true);
      const response = await rpcClient.listSSGDataStreams(product, limit, offset);

      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to fetch SSG data streams');
      }

      setData(response.payload);
      setError(null);
    } catch (err: any) {
      setError(err);
      logger.error('Error refetching SSG data streams', err, { product, limit, offset });
    } finally {
      setIsLoading(false);
    }
  };

  return { data, isLoading, error, refetch };
}

export function useSSGDataStream(dataStreamId: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!dataStreamId) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getSSGDataStream(dataStreamId);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch SSG data stream');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching SSG data stream', err, { dataStreamId });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [dataStreamId]);

  return { data, isLoading, error };
}

export function useDSProfiles(dataStreamId: string, limit?: number, offset?: number) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!dataStreamId) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listDSProfiles(dataStreamId, limit, offset);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch data stream profiles');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching data stream profiles', err, { dataStreamId, limit, offset });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [dataStreamId, limit, offset]);

  return { data, isLoading, error };
}

export function useDSProfile(profileId: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!profileId) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getDSProfile(profileId);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch data stream profile');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching data stream profile', err, { profileId });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [profileId]);

  return { data, isLoading, error };
}

export function useDSProfileRules(profileId: string, limit?: number, offset?: number) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!profileId) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getDSProfileRules(profileId, limit, offset);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch data stream profile rules');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching data stream profile rules', err, { profileId, limit, offset });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [profileId, limit, offset]);

  return { data, isLoading, error };
}

export function useDSGroups(dataStreamId: string, parentXccdfGroupId?: string, limit?: number, offset?: number) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!dataStreamId) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listDSGroups(dataStreamId, parentXccdfGroupId, limit, offset);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch data stream groups');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching data stream groups', err, { dataStreamId, parentXccdfGroupId, limit, offset });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [dataStreamId, parentXccdfGroupId, limit, offset]);

  return { data, isLoading, error };
}

export function useDSRules(dataStreamId: string, groupXccdfId?: string, severity?: string, limit?: number, offset?: number) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!dataStreamId) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listDSRules(dataStreamId, groupXccdfId, severity, limit, offset);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch data stream rules');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching data stream rules', err, { dataStreamId, groupXccdfId, severity, limit, offset });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [dataStreamId, groupXccdfId, severity, limit, offset]);

  return { data, isLoading, error };
}

export function useDSRule(ruleId: string) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!ruleId) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getDSRule(ruleId);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch data stream rule');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching data stream rule', err, { ruleId });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [ruleId]);

  return { data, isLoading, error };
}

// SSG Cross-Reference Hooks

export function useSSGCrossReferences(params: {
  sourceType?: string;
  sourceId?: string;
  targetType?: string;
  targetId?: string;
  limit?: number;
  offset?: number;
}) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    // Only fetch if we have either source or target params
    if ((!params.sourceType || !params.sourceId) && (!params.targetType || !params.targetId)) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getSSGCrossReferences(params);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch cross-references');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching cross-references', err, params);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [params.sourceType, params.sourceId, params.targetType, params.targetId, params.limit, params.offset]);

  return { data, isLoading, error };
}

export function useSSGRelatedObjects(params: {
  objectType: string;
  objectId: string;
  linkType?: string;
  limit?: number;
  offset?: number;
}) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!params.objectType || !params.objectId) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.findSSGRelatedObjects(params);

        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to find related objects');
        }

        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error finding related objects', err, params);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [params.objectType, params.objectId, params.linkType, params.limit, params.offset]);

  return { data, isLoading, error };
}

// ============================================================================
// UEE (Unified ETL Engine) Hooks
// ============================================================================

/**
 * Hook to fetch the ETL tree with automatic polling
 * @param pollingInterval - How often to poll in milliseconds (default: 5000)
 */
export function useEtlTree(pollingInterval = 5000) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let interval: NodeJS.Timeout;

    const fetchData = async () => {
      try {
        const response = await rpcClient.getEtlTree();
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch ETL tree');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching ETL tree', err);
      } finally {
        setIsLoading(false);
      }
    };

    // Initial fetch
    fetchData();

    // Set up polling
    if (pollingInterval > 0) {
      interval = setInterval(fetchData, pollingInterval);
    }

    return () => {
      if (interval) {
        clearInterval(interval);
      }
    };
  }, [pollingInterval]);

  return { data, isLoading, error };
}

/**
 * Hook to fetch kernel metrics with automatic polling
 * @param pollingInterval - How often to poll in milliseconds (default: 2000)
 */
export function useKernelMetrics(pollingInterval = 2000) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let interval: NodeJS.Timeout;

    const fetchData = async () => {
      try {
        const response = await rpcClient.getKernelMetrics();
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch kernel metrics');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching kernel metrics', err);
      } finally {
        setIsLoading(false);
      }
    };

    // Initial fetch
    fetchData();

    // Set up polling
    if (pollingInterval > 0) {
      interval = setInterval(fetchData, pollingInterval);
    }

    return () => {
      if (interval) {
        clearInterval(interval);
      }
    };
  }, [pollingInterval]);

  return { data, isLoading, error };
}

/**
 * Hook to fetch checkpoints for a specific provider
 */
export function useProviderCheckpoints(
  providerID: string,
  limit = 50,
  offset = 0
) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!providerID) {
      setIsLoading(false);
      return;
    }

    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getProviderCheckpoints(providerID, limit, offset);
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch checkpoints');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        logger.error('Error fetching checkpoints', err, { providerID, limit, offset });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    return () => {};
  }, [providerID, limit, offset]);

  return { data, isLoading, error };
}



// Re-export graph analysis hooks
export * from './hooks/useAnalysisGraph';

// ============================================================================
// Timeout and Progress Utilities
// ============================================================================

/**
 * Configuration for timeout behavior
 */
export interface TimeoutOptions {
  /** Timeout in milliseconds (default: 120000 = 2 minutes) */
  timeout?: number;
  /** Callback when timeout is about to occur (warning threshold) */
  onTimeoutWarning?: (timeRemaining: number) => void;
  /** Warning threshold in milliseconds before timeout (default: 10000 = 10 seconds) */
  warningThreshold?: number;
  /** Whether to show progress to user (default: false) */
  showProgress?: boolean;
}

/**
 * Progress state for long-running operations
 */
export interface ProgressState {
  /** Whether operation is in progress */
  isPending: boolean;
  /** Elapsed time in milliseconds */
  elapsed: number;
  /** Estimated remaining time (null if unknown) */
  remaining: number | null;
  /** Progress percentage (0-100, null if unknown) */
  progress: number | null;
  /** Whether timeout warning should be shown */
  isNearTimeout: boolean;
}

/**
 * Hook for handling long-running operations with timeout and progress
 *
 * Provides automatic timeout handling, progress tracking, and user-facing
 * indicators for operations that may take a long time to complete.
 *
 * @example
 * ```tsx
 * const { execute, state, cancel } = useLongRunningOperation({
 *   timeout: 300000, // 5 minutes
 *   warningThreshold: 30000, // warn at 30 seconds remaining
 *   showProgress: true,
 * });
 * ```
 */
export function useLongRunningOperation<T = unknown>(
  options: TimeoutOptions = {}
) {
  const {
    timeout = 120000, // 2 minutes default
    warningThreshold = 10000, // 10 seconds warning
    showProgress = false,
  } = options;

  const [state, setState] = useState<ProgressState>({
    isPending: false,
    elapsed: 0,
    remaining: null,
    progress: null,
    isNearTimeout: false,
  });

  const startTimeRef = useRef<number | null>(null);
  const timerRef = useRef<NodeJS.Timeout | null>(null);
  const progressTimerRef = useRef<NodeJS.Timeout | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  // Update progress state
  useEffect(() => {
    if (!state.isPending || !startTimeRef.current) return;

    const updateProgress = () => {
      const elapsed = Date.now() - (startTimeRef.current || 0);
      const remaining = Math.max(0, timeout - elapsed);
      const progress = Math.min(100, (elapsed / timeout) * 100);

      setState({
        isPending: true,
        elapsed,
        remaining,
        progress,
        isNearTimeout: remaining <= warningThreshold,
      });
    };

    // Update progress every 100ms for smooth UI
    progressTimerRef.current = setInterval(updateProgress, 100);

    return () => {
      if (progressTimerRef.current) {
        clearInterval(progressTimerRef.current);
        progressTimerRef.current = null;
      }
    };
  }, [state.isPending, timeout, warningThreshold]);

  /**
   * Execute a long-running operation with timeout handling
   */
  const execute = useCallback(
    async (
      operation: (signal?: AbortSignal) => Promise<T>,
      operationTimeout?: number
    ): Promise<T> => {
      // Cancel any existing operation
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }

      abortControllerRef.current = new AbortController();
      startTimeRef.current = Date.now();

      const actualTimeout = operationTimeout ?? timeout;

      setState({
        isPending: true,
        elapsed: 0,
        remaining: actualTimeout,
        progress: 0,
        isNearTimeout: false,
      });

      // Set up timeout warning
      if (warningThreshold > 0) {
        timerRef.current = setTimeout(() => {
          const elapsed = Date.now() - (startTimeRef.current || 0);
          const remaining = actualTimeout - elapsed;

          if (remaining > 0 && remaining <= warningThreshold) {
            setState((prev) => ({ ...prev, isNearTimeout: true }));
          }
        }, actualTimeout - warningThreshold);
      }

      // Set up operation timeout
      const timeoutPromise = new Promise<never>((_, reject) => {
        setTimeout(() => {
          abortControllerRef.current?.abort();
          reject(new Error(`Operation timed out after ${actualTimeout}ms`));
        }, actualTimeout);
      });

      try {
        const result = await Promise.race([
          operation(abortControllerRef.current.signal),
          timeoutPromise,
        ]);

        return result;
      } finally {
        // Cleanup
        if (timerRef.current) {
          clearTimeout(timerRef.current);
          timerRef.current = null;
        }

        setState({
          isPending: false,
          elapsed: Date.now() - (startTimeRef.current || 0),
          remaining: 0,
          progress: 100,
          isNearTimeout: false,
        });

        abortControllerRef.current = null;
        startTimeRef.current = null;
      }
    },
    [timeout, warningThreshold]
  );

  /**
   * Cancel the current operation
   */
  const cancel = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }

    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }

    setState({
      isPending: false,
      elapsed: 0,
      remaining: null,
      progress: null,
      isNearTimeout: false,
    });

    startTimeRef.current = null;
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      cancel();
    };
  }, [cancel]);

  return {
    execute,
    cancel,
    state,
  };
}

