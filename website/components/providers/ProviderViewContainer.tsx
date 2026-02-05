/**
 * Shared Provider View Container
 * Provides consistent Card-based wrapper for all provider views
 */

'use client';

import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';

interface ProviderViewContainerProps {
  title: string;
  description?: string;
  headerActions?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
  contentClassName?: string;
}

export function ProviderViewContainer({
  title,
  description,
  headerActions,
  children,
  className,
  contentClassName,
}: ProviderViewContainerProps) {
  return (
    <Card className={cn('w-full', className)}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex-1">
            <CardTitle>{title}</CardTitle>
            {description && <CardDescription>{description}</CardDescription>}
          </div>
          {headerActions && <div className="flex gap-2">{headerActions}</div>}
        </div>
      </CardHeader>
      <CardContent className={cn('overflow-auto', contentClassName)}>
        {children}
      </CardContent>
    </Card>
  );
}
