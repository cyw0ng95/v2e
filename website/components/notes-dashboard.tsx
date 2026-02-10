import React, { useState, useEffect } from 'react';
import { rpcClient } from '@/lib/rpc-client';
import { createLogger } from '@/lib/logger';

const logger = createLogger('notes-dashboard');

interface RecentItem {
  type: 'bookmark';
  title: string;
  date: string;
  item_type: string;
  item_id: string;
}

const NotesDashboard: React.FC = () => {
  const [stats, setStats] = useState({
    totalBookmarks: 0,
    totalNotes: 0,
    totalMemoryCards: 0,
    toReviewCards: 0,
    learningCards: 0,
    masteredCards: 0,
    todayReviews: 0,
  });
  const [recentItems, setRecentItems] = useState<RecentItem[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadDashboardData();
  }, []);

  const loadDashboardData = async () => {
    setLoading(true);
    try {
      // Load all bookmarks to calculate stats
      const bookmarksResponse = await rpcClient.listBookmarks({ limit: 100 });
      let totalBookmarks = 0;
      let totalNotes = 0;
      let totalMemoryCards = 0;
      let toReviewCards = 0;
      let learningCards = 0;
      let masteredCards = 0;
      let todayReviews = 0;

      if (bookmarksResponse.retcode === 0 && bookmarksResponse.payload) {
        totalBookmarks = bookmarksResponse.payload.total;
        
        // For each bookmark, get notes and memory cards
        for (const bookmark of bookmarksResponse.payload.bookmarks) {
          // Get notes for this bookmark
          const notesResponse = await rpcClient.getNotesByBookmark({
            bookmark_id: bookmark.id
          });
          if (notesResponse.retcode === 0 && notesResponse.payload) {
            totalNotes += notesResponse.payload.total;
          }

          // Get memory cards for this bookmark
          const cardsResponse = await rpcClient.listMemoryCards({
            bookmark_id: bookmark.id
          });
          if (
            cardsResponse.retcode === 0 &&
            cardsResponse.payload &&
            Array.isArray(cardsResponse.payload.memory_cards)
          ) {
            totalMemoryCards += cardsResponse.payload.total;
            // Count by learning state
            cardsResponse.payload.memory_cards.forEach(card => {
              switch (card.learning_state) {
                case 'to_review':
                  toReviewCards++;
                  // Check if it's due for review today
                  const today = new Date();
                  const nextReview = new Date(card.next_review_at);
                  if (nextReview <= today) {
                    todayReviews++;
                  }
                  break;
                case 'learning':
                  learningCards++;
                  break;
                case 'mastered':
                  masteredCards++;
                  break;
              }
            });
          }
        }
      }

      // Get recent items for activity feed
      const recentBookmarksResponse = await rpcClient.listBookmarks({ 
        limit: 5,
        offset: 0
      });
      const recentItemsData = [];
      if (recentBookmarksResponse.retcode === 0 && recentBookmarksResponse.payload) {
        for (const bookmark of recentBookmarksResponse.payload.bookmarks.slice(0, 5)) {
          recentItemsData.push({
            type: 'bookmark',
            title: bookmark.title,
            date: new Date(bookmark.created_at).toLocaleDateString(),
            item_type: bookmark.item_type,
            item_id: bookmark.item_id
          });
        }
      }

      // Defensive: ensure all values are numbers and not NaN
      function safeNumber(val: any) {
        return typeof val === 'number' && !isNaN(val) ? val : 0;
      }
      setStats({
        totalBookmarks: safeNumber(totalBookmarks),
        totalNotes: safeNumber(totalNotes),
        totalMemoryCards: safeNumber(totalMemoryCards),
        toReviewCards: safeNumber(toReviewCards),
        learningCards: safeNumber(learningCards),
        masteredCards: safeNumber(masteredCards),
        todayReviews: safeNumber(todayReviews),
      });
      setRecentItems(recentItemsData);
    } catch (err) {
      setError('Failed to load dashboard data');
      logger.error('Error loading dashboard data', err);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
        <span className="ml-2">Loading dashboard...</span>
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

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {/* Total Bookmarks Card */}
        <div className="bg-white rounded-lg shadow p-6 border">
          <div className="flex items-center">
            <div className="rounded-full bg-blue-100 p-3">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
              </svg>
            </div>
            <div className="ml-4">
              <h3 className="text-sm font-medium text-gray-500">Total Bookmarks</h3>
              <p className="text-2xl font-bold">{stats.totalBookmarks}</p>
            </div>
          </div>
        </div>

        {/* Total Notes Card */}
        <div className="bg-white rounded-lg shadow p-6 border">
          <div className="flex items-center">
            <div className="rounded-full bg-green-100 p-3">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
            </div>
            <div className="ml-4">
              <h3 className="text-sm font-medium text-gray-500">Total Notes</h3>
              <p className="text-2xl font-bold">{stats.totalNotes}</p>
            </div>
          </div>
        </div>

        {/* Total Memory Cards Card */}
        <div className="bg-white rounded-lg shadow p-6 border">
          <div className="flex items-center">
            <div className="rounded-full bg-yellow-100 p-3">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-yellow-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
              </svg>
            </div>
            <div className="ml-4">
              <h3 className="text-sm font-medium text-gray-500">Memory Cards</h3>
              <p className="text-2xl font-bold">{stats.totalMemoryCards}</p>
            </div>
          </div>
        </div>

        {/* Due for Review Card */}
        <div className="bg-white rounded-lg shadow p-6 border">
          <div className="flex items-center">
            <div className="rounded-full bg-red-100 p-3">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div className="ml-4">
              <h3 className="text-sm font-medium text-gray-500">Due for Review</h3>
              <p className="text-2xl font-bold">{stats.todayReviews}</p>
            </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Memory Card Status Distribution */}
        <div className="bg-white rounded-lg shadow p-6 border">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Memory Card Status</h3>
          <div className="space-y-4">
            <div>
              <div className="flex justify-between mb-1">
                <span className="text-sm font-medium text-blue-700">To Review ({stats.toReviewCards})</span>
                <span className="text-sm font-medium text-blue-700">
                  {stats.totalMemoryCards > 0 ? Math.round((stats.toReviewCards / stats.totalMemoryCards) * 100) : 0}%
                </span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2.5">
                <div 
                  className="bg-blue-600 h-2.5 rounded-full" 
                  style={{ width: `${stats.totalMemoryCards > 0 ? (stats.toReviewCards / stats.totalMemoryCards) * 100 : 0}%` }}
                ></div>
              </div>
            </div>
            <div>
              <div className="flex justify-between mb-1">
                <span className="text-sm font-medium text-yellow-700">Learning ({stats.learningCards})</span>
                <span className="text-sm font-medium text-yellow-700">
                  {stats.totalMemoryCards > 0 ? Math.round((stats.learningCards / stats.totalMemoryCards) * 100) : 0}%
                </span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2.5">
                <div 
                  className="bg-yellow-600 h-2.5 rounded-full" 
                  style={{ width: `${stats.totalMemoryCards > 0 ? (stats.learningCards / stats.totalMemoryCards) * 100 : 0}%` }}
                ></div>
              </div>
            </div>
            <div>
              <div className="flex justify-between mb-1">
                <span className="text-sm font-medium text-green-700">Mastered ({stats.masteredCards})</span>
                <span className="text-sm font-medium text-green-700">
                  {stats.totalMemoryCards > 0 ? Math.round((stats.masteredCards / stats.totalMemoryCards) * 100) : 0}%
                </span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2.5">
                <div 
                  className="bg-green-600 h-2.5 rounded-full" 
                  style={{ width: `${stats.totalMemoryCards > 0 ? (stats.masteredCards / stats.totalMemoryCards) * 100 : 0}%` }}
                ></div>
              </div>
            </div>
          </div>
        </div>

        {/* Recent Activity */}
        <div className="bg-white rounded-lg shadow p-6 border">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Recent Activity</h3>
          <div className="space-y-4">
            {recentItems.length > 0 ? (
              recentItems.map((item) => (
                <div key={`${item.type}-${item.urn}`} className="flex items-start">
                  <div className="shrink-0">
                    {item.type === 'bookmark' && (
                      <div className="bg-blue-100 rounded-full p-2">
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                        </svg>
                      </div>
                    )}
                  </div>
                  <div className="ml-3">
                    <p className="text-sm font-medium text-gray-900">{item.title}</p>
                    <p className="text-sm text-gray-500">{item.item_type}: {item.item_id} • {item.date}</p>
                  </div>
                </div>
              ))
            ) : (
              <p className="text-sm text-gray-500">No recent activity</p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default NotesDashboard;