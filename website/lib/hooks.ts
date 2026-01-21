/**
 * React Query hooks for v2e API
 */

import { useQuery, useMutation, useQueryClient, UseQueryResult } from '@tanstack/react-query';
import { rpcClient } from './rpc-client';
import { SysMetrics } from './types';

// ============================================================================
// Query Keys
// ============================================================================

export const queryKeys = {
  cves: {
    all: ['cves'] as const,
    list: (offset: number, limit: number) => ['cves', 'list', offset, limit] as const,
    count: () => ['cves', 'count'] as const,
    detail: (id: string) => ['cves', 'detail', id] as const,
  },
  cwes: {
    all: ['cwes'] as const,
    list: (params?: { offset?: number; limit?: number; search?: string }) => [
      'cwes',
      'list',
      params?.offset ?? 0,
      params?.limit ?? 100,
      params?.search ?? ''
    ] as const,
    detail: (id: string) => ['cwes', 'detail', id] as const,
  },
  session: {
    status: () => ['session', 'status'] as const,
  },
  health: () => ['health'] as const,
};

// ============================================================================
// CVE Queries
// ============================================================================

export function useCVE(cveId: string) {
  return useQuery({
    queryKey: queryKeys.cves.detail(cveId),
    queryFn: async () => {
      const response = await rpcClient.getCVE(cveId);
      if (response.retcode !== 0) {
        throw new Error(response.message);
      }
      return response.payload;
    },
    enabled: !!cveId,
  });
}

export function useCVEList(offset: number = 0, limit: number = 10) {
  return useQuery({
    queryKey: queryKeys.cves.list(offset, limit),
    queryFn: async () => {
      const response = await rpcClient.listCVEs(offset, limit);
      if (response.retcode !== 0) {
        throw new Error(response.message);
      }
      return response.payload;
    },
  });
}

export function useCVECount() {
  return useQuery({
    queryKey: queryKeys.cves.count(),
    queryFn: async () => {
      const response = await rpcClient.countCVEs();
      if (response.retcode !== 0) {
        throw new Error(response.message);
      }
      return response.payload;
    },
  });
}

// ============================================================================
// CVE Mutations
// ============================================================================

export function useCreateCVE() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (cveId: string) => {
      const response = await rpcClient.createCVE(cveId);
      if (response.retcode !== 0) {
        throw new Error(response.message);
      }
      return response.payload;
    },
    onSuccess: () => {
      // Invalidate CVE list to refetch
      queryClient.invalidateQueries({ queryKey: queryKeys.cves.all });
    },
  });
}

export function useUpdateCVE() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (cveId: string) => {
      const response = await rpcClient.updateCVE(cveId);
      if (response.retcode !== 0) {
        throw new Error(response.message);
      }
      return response.payload;
    },
    onSuccess: (_, cveId) => {
      // Invalidate specific CVE and list
      queryClient.invalidateQueries({ queryKey: queryKeys.cves.detail(cveId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.cves.all });
    },
  });
}

export function useDeleteCVE() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (cveId: string) => {
      const response = await rpcClient.deleteCVE(cveId);
      if (response.retcode !== 0) {
        throw new Error(response.message);
      }
      return response.payload;
    },
    onSuccess: () => {
      // Invalidate CVE list to refetch
      queryClient.invalidateQueries({ queryKey: queryKeys.cves.all });
    },
  });
}

// ============================================================================
// CWE Queries
// ============================================================================

import type { CWEItem, ListCWEsRequest, ListCWEsResponse } from './types';

export function useCWEList(params?: ListCWEsRequest) {
  return useQuery<ListCWEsResponse>({
    queryKey: queryKeys.cwes.list(params),
    queryFn: async () => {
      const response = await rpcClient.listCWEs(params);
      if (response.retcode !== 0) {
        throw new Error(response.message);
      }
      return response.payload || { cwes: [], offset: 0, limit: 0, total: 0 };
    },
  });
}

export function useCWE(cweId: string) {
  return useQuery({
    queryKey: queryKeys.cwes.detail(cweId),
    queryFn: async () => {
      const response = await rpcClient.getCWE(cweId);
      if (response.retcode !== 0) {
        throw new Error(response.message);
      }
      return response.payload?.cwe;
    },
    enabled: !!cweId,
  });
}

// ============================================================================
// Session Queries
// ============================================================================

export function useSessionStatus() {
  return useQuery({
    queryKey: queryKeys.session.status(),
    queryFn: async () => {
      const response = await rpcClient.getSessionStatus();
      if (response.retcode !== 0) {
        throw new Error(response.message);
      }
      return response.payload;
    },
    refetchInterval: 5000, // Refetch every 5 seconds to track progress
  });
}

// ============================================================================
// Session Mutations
// ============================================================================

export function useStartSession() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (params: {
      sessionId: string;
      startIndex?: number;
      resultsPerBatch?: number;
    }) => {
      const response = await rpcClient.startSession(
        params.sessionId,
        params.startIndex,
        params.resultsPerBatch
      );
      if (response.retcode !== 0) {
        throw new Error(response.message);
      }
      return response.payload;
    },
    onSuccess: () => {
      // Invalidate session status to refetch
      queryClient.invalidateQueries({ queryKey: queryKeys.session.status() });
    },
  });
}

export function useStopSession() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async () => {
      const response = await rpcClient.stopSession();
      if (response.retcode !== 0) {
        throw new Error(response.message);
      }
      return response.payload;
    },
    onSuccess: () => {
      // Invalidate session status and CVE list
      queryClient.invalidateQueries({ queryKey: queryKeys.session.status() });
      queryClient.invalidateQueries({ queryKey: queryKeys.cves.all });
    },
  });
}

export function usePauseJob() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async () => {
      const response = await rpcClient.pauseJob();
      if (response.retcode !== 0) {
        throw new Error(response.message);
      }
      return response.payload;
    },
    onSuccess: () => {
      // Invalidate session status
      queryClient.invalidateQueries({ queryKey: queryKeys.session.status() });
    },
  });
}

export function useResumeJob() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async () => {
      const response = await rpcClient.resumeJob();
      if (response.retcode !== 0) {
        throw new Error(response.message);
      }
      return response.payload;
    },
    onSuccess: () => {
      // Invalidate session status
      queryClient.invalidateQueries({ queryKey: queryKeys.session.status() });
    },
  });
}

// ============================================================================
// Health Check
// ============================================================================

export function useHealth() {
  return useQuery({
    queryKey: queryKeys.health(),
    queryFn: async () => {
      const response = await rpcClient.health();
      if (response.retcode !== 0) {
        throw new Error(response.message);
      }
      return response.payload;
    },
    refetchInterval: 30000, // Refetch every 30 seconds
  });
}

// ============================================================================
// System Metrics
// ============================================================================

export function useSysMetrics(): UseQueryResult<SysMetrics, Error> {
  return useQuery<SysMetrics, Error>({
    queryKey: ['sysMetrics'],
    queryFn: async () => {
      const response = await rpcClient.getSysMetrics();
      if (response.retcode !== 0) {
        throw new Error(response.message);
      }
      // Ensure payload is typed
      return response.payload as SysMetrics;
    },
  });
}
