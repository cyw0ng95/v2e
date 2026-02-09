'use client';

import { useState } from 'react';
import { useGLCStore } from '@/lib/glc/store';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { NodeTypeDefinition } from '@/lib/glc/types';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { ScrollArea } from '@/components/ui/scroll-area';
import { getNodeStyle } from '@/lib/glc/canvas/canvas-config';
import { iconMap } from './dynamic-node';
import * as Icons from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Search, GripHorizontal } from 'lucide-react';

interface NodePaletteProps {
  isOpen: boolean;
  onToggle: () => void;
}

export function NodePalette({ isOpen, onToggle }: NodePaletteProps) {
  const { currentPreset, addNode, nodes } = useGLCStore() as any;
  const [searchQuery, setSearchQuery] = useState('');
  const [expandedCategories, setExpandedCategories] = useState<Set<string>>(new Set());

  if (!currentPreset) {
    return null;
  }

  const filteredNodeTypes = currentPreset.nodeTypes.filter((nodeType: any) => {
    const searchLower = searchQuery.toLowerCase();
    return (
      nodeType.name.toLowerCase().includes(searchLower) ||
      nodeType.category.toLowerCase().includes(searchLower) ||
      nodeType.description.toLowerCase().includes(searchLower)
    );
  });

  const groupedByCategory = filteredNodeTypes.reduce((acc: any, nodeType: any) => {
    if (!acc[nodeType.category]) {
      acc[nodeType.category] = [];
    }
    acc[nodeType.category].push(nodeType);
    return acc;
  }, {} as Record<string, NodeTypeDefinition[]>);

  const categories = Object.keys(groupedByCategory);

  const toggleCategory = (category: string) => {
    setExpandedCategories(prev => {
      const newSet = new Set(prev);
      if (newSet.has(category)) {
        newSet.delete(category);
      } else {
        newSet.add(category);
      }
      return newSet;
    });
  };

  const handleDragStart = (event: React.DragEvent, nodeType: NodeTypeDefinition) => {
    event.dataTransfer.setData('nodeType', nodeType.id);
    event.dataTransfer.setData('nodeTypeData', JSON.stringify(nodeType));
    event.dataTransfer.effectAllowed = 'copy';
  };

  const handleDragEnd = (event: any) => {
    event.dataTransfer.clearData();
  };

  return (
    <Card className={`w-[300px] h-full border-l-0 border-r transition-all duration-300 ${
      isOpen ? 'translate-x-0' : '-translate-x-full'
    }`}>
      <CardHeader className="border-b">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg">Node Palette</CardTitle>
          <button
            onClick={onToggle}
            className="p-2 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-md"
          >
            <Icons.PanelLeftOpen className="h-5 w-5" />
          </button>
        </div>
        <div className="relative mt-3">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            type="text"
            placeholder="Search node types..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
      </CardHeader>

      <CardContent className="p-0">
        <ScrollArea className="h-[calc(100vh-140px)]">
          <div className="p-4 space-y-2">
            {categories.map(category => (
              <Accordion
                key={category}
                className="border rounded-lg"
                value={expandedCategories.has(category) ? category : undefined}
              >
                <AccordionTrigger
                  onClick={() => toggleCategory(category)}
                  className="px-4 py-3 hover:no-underline data-[state=open]:bg-muted/50"
                >
                  <div className="flex items-center justify-between w-full">
                    <div className="flex items-center gap-2">
                      <div className="w-2 h-2 rounded-full bg-primary" />
                      <span className="font-medium">{category}</span>
                      <span className="text-sm text-muted-foreground">
                        ({groupedByCategory[category].length})
                      </span>
                    </div>
                    <Icons.ChevronDown className="h-4 w-4 transition-transform duration-200" />
                  </div>
                </AccordionTrigger>
                <AccordionContent className="px-4 pt-2 pb-4">
                  <div className="space-y-2">
                    {groupedByCategory[category].map((nodeType: any) => {
                      const Icon = (iconMap as any)[nodeType.style.icon] || Icons.Box;
                      const style = getNodeStyle(currentPreset, nodeType.id);

                      return (
                        <div
                          key={nodeType.id}
                          draggable
                          onDragStart={(e) => handleDragStart(e, nodeType)}
                          onDragEnd={handleDragEnd}
                          className="group flex items-start gap-3 p-3 rounded-lg border-2 border-transparent hover:border-blue-300 dark:hover:border-blue-700 hover:bg-slate-50 dark:hover:bg-slate-800 cursor-grab active:cursor-grabbing transition-all"
                          style={{
                            background: `${style.backgroundColor}20`,
                          }}
                        >
                          <div
                            className="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0"
                            style={{
                              backgroundColor: style.backgroundColor,
                              color: style.color,
                            }}
                          >
                            <Icon className="w-5 h-5" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <div
                              className="font-medium text-sm"
                              style={{ color: style.color }}
                            >
                              {nodeType.name}
                            </div>
                            <div className="text-xs text-muted-foreground truncate">
                              {nodeType.description}
                            </div>
                            <div className="flex items-center gap-2 mt-1">
                              {nodeType.properties.length > 0 && (
                                <span className="text-xs text-muted-foreground">
                                  {nodeType.properties.length} properties
                                </span>
                              )}
                            </div>
                          </div>
                          <GripHorizontal className="w-4 h-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
                        </div>
                      );
                    })}
                  </div>
                </AccordionContent>
              </Accordion>
            ))}

            {filteredNodeTypes.length === 0 && (
              <div className="text-center py-8 text-muted-foreground">
                <Search className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>No node types found</p>
                <p className="text-sm mt-1">Try a different search term</p>
              </div>
            )}
          </div>
        </ScrollArea>
      </CardContent>
    </Card>
  );
}

export default NodePalette;
