import React, { useState, useEffect } from 'react';
import { rpcClient } from '@/lib/rpc-client';
import { MemoryCard } from '@/lib/types';

interface MemoryCardStudyProps {
  bookmarkId?: number;
  filterState?: string; // 'to_review', 'learning', 'mastered', 'archived'
}

const MemoryCardStudy: React.FC<MemoryCardStudyProps> = ({ 
  bookmarkId, 
  filterState = 'to_review' 
}) => {
  const [cards, setCards] = useState<MemoryCard[]>([]);
  const [currentCardIndex, setCurrentCardIndex] = useState<number>(0);
  const [showAnswer, setShowAnswer] = useState<boolean>(false);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [stats, setStats] = useState({
    total: 0,
    reviewed: 0,
    remaining: 0
  });

  useEffect(() => {
    loadCards();
  }, [bookmarkId, filterState]);

  const loadCards = async () => {
    setLoading(true);
    try {
      const params: any = { learning_state: filterState };
      if (bookmarkId) {
        params.bookmark_id = bookmarkId;
      }
      
      const response = await rpcClient.listMemoryCards(params);
      
      if (response.retcode === 0 && response.payload) {
        setCards(response.payload.memory_cards);
        setStats({
          total: response.payload.total,
          reviewed: 0,
          remaining: response.payload.total
        });
      } else {
        setError('Failed to load memory cards');
      }
    } catch (err) {
      setError('Error loading memory cards');
      console.error('Error loading memory cards:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleRateCard = async (rating: string) => {
    if (currentCardIndex >= cards.length) return;
    
    const currentCard = cards[currentCardIndex];
    
    try {
      const response = await rpcClient.rateMemoryCard({
        id: currentCard.id,
        rating: rating
      });
      
      if (response.retcode === 0 && response.payload?.memory_card) {
        // Update the card in our list
        const updatedCards = [...cards];
        updatedCards[currentCardIndex] = response.payload.memory_card;
        setCards(updatedCards);
        
        // Move to next card
        if (currentCardIndex < cards.length - 1) {
          setCurrentCardIndex(currentCardIndex + 1);
        } else {
          // If we're at the end, reload the list to get new cards to review
          loadCards();
          setCurrentCardIndex(0);
        }
        
        setShowAnswer(false);
        
        // Update stats
        setStats(prev => ({
          ...prev,
          reviewed: prev.reviewed + 1,
          remaining: Math.max(0, prev.remaining - 1)
        }));
      } else {
        setError('Failed to rate memory card');
      }
    } catch (err) {
      setError('Error rating memory card');
      console.error('Error rating memory card:', err);
    }
  };

  const currentCard = cards[currentCardIndex];

  if (loading) {
    return (
      <div className="flex justify-center items-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
        <span className="ml-2">Loading memory cards...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 bg-red-50 text-red-700 rounded-md">
        {error}
      </div>
    );
  }

  if (!currentCard) {
    return (
      <div className="p-6 text-center bg-white rounded-lg border">
        <h3 className="text-lg font-medium text-gray-900 mb-2">No cards to review</h3>
        <p className="text-gray-500">
          {filterState === 'to_review' 
            ? "No cards are ready for review at the moment." 
            : `No cards in "${filterState}" state.`}
        </p>
        <button 
          onClick={loadCards}
          className="mt-4 px-4 py-2 bg-blue-100 text-blue-700 rounded hover:bg-blue-200"
        >
          Refresh
        </button>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg border p-6 max-w-2xl mx-auto">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-xl font-bold text-gray-800">Memory Card Study</h2>
        <div className="text-sm text-gray-600">
          Card {currentCardIndex + 1} of {cards.length} | 
          Reviewed: {stats.reviewed} | 
          Remaining: {stats.remaining}
        </div>
      </div>

      <div className="mb-6">
        <div className="text-sm text-gray-500 mb-1">Learning State: {currentCard.learning_state}</div>
        <div className="text-lg font-medium bg-gray-50 p-4 rounded mb-2 min-h-[100px] flex items-center">
          {currentCard.front_content}
        </div>
      </div>

      {showAnswer && (
        <div className="mb-6 animate-fade-in">
          <div className="text-sm text-gray-500 mb-1">Answer:</div>
          <div className="bg-yellow-50 p-4 rounded border border-yellow-200 min-h-[100px] flex items-center">
            {currentCard.back_content}
          </div>
        </div>
      )}

      <div className="flex justify-between">
        <button
          onClick={() => setShowAnswer(!showAnswer)}
          className="px-4 py-2 bg-gray-100 text-gray-700 rounded hover:bg-gray-200"
        >
          {showAnswer ? 'Hide Answer' : 'Show Answer'}
        </button>

        {showAnswer && (
          <div className="flex space-x-2">
            <button
              onClick={() => handleRateCard('again')}
              className="px-4 py-2 bg-red-100 text-red-700 rounded hover:bg-red-200"
            >
              Again
            </button>
            <button
              onClick={() => handleRateCard('hard')}
              className="px-4 py-2 bg-orange-100 text-orange-700 rounded hover:bg-orange-200"
            >
              Hard
            </button>
            <button
              onClick={() => handleRateCard('good')}
              className="px-4 py-2 bg-green-100 text-green-700 rounded hover:bg-green-200"
            >
              Good
            </button>
            <button
              onClick={() => handleRateCard('easy')}
              className="px-4 py-2 bg-blue-100 text-blue-700 rounded hover:bg-blue-200"
            >
              Easy
            </button>
          </div>
        )}
      </div>

      <div className="mt-6 text-xs text-gray-500">
        <div>Interval: {currentCard.interval} days | Ease Factor: {currentCard.ease_factor.toFixed(2)}</div>
        <div>Next Review: {new Date(currentCard.next_review_at).toLocaleDateString()}</div>
      </div>
    </div>
  );
};

export default MemoryCardStudy;