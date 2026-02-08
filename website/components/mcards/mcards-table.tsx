'use client';

import { useState } from 'react';
import { Loader, ChevronLeft, ChevronRight } from 'lucide-react';
import {
  Table,
  TableHeader,
  TableBody,
  TableHead,
  TableRow,
  TableCell,
} from '@/components/ui/table';
import { Checkbox } from '@/components/ui/checkbox';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { useMemoryCards } from '@/lib/mcards/hooks';
import { LearningStateBadge } from './learning-state-badge';
import { CardActions } from './card-actions';

interface McardsTableProps {
  filters: {
    learningState?: string;
    majorClass?: string;
    bookmarkId?: string;
    search?: string;
  };
}

export function McardsTable({ filters }: McardsTableProps) {
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
  const [page, setPage] = useState(0);
  const limit = 20;

  const { data, isLoading, error } = useMemoryCards({
    learning_state: filters.learningState || undefined,
    bookmark_id: filters.bookmarkId ? Number(filters.bookmarkId) : undefined,
    limit,
    offset: page * limit,
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader className="w-6 h-6 animate-spin text-indigo-500" />
        <span className="ml-2 text-slate-600 dark:text-slate-400">Loading cards...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64 text-rose-600">
        Error loading cards: {error.message}
      </div>
    );
  }

  const cards = data?.payload?.memory_cards || [];
  const total = data?.payload?.total || 0;
  const totalPages = Math.ceil(total / limit);

  if (cards.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-center">
        <div className="text-6xl mb-4">📇</div>
        <h3 className="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-2">
          No memory cards yet
        </h3>
        <p className="text-slate-600 dark:text-slate-400 mb-4">
          Create your first card to start learning
        </p>
        <Button className="bg-indigo-500 hover:bg-indigo-600">
          Create Card
        </Button>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
      <Table>
        <TableHeader>
          <TableRow className="bg-slate-50 dark:bg-slate-900/50">
            <TableHead className="w-12">
              <Checkbox
                checked={selectedIds.size === cards.length && cards.length > 0}
                onCheckedChange={(checked) => {
                  if (checked) {
                    setSelectedIds(new Set(cards.map((c: any) => c.id)));
                  } else {
                    setSelectedIds(new Set());
                  }
                }}
              />
            </TableHead>
            <TableHead>Front Content</TableHead>
            <TableHead>Class</TableHead>
            <TableHead>State</TableHead>
            <TableHead>Next Review</TableHead>
            <TableHead className="w-32">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {cards.map((card: any) => (
            <TableRow
              key={card.id}
              className="cursor-pointer transition-colors duration-200 hover:bg-slate-50 dark:hover:bg-slate-900/50"
              data-testid={`card-row-${card.id}`}
            >
              <TableCell>
                <Checkbox
                  checked={selectedIds.has(card.id)}
                  onCheckedChange={(checked) => {
                    const newSelected = new Set(selectedIds);
                    if (checked) {
                      newSelected.add(card.id);
                    } else {
                      newSelected.delete(card.id);
                    }
                    setSelectedIds(newSelected);
                  }}
                />
              </TableCell>
              <TableCell>
                <div className="max-w-xs truncate">
                  {card.front_content || card.front || 'No content'}
                </div>
              </TableCell>
              <TableCell>
                <Badge variant="outline">{card.major_class}</Badge>
              </TableCell>
              <TableCell>
                <LearningStateBadge state={card.learning_state} />
              </TableCell>
              <TableCell>
                {card.next_review_at
                  ? new Date(card.next_review_at).toLocaleDateString()
                  : 'Now'}
              </TableCell>
              <TableCell>
                <CardActions
                  card={card}
                  onEdit={(c) => console.log('Edit', c)}
                  onDelete={(c) => console.log('Delete', c)}
                />
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>

      {/* Pagination */}
      <div className="flex items-center justify-between px-6 py-4 border-t border-slate-200 dark:border-slate-700">
        <p className="text-sm text-slate-600 dark:text-slate-400">
          Showing {cards.length} of {total} cards
        </p>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage(Math.max(0, page - 1))}
            disabled={page === 0}
          >
            <ChevronLeft className="w-4 h-4" />
          </Button>
          <span className="text-sm text-slate-600 dark:text-slate-400">
            Page {page + 1} of {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage(Math.min(totalPages - 1, page + 1))}
            disabled={page >= totalPages - 1}
          >
            <ChevronRight className="w-4 h-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
