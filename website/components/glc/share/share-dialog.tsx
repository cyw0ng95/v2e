'use client';

import { useState } from 'react';
import { Copy, Check, Code, Link2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { useGLCStore } from '@/lib/glc/store';
import { generateShareUrl, generateEmbedCode, copyToClipboard } from '@/lib/glc/share';

interface ShareDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ShareDialog({ open, onOpenChange }: ShareDialogProps) {
  const { graph, currentPreset } = useGLCStore();
  const [copiedUrl, setCopiedUrl] = useState(false);
  const [copiedEmbed, setCopiedEmbed] = useState(false);

  if (!graph || !currentPreset) return null;

  const shareUrl = generateShareUrl(graph, currentPreset.meta.id);
  const embedCode = generateEmbedCode(graph, currentPreset.meta.id);

  const handleCopyUrl = async () => {
    const success = await copyToClipboard(shareUrl);
    if (success) {
      setCopiedUrl(true);
      setTimeout(() => setCopiedUrl(false), 2000);
    }
  };

  const handleCopyEmbed = async () => {
    const success = await copyToClipboard(embedCode);
    if (success) {
      setCopiedEmbed(true);
      setTimeout(() => setCopiedEmbed(false), 2000);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Link2 className="w-5 h-5" />
            Share Graph
          </DialogTitle>
        </DialogHeader>

        <Tabs defaultValue="link" className="mt-4">
          <TabsList className="w-full">
            <TabsTrigger value="link" className="flex-1">
              <Link2 className="w-4 h-4 mr-2" />
              Share Link
            </TabsTrigger>
            <TabsTrigger value="embed" className="flex-1">
              <Code className="w-4 h-4 mr-2" />
              Embed Code
            </TabsTrigger>
          </TabsList>

          <TabsContent value="link" className="mt-4 space-y-4">
            <div>
              <Label>Share URL</Label>
              <div className="flex gap-2 mt-1">
                <Input
                  value={shareUrl}
                  readOnly
                  className="flex-1 text-sm"
                />
                <Button onClick={handleCopyUrl} variant="outline">
                  {copiedUrl ? (
                    <Check className="w-4 h-4 text-green-500" />
                  ) : (
                    <Copy className="w-4 h-4" />
                  )}
                </Button>
              </div>
              <p className="text-xs text-muted-foreground mt-2">
                Share this link to let others view your graph. The graph data is encoded in the URL.
              </p>
            </div>
          </TabsContent>

          <TabsContent value="embed" className="mt-4 space-y-4">
            <div>
              <Label>Embed Code</Label>
              <div className="flex gap-2 mt-1">
                <Input
                  value={embedCode}
                  readOnly
                  className="flex-1 text-xs font-mono"
                />
                <Button onClick={handleCopyEmbed} variant="outline">
                  {copiedEmbed ? (
                    <Check className="w-4 h-4 text-green-500" />
                  ) : (
                    <Copy className="w-4 h-4" />
                  )}
                </Button>
              </div>
              <p className="text-xs text-muted-foreground mt-2">
                Embed this graph in any webpage using the iframe code above.
              </p>
            </div>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}
