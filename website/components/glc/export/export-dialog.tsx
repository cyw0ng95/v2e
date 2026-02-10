'use client';

import { useState } from 'react';
import { Download, FileJson, Image, FileCode } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { useGLCStore } from '@/lib/glc/store';
import { downloadGraphJson, downloadCanvasPng, downloadCanvasSvg } from '@/lib/glc/io';

interface ExportDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  canvasRef?: React.RefObject<HTMLDivElement | null>;
}

export function ExportDialog({ open, onOpenChange, canvasRef }: ExportDialogProps) {
  const { graph, currentPreset } = useGLCStore();
  const [exporting, setExporting] = useState<string | null>(null);

  if (!graph || !currentPreset) return null;

  const filename = graph.metadata.name || 'graph';

  const handleExportJson = () => {
    setExporting('json');
    try {
      downloadGraphJson(graph);
    } finally {
      setExporting(null);
    }
  };

  const handleExportPng = async () => {
    if (!canvasRef?.current) return;
    setExporting('png');
    try {
      await downloadCanvasPng(canvasRef.current, `${filename}.png`, {
        backgroundColor: currentPreset.theme.background,
      });
    } finally {
      setExporting(null);
    }
  };

  const handleExportSvg = async () => {
    if (!canvasRef?.current) return;
    setExporting('svg');
    try {
      await downloadCanvasSvg(canvasRef.current, `${filename}.svg`, {
        backgroundColor: currentPreset.theme.background,
      });
    } finally {
      setExporting(null);
    }
  };

  const formats = [
    {
      id: 'json',
      label: 'JSON',
      description: 'Graph data file for importing later',
      icon: FileJson,
      handler: handleExportJson,
    },
    {
      id: 'png',
      label: 'PNG',
      description: 'High-resolution image (2x)',
      icon: Image,
      handler: handleExportPng,
      disabled: !canvasRef?.current,
    },
    {
      id: 'svg',
      label: 'SVG',
      description: 'Scalable vector graphics',
      icon: FileCode,
      handler: handleExportSvg,
      disabled: !canvasRef?.current,
    },
  ];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Download className="w-5 h-5" />
            Export Graph
          </DialogTitle>
        </DialogHeader>

        <div className="grid gap-3 mt-4">
          {formats.map((format) => (
            <Button
              key={format.id}
              variant="outline"
              className="justify-start h-auto py-3"
              onClick={format.handler}
              disabled={format.disabled || exporting === format.id}
            >
              <format.icon className="w-5 h-5 mr-3 flex-shrink-0" />
              <div className="text-left">
                <div className="font-medium">
                  {format.label}
                  {exporting === format.id && ' (exporting...)'}
                </div>
                <div className="text-xs text-muted-foreground font-normal">
                  {format.description}
                </div>
              </div>
            </Button>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}
