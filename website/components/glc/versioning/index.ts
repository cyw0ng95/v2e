/**
 * GLC Versioning Components
 *
 * Exports for version history, restore dialog, crash recovery, and save status.
 */

// Version history panel
export { VersionHistoryPanel, default as VersionHistoryPanelDefault } from './history-panel';

// Restore dialog
export { RestoreDialog, default as RestoreDialogDefault } from './restore-dialog';

// Crash recovery dialog
export {
  CrashRecoveryDialog,
  useCrashRecoveryDialog,
  default as CrashRecoveryDialogDefault,
} from './crash-recovery-dialog';

// Save status indicator
export {
  SaveStatusIndicator,
  SaveStatusDetailed,
  useAutoSaveStatus,
  default as SaveStatusIndicatorDefault,
} from './save-status-indicator';
