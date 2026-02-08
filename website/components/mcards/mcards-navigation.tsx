'use client';

import { LayoutGrid, BarChart3, BookOpen, Plus } from 'lucide-react';
import { cva, type VariantProps } from 'class-variance-authority';

const tabButtonVariants = cva(
  'inline-flex items-center gap-2 px-4 py-2 rounded-md font-medium transition-all duration-200 cursor-pointer',
  {
    variants: {
      active: {
        true: 'bg-white dark:bg-slate-800 text-indigo-600 dark:text-indigo-400 shadow-sm',
        false: 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200',
      },
    },
  }
);

interface TabButtonProps extends VariantProps<typeof tabButtonVariants> {
  children: React.ReactNode;
  onClick: () => void;
  icon: React.ReactNode;
  badge?: number;
  badgeColor?: 'indigo' | 'rose';
}

function TabButton({ active, onClick, icon, children, badge, badgeColor = 'indigo' }: TabButtonProps) {
  return (
    <button
      onClick={onClick}
      className={tabButtonVariants({ active })}
      aria-current={active ? 'page' : undefined}
    >
      {icon}
      <span>{children}</span>
      {badge !== undefined && (
        <span className={`px-2 py-0.5 text-xs font-medium rounded-full ${
          badgeColor === 'rose'
            ? 'bg-rose-100 dark:bg-rose-950 text-rose-700 dark:text-rose-400'
            : 'bg-indigo-100 dark:bg-indigo-950 text-indigo-700 dark:text-indigo-400'
        }`}>
          {badge}
        </span>
      )}
    </button>
  );
}

interface McardsNavigationProps {
  currentView: 'cards' | 'dashboard' | 'study';
  onViewChange: (view: 'cards' | 'dashboard' | 'study') => void;
  cardCount?: number;
  dueCount?: number;
  onCreateCard?: () => void;
}

export function McardsNavigation({ currentView, onViewChange, cardCount, dueCount, onCreateCard }: McardsNavigationProps) {
  return (
    <nav className="bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700">
      <div className="px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo/Title */}
          <div className="flex items-center gap-3">
            <div className="bg-indigo-100 dark:bg-indigo-950 p-2 rounded-lg">
              <BookOpen className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
            </div>
            <h1 className="text-xl font-semibold text-slate-900 dark:text-slate-100">
              Memory Cards
            </h1>
          </div>

          {/* Tabs */}
          <div className="flex items-center gap-1 bg-slate-100 dark:bg-slate-900 p-1 rounded-lg">
            <TabButton
              active={currentView === 'cards'}
              onClick={() => onViewChange('cards')}
              icon={<LayoutGrid className="w-4 h-4" />}
              badge={cardCount}
            >
              Cards
            </TabButton>
            <TabButton
              active={currentView === 'dashboard'}
              onClick={() => onViewChange('dashboard')}
              icon={<BarChart3 className="w-4 h-4" />}
            >
              Dashboard
            </TabButton>
            <TabButton
              active={currentView === 'study'}
              onClick={() => onViewChange('study')}
              icon={<BookOpen className="w-4 h-4" />}
              badge={dueCount}
              badgeColor="rose"
            >
              Study
            </TabButton>
          </div>

          {/* Actions */}
          <button
            onClick={onCreateCard}
            className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-500 hover:bg-indigo-600 text-white font-medium rounded-lg transition-colors duration-200 focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
          >
            <Plus className="w-4 h-4" />
            <span>New Card</span>
          </button>
        </div>
      </div>
    </nav>
  );
}
