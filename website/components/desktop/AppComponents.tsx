/**
 * v2e Portal - Desktop Application Components
 *
 * Component mapping for direct rendering in windows (SPA architecture)
 * Replaces iframe-based app loading with React component rendering
 */

'use client';

import { lazy, Suspense } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Loader2 } from 'lucide-react';
import { CVSSProvider } from '@/lib/cvss-context';
import { useEtlTree } from '@/lib/hooks';
import { rpcClient } from '@/lib/rpc-client';

// Lazy load application components for better performance
const CVSSCalculator = lazy(() => import('@/components/cvss/calculator').then(m => ({ default: m.default })));
const McardsStudy = lazy(() => import('@/components/mcards/mcards-study').then(m => ({ default: m.default })));
const BookmarkTable = lazy(() => import('@/components/bookmark-table').then(m => ({ default: m.default })));
const CAPECTable = lazy(() => import('@/components/capec-table').then(m => ({ default: m.CAPECTable })));
const CWE_TABLE = lazy(() => import('@/components/cwe-table').then(m => ({ default: m.CWETable })));
const ATTACKTable = lazy(() => import('@/components/attack-table').then(m => ({ default: m.AttackTable })));
const GraphAnalysisPage = lazy(() => import('@/components/graph-analysis-page').then(m => ({ default: m.default })));
const ETLTopologyViewer = lazy(() => import('@/components/etl-topology-viewer').then(m => ({ default: m.ETLTopologyViewer })));

/**
 * Loading component for lazy-loaded apps
 */
function AppLoading() {
  return (
    <div className="h-full flex items-center justify-center bg-gray-50">
      <div className="text-center">
        <Loader2 className="w-8 h-8 text-blue-600 animate-spin mx-auto mb-3" />
        <p className="text-sm text-gray-600">Loading application...</p>
      </div>
    </div>
  );
}

/**
 * Fallback component for unimplemented apps
 */
function AppPlaceholder({ appId, title }: { appId: string; title: string }) {
  return (
    <div className="h-full flex items-center justify-center bg-gray-50 p-6">
      <Card className="max-w-md">
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-gray-600">
            This application ({appId}) is currently loaded via iframe.
            Full component integration coming soon.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}

/**
 * Application component props
 */
export interface AppComponentProps {
  appId: string;
  title: string;
  windowId?: string;
}

/**
 * Application component mapping
 * Maps appId to React component for direct rendering
 */
export function AppComponent({ appId, title, windowId }: AppComponentProps) {
  return (
    <Suspense fallback={<AppLoading />}>
      {renderAppComponent(appId, title, windowId)}
    </Suspense>
  );
}

/**
 * Internal renderer that maps appId to component
 */
function renderAppComponent(appId: string, title: string, windowId?: string) {
  switch (appId) {
    case 'cvss':
      return (
        <CVSSProvider>
          <div className="h-full overflow-auto">
            <CVSSCalculator />
          </div>
        </CVSSProvider>
      );

    case 'mcards':
      return (
        <div className="h-full overflow-auto bg-background">
          <McardsStudy />
        </div>
      );

    case 'bookmarks':
      return (
        <div className="h-full overflow-auto bg-background">
          <BookmarkTable />
        </div>
      );

    case 'capec':
      return (
        <div className="h-full overflow-auto bg-background">
          <CAPECTable />
        </div>
      );

    case 'cwe':
      return (
        <div className="h-full overflow-auto bg-background">
          <CWE_TABLE />
        </div>
      );

    case 'attack':
      return (
        <div className="h-full overflow-auto bg-background">
          <ATTACKTable />
        </div>
      );

    case 'glc':
      return (
        <div className="h-full overflow-hidden bg-background">
          <GraphAnalysisPage />
        </div>
      );

    case 'etl': {
      const { data, isLoading, error } = useEtlTree(5000);
      
      const handleProviderAction = async (providerId: string, action: 'start' | 'pause' | 'stop') => {
        try {
          if (action === 'start') {
            await rpcClient.startProvider(providerId);
          } else if (action === 'pause') {
            await rpcClient.pauseProvider(providerId);
          } else {
            await rpcClient.stopProvider(providerId);
          }
        } catch (err) {
          console.error(`Failed to ${action} provider:`, err);
        }
      };

      return (
        <div className="h-full overflow-auto bg-background p-4">
          <ETLTopologyViewer 
            data={data} 
            isLoading={isLoading}
            onProviderAction={handleProviderAction}
          />
        </div>
      );
    }

    // Other apps - placeholder
    case 'sysmon':
    case 'cce':
    case 'ssg':
    case 'asvs':
    default:
      return <AppPlaceholder appId={appId} title={title} />;
  }
}

/**
 * Check if an app has a component implementation
 */
export function hasAppComponent(appId: string): boolean {
  const componentApps = ['cvss', 'mcards', 'bookmarks', 'capec', 'cwe', 'attack', 'glc', 'etl'];
  return componentApps.includes(appId);
}
