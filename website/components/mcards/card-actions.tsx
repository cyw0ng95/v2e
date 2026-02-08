'use client';

import { Pencil, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface CardActionsProps {
  card: any;
  onEdit: (card: any) => void;
  onDelete: (card: any) => void;
}

export function CardActions({ card, onEdit, onDelete }: CardActionsProps) {
  return (
    <div className="flex items-center gap-1">
      <Button
        variant="ghost"
        size="sm"
        onClick={() => onEdit(card)}
        className="h-8 w-8 p-0 hover:bg-slate-100 dark:hover:bg-slate-800"
        aria-label={`Edit card ${card.id}`}
      >
        <Pencil className="w-4 h-4" />
      </Button>
      <Button
        variant="ghost"
        size="sm"
        onClick={() => onDelete(card)}
        className="h-8 w-8 p-0 hover:bg-rose-50 dark:hover:bg-rose-950 text-rose-600 dark:text-rose-400 hover:text-rose-700 dark:hover:text-rose-300"
        aria-label={`Delete card ${card.id}`}
      >
        <Trash2 className="w-4 h-4" />
      </Button>
    </div>
  );
}
