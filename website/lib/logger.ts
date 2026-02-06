/**
 * Centralized logging utilities for v2e website
 * Provides structured logging with timestamps and context for Playwright bug detection
 */

type LogLevel = 'info' | 'warn' | 'error' | 'debug';

interface LogContext {
  [key: string]: unknown;
}

/** Get current timestamp in ISO format */
function getTimestamp(): string {
  return new Date().toISOString();
}

/** Format log message with timestamp, level, and component */
function formatMessage(level: string, component: string, message: string, context?: LogContext): string {
  const timestamp = getTimestamp();
  const contextStr = context ? ` | Context: ${JSON.stringify(context)}` : '';
  return `[${timestamp}] [${level.toUpperCase()}] [${component}] ${message}${contextStr}`;
}

/** Extract error stack trace safely */
function getStackTrace(error: unknown): string {
  if (error instanceof Error) {
    return error.stack || 'No stack trace available';
  }
  return String(error);
}

/** Extract error message safely */
function getErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  return String(error);
}

// ============================================================================
// Public Logging API
// ============================================================================

/**
 * Log error with stack trace and context
 * Use this for all errors - it ensures stack traces are captured for debugging
 */
export function logError(component: string, message: string, error: unknown, context?: LogContext): void {
  const errorStr = error instanceof Error
    ? `${error.message}\n${error.stack}`
    : String(error);
  const contextStr = context ? `\nContext: ${JSON.stringify(context, null, 2)}` : '';
  console.error(`[${getTimestamp()}] [ERROR] [${component}] ${message}\n${errorStr}${contextStr}`);
}

/**
 * Log warning with context
 * Use this for recoverable issues that don't prevent functionality
 */
export function logWarn(component: string, message: string, context?: LogContext): void {
  console.warn(formatMessage('WARN', component, message, context));
}

/**
 * Log info with context
 * Use this for important state changes or user actions
 */
export function logInfo(component: string, message: string, context?: LogContext): void {
  console.log(formatMessage('INFO', component, message, context));
}

/**
 * Log debug info (only in development)
 * Use this for detailed debugging information that shouldn't appear in production
 */
export function logDebug(component: string, message: string, context?: LogContext): void {
  if (process.env.NODE_ENV === 'development') {
    console.debug(formatMessage('DEBUG', component, message, context));
  }
}

/**
 * Log RPC-specific errors with full request/response context
 * This is optimized for debugging backend communication issues
 */
export function logRPCError(method: string, target: string, retcode: number, rpcMessage: string, requestContext?: LogContext): void {
  logError('rpc-client', `RPC call failed: ${method} -> ${target}`, new Error(rpcMessage || 'Unknown RPC error'), {
    request: requestContext || {},
    response: { retcode, message: rpcMessage },
  });
}

/**
 * Log empty/invalid data warnings
 * Use this when API returns unexpected data (empty arrays, null values, etc.)
 */
export function logDataWarning(component: string, dataType: string, context?: LogContext): void {
  logWarn(component, `Empty or invalid ${dataType} received`, context);
}

/**
 * Log user actions for debugging
 * Use this to track user interactions that may cause issues
 */
export function logUserAction(component: string, action: string, context?: LogContext): void {
  logInfo(component, `User action: ${action}`, context);
}

/**
 * Create a component-specific logger
 * Use this to avoid repeating component name in every log call
 *
 * @example
 * const logger = createLogger('MyComponent');
 * logger.error('Something went wrong', error);
 * logger.info('User clicked button', { buttonId: 'submit' });
 */
export function createLogger(componentName: string) {
  return {
    error: (message: string, error: unknown, context?: LogContext) => logError(componentName, message, error, context),
    warn: (message: string, context?: LogContext) => logWarn(componentName, message, context),
    info: (message: string, context?: LogContext) => logInfo(componentName, message, context),
    debug: (message: string, context?: LogContext) => logDebug(componentName, message, context),
    rpcError: (method: string, target: string, retcode: number, rpcMessage: string, requestContext?: LogContext) =>
      logRPCError(method, target, retcode, rpcMessage, requestContext),
    dataWarning: (dataType: string, context?: LogContext) => logDataWarning(componentName, dataType, context),
    userAction: (action: string, context?: LogContext) => logUserAction(componentName, action, context),
  };
}
