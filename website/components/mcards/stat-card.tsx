'use client';

import { LucideIcon } from 'lucide-react';
import { cva, type VariantProps } from 'class-variance-authority';
import { TrendingUp } from 'lucide-react';

const statCardVariants = cva(
  'bg-white dark:bg-slate-800 rounded-xl p-6 shadow-sm border',
  {
    variants: {
      color: {
        indigo: 'border-slate-200 dark:border-slate-700',
        rose: 'border-slate-200 dark:border-slate-700',
        amber: 'border-slate-200 dark:border-slate-700',
        emerald: 'border-slate-200 dark:border-slate-700',
      },
    },
  }
);

const iconBgVariants = cva('p-3 rounded-lg', {
  variants: {
    color: {
      indigo: 'bg-indigo-50 dark:bg-indigo-950 text-indigo-600 dark:text-indigo-400',
      rose: 'bg-rose-50 dark:bg-rose-950 text-rose-600 dark:text-rose-400',
      amber: 'bg-amber-50 dark:bg-amber-950 text-amber-600 dark:text-amber-400',
      emerald: 'bg-emerald-50 dark:bg-emerald-950 text-emerald-600 dark:text-emerald-400',
    },
  },
});

interface StatCardProps extends VariantProps<typeof statCardVariants>, VariantProps<typeof iconBgVariants> {
  title: string;
  value: number | string;
  icon: LucideIcon;
  trend?: number;
  trendLabel?: string;
}

export function StatCard({ title, value, icon: Icon, color = 'indigo', trend, trendLabel = 'from last week' }: StatCardProps) {
  return (
    <div className={statCardVariants({ color })}>
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-slate-600 dark:text-slate-400">{title}</p>
          <p className="text-3xl font-bold text-slate-900 dark:text-slate-100 mt-2">{value}</p>
        </div>
        <div className={iconBgVariants({ color })}>
          <Icon className="w-6 h-6" />
        </div>
      </div>
      {trend !== undefined && (
        <div className="mt-4 flex items-center text-sm">
          <span className={trend >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'}>
            <TrendingUp className={`w-4 h-4 mr-1 inline ${trend < 0 ? 'rotate-180' : ''}`} />
            {trend > 0 ? '+' : ''}{trend}%
          </span>
          <span className="text-slate-500 dark:text-slate-400 ml-2">{trendLabel}</span>
        </div>
      )}
    </div>
  );
}
