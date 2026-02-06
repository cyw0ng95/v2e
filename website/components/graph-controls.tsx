'use client';

import React, { useState, useCallback } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { useGraphControl, useGraphStats } from '@/lib/hooks';
import { Database, Trash2, Save, Upload, Play, Pause, AlertCircle } from 'lucide-react';
import { Skeleton } from '@/components/ui/skeleton';

interface GraphControlPanelProps {
  onFilterChange?: (types: string[]) => void;
  onSearchChange?: (query: string) => void;
}

export default function GraphControlPanel({ onFilterChange, onSearchChange }: GraphControlPanelProps) {
  const [buildLimit, setBuildLimit] = useState<number>(500);
  const [selectedTypes, setSelectedTypes] = useState<string[]>(['cve', 'cwe', 'capec', 'attack', 'ssg']);
  const [searchQuery, setSearchQuery] = useState<string>('');

  const { data: stats, isLoading: statsLoading, refetch: refetchStats } = useGraphStats();
  const { buildGraph, clearGraph, saveGraph, loadGraph, isLoading, error, buildResult } = useGraphControl();

  const handleBuildGraph = useCallback(async () => {
    await buildGraph(buildLimit);
    refetchStats();
  }, [buildLimit, buildGraph, refetchStats]);

  const handleClearGraph = useCallback(async () => {
    await clearGraph();
    refetchStats();
  }, [clearGraph, refetchStats]);

  const handleSaveGraph = useCallback(async () => {
    await saveGraph();
  }, [saveGraph]);

  const handleLoadGraph = useCallback(async () => {
    await loadGraph();
    refetchStats();
  }, [loadGraph, refetchStats]);

  const handleTypeToggle = useCallback((type: string, checked: boolean) => {
    const newTypes = checked
      ? [...selectedTypes, type]
      : selectedTypes.filter(t => t !== type);
    setSelectedTypes(newTypes);
    onFilterChange?.(newTypes);
  }, [selectedTypes, onFilterChange]);

  const handleSearchChange = useCallback((query: string) => {
    setSearchQuery(query);
    onSearchChange?.(query);
  }, [onSearchChange]);

  const nodeTypes = [
    { id: 'cve', label: 'CVE', color: '#EF4444' },
    { id: 'cwe', label: 'CWE', color: '#F97316' },
    { id: 'capec', label: 'CAPEC', color: '#EAB308' },
    { id: 'attack', label: 'ATT&CK', color: '#3B82F6' },
    { id: 'ssg', label: 'SSG', color: '#22C55E' },
  ];

  return (
    <div className="space-y-4">
      {/* Graph Statistics */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Database className="w-4 h-4" />
            Graph Statistics
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {statsLoading ? (
            <div className="space-y-2">
              <Skeleton className="h-8 w-full" />
              <Skeleton className="h-8 w-full" />
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-3">
              <div>
                <div className="text-2xl font-bold">{stats?.node_count?.toLocaleString() || 0}</div>
                <div className="text-xs text-muted-foreground">Nodes</div>
              </div>
              <div>
                <div className="text-2xl font-bold">{stats?.edge_count?.toLocaleString() || 0}</div>
                <div className="text-xs text-muted-foreground">Edges</div>
              </div>
            </div>
          )}

          {error && (
            <div className="flex items-start gap-2 text-sm text-destructive bg-destructive/10 p-2 rounded">
              <AlertCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {buildResult && (
            <div className="text-xs text-muted-foreground space-y-1">
              <div>Added: {buildResult.nodes_added} nodes, {buildResult.edges_added} edges</div>
              <div>Total: {buildResult.total_nodes} nodes, {buildResult.total_edges} edges</div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Graph Control */}
      <Card>
        <CardHeader>
          <CardTitle>Graph Control</CardTitle>
          <CardDescription>Build, save, load, or clear the graph</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="space-y-2">
            <Label htmlFor="build-limit">Build Limit (CVEs)</Label>
            <Input
              id="build-limit"
              type="number"
              value={buildLimit}
              onChange={(e) => setBuildLimit(parseInt(e.target.value) || 500)}
              min={10}
              max={10000}
              step={100}
            />
          </div>

          <div className="grid grid-cols-2 gap-2">
            <Button
              onClick={handleBuildGraph}
              disabled={isLoading}
              className="w-full"
            >
              {isLoading ? (
                <>
                  <Play className="w-4 h-4 mr-2 animate-spin" />
                  Building...
                </>
              ) : (
                <>
                  <Play className="w-4 h-4 mr-2" />
                  Build Graph
                </>
              )}
            </Button>

            <Button
              onClick={handleClearGraph}
              disabled={isLoading}
              variant="outline"
              className="w-full"
            >
              <Trash2 className="w-4 h-4 mr-2" />
              Clear
            </Button>

            <Button
              onClick={handleSaveGraph}
              disabled={isLoading}
              variant="outline"
              className="w-full"
            >
              <Save className="w-4 h-4 mr-2" />
              Save
            </Button>

            <Button
              onClick={handleLoadGraph}
              disabled={isLoading}
              variant="outline"
              className="w-full"
            >
              <Upload className="w-4 h-4 mr-2" />
              Load
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Node Type Filter */}
      <Card>
        <CardHeader>
          <CardTitle>Filter by Type</CardTitle>
          <CardDescription>Show or hide node types</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="space-y-2">
            {nodeTypes.map((type) => (
              <div key={type.id} className="flex items-center space-x-2">
                <Checkbox
                  id={`type-${type.id}`}
                  checked={selectedTypes.includes(type.id)}
                  onCheckedChange={(checked) => handleTypeToggle(type.id, checked as boolean)}
                />
                <Label
                  htmlFor={`type-${type.id}`}
                  className="flex items-center gap-2 cursor-pointer flex-1"
                >
                  <Badge
                    style={{
                      backgroundColor: type.color,
                      color: '#FFFFFF',
                      border: 'none',
                    }}
                  >
                    {type.label}
                  </Badge>
                </Label>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Search */}
      <Card>
        <CardHeader>
          <CardTitle>Search</CardTitle>
          <CardDescription>Search nodes by URN or name</CardDescription>
        </CardHeader>
        <CardContent>
          <Input
            placeholder="Search nodes..."
            value={searchQuery}
            onChange={(e) => handleSearchChange(e.target.value)}
          />
        </CardContent>
      </Card>
    </div>
  );
}
