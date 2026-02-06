'use client';

import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Search, RefreshCw } from 'lucide-react';

interface ProviderSearchHeaderProps {
  searchValue?: string;
  onSearchChange?: (value: string) => void;
  placeholder?: string;
  jobControls?: React.ReactNode;
  refreshAction?: () => void;
}

export function ProviderSearchHeader({
  searchValue,
  onSearchChange,
  placeholder = 'Search...',
  jobControls,
  refreshAction,
}: ProviderSearchHeaderProps) {
  return (
    <div className="flex items-center gap-2 mb-4">
      <div className="relative flex-1">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          type="search"
          placeholder={placeholder}
          value={searchValue}
          onChange={(e) => onSearchChange?.(e.target.value)}
          className="pl-9"
        />
      </div>
      {jobControls}
      {refreshAction && (
        <Button variant="outline" size="icon" onClick={refreshAction}>
          <RefreshCw className="h-4 w-4" />
        </Button>
      )}
    </div>
  );
}
