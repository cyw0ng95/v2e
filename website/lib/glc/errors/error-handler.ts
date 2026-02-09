import {
  GLCError,
  PresetValidationError,
  GraphValidationError,
  StateError,
  RPCTimeoutError,
  NetworkError,
  FileSystemError,
  SerializationError,
  isGLCError,
  getErrorCode,
  getErrorMessage,
} from './error-types';
import { toast } from 'sonner';

export type ErrorSeverity = 'info' | 'warning' | 'error' | 'critical';

export interface ErrorLogEntry {
  timestamp: string;
  severity: ErrorSeverity;
  code: string;
  message: string;
  context?: Record<string, any>;
  stack?: string;
}

class ErrorHandler {
  private errorLogs: ErrorLogEntry[] = [];
  private maxLogSize = 100;
  private logToConsole = true;
  private logToLocalStorage = true;

  constructor() {
    if (typeof window !== 'undefined') {
      this.loadLogsFromStorage();
    }
  }

  private loadLogsFromStorage(): void {
    try {
      const stored = localStorage.getItem('glc-error-logs');
      if (stored) {
        this.errorLogs = JSON.parse(stored);
      }
    } catch (error) {
      console.warn('Failed to load error logs from storage:', error);
    }
  }

  private saveLogsToStorage(): void {
    try {
      localStorage.setItem('glc-error-logs', JSON.stringify(this.errorLogs));
    } catch (error) {
      console.warn('Failed to save error logs to storage:', error);
    }
  }

  private addLog(entry: ErrorLogEntry): void {
    this.errorLogs.push(entry);
    
    if (this.errorLogs.length > this.maxLogSize) {
      this.errorLogs = this.errorLogs.slice(-this.maxLogSize);
    }

    if (this.logToLocalStorage) {
      this.saveLogsToStorage();
    }
  }

  public handleError(error: unknown, context?: Record<string, any>): void {
    const severity = this.getSeverity(error);
    const code = getErrorCode(error);
    const message = getErrorMessage(error);

    const logEntry: ErrorLogEntry = {
      timestamp: new Date().toISOString(),
      severity,
      code,
      message,
      context,
      stack: error instanceof Error ? error.stack : undefined,
    };

    this.addLog(logEntry);

    if (this.logToConsole) {
      this.logToConsole(severity, logEntry);
    }

    this.showUserNotification(severity, message);
  }

  private getSeverity(error: unknown): ErrorSeverity {
    if (error instanceof PresetValidationError || error instanceof GraphValidationError) {
      return 'warning';
    }
    if (error instanceof RPCTimeoutError || error instanceof NetworkError) {
      return 'critical';
    }
    if (error instanceof StateError || error instanceof SerializationError) {
      return 'error';
    }
    return 'error';
  }

  private logToConsole(severity: ErrorSeverity, entry: ErrorLogEntry): void {
    const consoleMethod = severity === 'critical' || severity === 'error' ? console.error :
                         severity === 'warning' ? console.warn :
                         console.info;
    
    consoleMethod(`[GLC ${severity.toUpperCase()}]`, {
      code: entry.code,
      message: entry.message,
      context: entry.context,
      timestamp: entry.timestamp,
    });
  }

  private showUserNotification(severity: ErrorSeverity, message: string): void {
    const duration = severity === 'critical' ? 10000 :
                     severity === 'error' ? 7000 :
                     severity === 'warning' ? 5000 :
                     3000;

    const title = severity === 'critical' ? 'Critical Error' :
                  severity === 'error' ? 'Error' :
                  severity === 'warning' ? 'Warning' :
                  'Information';

    toast[severity === 'critical' || severity === 'error' ? 'error' : 
         severity === 'warning' ? 'warning' : 'info'](title, {
      description: message,
      duration,
    });
  }

  public logError(message: string, context?: Record<string, any>): void {
    const error = new Error(message);
    this.handleError(error, context);
  }

  public showError(message: string, context?: Record<string, any>): void {
    toast.error(message, { duration: 7000 });
    this.logError(message, context);
  }

  public showWarning(message: string, context?: Record<string, any>): void {
    toast.warning(message, { duration: 5000 });
    this.handleError(new Error(message), { ...context, severity: 'warning' });
  }

  public showInfo(message: string, context?: Record<string, any>): void {
    toast.info(message, { duration: 3000 });
    this.handleError(new Error(message), { ...context, severity: 'info' });
  }

  public getLogs(): ErrorLogEntry[] {
    return [...this.errorLogs];
  }

  public getLogsBySeverity(severity: ErrorSeverity): ErrorLogEntry[] {
    return this.errorLogs.filter(log => log.severity === severity);
  }

  public getLogsByCode(code: string): ErrorLogEntry[] {
    return this.errorLogs.filter(log => log.code === code);
  }

  public clearLogs(): void {
    this.errorLogs = [];
    this.saveLogsToStorage();
  }

  public exportLogs(): string {
    return JSON.stringify(this.errorLogs, null, 2);
  }

  public importLogs(json: string): void {
    try {
      const logs = JSON.parse(json) as ErrorLogEntry[];
      if (Array.isArray(logs)) {
        this.errorLogs = logs;
        this.saveLogsToStorage();
      }
    } catch (error) {
      this.handleError(error, { action: 'import-logs' });
    }
  }
}

export const errorHandler = new ErrorHandler();

export const handleGLCError = (error: unknown, context?: Record<string, any>): void => {
  errorHandler.handleError(error, context);
};

export const logError = (message: string, context?: Record<string, any>): void => {
  errorHandler.logError(message, context);
};

export const showError = (message: string, context?: Record<string, any>): void => {
  errorHandler.showError(message, context);
};

export const showWarning = (message: string, context?: Record<string, any>): void => {
  errorHandler.showWarning(message, context);
};

export const showInfo = (message: string, context?: Record<string, any>): void => {
  errorHandler.showInfo(message, context);
};

export const getErrorLogs = (): ErrorLogEntry[] => {
  return errorHandler.getLogs();
};

export const clearErrorLogs = (): void => {
  errorHandler.clearLogs();
};

export const exportErrorLogs = (): string => {
  return errorHandler.exportLogs();
};

export default errorHandler;
