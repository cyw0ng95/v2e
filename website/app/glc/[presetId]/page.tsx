'use client';

import { useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useGLCStore } from '@/lib/glc/store';
import { Button } from '@/components/ui/button';
import { ArrowLeft, Loader2 } from 'lucide-react';

export default function CanvasPage() {
  const params = useParams();
  const router = useRouter();
  const { currentPreset, setCurrentPreset, getPresetById } = useGLCStore();

  useEffect(() => {
    const presetId = params.presetId as string;
    const preset = getPresetById(presetId);
    
    if (!preset) {
      router.push('/glc');
      return;
    }
    
    if (!currentPreset || currentPreset.id !== presetId) {
      setCurrentPreset(preset);
    }
  }, [params.presetId, currentPreset, getPresetById, setCurrentPreset, router]);

  if (!currentPreset) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-900">
        <div className="text-center">
          <Loader2 className="w-12 h-12 text-blue-500 animate-spin mx-auto mb-4" />
          <p className="text-white">Loading preset...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-900">
      <div className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <Button
            variant="ghost"
            onClick={() => router.push('/glc')}
            className="text-slate-300 hover:text-white hover:bg-slate-800"
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Presets
          </Button>
        </div>

        <div className="mb-8">
          <h1 className="text-4xl font-bold text-white mb-2">{currentPreset.name}</h1>
          <p className="text-slate-400">{currentPreset.description}</p>
        </div>

        <div className="bg-slate-800 border border-slate-700 rounded-lg p-8 text-center">
          <h2 className="text-2xl font-semibold text-white mb-4">Canvas Coming Soon</h2>
          <p className="text-slate-400 mb-6">
            The interactive canvas is under development. Check back soon for updates!
          </p>
          <div className="inline-flex items-center text-sm text-slate-500">
            <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            Phase 2 implementation in progress
          </div>
        </div>
      </div>
    </div>
  );
}
