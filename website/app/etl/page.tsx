'use client';

import { useEtlTree, useKernelMetrics } from '@/lib/hooks';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Activity, Cpu, Database, Gauge } from 'lucide-react';
import { ETLTopologyViewer } from '@/components/etl-topology-viewer';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { logInfo, logError } from '@/lib/logger';

export default function ETLEnginePage() {
  const { data: etlData, isLoading: etlLoading } = useEtlTree(5000);
  const { data: metricsData, isLoading: metricsLoading } = useKernelMetrics(2000);

  const tree = etlData?.tree;
  const metrics = metricsData?.metrics;

  const handleProviderAction = async (providerId: string, action: 'start' | 'pause' | 'stop') => {
    logInfo('ETLEnginePage', `Provider action: ${providerId} ${action}`);

    try {
      let response: RPCResponse<{ success: boolean }>;
      switch (action) {
        case 'start':
          response = await rpcClient.call<{ providerId: string }, { success: boolean }>(
            'RPCStartProvider',
            { providerId },
            'meta'
          );
          break;
        case 'pause':
          response = await rpcClient.call<{ providerId: string }, { success: boolean }>(
            'RPCPauseProvider',
            { providerId },
            'meta'
          );
          break;
        case 'stop':
          response = await rpcClient.call<{ providerId: string }, { success: boolean }>(
            'RPCStopProvider',
            { providerId },
            'meta'
          );
          break;
      }

      if (response.retcode !== 0) {
        logError('ETLEnginePage', 'Provider action failed', response.message);
        throw new Error(response.message || 'Provider action failed');
      }

      logInfo('ETLEnginePage', 'Provider action successful', { success: response.data.success });
      // Refetch ETL tree to update UI
      window.location.reload();
    } catch (error) {
      logError('ETLEnginePage', 'Error executing provider action', error);
      throw error;
    }
  };

  const handlePolicyUpdate = async (providerId: string, policy: any) => {
    logInfo('ETLEnginePage', `Policy update: ${providerId}`, { policy });

    try {
      const response = await rpcClient.call<{ providerId: string; policy: any }, { success: boolean }>(
        'RPCUpdatePerformancePolicy',
        { providerId, policy },
        'meta'
      );

      if (response.retcode !== 0) {
        logError('ETLEnginePage', 'Policy update failed', response.message);
        throw new Error(response.message || 'Policy update failed');
      }

      logInfo('ETLEnginePage', 'Policy updated successfully', { success: response.data.success });
    } catch (error) {
      logError('ETLEnginePage', 'Error updating performance policy', error);
      throw error;
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
          Unified ETL Engine monitoring - Master-Slave hierarchical FSM orchestration
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

      {/* Tabs for Different Views */}
      <Tabs defaultValue="topology" className="space-y-4">
        <TabsList>
          <TabsTrigger value="topology">Topology View</TabsTrigger>
          <TabsTrigger value="list">List View</TabsTrigger>
        </TabsList>

        <TabsContent value="topology" className="space-y-4">
          <ETLTopologyViewer 
            data={tree}
            isLoading={etlLoading}
            onProviderAction={handleProviderAction}
            onPolicyUpdate={handlePolicyUpdate}
          />
        </TabsContent>

        <TabsContent value="list" className="space-y-4">
          {/* Legacy list view (kept for backward compatibility) */}
          <Card>
            <CardHeader>
              <CardTitle>Provider List</CardTitle>
              <CardDescription>Table view of all ETL providers</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {tree?.macro?.providers?.map((provider: any) => (
                  <div key={provider.id} className="flex items-center justify-between border rounded p-3">
                    <div className="flex-1">
                      <div className="font-medium">{provider.providerType.toUpperCase()}</div>
                      <div className="text-xs text-muted-foreground font-mono">{provider.id}</div>
                    </div>
                    <div className="text-sm">
                      Processed: <span className="font-medium">{provider.processedCount}</span> | 
                      Errors: <span className="font-medium text-red-600">{provider.errorCount}</span>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

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
