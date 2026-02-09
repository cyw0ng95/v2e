'use client';

import { useGLCStore } from '@/lib/glc/store';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { CheckCircle2, Circle } from 'lucide-react';

interface PhaseProgress {
  sprint: string;
  title: string;
  tasks: { id: string; name: string; completed: boolean }[];
}

export function PhaseProgress({ phase = 1 }: { phase?: number }) {
  const { currentPreset } = useGLCStore();

  const phases: Record<number, PhaseProgress> = {
    1: {
      sprint: 'Sprint 1 (Weeks 1-2)',
      title: 'Project Initialization & Setup',
      tasks: [
        { id: '1.1', name: 'Project Initialization', completed: true },
        { id: '1.2', name: 'Core Dependencies Installation', completed: true },
        { id: '1.3', name: 'shadcn/ui Component Setup', completed: true },
        { id: '1.4', name: 'Basic Layout Structure', completed: true },
      ],
    },
    2: {
      sprint: 'Sprint 2 (Weeks 3-4)',
      title: 'State Management & Data Models',
      tasks: [
        { id: '1.5', name: 'Centralized State Management', completed: true },
        { id: '1.6', name: 'Complete TypeScript Type System', completed: true },
        { id: '1.7', name: 'Built-in Presets Implementation', completed: true },
        { id: '1.8', name: 'Landing Page & Navigation', completed: true },
      ],
    },
    3: {
      sprint: 'Sprint 3 (Weeks 5-6)',
      title: 'Preset System & Validation',
      tasks: [
        { id: '1.9', name: 'Preset Validation System', completed: false },
        { id: '1.10', name: 'Error Handling & Recovery', completed: false },
        { id: '1.11', name: 'Preset Management System', completed: false },
      ],
    },
    4: {
      sprint: 'Sprint 4 (Weeks 7-8)',
      title: 'Testing & Integration',
      tasks: [
        { id: '1.12', name: 'Unit Testing', completed: false },
        { id: '1.13', name: 'Integration Testing & Documentation', completed: false },
      ],
    },
  };

  const currentPhase = phases[phase];
  const completedTasks = currentPhase.tasks.filter(t => t.completed).length;
  const totalTasks = currentPhase.tasks.length;
  const progress = Math.round((completedTasks / totalTasks) * 100);

  return (
    <Card className="bg-slate-800 border-slate-700">
      <CardHeader>
        <CardTitle className="text-white flex items-center justify-between">
          <span>Phase 1 Progress</span>
          <span className="text-blue-400">{progress}%</span>
        </CardTitle>
        <CardDescription className="text-slate-400">{currentPhase.sprint}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-3 mb-4">
          {currentPhase.tasks.map((task) => (
            <div key={task.id} className="flex items-center space-x-3">
              {task.completed ? (
                <CheckCircle2 className="w-5 h-5 text-green-500 flex-shrink-0" />
              ) : (
                <Circle className="w-5 h-5 text-slate-600 flex-shrink-0" />
              )}
              <span className={`text-sm ${task.completed ? 'text-slate-300' : 'text-slate-500'}`}>
                {task.id} - {task.name}
              </span>
            </div>
          ))}
        </div>
        
        <div className="mt-6 pt-6 border-t border-slate-700">
          <p className="text-sm text-slate-400 mb-2">Current Preset:</p>
          <p className="text-white font-medium">
            {currentPreset ? currentPreset.name : 'None selected'}
          </p>
        </div>
      </CardContent>
    </Card>
  );
}

export default PhaseProgress;
