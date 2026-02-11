'use client';

/**
 * Global Error Boundary for RPC Failures
 *
 * This component catches React errors and provides user-friendly error recovery.
 * It's designed to handle:
 * - Network errors (fetch failures, timeouts)
 * - RPC errors (backend failures, malformed responses)
 * - Component rendering errors
 *
 * Usage: Wrap the application root with this component in app/layout.tsx
 */

import * as React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { createLogger } from '@/lib/logger';

const logger = createLogger('global-error-boundary');

interface GlobalErrorBoundaryProps {
  children: React.ReactNode;
}

interface GlobalErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorType: 'network' | 'rpc' | 'render' | 'unknown';
}

/**
 * Categorize error type for better user messaging
 */
function categorizeError(error: Error): GlobalErrorBoundaryState['errorType'] {
  const message = error.message.toLowerCase();

  // Network errors (fetch, timeout, CORS)
  if (
    message.includes('fetch') ||
    message.includes('network') ||
    message.includes('timeout') ||
    message.includes('aborterror') ||
    message.includes('failed to fetch') ||
    message.includes('connection')
  ) {
    return 'network';
  }

  // RPC errors (backend issues)
  if (
    message.includes('rpc') ||
    message.includes('backend') ||
    message.includes('server') ||
    message.includes('500') ||
    message.includes('502') ||
    message.includes('503')
  ) {
    return 'rpc';
  }

  return 'unknown';
}

/**
 * Get user-friendly error message based on error type
 */
function getErrorMessage(errorType: GlobalErrorBoundaryState['errorType'], error: Error | null): string {
  switch (errorType) {
    case 'network':
      return 'Unable to connect to the server. Please check your internet connection and try again.';
    case 'rpc':
      return 'The server encountered an error processing your request. Please try again.';
    case 'render':
      return 'A display error occurred. This has been logged for review.';
    default:
      return error?.message || 'An unexpected error occurred. Please try again.';
  }
}

class GlobalErrorBoundaryClass extends React.Component<
  GlobalErrorBoundaryProps,
  GlobalErrorBoundaryState
> {
  constructor(props: GlobalErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorType: 'unknown',
    };
  }

  static getDerivedStateFromError(error: Error): Partial<GlobalErrorBoundaryState> {
    const errorType = categorizeError(error);
    return {
      hasError: true,
      error,
      errorType,
    };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    const timestamp = new Date().toISOString();
    const errorType = this.state.errorType;

    logger.error(
      `Global Error Boundary caught ${errorType} error`,
      error,
      {
        errorType,
        componentStack: errorInfo.componentStack,
        timestamp,
      }
    );

    // Log to console for debugging
    console.error(
      `[${timestamp}] [ERROR] [GlobalErrorBoundary] ${errorType} error caught\n` +
        `Error: ${error.message}\n` +
        `Stack: ${error.stack}\n` +
        `Component Stack: ${errorInfo.componentStack}`
    );
  }

  handleReset = () => {
    logger.info('User requested error reset');
    this.setState({
      hasError: false,
      error: null,
      errorType: 'unknown',
    });
  };

  handleReload = () => {
    logger.info('User requested page reload');
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      const errorMessage = getErrorMessage(this.state.errorType, this.state.error);
      const showDetails = process.env.NODE_ENV === 'development';

      return (
        <div className="min-h-screen min-w-screen flex items-center justify-center bg-background p-4">
          <Card className="max-w-lg w-full shadow-lg">
            <CardHeader>
              <CardTitle className="text-destructive flex items-center gap-2">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-6 w-6"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                  />
                </svg>
                Something went wrong
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <Alert variant="destructive">
                <AlertDescription>{errorMessage}</AlertDescription>
              </Alert>

              {showDetails && this.state.error && (
                <details className="text-sm">
                  <summary className="cursor-pointer text-muted-foreground hover:text-foreground">
                    Error details
                  </summary>
                  <pre className="mt-2 p-2 bg-muted rounded text-xs overflow-auto max-h-32">
                    {this.state.error.stack || this.state.error.message}
                  </pre>
                </details>
              )}

              <div className="flex flex-col sm:flex-row gap-2">
                <Button onClick={this.handleReset} variant="default" className="flex-1">
                  Try Again
                </Button>
                <Button
                  onClick={this.handleReload}
                  variant="outline"
                  className="flex-1"
                >
                  Reload Page
                </Button>
              </div>

              {this.state.errorType === 'network' && (
                <p className="text-xs text-muted-foreground text-center">
                  If the problem persists, the server may be temporarily unavailable.
                  Please contact support if this continues.
                </p>
              )}
            </CardContent>
          </Card>
        </div>
      );
    }

    return this.props.children;
  }
}

export default GlobalErrorBoundaryClass;

export { GlobalErrorBoundaryClass as GlobalErrorBoundary };
