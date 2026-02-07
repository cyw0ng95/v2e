import React, { useState, useEffect } from 'react';
import { Star } from 'lucide-react';
import { rpcClient } from '@/lib/rpc-client';
import { Bookmark } from '@/lib/types';
import { createLogger } from '@/lib/logger';

const logger = createLogger('bookmark-star');

interface BookmarkStarProps {
  itemId: string;
  itemType: string;
  itemTitle: string;
  itemDescription?: string;
  viewMode: 'view' | 'learn';
  className?: string;
}

const BookmarkStar: React.FC<BookmarkStarProps> = ({ 
  itemId, 
  itemType, 
  itemTitle, 
  itemDescription = '',
  viewMode,
  className = ''
}) => {
  const [isBookmarked, setIsBookmarked] = useState<boolean>(false);
  const [bookmark, setBookmark] = useState<Bookmark | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  // Check if item is already bookmarked on component mount and when itemId/itemType changes
  // Only execute in Learn mode
  useEffect(() => {
    if (viewMode !== 'learn') {
      return;
    }

    const checkBookmarkStatus = async () => {
      setLoading(true);
      try {
        // First, try to find an existing bookmark for this item
        const listResponse = await rpcClient.listBookmarks({
          item_id: itemId,
          item_type: itemType
        });

        if (listResponse.retcode === 0 && listResponse.payload && listResponse.payload.bookmarks && listResponse.payload.bookmarks.length > 0) {
          const foundBookmark = listResponse.payload.bookmarks[0];
          setBookmark(foundBookmark);
          setIsBookmarked(true);
        } else {
          setIsBookmarked(false);
          setBookmark(null);
        }
      } catch (err) {
        setError('Failed to check bookmark status');
        logger.error('Error checking bookmark status', err);
      } finally {
        setLoading(false);
      }
    };

    checkBookmarkStatus();
  }, [itemId, itemType, viewMode]);

  // Only show in Learn mode
  if (viewMode !== 'learn') {
    return null;
  }

  const handleBookmarkToggle = async () => {
    if (loading) return;
    
    setLoading(true);
    setError(null);
    
    try {
      if (isBookmarked && bookmark) {
        // Unbookmark
        const response = await rpcClient.deleteBookmark({ id: bookmark.id });
        if (response.retcode === 0) {
          setIsBookmarked(false);
          setBookmark(null);
        } else {
          setError('Failed to remove bookmark');
        }
      } else {
        // Create bookmark
        const response = await rpcClient.createBookmark({
          global_item_id: `${itemType}-${itemId}`,
          item_type: itemType,
          item_id: itemId,
          title: itemTitle,
          description: itemDescription,
          is_private: false
        });

        if (response.retcode === 0 && response.payload?.bookmark) {
          setBookmark(response.payload.bookmark);
          setIsBookmarked(true);
        } else {
          setError('Failed to create bookmark');
        }
      }
    } catch (err) {
      setError('Failed to toggle bookmark');
      logger.error('Error toggling bookmark', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={`inline-flex items-center ${className}`}>
      <button
        onClick={handleBookmarkToggle}
        disabled={loading}
        aria-label={`${isBookmarked ? 'Remove' : 'Add'} bookmark for ${itemTitle}`}
        aria-pressed={isBookmarked}
        className={`
          p-1 rounded-full transition-colors duration-200
          ${isBookmarked 
            ? 'text-yellow-500 hover:text-yellow-600 bg-yellow-50 hover:bg-yellow-100' 
            : 'text-gray-400 hover:text-gray-600 hover:bg-gray-100'
          }
          ${loading ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
        `}
      >
        <Star 
          size={20} 
          className={isBookmarked ? 'fill-current' : ''}
        />
      </button>
      {error && (
        <span className="ml-2 text-xs text-red-500">{error}</span>
      )}
    </div>
  );
};

export default BookmarkStar;