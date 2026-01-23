'use client';

import React, { useState } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { LucideEye } from 'lucide-react';
import { AttackTable } from './attack-table';
import { AttackDetailContent } from './attack-detail-dialog';

export function AttackViews() {
  const [activeTab, setActiveTab] = useState<'techniques' | 'tactics' | 'mitigations' | 'software' | 'groups' | 'detail'>('techniques');
  const [previousTab, setPreviousTab] = useState<'techniques' | 'tactics' | 'mitigations' | 'software' | 'groups'>('techniques');
  const [selectedItem, setSelectedItem] = useState<any>(null);

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>ATT&CK Framework</CardTitle>
        <div className="flex flex-wrap gap-2 mt-4">
          <Button
            variant={activeTab === 'techniques' ? 'default' : 'outline'}
            onClick={() => setActiveTab('techniques')}
          >
            Techniques
          </Button>
          <Button
            variant={activeTab === 'tactics' ? 'default' : 'outline'}
            onClick={() => setActiveTab('tactics')}
          >
            Tactics
          </Button>
          <Button
            variant={activeTab === 'mitigations' ? 'default' : 'outline'}
            onClick={() => setActiveTab('mitigations')}
          >
            Mitigations
          </Button>
          <Button
            variant={activeTab === 'software' ? 'default' : 'outline'}
            onClick={() => setActiveTab('software')}
          >
            Software
          </Button>
          <Button
            variant={activeTab === 'groups' ? 'default' : 'outline'}
            onClick={() => setActiveTab('groups')}
          >
            Groups
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        {activeTab !== 'detail' ? (
          <AttackTable
            type={activeTab}
            onViewDetail={(item) => {
              setPreviousTab(activeTab as any);
              setSelectedItem(item);
              setActiveTab('detail');
            }}
          />
        ) : (
          <div>
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <Button variant="outline" size="sm" onClick={() => setActiveTab(previousTab)}>
                  ← Back
                </Button>
                <h3 className="text-lg font-semibold">Detail</h3>
              </div>
            </div>
            <AttackDetailContent item={selectedItem} type={previousTab} onBack={() => setActiveTab(previousTab)} />
          </div>
        )}
      </CardContent>
    </Card>
  );
}