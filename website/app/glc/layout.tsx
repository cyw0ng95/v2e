'use client';

import { ReactFlowProvider } from '@xyflow/react';
import { Toaster } from '@/components/ui/sonner';

export default function GLCLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <ReactFlowProvider>
      <div className="h-screen w-full overflow-hidden bg-background">
        {children}
      </div>
      <Toaster />
    </ReactFlowProvider>
  );
}
