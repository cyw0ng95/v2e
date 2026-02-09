'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useGLCStore } from '@/lib/glc/store';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Shield, Box, ArrowRight, FolderOpen } from 'lucide-react';

export default function GLCLandingPage() {
  const router = useRouter();
  const store = useGLCStore();
  const {
    builtInPresets = [],
    setCurrentPreset,
    getAllPresets,
  } = store as any;

  useEffect(() => {
    getAllPresets();
  }, [getAllPresets]);

  const handleSelectPreset = (presetId: string) => {
    const preset = builtInPresets.find((p: any) => p.id === presetId);
    if (preset) {
      setCurrentPreset(preset);
      router.push(`/glc/${presetId}`);
    }
  };

  const presetCards = [
    {
      id: 'd3fend',
      icon: Shield,
      title: 'D3FEND Canvas',
      description: 'MITRE D3FEND ontology for modeling cyber attack and defense strategies',
      category: 'Security',
      color: 'from-red-500 to-orange-500',
    },
    {
      id: 'topo-graph',
      icon: Box,
      title: 'Topology Graph',
      description: 'General-purpose topology and graph diagramming',
      category: 'General',
      color: 'from-blue-500 to-cyan-500',
    },
  ];

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900">
      <div className="container mx-auto px-4 py-16">
        <div className="text-center mb-16">
          <h1 className="text-6xl font-bold text-white mb-6">
            GLC
          </h1>
          <p className="text-xl text-slate-300 max-w-2xl mx-auto">
            Graphized Learning Canvas - A modern, interactive graph-based modeling platform
          </p>
        </div>

        <div className="grid md:grid-cols-2 gap-8 max-w-5xl mx-auto mb-16">
          {presetCards.map((card: any) => {
            const Icon = card.icon;
            return (
              <Card key={card.id} className="bg-slate-800 border-slate-700 hover:border-blue-500 transition-all cursor-pointer group">
                <CardHeader>
                  <div className={`w-16 h-16 rounded-lg bg-gradient-to-br ${card.color} flex items-center justify-center mb-4 group-hover:scale-110 transition-transform`}>
                    <Icon className="w-8 h-8 text-white" />
                  </div>
                  <CardTitle className="text-white text-2xl">{card.title}</CardTitle>
                  <CardDescription className="text-slate-400">{card.description}</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center space-x-2">
                    <span className="px-3 py-1 bg-slate-700 text-slate-300 rounded-full text-sm">
                      {card.category}
                    </span>
                  </div>
                </CardContent>
                <CardFooter>
                  <Button 
                    onClick={() => handleSelectPreset(card.id)}
                    className="w-full bg-blue-600 hover:bg-blue-700 text-white"
                  >
                    Open Canvas
                    <ArrowRight className="ml-2 h-4 w-4" />
                  </Button>
                </CardFooter>
              </Card>
            );
          })}
        </div>

        <div className="text-center">
          <Button 
            variant="outline" 
            className="border-slate-600 text-slate-300 hover:bg-slate-700"
            onClick={() => router.push('/')}
          >
            <FolderOpen className="mr-2 h-4 w-4" />
            Browse Recent Graphs
          </Button>
        </div>
      </div>
    </div>
  );
}
