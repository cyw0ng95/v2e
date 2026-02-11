import React, { useState } from 'react';
import { rpcClient } from '@/lib/rpc-client';
import { Bookmark, NoteModel as Note, MemoryCard, CrossReference, HistoryEntry } from '@/lib/types';

interface ExportImportData {
  bookmarks: Bookmark[];
  notes: Note[];
  memoryCards: MemoryCard[];
  crossReferences: CrossReference[];
  history: HistoryEntry[];
}

const NotesExportImport: React.FC = () => {
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [importProgress, setImportProgress] = useState<{current: number, total: number} | null>(null);

  const handleExport = async () => {
    setLoading(true);
    setError(null);
    setSuccess(null);
    
    try {
      // Fetch all data to export
      const bookmarksResponse = await rpcClient.listBookmarks({ limit: 1000 });
      let allBookmarks: Bookmark[] = [];
      if (bookmarksResponse.retcode === 0 && bookmarksResponse.payload) {
        allBookmarks = bookmarksResponse.payload.bookmarks;
      }

      // Fetch notes for each bookmark
      let allNotes: Note[] = [];
      for (const bookmark of allBookmarks) {
        const notesResponse = await rpcClient.getNotesByBookmark({ bookmark_id: bookmark.id });
        if (notesResponse.retcode === 0 && notesResponse.payload) {
          allNotes = allNotes.concat(notesResponse.payload.notes);
        }
      }

      // Fetch memory cards for each bookmark
      let allMemoryCards: MemoryCard[] = [];
      for (const bookmark of allBookmarks) {
        const cardsResponse = await rpcClient.listMemoryCards({ bookmark_id: bookmark.id });
        if (cardsResponse.retcode === 0 && cardsResponse.payload) {
          allMemoryCards = allMemoryCards.concat(cardsResponse.payload.memory_cards);
        }
      }

      // Fetch cross references
      const crossRefsResponse = await rpcClient.listCrossReferences({ limit: 1000 });
      let allCrossRefs: CrossReference[] = [];
      if (crossRefsResponse.retcode === 0 && crossRefsResponse.payload) {
        allCrossRefs = crossRefsResponse.payload.cross_references;
      }

      // Fetch history - get history for all items by iterating through bookmarks
      let allHistory: HistoryEntry[] = [];
      for (const bookmark of allBookmarks) {
        const historyResponse = await rpcClient.getHistory({ 
          item_id: bookmark.global_item_id, 
          item_type: bookmark.item_type 
        });
        if (historyResponse.retcode === 0 && historyResponse.payload) {
          allHistory = allHistory.concat(historyResponse.payload.history_entries);
        }
      }

      // Create export data object
      const exportData: ExportImportData = {
        bookmarks: allBookmarks,
        notes: allNotes,
        memoryCards: allMemoryCards,
        crossReferences: allCrossRefs,
        history: allHistory
      };

      // Convert to JSON and trigger download
      const jsonString = JSON.stringify(exportData, null, 2);
      const blob = new Blob([jsonString], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      
      const a = document.createElement('a');
      a.href = url;
      a.download = `v2e-notes-export-${new Date().toISOString().slice(0, 10)}.json`;
      document.body.appendChild(a);
      a.click();
      
      // Clean up
      setTimeout(() => {
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
      }, 100);

      setSuccess('Data exported successfully!');
    } catch (err) {
      setError('Failed to export data');
      console.error('Error exporting data:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleFileUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    setLoading(true);
    setError(null);
    setSuccess(null);
    setImportProgress({ current: 0, total: 0 });

    try {
      const text = await file.text();
      const importData: ExportImportData = JSON.parse(text);

      // Calculate total items to import
      const totalItems = 
        importData.bookmarks.length +
        importData.notes.length +
        importData.memoryCards.length +
        importData.crossReferences.length +
        importData.history.length;

      setImportProgress({ current: 0, total: totalItems });

      let processed = 0;

      // Import bookmarks
      for (const bookmark of importData.bookmarks) {
        await rpcClient.createBookmark({
          global_item_id: bookmark.global_item_id,
          item_type: bookmark.item_type,
          item_id: bookmark.item_id,
          title: bookmark.title,
          description: bookmark.description,
          author: bookmark.author,
          is_private: bookmark.is_private,
          metadata: bookmark.metadata
        });
        processed++;
        setImportProgress({ current: processed, total: totalItems });
      }

      // Import notes
      for (const note of importData.notes) {
        await rpcClient.addNote({
          bookmark_id: note.bookmark_id,
          content: note.content,
          author: note.author,
          is_private: note.is_private,
          metadata: note.metadata
        });
        processed++;
        setImportProgress({ current: processed, total: totalItems });
      }

      // Import memory cards
      for (const card of importData.memoryCards) {
        await rpcClient.createMemoryCard({
          bookmark_id: card.bookmark_id,
          front: card.front_content,
          back: card.back_content,
          card_type: card.card_type,
          author: card.author,
          is_private: card.is_private,
          metadata: card.metadata
        });
        processed++;
        setImportProgress({ current: processed, total: totalItems });
      }

      // Import cross references
      for (const ref of importData.crossReferences) {
        await rpcClient.createCrossReference({
          from_item_id: ref.from_item_id,
          from_item_type: ref.from_item_type,
          to_item_id: ref.to_item_id,
          to_item_type: ref.to_item_type,
          relationship_type: ref.relationship_type,
          description: ref.description,
          strength: ref.strength,
          author: ref.author,
          is_private: ref.is_private,
          metadata: ref.metadata
        });
        processed++;
        setImportProgress({ current: processed, total: totalItems });
      }

      // Import history
      for (const hist of importData.history) {
        await rpcClient.addHistory({
          item_id: hist.item_id,
          item_type: hist.item_type,
          action: hist.action,
          old_values: hist.old_values,
          new_values: hist.new_values,
          author: hist.author,
          metadata: hist.metadata
        });
        processed++;
        setImportProgress({ current: processed, total: totalItems });
      }

      setSuccess(`Successfully imported ${processed} items!`);
    } catch (err) {
      setError('Failed to import data. Please check the file format.');
      console.error('Error importing data:', err);
    } finally {
      setLoading(false);
      setImportProgress(null);
      // Reset file input
      event.target.value = '';
    }
  };

  return (
    <div className="bg-white rounded-lg shadow p-6 border">
      <h3 className="text-lg font-medium text-gray-900 mb-4">Export/Import Notes Data</h3>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Export Section */}
        <div className="border rounded-lg p-4 bg-gray-50">
          <h4 className="font-medium text-gray-800 mb-2">Export Data</h4>
          <p className="text-sm text-gray-600 mb-4">
            Export all your bookmarks, notes, memory cards, and related data to a JSON file.
          </p>
          <button
            onClick={handleExport}
            disabled={loading}
            className="px-4 py-2 bg-blue-100 text-blue-700 rounded hover:bg-blue-200 disabled:opacity-50"
          >
            {loading ? 'Exporting...' : 'Export All Data'}
          </button>
        </div>

        {/* Import Section */}
        <div className="border rounded-lg p-4 bg-gray-50">
          <h4 className="font-medium text-gray-800 mb-2">Import Data</h4>
          <p className="text-sm text-gray-600 mb-4">
            Import notes data from a JSON file. This will add to your existing data.
          </p>
          <label className="block">
            <input
              type="file"
              accept=".json"
              onChange={handleFileUpload}
              disabled={loading}
              className="hidden"
            />
            <div className="px-4 py-2 bg-green-100 text-green-700 rounded hover:bg-green-200 disabled:opacity-50 cursor-pointer inline-block">
              {loading ? 'Importing...' : 'Choose File to Import'}
            </div>
          </label>
        </div>
      </div>

      {/* Progress indicator */}
      {importProgress && (
        <div className="mt-4">
          <div className="flex justify-between text-sm text-gray-600 mb-1">
            <span>Importing...</span>
            <span>{importProgress.current} of {importProgress.total}</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2.5">
            <div 
              className="bg-blue-600 h-2.5 rounded-full transition-all duration-300" 
              style={{ width: `${(importProgress.current / importProgress.total) * 100}%` }}
            ></div>
          </div>
        </div>
      )}

      {/* Status messages */}
      {error && (
        <div className="mt-4 p-3 bg-red-50 text-red-700 rounded-md">
          {error}
        </div>
      )}

      {success && (
        <div className="mt-4 p-3 bg-green-50 text-green-700 rounded-md">
          {success}
        </div>
      )}
    </div>
  );
};

export default NotesExportImport;