'use client';

import React, { useState } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { LucideEye } from 'lucide-react';
import { AttackTable } from './attack-table';

export function AttackViews() {
  const [activeTab, setActiveTab] = useState<'techniques' | 'tactics' | 'mitigations' | 'software' | 'groups'>('techniques');

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
        <AttackTable type={activeTab} />
      </CardContent>
    </Card>
  );
}