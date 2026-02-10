/**
 * GLC Versioning Module
 *
 * Exports for versioning, auto-save, diff, and crash recovery.
 */

// Auto-saver
export {
  AutoSaver,
  getAutoSaver,
  destroyAutoSaver,
  cleanupAutoSavers,
  type AutoSaverConfig,
  type SaveResult,
  type SaveStatus,
  type AutoSaverState,
} from './auto-saver';

// Diff utilities
export {
  diffGraphs,
  formatDiffSummary,
  getChangeTypeColor,
  getChangeTypeBgColor,
  getChangeTypeIcon,
  type ChangeType,
  type ChangeRecord,
  type NodeChange,
  type EdgeChange,
  type ViewportChange,
  type GraphDiff,
} from './diff';

// Crash recovery
export {
  checkCrashRecovery,
  getAllCrashRecoveries,
  recoverFromCrash,
  clearCrashRecovery,
  clearAllCrashRecoveries,
  saveCrashRecovery,
  pruneOldRecoveries,
  formatRecoveryAge,
  useCrashRecovery,
  type CrashRecoveryData,
  type RecoveryMetadata,
  type CrashRecoveryResult,
} from './crash-recovery';
