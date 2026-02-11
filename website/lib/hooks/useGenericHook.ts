import { useState, useEffect, useCallback } from 'react';
import { createLogger } from '../logger';

const logger = createLogger('hooks');

interface UseQueryOptions<TParams, TData> {
  fetchFn: (params: TParams) => Promise<{ retcode: number; message: string; payload: TData }>;
  params?: TParams;
  enabled?: boolean;
  skip?: boolean;
  errorMessage?: string;
}

interface UseQueryResult<TData> {
  data: TData | null;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
}

export function useQuery<TParams = void, TData = unknown>(
  options: UseQueryOptions<TParams, TData>
): UseQueryResult<TData> {
  const { fetchFn, params, enabled = true, skip = false, errorMessage = 'Request failed' } = options;
  const [data, setData] = useState<TData | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  const execute = useCallback(async () => {
    if (!enabled || skip) {
      setIsLoading(false);
      return;
    }

    try {
      setIsLoading(true);
      const response = await fetchFn(params as TParams);

      if (response.retcode !== 0) {
        throw new Error(response.message || errorMessage);
      }

      setData(response.payload);
      setError(null);
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      setError(error);
    } finally {
      setIsLoading(false);
    }
  }, [fetchFn, params, enabled, skip, errorMessage]);

  useEffect(() => {
    execute();
  }, [execute]);

  return { data, isLoading, error, refetch: execute };
}

interface UseMutationOptions<TParams, TData, TError = Error> {
  mutateFn: (params: TParams) => Promise<{ retcode: number; message: string; payload: TData }>;
  onSuccess?: (data: TData, variables: TParams) => void;
  onError?: (error: TError, variables: TParams) => void;
  errorMessage?: string;
}

interface UseMutationResult<TParams, TData, TError = Error> {
  mutate: (variables: TParams, options?: { onSuccess?: (data: TData) => void; onError?: (error: TError) => void }) => void;
  isPending: boolean;
  error: TError | null;
  reset: () => void;
}

export function useMutation<TParams, TData, TError = Error>(
  options: UseMutationOptions<TParams, TData, TError>
): UseMutationResult<TParams, TData, TError> {
  const { mutateFn, onSuccess: globalOnSuccess, onError: globalOnError, errorMessage = 'Mutation failed' } = options;
  const [isPending, setIsPending] = useState<boolean>(false);
  const [error, setError] = useState<TError | null>(null);

  const mutate = useCallback(
    async (variables: TParams, localOptions?: { onSuccess?: (data: TData) => void; onError?: (error: TError) => void }) => {
      try {
        setIsPending(true);
        setError(null);

        const response = await mutateFn(variables);

        if (response.retcode !== 0) {
          throw new Error(response.message || errorMessage);
        }

        setError(null);
        localOptions?.onSuccess?.(response.payload);
        globalOnSuccess?.(response.payload, variables);
      } catch (err) {
        const mutationError = err instanceof Error ? (err as unknown as TError) : (new Error(String(err)) as unknown as TError);
        setError(mutationError);
        localOptions?.onError?.(mutationError);
        globalOnError?.(mutationError, variables);
      } finally {
        setIsPending(false);
      }
    },
    [mutateFn, errorMessage, globalOnSuccess, globalOnError]
  );

  const reset = useCallback(() => {
    setError(null);
  }, []);

  return { mutate, isPending, error, reset };
}
