'use client';

import React, { useState, useCallback } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { 
  Zap, 
  PlayCircle, 
  Pause, 
  StopCircle, 
  Settings,
  AlertCircle,
  CheckCircle,
  Clock,
  XCircle,
  Activity
} from 'lucide-react';

interface ProviderNode {
  id: string;
  providerType: string;
  state: string;
  processedCount: number;
  errorCount: number;
  permitsHeld: number;
  lastCheckpoint?: string;
}

interface MacroState {
  state: string;
  providers: ProviderNode[];
}

interface TopologyData {
  macro: MacroState;
  totalProviders: number;
  activeProviders: number;
}

interface ETLTopologyViewerProps {
  data?: TopologyData | null;
  isLoading?: boolean;
  onProviderAction?: (providerId: string, action: 'start' | 'pause' | 'stop') => void;
  onMacroAction?: (action: 'start' | 'stop' | 'pause' | 'resume') => void;
  onPolicyUpdate?: (providerId: string, policy: PerformancePolicy) => void;
}

interface PerformancePolicy {
  maxPermits?: number;
  batchSize?: number;
  backoffMs?: number;
  retryLimit?: number;
}

const getStateIcon = (state: string) => {
  switch (state) {
    case 'RUNNING':
    case 'ORCHESTRATING':
      return <Activity className="h-4 w-4 text-green-500" />;
    case 'IDLE':
    case 'ACQUIRING':
      return <Clock className="h-4 w-4 text-blue-500" />;
    case 'PAUSED':
    case 'WAITING_QUOTA':
    case 'WAITING_BACKOFF':
      return <Pause className="h-4 w-4 text-yellow-500" />;
    case 'TERMINATED':
    case 'DRAINING':
      return <XCircle className="h-4 w-4 text-red-500" />;
    default:
      return <AlertCircle className="h-4 w-4 text-gray-500" />;
  }
};

const getStateColor = (state: string): string => {
  switch (state) {
    case 'RUNNING':
    case 'ORCHESTRATING':
      return 'bg-green-100 border-green-400';
    case 'IDLE':
    case 'ACQUIRING':
      return 'bg-blue-100 border-blue-400';
    case 'PAUSED':
    case 'WAITING_QUOTA':
    case 'WAITING_BACKOFF':
      return 'bg-yellow-100 border-yellow-400';
    case 'TERMINATED':
    case 'DRAINING':
      return 'bg-red-100 border-red-400';
    default:
      return 'bg-gray-100 border-gray-400';
  }
};

const getStateBadgeVariant = (state: string): "default" | "secondary" | "destructive" | "outline" => {
  switch (state) {
    case 'RUNNING':
    case 'ORCHESTRATING':
      return 'default';
    case 'PAUSED':
    case 'WAITING_QUOTA':
    case 'WAITING_BACKOFF':
      return 'secondary';
    case 'TERMINATED':
    case 'DRAINING':
      return 'destructive';
    default:
      return 'outline';
  }
};

export function ETLTopologyViewer({ 
  data, 
  isLoading = false,
  onProviderAction,
  onMacroAction,
  onPolicyUpdate
}: ETLTopologyViewerProps) {
  const [selectedProvider, setSelectedProvider] = useState<ProviderNode | null>(null);
  const [policyDialogOpen, setPolicyDialogOpen] = useState(false);
  const [policyForm, setPolicyForm] = useState<PerformancePolicy>({
    maxPermits: 10,
    batchSize: 100,
    backoffMs: 1000,
    retryLimit: 3,
  });

  const handleProviderClick = useCallback((provider: ProviderNode) => {
    setSelectedProvider(provider);
  }, []);

  const handleProviderAction = useCallback((action: 'start' | 'pause' | 'stop') => {
    if (selectedProvider && onProviderAction) {
      onProviderAction(selectedProvider.id, action);
    }
  }, [selectedProvider, onProviderAction]);

  const handleMacroAction = useCallback((action: 'start' | 'stop' | 'pause' | 'resume') => {
    if (onMacroAction) {
      onMacroAction(action);
    }
  }, [onMacroAction]);

  const handlePolicySubmit = useCallback(() => {
    if (selectedProvider && onPolicyUpdate) {
      onPolicyUpdate(selectedProvider.id, policyForm);
      setPolicyDialogOpen(false);
    }
  }, [selectedProvider, policyForm, onPolicyUpdate]);

  if (isLoading || !data) {
    return (
      <Card>
        <CardContent className="p-8 text-center text-muted-foreground">
          {isLoading ? 'Loading topology...' : 'No topology data available'}
        </CardContent>
      </Card>
    );
  }

  const { macro, totalProviders, activeProviders } = data;

  return (
    <div className="space-y-6">
      {/* Macro FSM Node (Top Level) */}
      <Card className={`border-2 ${getStateColor(macro.state)}`}>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              {getStateIcon(macro.state)}
              <div>
                <CardTitle>Macro FSM Orchestrator</CardTitle>
                <CardDescription>High-level ETL coordination state machine</CardDescription>
              </div>
            </div>
            <Badge variant={getStateBadgeVariant(macro.state)} className="text-sm">
              {macro.state}
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-4 text-sm mb-4">
            <div>
              <span className="text-muted-foreground">Total Providers:</span>
              <span className="ml-2 font-medium">{totalProviders}</span>
            </div>
            <div>
              <span className="text-muted-foreground">Active:</span>
              <span className="ml-2 font-medium">{activeProviders}</span>
            </div>
          </div>
          {onMacroAction && (
            <div className="flex gap-2 pt-2 border-t">
              <Button 
                size="sm" 
                variant="outline" 
                className="flex-1"
                onClick={() => handleMacroAction('start')}
                disabled={macro.state === 'ORCHESTRATING'}
              >
                <PlayCircle className="h-4 w-4 mr-1" />
                Start All
              </Button>
              <Button 
                size="sm" 
                variant="outline" 
                className="flex-1"
                onClick={() => handleMacroAction('pause')}
                disabled={macro.state !== 'ORCHESTRATING'}
              >
                <Pause className="h-4 w-4 mr-1" />
                Pause All
              </Button>
              <Button 
                size="sm" 
                variant="outline" 
                className="flex-1"
                onClick={() => handleMacroAction('resume')}
                disabled={macro.state === 'ORCHESTRATING'}
              >
                <PlayCircle className="h-4 w-4 mr-1" />
                Resume All
              </Button>
              <Button 
                size="sm" 
                variant="outline" 
                className="flex-1"
                onClick={() => handleMacroAction('stop')}
                disabled={macro.state === 'IDLE'}
              >
                <StopCircle className="h-4 w-4 mr-1" />
                Stop All
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Provider FSM Nodes (Connected to Macro) */}
      <div className="relative">
        {/* Connection line indicator */}
        <div className="absolute left-1/2 top-0 w-px h-8 bg-border transform -translate-x-1/2" />
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 pt-8">
          {macro.providers?.map((provider) => (
            <Card 
              key={provider.id}
              className={`border-2 ${getStateColor(provider.state)} cursor-pointer hover:shadow-lg transition-shadow ${
                selectedProvider?.id === provider.id ? 'ring-2 ring-primary' : ''
              }`}
              onClick={() => handleProviderClick(provider)}
            >
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    {getStateIcon(provider.state)}
                    <CardTitle className="text-lg">
                      {(provider.providerType || provider.id || 'UNKNOWN').toUpperCase()}
                    </CardTitle>
                  </div>
                  <Badge variant={getStateBadgeVariant(provider.state)}>
                    {provider.state}
                  </Badge>
                </div>
                <CardDescription className="text-xs font-mono truncate">
                  {provider.id}
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-2">
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div>
                    <div className="text-muted-foreground text-xs">Processed</div>
                    <div className="font-medium">{provider.processedCount}</div>
                  </div>
                  <div>
                    <div className="text-muted-foreground text-xs">Errors</div>
                    <div className="font-medium text-red-600">{provider.errorCount}</div>
                  </div>
                  <div>
                    <div className="text-muted-foreground text-xs">Permits</div>
                    <div className="font-medium">{provider.permitsHeld}</div>
                  </div>
                  <div>
                    <div className="text-muted-foreground text-xs">Health</div>
                    <div className="font-medium">
                      {provider.errorCount === 0 ? (
                        <CheckCircle className="h-4 w-4 text-green-500 inline" />
                      ) : (
                        <AlertCircle className="h-4 w-4 text-yellow-500 inline" />
                      )}
                    </div>
                  </div>
                </div>
                
                {provider.lastCheckpoint && (
                  <div className="pt-2 border-t text-xs">
                    <div className="text-muted-foreground">Last Checkpoint:</div>
                    <div className="font-mono truncate" title={provider.lastCheckpoint}>
                      {provider.lastCheckpoint}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      </div>

      {/* Provider Control Dialog */}
      {selectedProvider && (
        <Dialog open={!!selectedProvider} onOpenChange={() => setSelectedProvider(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                {getStateIcon(selectedProvider.state)}
                Provider: {(selectedProvider.providerType || selectedProvider.id || 'UNKNOWN').toUpperCase()}
              </DialogTitle>
              <DialogDescription className="font-mono text-xs">
                {selectedProvider.id}
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <div className="text-sm text-muted-foreground">State</div>
                  <Badge variant={getStateBadgeVariant(selectedProvider.state)}>
                    {selectedProvider.state}
                  </Badge>
                </div>
                <div>
                  <div className="text-sm text-muted-foreground">Processed</div>
                  <div className="font-medium">{selectedProvider.processedCount}</div>
                </div>
                <div>
                  <div className="text-sm text-muted-foreground">Errors</div>
                  <div className="font-medium text-red-600">{selectedProvider.errorCount}</div>
                </div>
                <div>
                  <div className="text-sm text-muted-foreground">Permits</div>
                  <div className="font-medium">{selectedProvider.permitsHeld}</div>
                </div>
              </div>

              <div className="border-t pt-4">
                <h4 className="font-medium mb-3">Controls</h4>
                <div className="flex gap-2">
                  <Button 
                    size="sm" 
                    variant="outline" 
                    className="flex-1"
                    onClick={() => handleProviderAction('start')}
                    disabled={selectedProvider.state === 'RUNNING'}
                  >
                    <PlayCircle className="h-4 w-4 mr-1" />
                    Start
                  </Button>
                  <Button 
                    size="sm" 
                    variant="outline" 
                    className="flex-1"
                    onClick={() => handleProviderAction('pause')}
                    disabled={selectedProvider.state !== 'RUNNING'}
                  >
                    <Pause className="h-4 w-4 mr-1" />
                    Pause
                  </Button>
                  <Button 
                    size="sm" 
                    variant="outline" 
                    className="flex-1"
                    onClick={() => handleProviderAction('stop')}
                    disabled={selectedProvider.state === 'TERMINATED'}
                  >
                    <StopCircle className="h-4 w-4 mr-1" />
                    Stop
                  </Button>
                </div>
              </div>

              <div className="border-t pt-4">
                <Button 
                  size="sm" 
                  variant="secondary" 
                  className="w-full"
                  onClick={() => setPolicyDialogOpen(true)}
                >
                  <Settings className="h-4 w-4 mr-2" />
                  Update Performance Policy
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      )}

      {/* Performance Policy Dialog */}
      <Dialog open={policyDialogOpen} onOpenChange={setPolicyDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Performance Policy</DialogTitle>
            <DialogDescription>
              Configure performance parameters for {selectedProvider?.providerType}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label htmlFor="maxPermits">Max Permits</Label>
              <Input
                id="maxPermits"
                type="number"
                value={policyForm.maxPermits}
                onChange={(e) => setPolicyForm({ ...policyForm, maxPermits: parseInt(e.target.value) })}
              />
            </div>
            <div>
              <Label htmlFor="batchSize">Batch Size</Label>
              <Input
                id="batchSize"
                type="number"
                value={policyForm.batchSize}
                onChange={(e) => setPolicyForm({ ...policyForm, batchSize: parseInt(e.target.value) })}
              />
            </div>
            <div>
              <Label htmlFor="backoffMs">Backoff (ms)</Label>
              <Input
                id="backoffMs"
                type="number"
                value={policyForm.backoffMs}
                onChange={(e) => setPolicyForm({ ...policyForm, backoffMs: parseInt(e.target.value) })}
              />
            </div>
            <div>
              <Label htmlFor="retryLimit">Retry Limit</Label>
              <Input
                id="retryLimit"
                type="number"
                value={policyForm.retryLimit}
                onChange={(e) => setPolicyForm({ ...policyForm, retryLimit: parseInt(e.target.value) })}
              />
            </div>
            <div className="flex gap-2 pt-4">
              <Button variant="outline" className="flex-1" onClick={() => setPolicyDialogOpen(false)}>
                Cancel
              </Button>
              <Button className="flex-1" onClick={handlePolicySubmit}>
                Apply
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
