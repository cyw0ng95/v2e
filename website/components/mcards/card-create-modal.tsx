'use client';

import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { RichTextEditor } from './rich-text-editor';
import { useCreateCard } from '@/lib/mcards/hooks';

interface CardCreateModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  bookmarks?: Array<{ id: number; title: string }>;
}

export function CardCreateModal({ open, onOpenChange, bookmarks = [] }: CardCreateModalProps) {
  const [frontContent, setFrontContent] = useState('');
  const [backContent, setBackContent] = useState('');
  const [majorClass, setMajorClass] = useState('CWE');
  const [minorClass, setMinorClass] = useState('');
  const [isSaving, setIsSaving] = useState(false);

  const createCard = useCreateCard();

  const handleSubmit = async () => {
    if (!frontContent.trim() || !backContent.trim()) {
      return;
    }

    setIsSaving(true);
    try {
      await createCard.mutateAsync({
        front_content: frontContent,
        back_content: backContent,
        major_class: majorClass,
        minor_class: minorClass || majorClass,
        status: 'active',
        content: '{}',
        card_type: 'basic',
      });

      // Reset and close
      setFrontContent('');
      setBackContent('');
      setMajorClass('CWE');
      setMinorClass('');
      onOpenChange(false);
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create Memory Card</DialogTitle>
          <DialogDescription>
            Create a new memory card for spaced repetition learning.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* Front/Back Editor */}
          <div className="grid grid-cols-2 gap-6">
            <div>
              <Label htmlFor="front">Front Side (Question)</Label>
              <RichTextEditor
                content={frontContent}
                onChange={setFrontContent}
                placeholder="Enter your question here..."
              />
            </div>
            <div>
              <Label htmlFor="back">Back Side (Answer)</Label>
              <RichTextEditor
                content={backContent}
                onChange={setBackContent}
                placeholder="Enter the answer here..."
              />
            </div>
          </div>

          {/* Metadata */}
          <div className="grid grid-cols-3 gap-4">
            <div>
              <Label htmlFor="class">Card Class</Label>
              <Input
                id="class"
                value={majorClass}
                onChange={(e) => setMajorClass(e.target.value)}
                placeholder="CWE"
                required
              />
            </div>
            <div>
              <Label htmlFor="minor">Minor Class</Label>
              <Input
                id="minor"
                value={minorClass}
                onChange={(e) => setMinorClass(e.target.value)}
                placeholder="Injection"
              />
            </div>
            <div>
              <Label htmlFor="bookmark">Bookmark (Optional)</Label>
              <select
                id="bookmark"
                className="w-full px-3 py-2 border border-slate-200 dark:border-slate-700 rounded-lg"
              >
                <option value="">No bookmark</option>
                {bookmarks.map((bm) => (
                  <option key={bm.id} value={bm.id}>
                    {bm.title}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={isSaving || !frontContent.trim() || !backContent.trim()}
            className="bg-indigo-500 hover:bg-indigo-600"
          >
            {isSaving ? 'Creating...' : 'Create Card'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
