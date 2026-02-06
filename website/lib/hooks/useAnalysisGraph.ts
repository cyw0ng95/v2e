/**
 * Custom hooks for Graph Analysis
 */

import { useState, useEffect, useCallback } from 'react';
import { rpcClient } from '../rpc-client';
import type {
  GraphStats,
  GraphPath,
  GetGraphStatsResponse,
  GetNeighborsResponse,
  FindPathResponse,
  GetNodesByTypeResponse,
  BuildCVEGraphResponse,
  ClearGraphResponse,
  GetFSMStateResponse,
  SaveGraphResponse,
  LoadGraphResponse,
} from '../types';
import { logError, logDebug, createLogger } from '../logger';

const logger = createLogger('useAnalysisGraph');

// ============================================================================
// useGraphStats Hook
// ============================================================================

interface UseGraphStatsResult {
  data: GraphStats | null;
  isLoading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export function useGraphStats(): UseGraphStatsResult {
  const [data, setData] = useState<GraphStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchStats = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await rpcClient.getGraphStats();
      if (response.retcode === 0 && response.payload) {
         const stats = {
          node_count: response.payload.node_count,
          edge_count: response.payload.edge_count,
        };
        setData(stats);
        logger.debug('Graph stats fetched successfully', { stats });
      } else {
        setError(response.message || 'Failed to fetch graph stats');
        logger.error('Failed to fetch graph stats', response);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Unknown error';
      setError(message);
      logger.error('Error fetching graph stats', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStats();
  }, []);

  return { data, isLoading, error, refetch: fetchStats };
}

// ============================================================================
// useNeighbors Hook
// ============================================================================

interface UseNeighborsResult {
  neighbors: string[] | null;
  isLoading: boolean;
  error: string | null;
  fetchNeighbors: (urn: string) => Promise<void>;
}

export function useNeighbors(urn: string): UseNeighborsResult {
  const [neighbors, setNeighbors] = useState<string[] | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchNeighbors = useCallback(async (targetUrn: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await rpcClient.getNeighbors(targetUrn);
      if (response.retcode === 0 && response.payload) {
        setNeighbors(response.payload.neighbors);
        logger.debug('Neighbors fetched successfully', { data: response.payload.neighbors });
      } else {
        setError(response.message || 'Failed to fetch neighbors');
        logger.error('Failed to fetch neighbors', response);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Unknown error';
      setError(message);
      logger.error('Error fetching neighbors', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (urn) {
      fetchNeighbors(urn);
    }
  }, [urn, fetchNeighbors]);

  return { neighbors, isLoading, error, fetchNeighbors };
}

// ============================================================================
// useFindPath Hook
// ============================================================================

interface UseFindPathResult {
  path: GraphPath | null;
  isLoading: boolean;
  error: string | null;
  findPath: (from: string, to: string) => Promise<void>;
}

export function useFindPath(): UseFindPathResult {
  const [path, setPath] = useState<GraphPath | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const findPath = useCallback(async (from: string, to: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await rpcClient.findPath(from, to);
      if (response.retcode === 0 && response.payload) {
        setPath({
          path: response.payload.path,
          length: response.payload.length,
        });
         logger.debug('Path found successfully', { data: response.payload });
      } else {
        setError(response.message || 'Failed to find path');
        logger.error('Failed to find path', response);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Unknown error';
      setError(message);
      logger.error('Error finding path', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  return { path, isLoading, error, findPath };
}

// ============================================================================
// useNodesByType Hook
// ============================================================================

interface UseNodesByTypeResult {
  nodes: Array<{ urn: string; properties: Record<string, unknown> }> | null;
  count: number;
  isLoading: boolean;
  error: string | null;
  fetchNodes: (type: string) => Promise<void>;
}

export function useNodesByType(type?: string): UseNodesByTypeResult {
  const [nodes, setNodes] = useState<Array<{ urn: string; properties: Record<string, unknown> }> | null>(null);
  const [count, setCount] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchNodes = useCallback(async (nodeType: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await rpcClient.getNodesByType(nodeType);
      if (response.retcode === 0 && response.payload) {
        setNodes(response.payload.nodes);
        setCount(response.payload.count);
         logger.debug('Nodes fetched successfully', { data: response.payload });
      } else {
        setError(response.message || 'Failed to fetch nodes');
        logger.error('Failed to fetch nodes', response);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Unknown error';
      setError(message);
      logger.error('Error fetching nodes', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (type) {
      fetchNodes(type);
    }
  }, [type, fetchNodes]);

  return { nodes, count, isLoading, error, fetchNodes };
}

// ============================================================================
// useGraphControl Hook
// ============================================================================

interface UseGraphControlResult {
  buildGraph: (limit?: number) => Promise<void>;
  clearGraph: () => Promise<void>;
  saveGraph: () => Promise<void>;
  loadGraph: () => Promise<void>;
  isLoading: boolean;
  error: string | null;
  buildResult: BuildCVEGraphResponse | null;
  clearResult: ClearGraphResponse | null;
  saveResult: SaveGraphResponse | null;
  loadResult: LoadGraphResponse | null;
}

export function useGraphControl(): UseGraphControlResult {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [buildResult, setBuildResult] = useState<BuildCVEGraphResponse | null>(null);
  const [clearResult, setClearResult] = useState<ClearGraphResponse | null>(null);
  const [saveResult, setSaveResult] = useState<SaveGraphResponse | null>(null);
  const [loadResult, setLoadResult] = useState<LoadGraphResponse | null>(null);

  const buildGraph = useCallback(async (limit?: number) => {
    setIsLoading(true);
    setError(null);
    setBuildResult(null);
    try {
      const response = await rpcClient.buildCVEGraph(limit);
      if (response.retcode === 0 && response.payload) {
        setBuildResult(response.payload);
        logger.debug('Graph built successfully', { data: response.payload });
      } else {
        setError(response.message || 'Failed to build graph');
        logger.error('Failed to build graph', response);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Unknown error';
      setError(message);
      logger.error('Error building graph', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const clearGraph = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    setClearResult(null);
    try {
      const response = await rpcClient.clearGraph();
      if (response.retcode === 0 && response.payload) {
        setClearResult(response.payload);
        logger.debug('Graph cleared successfully', { data: response.payload });
      } else {
        setError(response.message || 'Failed to clear graph');
        logger.error('Failed to clear graph', response);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Unknown error';
      setError(message);
      logger.error('Error clearing graph', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const saveGraph = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    setSaveResult(null);
    try {
      const response = await rpcClient.saveGraph();
      if (response.retcode === 0 && response.payload) {
        setSaveResult(response.payload);
        logger.debug('Graph saved successfully', { data: response.payload });
      } else {
        setError(response.message || 'Failed to save graph');
        logger.error('Failed to save graph', response);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Unknown error';
      setError(message);
      logger.error('Error saving graph', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const loadGraph = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    setLoadResult(null);
    try {
      const response = await rpcClient.loadGraph();
      if (response.retcode === 0 && response.payload) {
        setLoadResult(response.payload);
        logger.debug('Graph loaded successfully', { data: response.payload });
      } else {
        setError(response.message || 'Failed to load graph');
        logger.error('Failed to load graph', response);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Unknown error';
      setError(message);
      logger.error('Error loading graph', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  return {
    buildGraph,
    clearGraph,
    saveGraph,
    loadGraph,
    isLoading,
    error,
    buildResult,
    clearResult,
    saveResult,
    loadResult,
  };
}
