'use client';

import React, { useState, useEffect } from 'react';
import { rpcClient } from '@/lib/rpc-client';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Skeleton } from '@/components/ui/skeleton';
import { 
  BookmarkIcon as Bookmark, 
  BookmarkCheckIcon as BookmarkCheck,
  FileTextIcon as FileText,
  BrainIcon as Brain,
  ChevronLeftIcon as ChevronLeft,
  ChevronRightIcon as ChevronRight,
  CheckCircleIcon as CheckCircle
} from '@/components/icons';
import { createLogger } from '@/lib/logger';

const logger = createLogger('learning-view');

interface LearningViewProps {
  urn: string;
  onClose?: () => void;
}

interface SecurityObject {
  urn: string;
  title: string;
  description: string;
  itemType: 'CVE' | 'CWE' | 'CAPEC' | 'ATTACK';
  itemId: string;
  metadata?: Record<string, any>;
}

const LearningView: React.FC<LearningViewProps> = ({ urn, onClose }) => {
  const [objectData, setObjectData] = useState<SecurityObject | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [isBookmarked, setIsBookmarked] = useState<boolean>(false);
  const [bookmarkId, setBookmarkId] = useState<number | null>(null);
  const [markAsLearned, setMarkAsLearned] = useState<boolean>(false);
  const [showNoteDialog, setShowNoteDialog] = useState<boolean>(false);
  const [showCardDialog, setShowCardDialog] = useState<boolean>(false);
  const [noteContent, setNoteContent] = useState<string>('');
  const [cardFront, setCardFront] = useState<string>('');
  const [cardBack, setCardBack] = useState<string>('');
  
  const [notesCount, setNotesCount] = useState<number>(0);
  const [cardsCount, setCardsCount] = useState<number>(0);

  useEffect(() => {
    loadObjectData();
  }, [urn]);

  const loadObjectData = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const parts = urn.split('::');
      if (parts.length < 4) {
        throw new Error('Invalid URN format');
      }
      
      const [, provider, type, itemId] = parts;
      const itemType = type.toUpperCase();
      
      let data: SecurityObject | null = null;
      
      switch (itemType) {
        case 'CVE': {
          const response = await rpcClient.getCVE(itemId);
          if (response.retcode === 0 && response.payload) {
            const cve = response.payload.cve;
            data = {
              urn,
              title: cve.id,
              description: cve.descriptions[0]?.value || 'No description',
              itemType: 'CVE',
              itemId: cve.id,
              metadata: { ...cve }
            };
          }
          break;
        }
        case 'CWE': {
          const response = await rpcClient.getCWE(itemId);
          if (response.retcode === 0 && response.payload) {
            const cwe = response.payload.cwe;
            data = {
              urn,
              title: `CWE-${cwe.id}`,
              description: cwe.name || 'No description',
              itemType: 'CWE',
              itemId: cwe.id.toString(),
              metadata: { ...cwe }
            };
          }
          break;
        }
        case 'CAPEC': {
          const response = await rpcClient.getCAPEC(itemId);
          if (response.retcode === 0 && response.payload) {
            const capec = response.payload;
            data = {
              urn,
              title: capec.name || itemId,
              description: capec.summary || capec.description || 'No description',
              itemType: 'CAPEC',
              itemId: itemId,
              metadata: { ...capec }
            };
          }
          break;
        }
        case 'ATTACK': {
          const response = await rpcClient.getAttackTechnique(itemId);
          if (response.retcode === 0 && response.payload) {
            const attack = response.payload;
            data = {
              urn,
              title: attack.name || itemId,
              description: attack.description || 'No description',
              itemType: 'ATTACK',
              itemId: itemId,
              metadata: { ...attack }
            };
          }
          break;
        }
        default:
          throw new Error(`Unsupported item type: ${itemType}`);
      }
      
      if (data) {
        setObjectData(data);
        await checkBookmarkStatus(data);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load object');
      logger.error('Error loading object data', err);
    } finally {
      setLoading(false);
    }
  };

  const checkBookmarkStatus = async (data: SecurityObject) => {
    try {
      const response = await rpcClient.listBookmarks({
        item_id: data.itemId,
        item_type: data.itemType
      });
      
      if (response.retcode === 0 && response.payload && response.payload.bookmarks.length > 0) {
        const bookmark = response.payload.bookmarks[0];
        setBookmarkId(bookmark.id);
        setIsBookmarked(true);
        
        const notesResponse = await rpcClient.getNotesByBookmark({
          bookmark_id: bookmark.id
        });
        if (notesResponse.retcode === 0 && notesResponse.payload) {
          setNotesCount(notesResponse.payload.total);
        }
        
        const cardsResponse = await rpcClient.listMemoryCards({
          bookmark_id: bookmark.id
        });
        if (cardsResponse.retcode === 0 && cardsResponse.payload) {
          setCardsCount(cardsResponse.payload.total);
        }
      }
    } catch (err) {
      logger.error('Error checking bookmark status', err);
    }
  };

  const handleBookmark = async () => {
    if (!objectData) return;
    
    try {
      if (isBookmarked && bookmarkId) {
        await rpcClient.deleteBookmark({ id: bookmarkId });
        setIsBookmarked(false);
        setBookmarkId(null);
        setNotesCount(0);
        setCardsCount(0);
      } else {
        const response = await rpcClient.createBookmark({
          global_item_id: objectData.itemId,
          item_type: objectData.itemType,
          item_id: objectData.itemId,
          title: objectData.title,
          description: objectData.description
        });
        
        if (response.retcode === 0 && response.payload) {
          setIsBookmarked(true);
          setBookmarkId(response.payload.bookmark.id);
          if (response.payload.memoryCard) {
            setCardsCount(1);
          }
        }
      }
    } catch (err) {
      logger.error('Error toggling bookmark', err);
      setError('Failed to bookmark item');
    }
  };

  const handleAddNote = async () => {
    if (!objectData || !noteContent.trim()) return;
    
    try {
      const response = await rpcClient.createBookmark({
        global_item_id: objectData.itemId,
        item_type: objectData.itemType,
        item_id: objectData.itemId,
        title: objectData.title,
        description: objectData.description
      });
      
      if (response.retcode === 0 && response.payload) {
        const addNoteResponse = await rpcClient.addNote({
          bookmark_id: response.payload.bookmark.id,
          content: noteContent
        });
        
        if (addNoteResponse.retcode === 0) {
          setNotesCount(prev => prev + 1);
          setNoteContent('');
          setShowNoteDialog(false);
          setIsBookmarked(true);
        }
      }
    } catch (err) {
      logger.error('Error adding note', err);
      setError('Failed to add note');
    }
  };

  const handleCreateCard = async () => {
    if (!objectData || !cardFront.trim() || !cardBack.trim()) return;
    
    try {
      const response = await rpcClient.createBookmark({
        global_item_id: objectData.itemId,
        item_type: objectData.itemType,
        item_id: objectData.itemId,
        title: objectData.title,
        description: objectData.description
      });
      
      if (response.retcode === 0 && response.payload) {
        const cardResponse = await rpcClient.createMemoryCard({
          bookmark_id: response.payload.bookmark.id,
          front: cardFront,
          back: cardBack,
          card_type: 'basic'
        });
        
        if (cardResponse.retcode === 0) {
          setCardsCount(prev => prev + 1);
          setCardFront('');
          setCardBack('');
          setShowCardDialog(false);
          setIsBookmarked(true);
        }
      }
    } catch (err) {
      logger.error('Error creating card', err);
      setError('Failed to create card');
    }
  };

  if (loading) {
    return (
      <Card className="w-full">
        <CardHeader>
          <Skeleton className="h-6 w-3/4 mb-2" />
          <Skeleton className="h-4 w-1/2" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-32 w-full mb-4" />
          <div className="flex space-x-2">
            <Skeleton className="h-10 w-24" />
            <Skeleton className="h-10 w-24" />
            <Skeleton className="h-10 w-24" />
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="w-full">
        <CardContent className="p-6">
          <div className="text-red-600">{error}</div>
        </CardContent>
      </Card>
    );
  }

  if (!objectData) {
    return null;
  }

  return (
    <Card className="w-full">
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center space-x-2 mb-2">
              <Badge variant="outline">{objectData.itemType}</Badge>
              {markAsLearned && (
                <Badge variant="default" className="bg-green-100 text-green-700">
                  <CheckCircle className="w-3 h-3 mr-1" />
                  Learned
                </Badge>
              )}
            </div>
            <CardTitle className="text-xl">{objectData.title}</CardTitle>
            <CardDescription className="mt-2 font-mono text-xs">
              URN: {objectData.urn}
            </CardDescription>
          </div>
          <div className="flex space-x-2">
            <Button
              variant={isBookmarked ? "default" : "outline"}
              size="sm"
              onClick={handleBookmark}
            >
              {isBookmarked ? <BookmarkCheck className="w-4 h-4 mr-1" /> : <Bookmark className="w-4 h-4 mr-1" />}
              {isBookmarked ? 'Bookmarked' : 'Bookmark'}
            </Button>
            {onClose && (
              <Button variant="ghost" size="sm" onClick={onClose}>
                Close
              </Button>
            )}
          </div>
        </div>
      </CardHeader>
      
      <CardContent className="space-y-6">
        <div>
          <h3 className="text-sm font-medium text-gray-500 mb-2">Description</h3>
          <p className="text-sm text-gray-700 leading-relaxed">{objectData.description}</p>
        </div>

        <div className="flex items-center justify-between border-t pt-6">
          <div className="flex space-x-6 text-sm text-gray-500">
            <div className="flex items-center">
              <FileText className="w-4 h-4 mr-1" />
              {notesCount} Notes
            </div>
            <div className="flex items-center">
              <Brain className="w-4 h-4 mr-1" />
              {cardsCount} Cards
            </div>
          </div>
          
          <div className="flex space-x-2">
            <Dialog open={showNoteDialog} onOpenChange={setShowNoteDialog}>
              <DialogTrigger asChild>
                <Button variant="outline" size="sm">
                  <FileText className="w-4 h-4 mr-1" />
                  Add Note
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Add Note</DialogTitle>
                  <DialogDescription>
                    Create a note for {objectData.title}
                  </DialogDescription>
                </DialogHeader>
                <div className="space-y-4 py-4">
                  <div>
                    <Label htmlFor="note-content">Content</Label>
                    <Textarea
                      id="note-content"
                      placeholder="Enter your note..."
                      value={noteContent}
                      onChange={(e) => setNoteContent(e.target.value)}
                      className="min-h-32"
                    />
                  </div>
                  <div className="flex justify-end space-x-2">
                    <Button variant="outline" onClick={() => setShowNoteDialog(false)}>
                      Cancel
                    </Button>
                    <Button onClick={handleAddNote}>
                      Save Note
                    </Button>
                  </div>
                </div>
              </DialogContent>
            </Dialog>

            <Dialog open={showCardDialog} onOpenChange={setShowCardDialog}>
              <DialogTrigger asChild>
                <Button variant="outline" size="sm">
                  <Brain className="w-4 h-4 mr-1" />
                  Create Card
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Create Memory Card</DialogTitle>
                  <DialogDescription>
                    Create a memory card for {objectData.title}
                  </DialogDescription>
                </DialogHeader>
                <div className="space-y-4 py-4">
                  <div>
                    <Label htmlFor="card-front">Front (Question)</Label>
                    <Input
                      id="card-front"
                      placeholder="Enter the question..."
                      value={cardFront}
                      onChange={(e) => setCardFront(e.target.value)}
                    />
                  </div>
                  <div>
                    <Label htmlFor="card-back">Back (Answer)</Label>
                    <Textarea
                      id="card-back"
                      placeholder="Enter the answer..."
                      value={cardBack}
                      onChange={(e) => setCardBack(e.target.value)}
                      className="min-h-32"
                    />
                  </div>
                  <div className="flex justify-end space-x-2">
                    <Button variant="outline" onClick={() => setShowCardDialog(false)}>
                      Cancel
                    </Button>
                    <Button onClick={handleCreateCard}>
                      Create Card
                    </Button>
                  </div>
                </div>
              </DialogContent>
            </Dialog>

            {!markAsLearned && isBookmarked && (
              <Button
                variant="default"
                size="sm"
                onClick={() => setMarkAsLearned(true)}
              >
                <CheckCircle className="w-4 h-4 mr-1" />
                Mark as Learned
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default LearningView;
