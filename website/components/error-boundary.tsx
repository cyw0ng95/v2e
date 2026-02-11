'use client';

import * as React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

interface ErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
  fallbackComponent?: React.ComponentType<{ error: Error; resetError: () => void }>;
  onReset?: () => void;
  showReload?: boolean; // If true, shows "Reload Page" instead of "Try Again"
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

class ErrorBoundaryClass extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    const timestamp = new Date().toISOString();
    console.error(`[${timestamp}] [ERROR] [ErrorBoundary] React component error caught\n` +
      `Error: ${error.message}\n` +
      `Stack: ${error.stack}\n` +
      `Component Stack: ${errorInfo.componentStack}`);
  }

  resetError = () => {
    this.setState({ hasError: false, error: null });
    this.props.onReset?.();
  };

  render() {
    if (this.state.hasError) {
      // Use custom component fallback if provided
      if (this.props.fallbackComponent && this.state.error) {
        const FallbackComponent = this.props.fallbackComponent;
        return <FallbackComponent error={this.state.error} resetError={this.resetError} />;
      }

      // Use simple fallback if provided
      if (this.props.fallback) {
        return this.props.fallback;
      }

      // Default error UI
      const errorMessage = this.state.error?.message || 'An unexpected error occurred';

      if (this.props.showReload) {
        // Reload Page variant
        return (
          <Card className="m-4">
            <CardHeader>
              <CardTitle className="text-destructive">Something went wrong</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground mb-4">{errorMessage}</p>
              <Button
                onClick={() => window.location.reload()}
                variant="outline"
              >
                Reload Page
              </Button>
            </CardContent>
          </Card>
        );
      }

      // Try Again variant with centered layout
      return (
        <div className="min-h-screen flex items-center justify-center bg-background">
          <div className="max-w-md w-full space-y-4 p-6 rounded-lg border border-border bg-card">
            <div className="text-center">
              <h2 className="text-lg font-semibold text-foreground mb-2">Something went wrong</h2>
              <p className="text-muted-foreground mb-4">{errorMessage}</p>
              <Button onClick={this.resetError} variant="default">
                Try Again
              </Button>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

// Export as default for compatibility
export default ErrorBoundaryClass;

// Also export as named export
export { ErrorBoundaryClass as ErrorBoundary };
