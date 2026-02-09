'use client';

import Link from 'next/link';
import { useGLCStore } from '@/lib/glc/store';
import { Network, GitBranch, Shield, ArrowRight, Plus, FolderOpen } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

export default function GLCLandingPage() {
  const builtInPresets = useGLCStore((state) => state.builtInPresets);
  const userPresets = useGLCStore((state) => state.userPresets);

  return (
    <div className="min-h-screen bg-background">
      {/* Hero Section */}
      <div className="border-b border-border bg-surface/50">
        <div className="container mx-auto px-6 py-16">
          <div className="flex items-center gap-3 mb-4">
            <div className="p-3 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600">
              <Network className="w-8 h-8 text-white" />
            </div>
            <h1 className="text-4xl font-bold text-text">
              Graphized Learning Canvas
            </h1>
          </div>
          <p className="text-xl text-textMuted max-w-2xl mb-8">
            Create interactive graph-based models for cyber attack/defense analysis,
            network topology visualization, and more.
          </p>
          <div className="flex gap-4">
            <Button asChild size="lg">
              <Link href="/glc/d3fend">
                <Shield className="w-5 h-5 mr-2" />
                D3FEND Canvas
              </Link>
            </Button>
            <Button asChild variant="outline" size="lg">
              <Link href="/glc/topo">
                <GitBranch className="w-5 h-5 mr-2" />
                Topo-Graph
              </Link>
            </Button>
          </div>
        </div>
      </div>

      {/* Presets Section */}
      <div className="container mx-auto px-6 py-12">
        <h2 className="text-2xl font-semibold text-text mb-6">Choose a Canvas Preset</h2>

        {/* Built-in Presets */}
        <div className="grid md:grid-cols-2 gap-6 mb-12">
          {builtInPresets.map((preset) => (
            <Card key={preset.meta.id} className="hover:border-accent/50 transition-colors">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="flex items-center gap-2">
                    {preset.meta.id === 'd3fend' ? (
                      <Shield className="w-5 h-5 text-success" />
                    ) : (
                      <GitBranch className="w-5 h-5 text-primary" />
                    )}
                    {preset.meta.name}
                  </CardTitle>
                  <Badge variant="secondary">v{preset.meta.version}</Badge>
                </div>
                <CardDescription>{preset.meta.description}</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex gap-4 text-sm text-textMuted">
                  <span>{preset.nodeTypes.length} node types</span>
                  <span>{preset.relationships.length} relationships</span>
                </div>
              </CardContent>
              <CardFooter>
                <Button asChild className="w-full">
                  <Link href={`/glc/${preset.meta.id}`}>
                    Open Canvas
                    <ArrowRight className="w-4 h-4 ml-2" />
                  </Link>
                </Button>
              </CardFooter>
            </Card>
          ))}
        </div>

        {/* User Presets */}
        {userPresets.length > 0 && (
          <>
            <h3 className="text-lg font-medium text-text mb-4">Your Presets</h3>
            <div className="grid md:grid-cols-3 gap-4">
              {userPresets.map((preset) => (
                <Card key={preset.meta.id}>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-base">{preset.meta.name}</CardTitle>
                    <CardDescription className="text-xs">
                      {preset.meta.description}
                    </CardDescription>
                  </CardHeader>
                  <CardFooter className="pt-2">
                    <Button asChild variant="outline" size="sm" className="w-full">
                      <Link href={`/glc/${preset.meta.id}`}>
                        Open
                      </Link>
                    </Button>
                  </CardFooter>
                </Card>
              ))}
            </div>
          </>
        )}

        {/* Quick Actions */}
        <div className="mt-12 flex gap-4">
          <Button variant="outline" asChild>
            <Link href="/glc/d3fend">
              <Plus className="w-4 h-4 mr-2" />
              New Graph
            </Link>
          </Button>
          <Button variant="ghost" disabled>
            <FolderOpen className="w-4 h-4 mr-2" />
            My Graphs (coming soon)
          </Button>
        </div>
      </div>
    </div>
  );
}
