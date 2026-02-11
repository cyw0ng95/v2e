'use client';

import { useState, useMemo } from 'react';
import { McardsNavigation } from './mcards-navigation';
import { McardsSidebar } from './mcards-sidebar';
import { McardsTable } from './mcards-table';
import { McardsDashboard } from './mcards-dashboard';
import { McardsStudy } from './mcards-study';
import { CardCreateModal } from './card-create-modal';

type View = 'cards' | 'dashboard' | 'study';

export function McardsPage() {
  const [currentView, setCurrentView] = useState<View>('cards');
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [filters, setFilters] = useState({
    learningState: '',
    majorClass: '',
    bookmarkId: '',
  });

  const renderContent = useMemo(() => {
    switch (currentView) {
      case 'cards':
        return <McardsTable filters={filters} />;
      case 'dashboard':
        return <McardsDashboard />;
      case 'study':
        return <McardsStudy />;
      default:
        return <McardsTable filters={filters} />;
    }
  }, [currentView, filters]);

  return (
    <div className="min-h-screen bg-slate-50">
      <McardsNavigation
        currentView={currentView}
        onViewChange={setCurrentView}
        onCreateCard={() => setCreateModalOpen(true)}
      />
      <div className="flex">
        <McardsSidebar
          filters={filters}
          onFilterChange={setFilters}
        />
        <main className="flex-1 p-6">
          {renderContent}
        </main>
      </div>
      <CardCreateModal
        open={createModalOpen}
        onOpenChange={setCreateModalOpen}
      />
    </div>
  );
}
