'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  FileText,
  Save,
  FolderOpen,
  Download,
  Share2,
  FolderTree,
  LayoutTemplate,
} from 'lucide-react';
import ExampleDialog from './glc-examples/example-dialog';

interface FileMenuProps {
  preset?: string;
  onSave?: () => void;
  onOpen?: () => void;
  onExport?: (format: string) => void;
  onShare?: () => void;
}

export default function FileMenu({
  preset,
  onSave,
  onOpen,
  onExport,
  onShare,
}: FileMenuProps) {
  const [exampleDialogOpen, setExampleDialogOpen] = useState(false);

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size="sm">
            <FileText className="mr-2 h-4 w-4" />
            File
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start">
          <DropdownMenuLabel>File Actions</DropdownMenuLabel>
          <DropdownMenuSeparator />

          <DropdownMenuItem onClick={onSave}>
            <Save className="mr-2 h-4 w-4" />
            Save Graph
          </DropdownMenuItem>

          <DropdownMenuItem onClick={onOpen}>
            <FolderOpen className="mr-2 h-4 w-4" />
            Open Graph
          </DropdownMenuItem>

          <DropdownMenuSeparator />

          <DropdownMenuItem
            onClick={() => setExampleDialogOpen(true)}
          >
            <FolderTree className="mr-2 h-4 w-4" />
            Load Example
          </DropdownMenuItem>

          <DropdownMenuSeparator />

          <DropdownMenuItem onClick={() => onExport?.('png')}>
            <Download className="mr-2 h-4 w-4" />
            Export as PNG
          </DropdownMenuItem>

          <DropdownMenuItem onClick={() => onExport?.('svg')}>
            <Download className="mr-2 h-4 w-4" />
            Export as SVG
          </DropdownMenuItem>

          <DropdownMenuItem onClick={() => onExport?.('pdf')}>
            <Download className="mr-2 h-4 w-4" />
            Export as PDF
          </DropdownMenuItem>

          <DropdownMenuItem onClick={() => onExport?.('json')}>
            <Download className="mr-2 h-4 w-4" />
            Export as JSON
          </DropdownMenuItem>

          <DropdownMenuSeparator />

          <DropdownMenuItem onClick={onShare}>
            <Share2 className="mr-2 h-4 w-4" />
            Share Graph
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <ExampleDialog
        open={exampleDialogOpen}
        onOpenChange={setExampleDialogOpen}
        preset={preset}
      />
    </>
  );
}
