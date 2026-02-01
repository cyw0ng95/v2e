import { useState, useEffect, useMemo } from 'react';
import { rpcClient } from './rpc-client';

interface AttackQueryParams {
  offset?: number;
  limit?: number;
  search?: string;
}

export function useAttackTechniques(params: AttackQueryParams = {}) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  const { offset = 0, limit = 100, search = '' } = params;

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listAttackTechniques(offset, limit);
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch ATT&CK techniques');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        console.error('Error fetching ATT&CK techniques:', err);
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

export function useAttackTactics(params: AttackQueryParams = {}) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  const { offset = 0, limit = 100, search = '' } = params;

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listAttackTactics(offset, limit);
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch ATT&CK tactics');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        console.error('Error fetching ATT&CK tactics:', err);
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

export function useAttackMitigations(params: AttackQueryParams = {}) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  const { offset = 0, limit = 100, search = '' } = params;

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listAttackMitigations(offset, limit);
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch ATT&CK mitigations');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        console.error('Error fetching ATT&CK mitigations:', err);
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
        console.error('Error fetching CAPEC:', err);
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

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listCAPECs(offset, limit);
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch CAPEC list');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        console.error('Error fetching CAPEC list:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    // Cleanup function
    return () => {};
  }, [offset, limit]);

  return { data, isLoading, error };
}

// Session Management Hooks
export function useSessionStatus() {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getSessionStatus();
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch session status');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        console.error('Error fetching session status:', err);
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

export function useStartSession() {
  const [isPending, setIsPending] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);

  const mutate = async (params: { sessionId: string; startIndex?: number; resultsPerBatch?: number }, options?: { onSuccess?: (data: any) => void; onError?: (error: Error) => void }) => {
    try {
      setIsPending(true);
      setError(null);
      
      const { sessionId, startIndex, resultsPerBatch } = params;
      const response = await rpcClient.startSession(sessionId, startIndex, resultsPerBatch);
      
      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to start session');
      }
      
      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      console.error('Error starting session:', err);
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
    try {
      setIsPending(true);
      setError(null);
      
      const { sessionId, dataType, startIndex, resultsPerBatch, params: extraParams } = params;
      const response = await rpcClient.startTypedSession(sessionId, dataType, startIndex, resultsPerBatch, extraParams);
      
      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to start typed session');
      }
      
      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      console.error('Error starting typed session:', err);
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
      
      const response = await rpcClient.startCWEImport(params);
      
      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to start CWE import');
      }
      
      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      console.error('Error starting CWE import:', err);
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
      
      const response = await rpcClient.startCAPECImport(params);
      
      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to start CAPEC import');
      }
      
      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      console.error('Error starting CAPEC import:', err);
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
      
      const response = await rpcClient.startATTACKImport(params);
      
      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to start ATT&CK import');
      }
      
      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      console.error('Error starting ATT&CK import:', err);
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
        console.warn('Skipping stop session request - already saw "run not active" error');
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
      console.error('Error stopping session:', err);
      
      // Prevent infinite retry loop for "run not active" errors
      if (err.message && err.message.includes('run not active')) {
        console.warn('Session already inactive, preventing future retries');
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
      console.error('Error pausing job:', err);
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
      console.error('Error resuming job:', err);
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

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listCWEViews(offset, limit);
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch CWE views');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        console.error('Error fetching CWE views:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    // Cleanup function
    return () => {};
  }, [offset, limit]);

  const refetch = async () => {
    try {
      setIsLoading(true);
      const response = await rpcClient.listCWEViews(offset, limit);
      
      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to fetch CWE views');
      }
      
      setData(response.payload);
      setError(null);
    } catch (err: any) {
      setError(err);
      console.error('Error fetching CWE views:', err);
    } finally {
      setIsLoading(false);
    }
  };

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
        console.error('Error fetching CWE job status:', err);
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
    try {
      setIsPending(true);
      setError(null);
      
      const { sessionId, startIndex, resultsPerBatch } = params;
      const response = await rpcClient.startCWEViewJob(sessionId, startIndex, resultsPerBatch);
      
      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to start CWE view job');
      }
      
      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      console.error('Error starting CWE view job:', err);
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
    try {
      setIsPending(true);
      setError(null);
      
      const { sessionId } = params;
      const response = await rpcClient.stopCWEViewJob(sessionId);
      
      if (response.retcode !== 0) {
        throw new Error(response.message || 'Failed to stop CWE view job');
      }
      
      if (options?.onSuccess) {
        options.onSuccess(response.payload);
      }
    } catch (err: any) {
      setError(err);
      console.error('Error stopping CWE view job:', err);
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

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.getSysMetrics();
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch system metrics');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        console.error('Error fetching system metrics:', err);
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

// CVE Hooks
export function useCVEList(offset: number = 0, limit: number = 100) {
  const [data, setData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listCVEs(offset, limit);
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch CVE list');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        console.error('Error fetching CVE list:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    // Cleanup function
    return () => {};
  }, [offset, limit]);

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
        console.error('Error fetching CVE count:', err);
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
        console.error('Error fetching ATT&CK software:', err);
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
        console.error('Failed to fetch ATT&CK groups:', err);
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

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const response = await rpcClient.listCWEs({ offset, limit, search });
        
        if (response.retcode !== 0) {
          throw new Error(response.message || 'Failed to fetch CWE list');
        }
        
        setData(response.payload);
        setError(null);
      } catch (err: any) {
        setError(err);
        console.error('Error fetching CWE list:', err);
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
        console.error('Error fetching attack technique:', err);
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
        console.error('Error fetching attack tactic:', err);
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
        console.error('Error fetching attack mitigation:', err);
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
        console.error('Error fetching attack software:', err);
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
        console.error('Error fetching attack group:', err);
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