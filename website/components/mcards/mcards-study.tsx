'use client';

import { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { useMemoryCards } from '@/lib/mcards/hooks';
import { Flashcard } from './flashcard';
import { RatingButtons, type Rating } from './rating-buttons';
import { ChevronLeft, ChevronRight, Play, Pause, ArrowLeft } from 'lucide-react';

interface KeyboardHandlers {
  onSpace?: () => void;
  on1?: () => void;
  on2?: () => void;
  on3?: () => void;
  on4?: () => void;
  onArrowLeft?: () => void;
  onArrowRight?: () => void;
  onEscape?: () => void;
  onP?: () => void;
  onF?: () => void;
}

function useKeyboardShortcuts(handlers: KeyboardHandlers, enabled: boolean = true) {
  useEffect(() => {
    if (!enabled) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      // Ignore if in input field
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLTextAreaElement
      ) {
        return;
      }

      switch (e.key) {
        case ' ':
          e.preventDefault();
          handlers.onSpace?.();
          break;
        case '1':
          handlers.on1?.();
          break;
        case '2':
          handlers.on2?.();
          break;
        case '3':
          handlers.on3?.();
          break;
        case '4':
          handlers.on4?.();
          break;
        case 'ArrowLeft':
          handlers.onArrowLeft?.();
          break;
        case 'ArrowRight':
          handlers.onArrowRight?.();
          break;
        case 'Escape':
          handlers.onEscape?.();
          break;
        case 'p':
        case 'P':
          handlers.onP?.();
          break;
        case 'f':
        case 'F':
          handlers.onF?.();
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handlers, enabled]);
}

function formatTime(seconds: number): string {
  const hrs = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  if (hrs > 0) {
    return `${hrs}:${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  }
  return `${mins}:${secs.toString().padStart(2, '0')}`;
}

function calculateIntervals(card: any): Record<Rating, string> {
  // Calculate next intervals based on SM-2 parameters
  const interval = card.interval || 0;
  const ease = card.ease_factor || 2.5;

  // Simplified interval calculations
  const again = interval < 1 ? '<1m' : interval < 10 ? '<1m' : '<1m';
  const hard = interval < 1 ? '<10m' : interval < 2 ? '<10m' : `${Math.floor(interval * 1.2)}d`;
  const good = interval < 1 ? '<2d' : `${Math.floor(interval * ease)}d`;
  const easy = interval < 1 ? '<4d' : `${Math.floor(interval * ease * 1.3)}d`;

  return { again, hard, good, easy };
}

export default function McardsStudy() {
  // Fetch cards with learning_state filter
  const { data: cardsData, isLoading } = useMemoryCards({});
 
  const cards = useMemo(() => {
    if (!cardsData) return [];
    // Handle RPCResponse structure
    const response = cardsData as unknown as { payload?: { memory_cards?: any[] } };
    return response.payload?.memory_cards || [];
  }, [cardsData]);
 
  // Use ref to store latest cards to avoid closure issues
  const cardsRef = useRef<CardType[]>(cards);
  useEffect(() => {
    cardsRef.current = cards;
  }, [cards]);
 
  // Study session state
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isFlipped, setIsFlipped] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [isComplete, setIsComplete] = useState(false);
  const [elapsed, setElapsed] = useState(0);
  const [startTime] = useState(() => Date.now());
  const [reviewed, setReviewed] = useState<number[]>([]);

  // Timer effect
  useEffect(() => {
    if (!isPaused && !isComplete) {
      const interval = setInterval(() => {
        setElapsed(Math.floor((Date.now() - startTime) / 1000));
      }, 1000);
      return () => clearInterval(interval);
    }
  }, [isPaused, isComplete, startTime]);

  // Reset when cards change
  useEffect(() => {
    setCurrentIndex(0);
    setIsFlipped(false);
    setIsComplete(false);
    setReviewed([]);
    setElapsed(0);
  }, [cards]);

  // Rate card mutation
  const rateCard = useCallback(async (rating: Rating) => {
    // Use ref to access latest cards to avoid closure issues
    const currentCards = cardsRef.current;
    if (currentIndex >= currentCards.length) return;
    
    const card = currentCards[currentIndex];
    await fetch('/restful/rpc', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        method: 'RPCRateMemoryCard',
        target: 'local',
        params: { card_id: card.id, rating },
      }),
    });

    setReviewed([...reviewed, card.id]);

    // Move to next card - use ref to access latest cards
    if (currentIndex + 1 >= currentCards.length) {
      setIsComplete(true);
    } else {
      setCurrentIndex(currentIndex + 1);
      setIsFlipped(false);
    }
  }, [currentIndex, reviewed]);

  const currentCard = useMemo(() => {
    if (currentIndex >= cards.length || currentIndex < 0) return null;
    return cards[currentIndex];
  }, [currentIndex, cards]);

  const progress = useMemo(() => {
    if (cards.length === 0) return 0;
    return ((currentIndex + 1) / cards.length) * 100;
  }, [currentIndex, cards.length]);

  // Keyboard shortcuts
  const keyboardHandlers: KeyboardHandlers = useMemo(
    () => ({
      onSpace: () => !isComplete && setIsFlipped(!isFlipped),
      on1: () => isFlipped && rateCard('again'),
      on2: () => isFlipped && rateCard('hard'),
      on3: () => isFlipped && rateCard('good'),
      on4: () => isFlipped && rateCard('easy'),
      onArrowLeft: () => setCurrentIndex(Math.max(0, currentIndex - 1)),
      onArrowRight: () => setCurrentIndex(Math.min(cards.length - 1, currentIndex + 1)),
      onP: () => setIsPaused(!isPaused),
    }),
    [isFlipped, isComplete, rateCard, currentIndex, cards.length, isPaused]
  );

  useKeyboardShortcuts(keyboardHandlers, !isLoading && cards.length > 0);

  // Empty state
  if (!isLoading && cards.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-64">
        <p className="text-slate-600 dark:text-slate-400 mb-4">No cards available to study</p>
        <p className="text-sm text-slate-500 dark:text-slate-500">
          Create some memory cards to start studying
        </p>
      </div>
    );
  }

  // Loading state
  if (isLoading || !currentCard) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-slate-600 dark:text-slate-400">Loading cards...</div>
      </div>
    );
  }

  // Complete state
  if (isComplete) {
    return (
      <div className="flex flex-col items-center justify-center h-96 space-y-6">
        <div className="text-6xl">🎉</div>
        <h2 className="text-2xl font-semibold text-slate-900 dark:text-slate-100">
          Study Session Complete!
        </h2>
        <p className="text-slate-600 dark:text-slate-400">
          You reviewed {reviewed.length} cards in {formatTime(elapsed)}
        </p>
        <button
          onClick={() => {
            setCurrentIndex(0);
            setIsComplete(false);
            setIsFlipped(false);
            setReviewed([]);
            setElapsed(0);
          }}
          className="px-6 py-3 bg-indigo-500 hover:bg-indigo-600 text-white font-medium rounded-lg transition-colors"
        >
          Start New Session
        </button>
      </div>
    );
  }

  const intervals = calculateIntervals(currentCard);

  return (
    <div className="max-w-4xl mx-auto">
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-center justify-between mb-4">
          <button
            onClick={() => window.history.back()}
            className="inline-flex items-center gap-2 px-3 py-2 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            <span className="text-sm font-medium">Exit</span>
          </button>
          <div className="text-center">
            <p className="text-sm text-slate-600 dark:text-slate-400">
              Card {currentIndex + 1} of {cards.length}
            </p>
          </div>
          <div className="w-16" />
        </div>

        {/* Progress Bar */}
        <div className="h-2 bg-slate-200 dark:bg-slate-700 rounded-full overflow-hidden">
          <div
            className="h-full bg-indigo-500 transition-all duration-300 ease-out"
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>

      {/* Flashcard */}
      <div className="flex justify-center mb-12">
        <Flashcard
          card={currentCard}
          isFlipped={isFlipped}
          onFlip={() => setIsFlipped(!isFlipped)}
          showAnswer={isFlipped}
        />
      </div>

      {/* Rating Buttons */}
      {isFlipped && (
        <div className="mb-8">
          <RatingButtons
            onRate={rateCard}
            intervals={intervals}
            disabled={false}
          />

          {/* SM-2 Stats */}
          <div className="mt-4 text-center text-sm text-slate-600 dark:text-slate-400">
            Interval: {currentCard.interval}d | Ease: {currentCard.ease_factor?.toFixed(2) || '2.50'} | Reps: {currentCard.repetitions || 0}
          </div>
        </div>
      )}

      {/* Keyboard Shortcuts Hint */}
      {!isFlipped && (
        <div className="mb-12 text-center text-sm text-slate-500 dark:text-slate-400">
          <span className="inline-flex items-center gap-1">
            <kbd className="px-2 py-1 bg-slate-100 dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded text-xs font-mono">
              Space
            </kbd>
            <span>Show/Hide</span>
          </span>
          <span className="mx-2">•</span>
          <span className="inline-flex items-center gap-1">
            <kbd className="px-2 py-1 bg-slate-100 dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded text-xs font-mono">
              1-4
            </kbd>
            <span>Rate</span>
          </span>
          <span className="mx-2">•</span>
          <span className="inline-flex items-center gap-1">
            <kbd className="px-2 py-1 bg-slate-100 dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded text-xs font-mono">
              P
            </kbd>
            <span>Pause</span>
          </span>
        </div>
      )}

      {/* Footer Stats */}
      <div className="fixed bottom-0 left-0 right-0 bg-white dark:bg-slate-800 border-t border-slate-200 dark:border-slate-700 px-4 py-3">
        <div className="flex items-center justify-between max-w-4xl mx-auto text-sm">
          <div className="flex items-center gap-4">
            <span className="text-slate-600 dark:text-slate-400">
              Time: {formatTime(elapsed)}
            </span>
            <span className="text-slate-600 dark:text-slate-400">
              Reviewed: {reviewed.length}
            </span>
            {isPaused && (
              <span className="text-amber-600 dark:text-amber-400 font-medium">
                Paused
              </span>
            )}
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setCurrentIndex(Math.max(0, currentIndex - 1))}
              disabled={currentIndex === 0}
              className="p-2 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              aria-label="Previous card"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            <button
              onClick={() => setIsPaused(!isPaused)}
              className="p-2 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
              aria-label={isPaused ? 'Resume' : 'Pause'}
            >
              {isPaused ? <Play className="w-4 h-4" /> : <Pause className="w-4 h-4" />}
            </button>
            <button
              onClick={() => setCurrentIndex(Math.min(cards.length - 1, currentIndex + 1))}
              disabled={currentIndex >= cards.length - 1}
              className="p-2 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              aria-label="Next card"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
