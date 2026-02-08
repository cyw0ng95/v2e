'use client';

import { RotateCcw, BookOpen, GraduationCap, Archive } from 'lucide-react';
import { cva, type VariantProps } from 'class-variance-authority';
import { Badge } from '@/components/ui/badge';

const badgeVariants = cva(
  '',
  {
    variants: {
      state: {
        to_review: 'bg-rose-50 dark:bg-rose-950 text-rose-700 dark:text-rose-400 border-rose-200 dark:border-rose-800',
        learning: 'bg-amber-50 dark:bg-amber-950 text-amber-700 dark:text-amber-400 border-amber-200 dark:border-amber-800',
        mastered: 'bg-emerald-50 dark:bg-emerald-950 text-emerald-700 dark:text-emerald-400 border-emerald-200 dark:border-emerald-800',
        archived: 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border-slate-200 dark:border-slate-700',
      },
    },
  }
);

const stateConfig = {
  to_review: { label: 'To Review', icon: RotateCcw },
  learning: { label: 'Learning', icon: BookOpen },
  mastered: { label: 'Mastered', icon: GraduationCap },
  archived: { label: 'Archived', icon: Archive },
};

interface LearningStateBadgeProps extends VariantProps<typeof badgeVariants> {
  state: keyof typeof stateConfig;
}

export function LearningStateBadge({ state }: LearningStateBadgeProps) {
  const config = stateConfig[state || 'to_review'];
  const Icon = config.icon;

  return (
    <Badge variant="outline" className={badgeVariants({ state })}>
      <span className="flex items-center gap-1">
        <Icon className="w-3 h-3" />
        {config.label}
      </span>
    </Badge>
  );
}
