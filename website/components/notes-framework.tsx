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
  const [crossReferences, setCrossReferences] = useState<CrossReference[]>([]);
  const [histories, setHistories] = useState<HistoryEntry[]>([]);
  const [showNotes, setShowNotes] = useState<boolean>(false);
  const [showMemoryCards, setShowMemoryCards] = useState<boolean>(false);
  const [showCrossRefs, setShowCrossRefs] = useState<boolean>(false);
  const [showHistory, setShowHistory] = useState<boolean>(false);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [editingNoteId, setEditingNoteId] = useState<number | null>(null);
  const [editingNoteContent, setEditingNoteContent] = useState<string>('');
  const [editingCardId, setEditingCardId] = useState<number | null>(null);
  const [editingCardFront, setEditingCardFront] = useState<string>('');
  const [editingCardBack, setEditingCardBack] = useState<string>('');

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
        
        // Load cross-references
        const refsResponse = await rpcClient.listCrossReferences({
          from_item_id: itemId,
          from_item_type: itemType
        });
        
        if (refsResponse.retcode === 0 && refsResponse.payload) {
          setCrossReferences(refsResponse.payload.cross_references);
        }
        
        // Load history
        const historyResponse = await rpcClient.getHistory({
          item_id: `${itemType}-${itemId}`,
          item_type: itemType
        });
        
        if (historyResponse.retcode === 0 && historyResponse.payload) {
          setHistories(historyResponse.payload.history_entries);
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
        setCrossReferences([]);
        setHistories([]);
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
      // Add to history
      await rpcClient.addHistory({
        item_id: `${itemType}-${itemId}`,
        item_type: itemType,
        action: 'note_added',
        new_values: { note_id: response.payload.note.id, content: newNote.trim() }
      });
      const newHistoryEntry: HistoryEntry = {
        id: histories.length + 1,
        item_id: `${itemType}-${itemId}`,
        item_type: itemType,
        action: 'note_added',
        timestamp: new Date().toISOString(),
        new_values: { note_id: response.payload.note.id, content: newNote.trim() },
        metadata: {}
      };
      setHistories(prev => [...prev, newHistoryEntry]);
    } else {
      setError('Failed to add note');
    }
  };

  const handleUpdateNote = async (noteId: number) => {
    if (!editingNoteContent.trim()) return;

    const response = await rpcClient.updateNote({
      id: noteId,
      content: editingNoteContent
    });

    if (response.retcode === 0 && response.payload?.note) {
      setNotes(notes.map(note => note.id === noteId ? response.payload!.note : note));
      setEditingNoteId(null);
      setEditingNoteContent('');
    } else {
      setError('Failed to update note');
    }
  };

  const handleDeleteNote = async (noteId: number) => {
    const response = await rpcClient.deleteNote({ id: noteId });

    if (response.retcode === 0) {
      setNotes(notes.filter(note => note.id !== noteId));
    } else {
      setError('Failed to delete note');
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
      // Add to history
      await rpcClient.addHistory({
        item_id: `${itemType}-${itemId}`,
        item_type: itemType,
        action: 'memory_card_created',
        new_values: { card_id: response.payload.memory_card.id, front: front.trim() }
      });
      const newHistoryEntry: HistoryEntry = {
        id: histories.length + 1,
        item_id: `${itemType}-${itemId}`,
        item_type: itemType,
        action: 'memory_card_created',
        timestamp: new Date().toISOString(),
        new_values: { card_id: response.payload.memory_card.id, front: front.trim() },
        metadata: {}
      };
      setHistories(prev => [...prev, newHistoryEntry]);
    } else {
      setError('Failed to create memory card');
    }
  };

  const handleUpdateMemoryCard = async (cardId: number) => {
    if (!editingCardFront.trim() || !editingCardBack.trim()) return;

    const response = await rpcClient.updateMemoryCard({
      id: cardId,
      front_content: editingCardFront,
      back_content: editingCardBack
    });

    if (response.retcode === 0 && response.payload?.memory_card) {
      setMemoryCards(memoryCards.map(card => card.id === cardId ? response.payload!.memory_card : card));
      setEditingCardId(null);
      setEditingCardFront('');
      setEditingCardBack('');
    } else {
      setError('Failed to update memory card');
    }
  };

  const handleDeleteMemoryCard = async (cardId: number) => {
    const response = await rpcClient.deleteMemoryCard({ id: cardId });

    if (response.retcode === 0) {
      setMemoryCards(memoryCards.filter(card => card.id !== cardId));
    } else {
      setError('Failed to delete memory card');
    }
  };

  const handleRateMemoryCard = async (cardId: number, rating: string) => {
    const response = await rpcClient.rateMemoryCard({
      id: cardId,
      rating: rating
    });

    if (response.retcode === 0 && response.payload?.memory_card) {
      setMemoryCards(memoryCards.map(card => card.id === cardId ? response.payload!.memory_card : card));
    } else {
      setError('Failed to rate memory card');
    }
  };

  const handleCreateCrossReference = async (toItemId: string, toItemtype: string, relationshipType: string) => {
    if (!toItemId.trim() || !toItemtype.trim() || !relationshipType.trim() || !bookmark) return;

    const response = await rpcClient.createCrossReference({
      from_item_id: `${itemType}-${itemId}`,
      from_item_type: itemType,
      to_item_id: toItemId,
      to_item_type: toItemtype,
      relationship_type: relationshipType
    });

    if (response.retcode === 0 && response.payload?.cross_reference) {
      setCrossReferences([...crossReferences, response.payload.cross_reference]);
    } else {
      setError('Failed to create cross-reference');
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
              <h4 className="font-medium text-gray-700">Notes ({notes.length})</h4>
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
                <div className="space-y-2 max-h-60 overflow-y-auto">
                  {notes.length === 0 ? (
                    <p className="text-sm text-gray-500 italic">No notes yet</p>
                  ) : (
                    notes.map((note) => (
                      <div key={note.id} className="p-2 bg-gray-50 rounded text-sm border">
                        {editingNoteId === note.id ? (
                          <div className="space-y-2">
                            <textarea
                              value={editingNoteContent}
                              onChange={(e) => setEditingNoteContent(e.target.value)}
                              className="w-full border rounded px-2 py-1 text-sm"
                              rows={2}
                            />
                            <div className="flex space-x-1">
                              <button
                                onClick={() => handleUpdateNote(note.id)}
                                className="px-2 py-1 bg-blue-100 text-blue-700 rounded text-xs"
                              >
                                Save
                              </button>
                              <button
                                onClick={() => {
                                  setEditingNoteId(null);
                                  setEditingNoteContent('');
                                }}
                                className="px-2 py-1 bg-gray-100 text-gray-700 rounded text-xs"
                              >
                                Cancel
                              </button>
                            </div>
                          </div>
                        ) : (
                          <>
                            <div>{note.content}</div>
                            <div className="flex justify-between items-center text-xs text-gray-500 mt-1">
                              <span>{new Date(note.created_at).toLocaleDateString()}</span>
                              <div className="space-x-1">
                                <button
                                  onClick={() => {
                                    setEditingNoteId(note.id);
                                    setEditingNoteContent(note.content);
                                  }}
                                  className="text-blue-600 hover:text-blue-800"
                                >
                                  Edit
                                </button>
                                <button
                                  onClick={() => handleDeleteNote(note.id)}
                                  className="text-red-600 hover:text-red-800"
                                >
                                  Delete
                                </button>
                              </div>
                            </div>
                          </>
                        )}
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
              <h4 className="font-medium text-gray-700">Memory Cards ({memoryCards.length})</h4>
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
                <div className="grid grid-cols-1 gap-2 max-h-80 overflow-y-auto">
                  {memoryCards.length === 0 ? (
                    <p className="text-sm text-gray-500 italic">No memory cards yet</p>
                  ) : (
                    memoryCards.map((card) => (
                      <div key={card.id} className="p-3 bg-yellow-50 border rounded">
                        {editingCardId === card.id ? (
                          <div className="space-y-2">
                            <input
                              type="text"
                              value={editingCardFront}
                              onChange={(e) => setEditingCardFront(e.target.value)}
                              placeholder="Front of card..."
                              className="w-full border rounded px-2 py-1 text-sm"
                            />
                            <input
                              type="text"
                              value={editingCardBack}
                              onChange={(e) => setEditingCardBack(e.target.value)}
                              placeholder="Back of card..."
                              className="w-full border rounded px-2 py-1 text-sm"
                            />
                            <div className="flex space-x-1">
                              <button
                                onClick={() => handleUpdateMemoryCard(card.id)}
                                className="px-2 py-1 bg-blue-100 text-blue-700 rounded text-xs"
                              >
                                Save
                              </button>
                              <button
                                onClick={() => {
                                  setEditingCardId(null);
                                  setEditingCardFront('');
                                  setEditingCardBack('');
                                }}
                                className="px-2 py-1 bg-gray-100 text-gray-700 rounded text-xs"
                              >
                                Cancel
                              </button>
                            </div>
                          </div>
                        ) : (
                          <>
                            <div className="font-medium">{card.front_content}</div>
                            <div className="mt-2 text-sm">
                              <div>{card.back_content}</div>
                              <div className="text-xs text-gray-500 mt-1">
                                State: {card.learning_state} | Next review: {new Date(card.next_review_at).toLocaleDateString()}
                              </div>
                              <div className="flex justify-between items-center mt-2">
                                <div className="text-xs">
                                  Interval: {card.interval} days | EF: {card.ease_factor.toFixed(2)}
                                </div>
                                <div className="space-x-1">
                                  <button
                                    onClick={() => {
                                      setEditingCardId(card.id);
                                      setEditingCardFront(card.front_content);
                                      setEditingCardBack(card.back_content);
                                    }}
                                    className="text-blue-600 hover:text-blue-800 text-xs"
                                  >
                                    Edit
                                  </button>
                                  <button
                                    onClick={() => handleDeleteMemoryCard(card.id)}
                                    className="text-red-600 hover:text-red-800 text-xs"
                                  >
                                    Delete
                                  </button>
                                  <RatingButtons onRate={(rating) => handleRateMemoryCard(card.id, rating)} />
                                </div>
                              </div>
                            </div>
                          </>
                        )}
                      </div>
                    ))
                  )}
                </div>
              </div>
            )}
          </div>

          {/* Cross References Section */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <h4 className="font-medium text-gray-700">Cross References ({crossReferences.length})</h4>
              <button
                onClick={() => setShowCrossRefs(!showCrossRefs)}
                className="text-sm text-blue-600 hover:text-blue-800"
              >
                {showCrossRefs ? 'Hide' : 'Show'} References
              </button>
            </div>

            {showCrossRefs && (
              <div className="space-y-3">
                <CrossReferenceForm onCreate={handleCreateCrossReference} />
                
                <div className="space-y-2 max-h-40 overflow-y-auto">
                  {crossReferences.length === 0 ? (
                    <p className="text-sm text-gray-500 italic">No cross-references yet</p>
                  ) : (
                    crossReferences.map((ref) => (
                      <div key={ref.id} className="p-2 bg-blue-50 rounded text-sm border">
                        <div className="font-medium">{ref.relationship_type}</div>
                        <div className="text-xs text-gray-600">
                          From: {ref.from_item_id} ({ref.from_item_type}) 
                          {' '}→ To: {ref.to_item_id} ({ref.to_item_type})
                        </div>
                      </div>
                    ))
                  )}
                </div>
              </div>
            )}
          </div>

          {/* History Section */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <h4 className="font-medium text-gray-700">History ({histories.length})</h4>
              <button
                onClick={() => setShowHistory(!showHistory)}
                className="text-sm text-blue-600 hover:text-blue-800"
              >
                {showHistory ? 'Hide' : 'Show'} History
              </button>
            </div>

            {showHistory && (
              <div className="space-y-2 max-h-40 overflow-y-auto">
                {histories.length === 0 ? (
                  <p className="text-sm text-gray-500 italic">No history yet</p>
                ) : (
                  histories.map((entry, index) => (
                    <div key={index} className="p-2 bg-gray-100 rounded text-xs">
                      <div className="font-medium">{entry.action}</div>
                      <div className="text-gray-600">
                        {new Date(entry.timestamp).toLocaleString()} | {entry.item_type}: {entry.item_id}
                      </div>
                    </div>
                  ))
                )}
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
    <form onSubmit={handleSubmit} className="space-y-2 p-2 bg-gray-50 rounded border">
      <div className="text-xs font-medium mb-1">Create New Memory Card:</div>
      <input
        type="text"
        value={front}
        onChange={(e) => setFront(e.target.value)}
        placeholder="Front of card..."
        className="w-full border rounded px-2 py-1 text-sm"
      />
      <input
        type="text"
        value={back}
        onChange={(e) => setBack(e.target.value)}
        placeholder="Back of card..."
        className="w-full border rounded px-2 py-1 text-sm"
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

interface RatingButtonsProps {
  onRate: (rating: string) => void;
}

const RatingButtons: React.FC<RatingButtonsProps> = ({ onRate }) => {
  return (
    <div className="flex space-x-1">
      {['again', 'hard', 'good', 'easy'].map((rating) => (
        <button
          key={rating}
          onClick={() => onRate(rating)}
          className={`text-xs px-2 py-1 rounded ${
            rating === 'again' ? 'bg-red-100 text-red-700' :
            rating === 'hard' ? 'bg-orange-100 text-orange-700' :
            rating === 'good' ? 'bg-green-100 text-green-700' :
            'bg-blue-100 text-blue-700'
          }`}
        >
          {rating.charAt(0).toUpperCase()}
        </button>
      ))}
    </div>
  );
};

interface CrossReferenceFormProps {
  onCreate: (toItemId: string, toItemtype: string, relationshipType: string) => void;
}

const CrossReferenceForm: React.FC<CrossReferenceFormProps> = ({ onCreate }) => {
  const [toItemId, setToItemId] = useState('');
  const [toItemtype, setToItemtype] = useState('CVE');
  const [relationshipType, setRelationshipType] = useState('related_to');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (toItemId.trim() && toItemtype.trim() && relationshipType.trim()) {
      onCreate(toItemId.trim(), toItemtype.trim(), relationshipType.trim());
      setToItemId('');
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-2 p-2 bg-gray-50 rounded border">
      <div className="text-xs font-medium mb-1">Create Cross Reference:</div>
      <div className="grid grid-cols-3 gap-2">
        <select
          value={toItemtype}
          onChange={(e) => setToItemtype(e.target.value)}
          className="border rounded px-2 py-1 text-sm"
        >
          <option value="CVE">CVE</option>
          <option value="CWE">CWE</option>
          <option value="CAPEC">CAPEC</option>
          <option value="ATT&CK">ATT&CK</option>
        </select>
        <input
          type="text"
          value={toItemId}
          onChange={(e) => setToItemId(e.target.value)}
          placeholder="Target ID"
          className="border rounded px-2 py-1 text-sm"
        />
        <select
          value={relationshipType}
          onChange={(e) => setRelationshipType(e.target.value)}
          className="border rounded px-2 py-1 text-sm"
        >
          <option value="related_to">Related To</option>
          <option value="depends_on">Depends On</option>
          <option value="similar_to">Similar To</option>
          <option value="opposite_of">Opposite Of</option>
          <option value="causes">Causes</option>
          <option value="mitigates">Mitigates</option>
        </select>
      </div>
      <button
        type="submit"
        disabled={!toItemId.trim() || !toItemtype.trim() || !relationshipType.trim()}
        className="px-3 py-1 bg-indigo-100 text-indigo-700 rounded hover:bg-indigo-200 disabled:opacity-50 text-sm"
      >
        Create Reference
      </button>
    </form>
  );
};

export default NotesFramework;