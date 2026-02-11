'use client';

/**
 * CVSS Calculator Page - Dynamic Route
 * /cvss/[version] where version is 3.0, 3.1, or 4.0
 */

import { notFound } from 'next/navigation';
import { CVSSProvider } from '@/lib/cvss-context';
import CVSSCalculator from '@/components/cvss/calculator';

interface PageProps {
  params: Promise<{ version: string }>;
}

export default async function CVSSVersionPage({ params }: PageProps) {
  const { version } = await params;

  // Validate version
  if (!['3.0', '3.1', '4.0'].includes(version)) {
    notFound();
  }

  return (
    <CVSSProvider initialVersion={version as '3.0' | '3.1' | '4.0'}>
      <CVSSCalculator />
    </CVSSProvider>
  );
}
