'use client';

import { Search, Grid, List, X, Calendar } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { cn } from '@/lib/utils';

interface DateRange {
  from: Date | undefined;
  to: Date | undefined;
}

interface GraphFiltersProps {
  viewMode: 'grid' | 'list';
  onViewModeChange: (mode: 'grid' | 'list') => void;
  search: string;
  onSearchChange: (value: string) => void;
  presetFilter: string;
  onPresetFilterChange: (value: string) => void;
  dateRange: DateRange | undefined;
  onDateRangeChange: (range: DateRange | undefined) => void;
  pageSize: number;
  onPageSizeChange: (size: number) => void;
  presets: { id: string; name: string }[];
  hasFilters: boolean;
  onClearFilters: () => void;
}

function formatDateForInput(date: Date | undefined): string {
  if (!date) return '';
  return date.toISOString().split('T')[0];
}

function parseDateFromInput(value: string): Date | undefined {
  if (!value) return undefined;
  const date = new Date(value + 'T00:00:00');
  return isNaN(date.getTime()) ? undefined : date;
}

export function GraphFilters({
  viewMode,
  onViewModeChange,
  search,
  onSearchChange,
  presetFilter,
  onPresetFilterChange,
  dateRange,
  onDateRangeChange,
  pageSize,
  onPageSizeChange,
  presets,
  hasFilters,
  onClearFilters,
}: GraphFiltersProps) {
  const dateRangeText = dateRange?.from
    ? dateRange.to
      ? `${dateRange.from.toLocaleDateString()} - ${dateRange.to.toLocaleDateString()}`
      : dateRange.from.toLocaleDateString()
    : 'Date range';

  const handleFromChange = (value: string) => {
    const from = parseDateFromInput(value);
    onDateRangeChange({
      from,
      to: dateRange?.to,
    });
  };

  const handleToChange = (value: string) => {
    const to = parseDateFromInput(value);
    onDateRangeChange({
      from: dateRange?.from,
      to,
    });
  };

  return (
    <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
      {/* Left: Search and Filters */}
      <div className="flex flex-wrap gap-2 items-center flex-1">
        {/* Search */}
        <div className="relative flex-1 min-w-[200px] max-w-[300px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-textMuted" />
          <Input
            placeholder="Search graphs..."
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            className="pl-9 pr-9"
          />
          {search && (
            <Button
              variant="ghost"
              size="icon"
              className="absolute right-1 top-1/2 -translate-y-1/2 h-6 w-6"
              onClick={() => onSearchChange('')}
            >
              <X className="w-3 h-3" />
            </Button>
          )}
        </div>

        {/* Preset Filter */}
        <Select value={presetFilter} onValueChange={onPresetFilterChange}>
          <SelectTrigger className="w-[150px]">
            <SelectValue placeholder="All presets" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All presets</SelectItem>
            {presets.map((preset) => (
              <SelectItem key={preset.id} value={preset.id}>
                {preset.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {/* Date Range Filter */}
        <Popover>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              className={cn(
                'w-[200px] justify-start text-left font-normal',
                !dateRange?.from && 'text-textMuted'
              )}
            >
              <Calendar className="mr-2 h-4 w-4" />
              {dateRangeText}
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-72" align="start">
            <div className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">From</label>
                <Input
                  type="date"
                  value={formatDateForInput(dateRange?.from)}
                  onChange={(e) => handleFromChange(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">To</label>
                <Input
                  type="date"
                  value={formatDateForInput(dateRange?.to)}
                  onChange={(e) => handleToChange(e.target.value)}
                />
              </div>
              {dateRange && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="w-full"
                  onClick={() => onDateRangeChange(undefined)}
                >
                  Clear dates
                </Button>
              )}
            </div>
          </PopoverContent>
        </Popover>

        {/* Clear Filters */}
        {hasFilters && (
          <Button variant="ghost" size="sm" onClick={onClearFilters}>
            <X className="w-4 h-4 mr-1" />
            Clear filters
          </Button>
        )}
      </div>

      {/* Right: View Toggle and Page Size */}
      <div className="flex items-center gap-2">
        {/* Page Size */}
        <Select value={String(pageSize)} onValueChange={(v) => onPageSizeChange(Number(v))}>
          <SelectTrigger className="w-[100px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="12">12 / page</SelectItem>
            <SelectItem value="24">24 / page</SelectItem>
            <SelectItem value="48">48 / page</SelectItem>
          </SelectContent>
        </Select>

        {/* View Toggle */}
        <div className="flex items-center border rounded-md">
          <Button
            variant={viewMode === 'grid' ? 'secondary' : 'ghost'}
            size="icon"
            className="h-9 w-9 rounded-r-none"
            onClick={() => onViewModeChange('grid')}
          >
            <Grid className="h-4 w-4" />
          </Button>
          <Button
            variant={viewMode === 'list' ? 'secondary' : 'ghost'}
            size="icon"
            className="h-9 w-9 rounded-l-none border-l"
            onClick={() => onViewModeChange('list')}
          >
            <List className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}

export type { DateRange };
