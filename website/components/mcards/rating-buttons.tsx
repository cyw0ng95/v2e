'use client';

import { RotateCcw, Zap, ThumbsUp, Sparkles } from 'lucide-react';
import { cva, type VariantProps } from 'class-variance-authority';

export type Rating = 'again' | 'hard' | 'good' | 'easy';

interface RatingButtonsProps {
  onRate: (rating: Rating) => void;
  intervals: Record<Rating, string>;
  disabled?: boolean;
}

const ratingButtonVariants = cva(
  'flex flex-col items-center justify-center rounded-xl text-white shadow-md transition-all duration-200 ease-out cursor-pointer focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed disabled:shadow-none',
  {
    variants: {
      color: {
        rose: 'bg-rose-500 hover:bg-rose-600 active:bg-rose-700 shadow-rose-200 dark:shadow-rose-900 hover:shadow-lg hover:shadow-rose-300 dark:hover:shadow-rose-950 active:shadow-sm active:scale-95 focus:ring-rose-500',
        amber: 'bg-amber-500 hover:bg-amber-600 active:bg-amber-700 shadow-amber-200 dark:shadow-amber-900 hover:shadow-lg hover:shadow-amber-300 dark:hover:shadow-amber-950 active:shadow-sm active:scale-95 focus:ring-amber-500',
        emerald: 'bg-emerald-500 hover:bg-emerald-600 active:bg-emerald-700 shadow-emerald-200 dark:shadow-emerald-900 hover:shadow-lg hover:shadow-emerald-300 dark:hover:shadow-emerald-950 active:shadow-sm active:scale-95 focus:ring-emerald-500',
        indigo: 'bg-indigo-500 hover:bg-indigo-600 active:bg-indigo-700 shadow-indigo-200 dark:shadow-indigo-900 hover:shadow-lg hover:shadow-indigo-300 dark:hover:shadow-indigo-950 active:shadow-sm active:scale-95 focus:ring-indigo-500',
      },
      size: {
        sm: 'w-24 h-18',
        md: 'w-28 h-20 sm:w-32 sm:h-24',
        lg: 'w-32 h-24 sm:w-36 sm:h-28',
      },
    },
    defaultVariants: {
      size: 'md',
    },
  }
);

interface RatingButtonProps extends VariantProps<typeof ratingButtonVariants> {
  label: string;
  interval: string;
  icon: React.ReactNode;
  shortcut: string;
  onClick: () => void;
  disabled?: boolean;
  color: 'rose' | 'amber' | 'emerald' | 'indigo';
}

function RatingButton({
  label,
  interval,
  icon,
  shortcut,
  onClick,
  disabled = false,
  color,
  size = 'md',
}: RatingButtonProps) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={ratingButtonVariants({ color, size })}
      aria-label={`Rate as ${label}: ${interval}`}
    >
      <div className="mb-1">{icon}</div>
      <div className="font-medium text-sm">{label}</div>
      <div className="text-xs text-white/80">{interval}</div>
      <div className="mt-1 text-[10px] text-white/60 font-mono">
        <kbd className="px-1 py-0.5 bg-white/20 rounded">{shortcut}</kbd>
      </div>
    </button>
  );
}

export function RatingButtons({ onRate, intervals, disabled = false }: RatingButtonsProps) {
  return (
    <div className="flex items-center justify-center gap-3 flex-wrap">
      <RatingButton
        label="Again"
        interval={intervals.again}
        icon={<RotateCcw className="w-6 h-6" />}
        shortcut="1"
        color="rose"
        onClick={() => onRate('again')}
        disabled={disabled}
      />
      <RatingButton
        label="Hard"
        interval={intervals.hard}
        icon={<Zap className="w-6 h-6" />}
        shortcut="2"
        color="amber"
        onClick={() => onRate('hard')}
        disabled={disabled}
      />
      <RatingButton
        label="Good"
        interval={intervals.good}
        icon={<ThumbsUp className="w-6 h-6" />}
        shortcut="3"
        color="emerald"
        onClick={() => onRate('good')}
        disabled={disabled}
      />
      <RatingButton
        label="Easy"
        interval={intervals.easy}
        icon={<Sparkles className="w-6 h-6" />}
        shortcut="4"
        color="indigo"
        onClick={() => onRate('easy')}
        disabled={disabled}
      />
    </div>
  );
}
