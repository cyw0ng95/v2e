'use client';

import { useMemo } from 'react';

export interface MemoryCardContent {
  id: number;
  front: string;
  back: string;
  front_content?: string;
  back_content?: string;
  major_class: string;
}

interface FlashcardProps {
  card: MemoryCardContent;
  isFlipped: boolean;
  onFlip: () => void;
  showAnswer?: boolean;
}

export function Flashcard({ card, isFlipped, onFlip, showAnswer = false }: FlashcardProps) {
  const frontContent = useMemo(() => card.front_content || card.front, [card]);
  const backContent = useMemo(() => card.back_content || card.back, [card]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onFlip();
    }
  };

  return (
    <div className="relative w-full max-w-2xl h-80 perspective-1000 group">
      <div
        className={`relative w-full h-full transition-transform duration-400 ease-[cubic-bezier(0.4,0,0.2,1)] transform-style-3d cursor-pointer ${
          isFlipped ? 'rotate-y-180' : ''
        }`}
        onClick={onFlip}
        onKeyDown={handleKeyDown}
        tabIndex={0}
        role="button"
        aria-label={isFlipped ? 'Show question' : 'Show answer'}
      >
        {/* Front Side */}
        <div className="absolute inset-0 backface-hidden">
          <div className="w-full h-full bg-white dark:bg-slate-800 rounded-2xl shadow-lg border-2 border-slate-200 dark:border-slate-700 p-8 flex flex-col">
            <div className="flex items-center justify-between mb-4">
              <span className="px-2 py-1 text-xs font-medium bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 rounded-md border border-slate-200 dark:border-slate-600">
                {card.major_class}
              </span>
              <span className="text-sm text-slate-400 dark:text-slate-500">Front</span>
            </div>
            <div className="flex-1 flex items-center justify-center overflow-auto">
              <div className="prose prose-slate dark:prose-invert max-w-none">
                {frontContent}
              </div>
            </div>
          </div>
        </div>

        {/* Back Side */}
        <div className="absolute inset-0 backface-hidden rotate-y-180">
          <div className="w-full h-full bg-white dark:bg-slate-800 rounded-2xl shadow-lg border-2 border-indigo-200 dark:border-indigo-800 p-8 flex flex-col">
            <div className="flex items-center justify-between mb-4">
              <span className="px-2 py-1 text-xs font-medium bg-indigo-50 dark:bg-indigo-950 text-indigo-700 dark:text-indigo-400 rounded-md border border-indigo-200 dark:border-indigo-800">
                {card.major_class}
              </span>
              <span className="text-sm text-indigo-500 dark:text-indigo-400">Back</span>
            </div>
            <div className="flex-1 flex items-center justify-center overflow-auto">
              <div className="prose prose-slate dark:prose-invert max-w-none">
                {backContent}
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Click hint */}
      {!showAnswer && (
        <div className="absolute -bottom-8 left-1/2 -translate-x-1/2 text-sm text-slate-400 dark:text-slate-500 opacity-0 group-hover:opacity-100 transition-opacity">
          Click or press Space to flip
        </div>
      )}
    </div>
  );
}
