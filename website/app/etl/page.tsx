'use client';

import { useEtlTree, useKernelMetrics } from '@/lib/hooks';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Activity, Cpu, Database, Gauge, Layers, PlayCircle, Pause, StopCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';

export default function ETLEnginePage() {
  const { data: etlData, isLoading: etlLoading } = useEtlTree(5000);
  const { data: metricsData, isLoading: metricsLoading } = useKernelMetrics(2000);

  const tree = etlData?.tree;
  const metrics = metricsData?.metrics;

  // Helper to get state badge variant
  const getStateVariant = (state: string) => {
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

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* Page Header */}
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <Activity className="h-6 w-6" />
          <h1 className="text-3xl font-bold tracking-tight">ETL Engine</h1>
        </div>
        <p className="text-muted-foreground">
          Unified ETL Engine monitoring - Master-Slave FSM orchestration
        </p>
      </div>

      {/* Kernel Metrics */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">P99 Latency</CardTitle>
            <Gauge className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {metricsLoading ? '...' : `${metrics?.p99Latency?.toFixed(1) || '0.0'}ms`}
            </div>
            <p className="text-xs text-muted-foreground">
              {metrics?.p99Latency > 30 ? '⚠️ Above threshold (30ms)' : '✓ Within limits'}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Buffer Saturation</CardTitle>
            <Database className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {metricsLoading ? '...' : `${metrics?.bufferSaturation?.toFixed(0) || '0'}%`}
            </div>
            <p className="text-xs text-muted-foreground">
              {metrics?.bufferSaturation > 80 ? '⚠️ High saturation' : '✓ Normal'}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Message Rate</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {metricsLoading ? '...' : `${metrics?.messageRate?.toFixed(0) || '0'}`}
            </div>
            <p className="text-xs text-muted-foreground">messages/sec</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Error Rate</CardTitle>
            <Cpu className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {metricsLoading ? '...' : `${metrics?.errorRate?.toFixed(2) || '0.00'}`}
            </div>
            <p className="text-xs text-muted-foreground">errors/sec</p>
          </CardContent>
        </Card>
      </div>

      {/* Macro FSM State */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Macro FSM Orchestrator</CardTitle>
              <CardDescription>High-level ETL coordination state machine</CardDescription>
            </div>
            <Badge variant={getStateVariant(tree?.macro?.state)}>
              {etlLoading ? 'Loading...' : tree?.macro?.state || 'UNKNOWN'}
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Total Providers:</span>
              <span className="font-medium">{tree?.totalProviders || 0}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Active Providers:</span>
              <span className="font-medium">{tree?.activeProviders || 0}</span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Provider FSMs */}
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <Layers className="h-5 w-5" />
          <h2 className="text-2xl font-bold">Provider State Machines</h2>
        </div>

        {etlLoading ? (
          <Card>
            <CardContent className="p-6">
              <p className="text-center text-muted-foreground">Loading providers...</p>
            </CardContent>
          </Card>
        ) : (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {tree?.macro?.providers?.map((provider: any) => (
              <Card key={provider.id}>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-lg">{provider.providerType.toUpperCase()}</CardTitle>
                    <Badge variant={getStateVariant(provider.state)}>
                      {provider.state}
                    </Badge>
                  </div>
                  <CardDescription className="text-xs font-mono">
                    {provider.id}
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="space-y-1">
                    <div className="flex justify-between text-sm">
                      <span className="text-muted-foreground">Processed:</span>
                      <span className="font-medium">{provider.processedCount}</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-muted-foreground">Errors:</span>
                      <span className="font-medium">{provider.errorCount}</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-muted-foreground">Permits:</span>
                      <span className="font-medium">{provider.permitsHeld}</span>
                    </div>
                  </div>

                  {provider.lastCheckpoint && (
                    <div className="pt-2 border-t">
                      <p className="text-xs text-muted-foreground">Last Checkpoint:</p>
                      <p className="text-xs font-mono truncate" title={provider.lastCheckpoint}>
                        {provider.lastCheckpoint}
                      </p>
                    </div>
                  )}

                  <div className="flex gap-2 pt-2">
                    <Button size="sm" variant="outline" className="flex-1" disabled>
                      <PlayCircle className="h-3 w-3 mr-1" />
                      Start
                    </Button>
                    <Button size="sm" variant="outline" className="flex-1" disabled>
                      <Pause className="h-3 w-3 mr-1" />
                      Pause
                    </Button>
                    <Button size="sm" variant="outline" className="flex-1" disabled>
                      <StopCircle className="h-3 w-3 mr-1" />
                      Stop
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>

      {/* Info Card */}
      <Card className="bg-muted/50">
        <CardHeader>
          <CardTitle className="text-sm">About the ETL Engine</CardTitle>
        </CardHeader>
        <CardContent className="text-sm space-y-2">
          <p>
            The Unified ETL Engine uses a <strong>Master-Slave hierarchical FSM model</strong> where:
          </p>
          <ul className="list-disc list-inside space-y-1 text-muted-foreground">
            <li>The <strong>Broker (Master)</strong> manages worker permits and monitors kernel metrics</li>
            <li>The <strong>Meta Service (Slave)</strong> orchestrates provider FSMs for ETL tasks</li>
            <li>Providers use <strong>URN-based checkpointing</strong> for resumability</li>
            <li>Auto-recovery resumes jobs after service restarts</li>
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}
