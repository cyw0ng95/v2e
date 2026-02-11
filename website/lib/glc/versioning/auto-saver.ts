/**
 * GLC Auto-Saver
 *
 * Enhanced auto-save with debouncing, idle detection, and beforeunload handling.
 */

import type { Graph } from '../types';
import { rpcClient } from '../../rpc-client';
import type { UpdateGLCGraphRequest } from '../../types';
import { createLogger } from '../../logger';

const logger = createLogger('glc-auto-saver');

export interface AutoSaverConfig {
  /** Debounce delay in milliseconds (default: 500ms) */
  debounceDelay: number;
  /** Idle timeout in milliseconds (default: 30000ms = 30 seconds) */
  idleTimeout: number;
  /** Maximum number of versions to keep (default: 50) */
  maxVersions: number;
  /** Enable idle detection */
  enableIdleDetection: boolean;
  /** Enable beforeunload save */
  enableBeforeUnload: boolean;
}

const DEFAULT_CONFIG: AutoSaverConfig = {
  debounceDelay: 500,
  idleTimeout: 30000,
  maxVersions: 50,
  enableIdleDetection: true,
  enableBeforeUnload: true,
};

export interface SaveResult {
  success: boolean;
  version?: number;
  timestamp: string;
  error?: string;
}

export type SaveStatus = 'idle' | 'pending' | 'saving' | 'saved' | 'error';

export interface AutoSaverState {
  status: SaveStatus;
  lastSavedAt: string | null;
  lastVersion: number | null;
  error: string | null;
  pendingChanges: boolean;
}

type StateListener = (state: AutoSaverState) => void;

/**
 * Auto-saver class for GLC graphs
 */
export class AutoSaver {
  private config: AutoSaverConfig;
  private debounceTimer: ReturnType<typeof setTimeout> | null = null;
  private idleTimer: ReturnType<typeof setTimeout> | null = null;
  private lastActivity: number = Date.now();
  private state: AutoSaverState = {
    status: 'idle',
    lastSavedAt: null,
    lastVersion: null,
    error: null,
    pendingChanges: false,
  };
  private listeners: Set<StateListener> = new Set();
  private pendingGraph: Graph | null = null;
  private isDestroyed = false;

  constructor(config: Partial<AutoSaverConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
    this.setupEventListeners();
  }

  /**
   * Set up event listeners for idle detection and beforeunload
   */
  private setupEventListeners(): void {
    if (this.config.enableIdleDetection && typeof window !== 'undefined') {
      // Track user activity
      const activityEvents = ['mousedown', 'keydown', 'touchstart', 'scroll'];
      const handleActivity = () => this.onUserActivity();

      activityEvents.forEach((event) => {
        window.addEventListener(event, handleActivity, { passive: true });
      });

      // Start idle timer
      this.resetIdleTimer();
    }

    if (this.config.enableBeforeUnload && typeof window !== 'undefined') {
      window.addEventListener('beforeunload', (e) => this.onBeforeUnload(e));
      // Also save on visibility change (mobile/tab switch)
      document.addEventListener('visibilitychange', () => {
        if (document.visibilityState === 'hidden') {
          this.saveImmediately();
        }
      });
    }
  }

  /**
   * Handle user activity for idle detection
   */
  private onUserActivity(): void {
    this.lastActivity = Date.now();
    this.resetIdleTimer();
  }

  /**
   * Reset the idle timer
   */
  private resetIdleTimer(): void {
    if (this.idleTimer) {
      clearTimeout(this.idleTimer);
    }

    this.idleTimer = setTimeout(() => {
      if (this.pendingGraph && this.state.pendingChanges) {
        logger.debug('User idle, triggering auto-save');
        this.saveImmediately();
      }
    }, this.config.idleTimeout);
  }

  /**
   * Handle beforeunload event
   */
  private onBeforeUnload(e: BeforeUnloadEvent): void {
    if (this.state.pendingChanges && this.pendingGraph) {
      // Save immediately (synchronous attempt)
      this.saveToLocalStorage(this.pendingGraph);

      // Show confirmation dialog
      e.preventDefault();
      e.returnValue = 'You have unsaved changes. Are you sure you want to leave?';
      return e.returnValue;
    }
  }

  /**
   * Schedule a debounced save
   */
  scheduleSave(graph: Graph): void {
    this.pendingGraph = graph;
    this.updateState({ pendingChanges: true });

    if (this.debounceTimer) {
      clearTimeout(this.debounceTimer);
    }

    this.debounceTimer = setTimeout(() => {
      this.saveImmediately();
    }, this.config.debounceDelay);
  }

  /**
   * Save immediately without debouncing
   */
  async saveImmediately(): Promise<SaveResult> {
    if (!this.pendingGraph) {
      return {
        success: false,
        timestamp: new Date().toISOString(),
        error: 'No graph to save',
      };
    }

    // Clear debounce timer
    if (this.debounceTimer) {
      clearTimeout(this.debounceTimer);
      this.debounceTimer = null;
    }

    return this.save(this.pendingGraph);
  }

  /**
   * Save graph to backend
   */
  async save(graph: Graph): Promise<SaveResult> {
    if (this.isDestroyed) {
      return {
        success: false,
        timestamp: new Date().toISOString(),
        error: 'AutoSaver has been destroyed',
      };
    }

    this.updateState({ status: 'saving', error: null });
    const timestamp = new Date().toISOString();

    try {
      // First save to localStorage for crash recovery
      this.saveToLocalStorage(graph);

      // Then save to backend
      const params: UpdateGLCGraphRequest = {
        graph_id: graph.metadata.id,
        name: graph.metadata.name,
        description: graph.metadata.description,
        nodes: JSON.stringify(graph.nodes),
        edges: JSON.stringify(graph.edges),
        viewport: graph.viewport ? JSON.stringify(graph.viewport) : undefined,
        tags: graph.metadata.tags?.join(','),
      };

      const response = await rpcClient.updateGLCGraph(params);

      if (response.retcode !== 0 || !response.payload?.success) {
        throw new Error(response.message || 'Failed to save graph');
      }

      const newVersion = response.payload.graph?.version ?? graph.metadata.version;
      this.pendingGraph = null;

      this.updateState({
        status: 'saved',
        lastSavedAt: timestamp,
        lastVersion: newVersion,
        pendingChanges: false,
      });

      logger.debug('Graph saved successfully', {
        graphId: graph.metadata.id,
        version: newVersion,
      });

      return {
        success: true,
        version: newVersion,
        timestamp,
      };
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      this.updateState({
        status: 'error',
        error: errorMessage,
      });

      logger.error('Failed to save graph', error, {
        graphId: graph.metadata.id,
      });

      return {
        success: false,
        timestamp,
        error: errorMessage,
      };
    }
  }

  /**
   * Save graph to localStorage for crash recovery
   */
  private saveToLocalStorage(graph: Graph): void {
    if (typeof window === 'undefined') return;

    const key = `glc-crash-recovery-${graph.metadata.id}`;
    const data = {
      graph,
      savedAt: Date.now(),
      version: graph.metadata.version,
    };

    try {
      localStorage.setItem(key, JSON.stringify(data));
    } catch (error) {
      logger.warn('Failed to save to localStorage', { error });
    }
  }

  /**
   * Update internal state and notify listeners
   */
  private updateState(updates: Partial<AutoSaverState>): void {
    this.state = { ...this.state, ...updates };
    this.notifyListeners();
  }

  /**
   * Notify all state listeners
   */
  private notifyListeners(): void {
    this.listeners.forEach((listener) => {
      try {
        listener(this.state);
      } catch (error) {
        logger.error('Listener error', error);
      }
    });
  }

  /**
   * Subscribe to state changes
   */
  subscribe(listener: StateListener): () => void {
    this.listeners.add(listener);
    // Immediately notify with current state
    listener(this.state);

    return () => {
      this.listeners.delete(listener);
    };
  }

  /**
   * Get current state
   */
  getState(): AutoSaverState {
    return { ...this.state };
  }

  /**
   * Force a save (useful for manual save button)
   */
  async forceSave(graph: Graph): Promise<SaveResult> {
    this.pendingGraph = graph;
    return this.saveImmediately();
  }

  /**
   * Mark changes as saved (useful after successful backend sync)
   */
  markSaved(version: number): void {
    this.updateState({
      status: 'saved',
      lastSavedAt: new Date().toISOString(),
      lastVersion: version,
      pendingChanges: false,
    });
    this.pendingGraph = null;
  }

  /**
   * Check if there are pending changes
   */
  hasPendingChanges(): boolean {
    return this.state.pendingChanges;
  }

  /**
   * Destroy the auto-saver and cleanup
   */
  destroy(): void {
    this.isDestroyed = true;

    if (this.debounceTimer) {
      clearTimeout(this.debounceTimer);
      this.debounceTimer = null;
    }

    if (this.idleTimer) {
      clearTimeout(this.idleTimer);
      this.idleTimer = null;
    }

    this.listeners.clear();
    this.pendingGraph = null;
  }
}

/**
 * Create a singleton auto-saver instance per graph
 */
const autoSavers = new Map<string, AutoSaver>();

export function getAutoSaver(
  graphId: string,
  config?: Partial<AutoSaverConfig>
): AutoSaver {
  let saver = autoSavers.get(graphId);

  if (!saver) {
    saver = new AutoSaver(config);
    autoSavers.set(graphId, saver);
  }

  return saver;
}

export function destroyAutoSaver(graphId: string): void {
  const saver = autoSavers.get(graphId);
  if (saver) {
    saver.destroy();
    autoSavers.delete(graphId);
  }
}

/**
 * Cleanup all auto-savers
 */
export function cleanupAutoSavers(): void {
  autoSavers.forEach((saver) => saver.destroy());
  autoSavers.clear();
}
