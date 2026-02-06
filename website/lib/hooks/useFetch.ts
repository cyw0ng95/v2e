import { useState, useEffect } from 'react';
import { createLogger } from '../logger';

const logger = createLogger('hooks');

interface UseFetchOptions<TParams, TResponse> {
  fetchFn: (params: TParams) => Promise<{ retcode: number; message: string; payload: TResponse }>;
  params?: TParams;
  enabled?: boolean;
  onSuccess?: (data: TResponse) => void;
  onError?: (error: Error) => void;
}

interface UseFetchResult<TData> {
  data: TData | null;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
}

export function useFetch<TParams, TData>(
  options: UseFetchOptions<TParams, TData>
): UseFetchResult<TData> {
  const { fetchFn, params, enabled = true, onSuccess, onError } = options;
  const [data, setData] = useState<TData | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  const execute = async () => {
    if (!enabled) {
      setIsLoading(false);
      return;
    }

    try {
      setIsLoading(true);
      const response = await fetchFn(params);

      if (response.retcode !== 0) {
        throw new Error(response.message || 'Request failed');
      }

      setData(response.payload);
      setError(null);
      onSuccess?.(response.payload);
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      setError(error);
      onError?.(error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    execute();
  }, [enabled]);

  const refetch = async () => {
    await execute();
  };

  return { data, isLoading, error, refetch };
}
