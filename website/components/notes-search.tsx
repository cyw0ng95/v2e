import React, { useState, useEffect } from 'react';
import { rpcClient } from '@/lib/rpc-client';
import { Bookmark, NoteModel as Note, MemoryCard } from '@/lib/types';

interface NotesSearchProps {
  onResultClick?: (bookmark: Bookmark, note?: Note, card?: MemoryCard) => void;
}

const NotesSearch: React.FC<NotesSearchProps> = ({ onResultClick }) => {
  const [query, setQuery] = useState<string>('');
  const [results, setResults] = useState<Array<{type: string, item: any}>>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState({
    includeBookmarks: true,
    includeNotes: true,
    includeCards: true,
    itemType: 'all',
    dateRange: 'all'
  });

  useEffect(() => {
    if (query.trim()) {
      performSearch();
    } else {
      setResults([]);
    }
  }, [query, filters]);

  const performSearch = async () => {
    setLoading(true);
    setError(null);
    try {
      const searchResults = [];

      // Search bookmarks if enabled
      if (filters.includeBookmarks) {
        const bookmarkResponse = await rpcClient.listBookmarks({ 
          limit: 10
        });
        if (bookmarkResponse.retcode === 0 && bookmarkResponse.payload) {
          const bookmarks = bookmarkResponse.payload.bookmarks.filter(b => 
            b.title.toLowerCase().includes(query.toLowerCase()) || 
            b.description.toLowerCase().includes(query.toLowerCase())
          );
          for (const bookmark of bookmarks) {
            searchResults.push({
              type: 'bookmark',
              item: bookmark
            });
          }
        }
      }

      // Search notes if enabled
      if (filters.includeNotes) {
        // We need to get all bookmarks first and then search their notes
        const bookmarkResponse = await rpcClient.listBookmarks({ limit: 100 });
        if (bookmarkResponse.retcode === 0 && bookmarkResponse.payload) {
          for (const bookmark of bookmarkResponse.payload.bookmarks) {
            const notesResponse = await rpcClient.getNotesByBookmark({
              bookmark_id: bookmark.id
            });
            if (notesResponse.retcode === 0 && notesResponse.payload) {
              const matchingNotes = notesResponse.payload.notes.filter(note => 
                note.content.toLowerCase().includes(query.toLowerCase())
              );
              for (const note of matchingNotes) {
                searchResults.push({
                  type: 'note',
                  item: { ...note, bookmark }
                });
              }
            }
          }
        }
      }

      // Search memory cards if enabled
      if (filters.includeCards) {
        // Get all bookmarks and search their memory cards
        const bookmarkResponse = await rpcClient.listBookmarks({ limit: 100 });
        if (bookmarkResponse.retcode === 0 && bookmarkResponse.payload) {
          for (const bookmark of bookmarkResponse.payload.bookmarks) {
            const cardsResponse = await rpcClient.listMemoryCards({
              bookmark_id: bookmark.id
            });
            if (cardsResponse.retcode === 0 && cardsResponse.payload) {
              const matchingCards = cardsResponse.payload.memory_cards.filter(card => 
                card.front_content.toLowerCase().includes(query.toLowerCase()) ||
                card.back_content.toLowerCase().includes(query.toLowerCase())
              );
              for (const card of matchingCards) {
                searchResults.push({
                  type: 'card',
                  item: { ...card, bookmark }
                });
              }
            }
          }
        }
      }

      setResults(searchResults);
    } catch (err) {
      setError('Failed to perform search');
      console.error('Error performing search:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleFilterChange = (filterName: string, value: any) => {
    setFilters(prev => ({
      ...prev,
      [filterName]: value
    }));
  };

  const handleResultClick = (result: {type: string, item: any}) => {
    if (onResultClick) {
      if (result.type === 'bookmark') {
        onResultClick(result.item);
      } else if (result.type === 'note') {
        onResultClick(result.item.bookmark, result.item);
      } else if (result.type === 'card') {
        onResultClick(result.item.bookmark, undefined, result.item);
      }
    }
  };

  return (
    <div className="bg-white rounded-lg shadow p-4 border">
      <div className="mb-4">
        <div className="flex space-x-2">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search bookmarks, notes, and memory cards..."
            className="flex-1 border rounded px-3 py-2 text-sm"
          />
          <button 
            onClick={performSearch}
            disabled={loading}
            className="px-4 py-2 bg-blue-100 text-blue-700 rounded hover:bg-blue-200 disabled:opacity-50 text-sm"
          >
            {loading ? 'Searching...' : 'Search'}
          </button>
        </div>
      </div>

      {/* Filters */}
      <div className="mb-4 grid grid-cols-1 md:grid-cols-4 gap-2 text-sm">
        <div className="flex items-center">
          <input
            type="checkbox"
            id="includeBookmarks"
            checked={filters.includeBookmarks}
            onChange={(e) => handleFilterChange('includeBookmarks', e.target.checked)}
            className="mr-2"
          />
          <label htmlFor="includeBookmarks">Bookmarks</label>
        </div>
        <div className="flex items-center">
          <input
            type="checkbox"
            id="includeNotes"
            checked={filters.includeNotes}
            onChange={(e) => handleFilterChange('includeNotes', e.target.checked)}
            className="mr-2"
          />
          <label htmlFor="includeNotes">Notes</label>
        </div>
        <div className="flex items-center">
          <input
            type="checkbox"
            id="includeCards"
            checked={filters.includeCards}
            onChange={(e) => handleFilterChange('includeCards', e.target.checked)}
            className="mr-2"
          />
          <label htmlFor="includeCards">Memory Cards</label>
        </div>
        <div>
          <select
            value={filters.itemType}
            onChange={(e) => handleFilterChange('itemType', e.target.value)}
            className="border rounded px-2 py-1 text-sm w-full"
          >
            <option value="all">All Types</option>
            <option value="CVE">CVE</option>
            <option value="CWE">CWE</option>
            <option value="CAPEC">CAPEC</option>
            <option value="ATT&CK">ATT&CK</option>
          </select>
        </div>
      </div>

      {/* Results */}
      <div>
        {error && (
          <div className="p-3 bg-red-50 text-red-700 rounded-md mb-4">
            {error}
          </div>
        )}

        {query && (
          <div className="text-sm text-gray-600 mb-2">
            Found {results.length} results for "{query}"
          </div>
        )}

        <div className="max-h-96 overflow-y-auto">
          {results.length === 0 && !loading && query ? (
            <div className="text-center py-8 text-gray-500">
              No results found for "{query}"
            </div>
          ) : (
            <ul className="space-y-2">
              {results.map((result, index) => (
                <li 
                  key={index} 
                  className="p-3 border rounded hover:bg-gray-50 cursor-pointer transition-colors"
                  onClick={() => handleResultClick(result)}
                >
                  <div className="flex justify-between items-start">
                    <div>
                      <div className="font-medium text-sm">
                        {result.type === 'bookmark' && (
                          <span className="inline-flex items-center">
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                            </svg>
                            Bookmark
                          </span>
                        )}
                        {result.type === 'note' && (
                          <span className="inline-flex items-center">
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                            </svg>
                            Note
                          </span>
                        )}
                        {result.type === 'card' && (
                          <span className="inline-flex items-center">
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1 text-yellow-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                            </svg>
                            Memory Card
                          </span>
                        )}
                      </div>
                      <div className="font-medium text-gray-900 text-sm mt-1">
                        {result.type === 'bookmark' && result.item.title}
                        {result.type === 'note' && `"${result.item.content.substring(0, 50)}${result.item.content.length > 50 ? '...' : ''}"`}
                        {result.type === 'card' && `"${result.item.front_content.substring(0, 50)}${result.item.front_content.length > 50 ? '...' : ''}"`}
                      </div>
                      <div className="text-xs text-gray-500">
                        {result.type === 'bookmark' && `${result.item.item_type}: ${result.item.item_id}`}
                        {result.type === 'note' && `in Bookmark: ${result.item.bookmark.title}`}
                        {result.type === 'card' && `Learning State: ${result.item.learning_state}`}
                      </div>
                    </div>
                    <div className="text-xs text-gray-400">
                      {new Date(result.item.created_at).toLocaleDateString()}
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
};

export default NotesSearch;