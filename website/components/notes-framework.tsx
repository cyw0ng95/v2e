import React, { useState, useEffect } from 'react';
import { rpcClient } from '@/lib/rpc-client';
import { 
  Bookmark, 
  NoteModel as Note, 
  MemoryCard, 
  CrossReference, 
  HistoryEntry 
} from '@/lib/types';

interface NotesFrameworkProps {
  itemId: string;
  itemType: string;
  itemTitle: string;
  itemDescription?: string;
}

const NotesFramework: React.FC<NotesFrameworkProps> = ({ 
  itemId, 
  itemType, 
  itemTitle, 
  itemDescription = '' 
}) => {
  const [isBookmarked, setIsBookmarked] = useState<boolean>(false);
  const [bookmark, setBookmark] = useState<Bookmark | null>(null);
  const [notes, setNotes] = useState<Note[]>([]);
  const [newNote, setNewNote] = useState<string>('');
  const [memoryCards, setMemoryCards] = useState<MemoryCard[]>([]);
  const [showNotes, setShowNotes] = useState<boolean>(false);
  const [showMemoryCards, setShowMemoryCards] = useState<boolean>(false);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  // Check if item is already bookmarked on component mount
  useEffect(() => {
    checkBookmarkStatus();
  }, [itemId, itemType]);

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
        
        // Load associated notes
        const notesResponse = await rpcClient.getNotesByBookmark({
          bookmark_id: foundBookmark.id
        });
        
        if (notesResponse.retcode === 0 && notesResponse.payload) {
          setNotes(notesResponse.payload.notes);
        }
        
        // Load memory cards
        const cardsResponse = await rpcClient.listMemoryCards({
          bookmark_id: foundBookmark.id
        });
        
        if (cardsResponse.retcode === 0 && cardsResponse.payload) {
          setMemoryCards(cardsResponse.payload.memory_cards);
        }
      } else {
        setIsBookmarked(false);
        setBookmark(null);
      }
    } catch (err) {
      setError('Failed to check bookmark status');
      console.error('Error checking bookmark status:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleBookmarkToggle = async () => {
    if (isBookmarked && bookmark) {
      // Unbookmark
      const response = await rpcClient.deleteBookmark({ id: bookmark.id });
      if (response.retcode === 0) {
        setIsBookmarked(false);
        setBookmark(null);
        setNotes([]);
        setMemoryCards([]);
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
  };

  const handleAddNote = async () => {
    if (!newNote.trim() || !bookmark) return;

    const response = await rpcClient.addNote({
      bookmark_id: bookmark.id,
      content: newNote.trim()
    });

    if (response.retcode === 0 && response.payload?.note) {
      setNotes([...notes, response.payload.note]);
      setNewNote('');
    } else {
      setError('Failed to add note');
    }
  };

  const handleCreateMemoryCard = async (front: string, back: string) => {
    if (!front.trim() || !back.trim() || !bookmark) return;

    const response = await rpcClient.createMemoryCard({
      bookmark_id: bookmark.id,
      front_content: front.trim(),
      back_content: back.trim()
    });

    if (response.retcode === 0 && response.payload?.memory_card) {
      setMemoryCards([...memoryCards, response.payload.memory_card]);
    } else {
      setError('Failed to create memory card');
    }
  };

  return (
    <div className="border rounded-lg p-4 bg-white shadow-sm">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold">Notes & Bookmarks</h3>
        <button
          onClick={handleBookmarkToggle}
          disabled={loading}
          className={`px-4 py-2 rounded-md ${
            isBookmarked
              ? 'bg-red-100 text-red-700 hover:bg-red-200'
              : 'bg-blue-100 text-blue-700 hover:bg-blue-200'
          } transition-colors disabled:opacity-50`}
        >
          {loading ? 'Loading...' : isBookmarked ? 'Unbookmark' : 'Bookmark'}
        </button>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-50 text-red-700 rounded-md">
          {error}
        </div>
      )}

      {isBookmarked && bookmark && (
        <div className="space-y-4">
          {/* Notes Section */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <h4 className="font-medium text-gray-700">Notes</h4>
              <button
                onClick={() => setShowNotes(!showNotes)}
                className="text-sm text-blue-600 hover:text-blue-800"
              >
                {showNotes ? 'Hide' : 'Show'} Notes
              </button>
            </div>

            {showNotes && (
              <div className="space-y-3">
                {/* Add Note Form */}
                <div className="flex space-x-2">
                  <input
                    type="text"
                    value={newNote}
                    onChange={(e) => setNewNote(e.target.value)}
                    placeholder="Add a note..."
                    className="flex-1 border rounded px-3 py-2 text-sm"
                    onKeyPress={(e) => e.key === 'Enter' && handleAddNote()}
                  />
                  <button
                    onClick={handleAddNote}
                    disabled={!newNote.trim()}
                    className="px-3 py-2 bg-green-100 text-green-700 rounded hover:bg-green-200 disabled:opacity-50 text-sm"
                  >
                    Add
                  </button>
                </div>

                {/* Notes List */}
                <div className="space-y-2 max-h-40 overflow-y-auto">
                  {notes.length === 0 ? (
                    <p className="text-sm text-gray-500 italic">No notes yet</p>
                  ) : (
                    notes.map((note) => (
                      <div key={note.id} className="p-2 bg-gray-50 rounded text-sm">
                        {note.content}
                        <div className="text-xs text-gray-500 mt-1">
                          {new Date(note.created_at).toLocaleDateString()}
                        </div>
                      </div>
                    ))
                  )}
                </div>
              </div>
            )}
          </div>

          {/* Memory Cards Section */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <h4 className="font-medium text-gray-700">Memory Cards</h4>
              <button
                onClick={() => setShowMemoryCards(!showMemoryCards)}
                className="text-sm text-blue-600 hover:text-blue-800"
              >
                {showMemoryCards ? 'Hide' : 'Show'} Cards
              </button>
            </div>

            {showMemoryCards && (
              <div className="space-y-3">
                {/* Create Memory Card Form */}
                <CreateMemoryCardForm onCreate={handleCreateMemoryCard} />

                {/* Memory Cards List */}
                <div className="grid grid-cols-1 gap-2 max-h-60 overflow-y-auto">
                  {memoryCards.length === 0 ? (
                    <p className="text-sm text-gray-500 italic">No memory cards yet</p>
                  ) : (
                    memoryCards.map((card) => (
                      <MemoryCardDisplay key={card.id} card={card} />
                    ))
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {!isBookmarked && (
        <p className="text-gray-500 text-sm">
          Bookmark this item to save notes and create memory cards.
        </p>
      )}
    </div>
  );
};

interface CreateMemoryCardFormProps {
  onCreate: (front: string, back: string) => void;
}

const CreateMemoryCardForm: React.FC<CreateMemoryCardFormProps> = ({ onCreate }) => {
  const [front, setFront] = useState('');
  const [back, setBack] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (front.trim() && back.trim()) {
      onCreate(front.trim(), back.trim());
      setFront('');
      setBack('');
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-2">
      <input
        type="text"
        value={front}
        onChange={(e) => setFront(e.target.value)}
        placeholder="Front of card..."
        className="w-full border rounded px-3 py-1 text-sm"
      />
      <input
        type="text"
        value={back}
        onChange={(e) => setBack(e.target.value)}
        placeholder="Back of card..."
        className="w-full border rounded px-3 py-1 text-sm"
      />
      <button
        type="submit"
        disabled={!front.trim() || !back.trim()}
        className="px-3 py-1 bg-purple-100 text-purple-700 rounded hover:bg-purple-200 disabled:opacity-50 text-sm"
      >
        Create Card
      </button>
    </form>
  );
};

interface MemoryCardDisplayProps {
  card: MemoryCard;
}

const MemoryCardDisplay: React.FC<MemoryCardDisplayProps> = ({ card }) => {
  const [showAnswer, setShowAnswer] = useState(false);

  return (
    <div 
      className="p-3 bg-yellow-50 border rounded cursor-pointer hover:bg-yellow-100"
      onClick={() => setShowAnswer(!showAnswer)}
    >
      <div className="font-medium">{card.front_content}</div>
      {showAnswer && (
        <div className="mt-2 pt-2 border-t text-sm">
          <div>{card.back_content}</div>
          <div className="text-xs text-gray-500 mt-1">
            State: {card.learning_state} | Next review: {new Date(card.next_review_at).toLocaleDateString()}
          </div>
        </div>
      )}
    </div>
  );
};

export default NotesFramework;