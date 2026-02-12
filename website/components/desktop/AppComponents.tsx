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

// Lazy load application components for better performance
const CVSSCalculator = lazy(() => import('@/components/cvss/calculator').then(m => ({ default: m.CVSSCalculator })));
const ETLEnginePage = lazy(() => import('@/app/etl/page').then(m => ({ default: m.default })));
const GLCLandingPage = lazy(() => import('@/app/glc/page').then(m => ({ default: m.default })));
const McardsStudy = lazy(() => import('@/components/mcards/mcards-study').then(m => ({ default: m.McardsStudy })));
const BookmarkTable = lazy(() => import('@/components/bookmark-table').then(m => ({ default: m.BookmarkTable })));

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
        <div className="h-full overflow-auto">
          <CVSSCalculator />
        </div>
      );

    case 'etl':
      return (
        <div className="h-full overflow-auto bg-background">
          <ETLEnginePage />
        </div>
      );

    case 'glc':
      return (
        <div className="h-full overflow-auto bg-background">
          <GLCLandingPage />
        </div>
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

    // Database apps (CVE, CWE, CAPEC, ATT&CK) - still using route-based pages
    // These can be progressively migrated to components
    case 'cve':
    case 'cwe':
    case 'capec':
    case 'attack':
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
  const componentApps = ['cvss', 'etl', 'glc', 'mcards', 'bookmarks'];
  return componentApps.includes(appId);
}
