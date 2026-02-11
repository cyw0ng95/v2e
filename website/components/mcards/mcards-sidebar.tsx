'use client';

import { Search, X } from 'lucide-react';

interface CardFilters {
  learningState: string;
  majorClass: string;
  bookmarkId: string;
  search?: string;
}

interface McardsSidebarProps {
  filters: CardFilters;
  onFilterChange: (filters: CardFilters) => void;
  learningStates?: { value: string; label: string }[];
  classes?: { value: string; label: string }[];
  bookmarks?: { id: number; title: string }[];
  cardCounts?: Record<string, number>;
}

export function McardsSidebar({
  filters,
  onFilterChange,
  learningStates = [
    { value: '', label: 'All States' },
    { value: 'to_review', label: 'To Review' },
    { value: 'learning', label: 'Learning' },
    { value: 'mastered', label: 'Mastered' },
    { value: 'archived', label: 'Archived' },
  ],
  classes = [
    { value: '', label: 'All Classes' },
    { value: 'CWE', label: 'CWE' },
    { value: 'CVE', label: 'CVE' },
    { value: 'CAPEC', label: 'CAPEC' },
  ],
  bookmarks = [],
  cardCounts = {},
}: McardsSidebarProps) {
  const hasActiveFilters = filters.learningState || filters.majorClass || filters.bookmarkId || filters.search;

  return (
    <aside className="w-72 bg-white dark:bg-slate-800 border-r border-slate-200 dark:border-slate-700 p-4 hidden lg:block">
      <div className="space-y-6">
        {/* Search */}
        <div>
          <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
            Search Cards
          </label>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
            <input
              type="text"
              placeholder="Search front or back..."
              value={filters.search || ''}
              onChange={(e) => onFilterChange({ ...filters, search: e.target.value })}
              className="w-full pl-10 pr-4 py-2 border border-slate-200 dark:border-slate-700 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            />
          </div>
        </div>

        {/* Learning State Filter */}
        <div>
          <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
            Learning State
          </label>
          <select
            value={filters.learningState}
            onChange={(e) => onFilterChange({ ...filters, learningState: e.target.value })}
            className="w-full px-3 py-2 border border-slate-200 dark:border-slate-700 rounded-lg focus:ring-2 focus:ring-indigo-500"
          >
            {learningStates.map((state) => (
              <option key={state.value} value={state.value}>
                {state.label} ({cardCounts[state.value] || 0})
              </option>
            ))}
          </select>
        </div>

        {/* Class Filter */}
        <div>
          <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
            Card Class
          </label>
          <select
            value={filters.majorClass}
            onChange={(e) => onFilterChange({ ...filters, majorClass: e.target.value })}
            className="w-full px-3 py-2 border border-slate-200 dark:border-slate-700 rounded-lg focus:ring-2 focus:ring-indigo-500"
          >
            {classes.map((cls) => (
              <option key={cls.value} value={cls.value}>
                {cls.label} ({cardCounts[cls.value] || 0})
              </option>
            ))}
          </select>
        </div>

        {/* Clear Filters */}
        {hasActiveFilters && (
          <div className="pt-4 border-t border-slate-200 dark:border-slate-700">
            <button
              onClick={() => onFilterChange({ learningState: '', majorClass: '', bookmarkId: '', search: '' })}
              className="w-full px-4 py-2 text-sm font-medium text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-900 transition-colors duration-200 flex items-center justify-center gap-2"
            >
              <X className="w-4 h-4" />
              Clear All Filters
            </button>
          </div>
        )}
      </div>
    </aside>
  );
}
