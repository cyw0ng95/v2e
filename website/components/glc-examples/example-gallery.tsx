'use client';

import { useState, useEffect } from 'react';
import { ExampleGraph } from '@/lib/glc/lib/examples/example-types';
import { loadExamples, getExamplesByPreset, searchExamples, getCategories } from '@/lib/glc/lib/examples/examples-loader';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog';
import { Search, Grid, List, ExternalLink } from 'lucide-react';
import ExampleCard from './example-card';

interface ExampleGalleryProps {
  preset?: string;
  onOpenExample?: (example: ExampleGraph) => void;
  onClose?: () => void;
}

export default function ExampleGallery({ preset, onOpenExample, onClose }: ExampleGalleryProps) {
  const [examples, setExamples] = useState<ExampleGraph[]>([]);
  const [filteredExamples, setFilteredExamples] = useState<ExampleGraph[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const [categories, setCategories] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [selectedExample, setSelectedExample] = useState<ExampleGraph | null>(null);

  useEffect(() => {
    async function load() {
      try {
        setLoading(true);
        const allExamples = await loadExamples();
        const presetExamples = preset ? await getExamplesByPreset(preset) : allExamples;
        const cats = await getCategories();
        setExamples(presetExamples);
        setFilteredExamples(presetExamples);
        setCategories(cats);
      } catch (error) {
        console.error('Error loading examples:', error);
      } finally {
        setLoading(false);
      }
    }

    load();
  }, [preset]);

  useEffect(() => {
    let filtered = examples;

    if (searchQuery) {
      filtered = filtered.filter(example =>
        example.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        example.description.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }

    if (selectedCategory) {
      filtered = filtered.filter(example => example.category === selectedCategory);
    }

    setFilteredExamples(filtered);
  }, [searchQuery, selectedCategory, examples]);

  function handleOpenExample(example: ExampleGraph) {
    setSelectedExample(example);
  }

  function handleLoadExample() {
    if (selectedExample && onOpenExample) {
      onOpenExample(selectedExample);
      handleClose();
    }
  }

  function handleClose() {
    if (onClose) {
      onClose();
    }
    setSelectedExample(null);
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Example Graphs</h2>
          <p className="text-muted-foreground">
            Browse and load example graphs to get started
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant={viewMode === 'grid' ? 'default' : 'outline'}
            size="icon"
            onClick={() => setViewMode('grid')}
          >
            <Grid className="h-4 w-4" />
          </Button>
          <Button
            variant={viewMode === 'list' ? 'default' : 'outline'}
            size="icon"
            onClick={() => setViewMode('list')}
          >
            <List className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <div className="space-y-4">
        <div className="flex gap-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search examples..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
        </div>

        <div className="flex gap-2 flex-wrap">
          <Badge
            variant={selectedCategory === null ? 'default' : 'outline'}
            className="cursor-pointer"
            onClick={() => setSelectedCategory(null)}
          >
            All Categories
          </Badge>
          {categories.map((category) => (
            <Badge
              key={category}
              variant={selectedCategory === category ? 'default' : 'outline'}
              className="cursor-pointer"
              onClick={() => setSelectedCategory(category)}
            >
              {category}
            </Badge>
          ))}
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center h-64">
          <div className="text-muted-foreground">Loading examples...</div>
        </div>
      ) : filteredExamples.length === 0 ? (
        <div className="flex items-center justify-center h-64">
          <div className="text-muted-foreground">No examples found</div>
        </div>
      ) : (
        <div
          className={
            viewMode === 'grid'
              ? 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4'
              : 'space-y-4'
          }
        >
          {filteredExamples.map((example) => (
            <ExampleCard
              key={example.id}
              example={example}
              viewMode={viewMode}
              onClick={() => handleOpenExample(example)}
            />
          ))}
        </div>
      )}

      <Dialog open={selectedExample !== null} onOpenChange={handleClose}>
        {selectedExample && (
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                {selectedExample.name}
                <Badge variant="outline">{selectedExample.metadata.complexity}</Badge>
              </DialogTitle>
              <DialogDescription>{selectedExample.description}</DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <h4 className="text-sm font-medium mb-2">Statistics</h4>
                  <div className="space-y-2 text-sm text-muted-foreground">
                    <p>Nodes: {selectedExample.metadata.nodeCount}</p>
                    <p>Edges: {selectedExample.metadata.edgeCount}</p>
                    <p>Preset: {selectedExample.preset.toUpperCase()}</p>
                  </div>
                </div>
                <div>
                  <h4 className="text-sm font-medium mb-2">Category</h4>
                  <Badge variant="secondary">{selectedExample.category}</Badge>
                </div>
              </div>

              <div className="flex gap-2 justify-end">
                <Button variant="outline" onClick={handleClose}>
                  Cancel
                </Button>
                <Button onClick={handleLoadExample}>
                  <ExternalLink className="h-4 w-4 mr-2" />
                  Load Example
                </Button>
              </div>
            </div>
          </DialogContent>
        )}
      </Dialog>
    </div>
  );
}
