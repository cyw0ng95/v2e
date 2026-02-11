'use client';

import * as React from 'react';
import { ChevronRight, Home } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface BreadcrumbItem {
  label: string;
  href?: string;
  onClick?: () => void;
}

interface BreadcrumbProps {
  items: BreadcrumbItem[];
  className?: string;
}

export function Breadcrumb({ items, className }: BreadcrumbProps) {
  if (items.length === 0) {
    return null;
  }

  return (
    <nav className={`flex items-center ${className}`} aria-label="Breadcrumb">
      <ol className="flex items-center space-x-2">
        <li>
          <Button
            variant="ghost"
            size="sm"
            className="px-2 py-1 h-auto text-muted-foreground hover:text-foreground"
            onClick={items[0].onClick}
            aria-label="Home"
          >
            <Home className="h-4 w-4" />
          </Button>
        </li>
        
        {items.map((item, index) => (
          <React.Fragment key={index}>
            <ChevronRight className="h-4 w-4 text-muted-foreground" />
            <li>
              {index === items.length - 1 ? (
                <span className="text-sm font-medium text-foreground px-2 py-1">
                  {item.label}
                </span>
              ) : (
                <Button
                  variant="ghost"
                  size="sm"
                  className="px-2 py-1 h-auto text-muted-foreground hover:text-foreground"
                  onClick={item.onClick}
                >
                  {item.label}
                </Button>
              )}
            </li>
          </React.Fragment>
        ))}
      </ol>
    </nav>
  );
}