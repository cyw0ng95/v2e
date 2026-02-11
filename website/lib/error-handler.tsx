'use client';

import { toast } from 'sonner';

export const errorTypes = {
  NETWORK_ERROR: 'network_error',
  VALIDATION_ERROR: 'validation_error',
  AUTH_ERROR: 'auth_error',
  UNKNOWN_ERROR: 'unknown_error',
};

export const showError = (error: Error | string, context?: string) => {
  const errorMessage = error instanceof Error ? error.message : error;
  console.error(context ? `[${context}] ${errorMessage}` : errorMessage);

  toast.error(errorMessage, {
    description: context,
    action: {
      label: 'Dismiss',
      onClick: () => {},
    },
  });
};

export const showSuccess = (message: string) => {
  toast.success(message);
};

export const showWarning = (message: string) => {
  toast.warning(message);
};

export const showInfo = (message: string) => {
  toast.info(message);
};

// Re-export ErrorBoundary from components for convenience
export { ErrorBoundary as ErrorBoundary } from '@/components/error-boundary';
export { default as ErrorBoundaryDefault } from '@/components/error-boundary';
