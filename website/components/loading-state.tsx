'use client';

import * as React from 'react';
import { Loader2 } from 'lucide-react';

interface LoadingStateProps {
  message?: string;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export function LoadingState({ message, size = 'md', className }: LoadingStateProps) {
  const sizeClasses = {
    sm: 'h-6 w-6',
    md: 'h-8 w-8',
    lg: 'h-12 w-12',
  };

  return (
    <div className={`flex flex-col items-center justify-center py-12 ${className}`}>
      <Loader2 className={`animate-spin text-primary mb-4 ${sizeClasses[size]}`} />
      {message && <p className="text-sm text-muted-foreground">{message}</p>}
    </div>
  );
}

// Skeleton components for different content types
export function CardSkeleton() {
  return (
    <div className="rounded-xl border border-border bg-card p-6 space-y-4">
      <div className="flex items-center justify-between">
        <div className="h-6 bg-muted rounded w-1/3"></div>
        <div className="h-4 bg-muted rounded w-16"></div>
      </div>
      <div className="space-y-2">
        <div className="h-4 bg-muted rounded w-full"></div>
        <div className="h-4 bg-muted rounded w-5/6"></div>
        <div className="h-4 bg-muted rounded w-4/6"></div>
      </div>
      <div className="flex justify-end pt-4">
        <div className="h-8 bg-muted rounded w-20"></div>
      </div>
    </div>
  );
}

export function TableSkeleton() {
  return (
    <div className="border rounded-lg overflow-hidden">
      <div className="bg-muted border-b px-4 py-3">
        <div className="h-5 bg-muted-foreground/20 rounded w-1/4"></div>
      </div>
      <div className="divide-y">
        {[...Array(5)].map((_, i) => (
          <div key={i} className="px-4 py-3 flex items-center space-x-4">
            <div className="h-4 bg-muted rounded w-1/6"></div>
            <div className="h-4 bg-muted rounded w-1/4"></div>
            <div className="h-4 bg-muted rounded w-2/5"></div>
            <div className="h-4 bg-muted rounded w-1/6"></div>
            <div className="h-4 bg-muted rounded w-16 ml-auto"></div>
          </div>
        ))}
      </div>
    </div>
  );
}

export function ListSkeleton() {
  return (
    <div className="space-y-3">
      {[...Array(6)].map((_, i) => (
        <div key={i} className="flex items-center space-x-3 p-3 rounded-lg border">
          <div className="h-10 w-10 bg-muted rounded"></div>
          <div className="flex-1 space-y-2">
            <div className="h-4 bg-muted rounded w-3/4"></div>
            <div className="h-3 bg-muted rounded w-1/2"></div>
          </div>
          <div className="h-6 bg-muted rounded w-16"></div>
        </div>
      ))}
    </div>
  );
}