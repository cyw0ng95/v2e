'use client';

import React from 'react';
import { useSSGProfiles, useSSGRules, useSSGMetadata } from '@/lib/hooks';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import SSGDetailDialog from './ssg-detail-dialog';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';

export function SSGTable() {
  const [page, setPage] = React.useState(0);
  const pageSize = 20; // Fixed page size
  const [activeTab, setActiveTab] = React.useState<'profiles' | 'rules'>('profiles');
  const offset = page * pageSize;

  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [selectedId, setSelectedId] = React.useState<string | null>(null);
  const [selectedType, setSelectedType] = React.useState<'profile' | 'rule'>('profile');
  const [selectedRow, setSelectedRow] = React.useState<unknown>(null);

  const { data: profileData, isLoading: profilesLoading } = useSSGProfiles(offset, pageSize);
  const { data: ruleData, isLoading: rulesLoading } = useSSGRules(offset, pageSize);
  const { data: metadataData } = useSSGMetadata();

  const profiles = profileData?.profiles || [];
  const rules = ruleData?.rules || [];
  const metadata = metadataData?.metadata || null;

  const isLoading = activeTab === 'profiles' ? profilesLoading : rulesLoading;
  const total = activeTab === 'profiles' ? (profileData?.total ?? 0) : (ruleData?.total ?? 0);
  const items = activeTab === 'profiles' ? profiles : rules;

  const pageCount = Math.max(1, Math.ceil(total / pageSize));
  const visiblePages = 7;
  const getPageRange = () => {
    const half = Math.floor(visiblePages / 2);
    let start = Math.max(0, page - half);
    const end = Math.min(pageCount - 1, start + visiblePages - 1);
    if (end - start + 1 < visiblePages) start = Math.max(0, end - visiblePages + 1);
    return { start, end };
  };
  const { start, end } = getPageRange();
  const pages: number[] = [];
  for (let p = start; p <= end; p++) pages.push(p);

  const getSeverityVariant = (severity: string): "default" | "secondary" | "destructive" | "outline" => {
    switch (severity?.toLowerCase()) {
      case 'critical':
        return 'destructive';
      case 'high':
        return 'default';
      case 'medium':
        return 'secondary';
      case 'low':
        return 'outline';
      default:
        return 'outline';
    }
  };

  return (
    <Card className="h-full flex flex-col">
      <CardHeader>
        <CardTitle>SSG (SCAP Security Guide) Database</CardTitle>
        <CardDescription>
          Browse SSG profiles and security rules
          {metadata && (
            <span className="block mt-1 text-xs">
              {metadata.title} • Version {metadata.version} • {metadata.profileCount} profiles, {metadata.ruleCount} rules
            </span>
          )}
        </CardDescription>
        <div className="mt-3">
          <Input
            className="w-full"
            placeholder="Search SSG content"
            // Add search functionality here if needed
          />
        </div>
      </CardHeader>
      <CardContent className="flex-1 min-h-0 overflow-auto">
        <Tabs value={activeTab} onValueChange={(v) => { setActiveTab(v as 'profiles' | 'rules'); setPage(0); }}>
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="profiles">Profiles</TabsTrigger>
            <TabsTrigger value="rules">Rules</TabsTrigger>
          </TabsList>

          <TabsContent value="profiles" className="mt-4">
            <table className="min-w-full text-sm">
              <thead>
                <tr className="border-b">
                  <th className="text-left p-2">Profile ID</th>
                  <th className="text-left p-2">Title</th>
                  <th className="text-left p-2">Description</th>
                  <th className="text-left p-2">Rules</th>
                  <th className="text-left p-2">Actions</th>
                </tr>
              </thead>
              <tbody>
                {isLoading ? (
                  <tr>
                    <td colSpan={5} className="p-4 text-sm text-muted-foreground">Loading...</td>
                  </tr>
                ) : profiles.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="p-4 text-sm text-muted-foreground">No SSG profiles</td>
                  </tr>
                ) : (
                  profiles.map((profile: { id: string; title: string; description: string; ruleCount: number }) => (
                    <tr key={profile.id} className="hover:bg-muted">
                      <td className="p-2 font-mono text-xs">{profile.id}</td>
                      <td className="p-2 font-medium">{profile.title}</td>
                      <td className="p-2 max-w-xs truncate" title={profile.description}>{profile.description}</td>
                      <td className="p-2">
                        <Badge variant="outline">{profile.ruleCount} rules</Badge>
                      </td>
                      <td className="p-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => {
                            setSelectedRow(profile);
                            setSelectedId(profile.id);
                            setSelectedType('profile');
                            setDialogOpen(true);
                          }}
                        >
                          View Detail
                        </Button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </TabsContent>

          <TabsContent value="rules" className="mt-4">
            <table className="min-w-full text-sm">
              <thead>
                <tr className="border-b">
                  <th className="text-left p-2">Rule ID</th>
                  <th className="text-left p-2">Title</th>
                  <th className="text-left p-2">Severity</th>
                  <th className="text-left p-2">Description</th>
                  <th className="text-left p-2">Actions</th>
                </tr>
              </thead>
              <tbody>
                {isLoading ? (
                  <tr>
                    <td colSpan={5} className="p-4 text-sm text-muted-foreground">Loading...</td>
                  </tr>
                ) : rules.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="p-4 text-sm text-muted-foreground">No SSG rules</td>
                  </tr>
                ) : (
                  rules.map((rule: { id: string; title: string; severity: string; description: string }) => (
                    <tr key={rule.id} className="hover:bg-muted">
                      <td className="p-2 font-mono text-xs">{rule.id}</td>
                      <td className="p-2 font-medium">{rule.title}</td>
                      <td className="p-2">
                        <Badge variant={getSeverityVariant(rule.severity)}>
                          {rule.severity || 'unknown'}
                        </Badge>
                      </td>
                      <td className="p-2 max-w-xs truncate" title={rule.description}>{rule.description}</td>
                      <td className="p-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => {
                            setSelectedRow(rule);
                            setSelectedId(rule.id);
                            setSelectedType('rule');
                            setDialogOpen(true);
                          }}
                        >
                          View Detail
                        </Button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </TabsContent>
        </Tabs>
        
        <div className="pt-4 flex items-center justify-between">
          <div className="text-sm text-muted-foreground">
            {isLoading ? (
              'Loading...'
            ) : (
              `Showing ${Math.min(offset + 1, total || 0)}-${Math.min(offset + items.length, total || offset + items.length)} of ${total}`
            )}
          </div>
          <div className="flex items-center gap-2">
            <Button size="sm" variant="outline" onClick={() => setPage(0)} disabled={page === 0}>First</Button>
            <Button size="sm" variant="outline" onClick={() => setPage((p) => Math.max(0, p - 1))} disabled={page === 0}>Prev</Button>
            {pages.map((p) => (
              <Button
                key={p}
                size="sm"
                variant={p === page ? 'default' : 'outline'}
                onClick={() => setPage(p)}
              >
                {p + 1}
              </Button>
            ))}
            <Button size="sm" variant="outline" onClick={() => setPage((p) => Math.min(pageCount - 1, p + 1))} disabled={page >= pageCount - 1}>Next</Button>
            <Button size="sm" variant="outline" onClick={() => setPage(pageCount - 1)} disabled={page >= pageCount - 1}>Last</Button>
          </div>
        </div>
      </CardContent>

      <SSGDetailDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        itemId={selectedId}
        itemType={selectedType}
        initial={selectedRow}
      />
    </Card>
  );
}
