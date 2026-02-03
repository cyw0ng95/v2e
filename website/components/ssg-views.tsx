import React, { useState } from "react";
import {
  useSSGGuides,
  useSSGTables,
  useSSGImportStatus,
  useStartSSGImportJob,
  useStopSSGImportJob,
  usePauseSSGImportJob,
  useResumeSSGImportJob,
  useSSGTree,
  useSSGTableEntries,
} from "@/lib/hooks";
import { Button } from "./ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "./ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./ui/tabs";
import { Skeleton } from "./ui/skeleton";
import { Badge } from "./ui/badge";
import { ChevronRight, ChevronDown, Folder, FileText, Play, Pause, Square, RotateCcw, Table } from "lucide-react";
import type { TreeNode } from "@/lib/types";

// Tree Node Component for recursive rendering
function TreeViewNode({ node, level = 0, onNodeClick }: { node: TreeNode; level?: number; onNodeClick: (node: TreeNode) => void }) {
  const [isExpanded, setIsExpanded] = useState(level < 2); // Auto-expand first 2 levels

  const hasChildren = node.children && node.children.length > 0;
  const isGroup = node.type === 'group';

  return (
    <div className="select-none">
      <div
        className="flex items-center gap-1 py-1 px-2 hover:bg-muted rounded cursor-pointer"
        style={{ paddingLeft: `${level * 16 + 8}px` }}
        onClick={() => {
          if (hasChildren) {
            setIsExpanded(!isExpanded);
          }
          onNodeClick(node);
        }}
      >
        {hasChildren ? (
          isExpanded ? (
            <ChevronDown className="w-4 h-4 text-muted-foreground" />
          ) : (
            <ChevronRight className="w-4 h-4 text-muted-foreground" />
          )
        ) : (
          <span className="w-4 h-4" />
        )}
        {isGroup ? (
          <Folder className="w-4 h-4 text-blue-500 shrink-0" />
        ) : (
          <FileText className="w-4 h-4 text-green-500 shrink-0" />
        )}
        <span className="text-sm truncate flex-1" title={node.group?.title || node.rule?.title}>
          {node.group?.title || node.rule?.title || node.id}
        </span>
        {node.rule?.severity && (
          <Badge
            variant={
              node.rule.severity === 'high'
                ? 'destructive'
                : node.rule.severity === 'medium'
                  ? 'default'
                  : 'secondary'
            }
            className="text-xs"
          >
            {node.rule.severity}
          </Badge>
        )}
      </div>
      {isExpanded && hasChildren && (
        <div>
          {node.children.map((child) => (
            <TreeViewNode key={child.id} node={child} level={level + 1} onNodeClick={onNodeClick} />
          ))}
        </div>
      )}
    </div>
  );
}

// Detail Panel Component
function DetailPanel({ selectedNode, onClose }: { selectedNode: TreeNode | null; onClose: () => void }) {
  if (!selectedNode) {
    return (
      <div className="flex items-center justify-center h-full text-muted-foreground">
        Select a group or rule to view details
      </div>
    );
  }

  if (selectedNode.type === 'group' && selectedNode.group) {
    const group = selectedNode.group;
    return (
      <div className="h-full flex flex-col">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold flex items-center gap-2">
            <Folder className="w-5 h-5 text-blue-500" />
            {group.title}
          </h3>
          <Button variant="ghost" size="sm" onClick={onClose}>
            Close
          </Button>
        </div>
        <div className="flex-1 overflow-auto space-y-4">
          <div>
            <h4 className="text-sm font-medium text-muted-foreground mb-1">ID</h4>
            <p className="text-sm font-mono">{group.id}</p>
          </div>
          <div>
            <h4 className="text-sm font-medium text-muted-foreground mb-1">Description</h4>
            <p className="text-sm whitespace-pre-wrap">{group.description || 'No description'}</p>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <h4 className="text-sm font-medium text-muted-foreground mb-1">Level</h4>
              <p className="text-sm">{group.level}</p>
            </div>
            <div>
              <h4 className="text-sm font-medium text-muted-foreground mb-1">Contains</h4>
              <p className="text-sm">{group.groupCount || 0} groups, {group.ruleCount || 0} rules</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (selectedNode.type === 'rule' && selectedNode.rule) {
    const rule = selectedNode.rule;
    return (
      <div className="h-full flex flex-col">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold flex items-center gap-2">
            <FileText className="w-5 h-5 text-green-500" />
            {rule.title}
          </h3>
          <Badge
            variant={
              rule.severity === 'high'
                ? 'destructive'
                : rule.severity === 'medium'
                  ? 'default'
                  : 'secondary'
            }
          >
            {rule.severity}
          </Badge>
        </div>
        <div className="flex-1 overflow-auto space-y-4">
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="sm" onClick={onClose}>
              Close
            </Button>
          </div>
          <div>
            <h4 className="text-sm font-medium text-muted-foreground mb-1">ID</h4>
            <p className="text-sm font-mono">{rule.id}</p>
          </div>
          <div>
            <h4 className="text-sm font-medium text-muted-foreground mb-1">Description</h4>
            <p className="text-sm whitespace-pre-wrap">{rule.description || 'No description'}</p>
          </div>
          <div>
            <h4 className="text-sm font-medium text-muted-foreground mb-1">Rationale</h4>
            <p className="text-sm whitespace-pre-wrap">{rule.rationale || 'No rationale provided'}</p>
          </div>
          {rule.references && rule.references.length > 0 && (
            <div>
              <h4 className="text-sm font-medium text-muted-foreground mb-2">References</h4>
              <ul className="space-y-2">
                {rule.references.map((ref, idx) => (
                  <li key={idx} className="text-sm">
                    <a
                      href={ref.href}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-blue-500 hover:underline"
                    >
                      {ref.label || ref.value}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      </div>
    );
  }

  return null;
}

export function SSGViews() {
  const [activeTab, setActiveTab] = useState<'guides' | 'tables'>('guides');
  const [selectedGuide, setSelectedGuide] = useState<any | null>(null);
  const [selectedTable, setSelectedTable] = useState<any | null>(null);
  const [selectedNode, setSelectedNode] = useState<TreeNode | null>(null);

  // Guide hooks
  const { data: guidesData, isLoading: guidesLoading, error: guidesError, refetch: refetchGuides } = useSSGGuides();
  
  // Table hooks
  const { data: tablesData, isLoading: tablesLoading, error: tablesError, refetch: refetchTables } = useSSGTables();

  // Job control hooks
  const { data: jobStatus, isLoading: jobLoading, refetch: refetchJobStatus } = useSSGImportStatus();
  const startJob = useStartSSGImportJob();
  const stopJob = useStopSSGImportJob();
  const pauseJob = usePauseSSGImportJob();
  const resumeJob = useResumeSSGImportJob();

  // Poll job status and data when running
  React.useEffect(() => {
    if (jobStatus?.state === 'running' || jobStatus?.state === 'queued') {
      const interval = setInterval(() => {
        refetchJobStatus();
        refetchGuides();
        refetchTables();
      }, 2000);
      return () => clearInterval(interval);
    }
  }, [jobStatus?.state, refetchJobStatus, refetchGuides, refetchTables]);

  // Extract guides and tables from response
  const guides: any[] = (guidesData && Array.isArray(guidesData.guides)) ? guidesData.guides : [];
  const tables: any[] = (tablesData && Array.isArray(tablesData.tables)) ? tablesData.tables : [];

  const jobProgress = jobStatus?.progress;

  return (
    <div className="w-full h-full flex gap-4">
      {/* Left Panel: Tabs for Guides/Tables and Job Control */}
      <Card className="w-80 flex flex-col">
        <CardHeader>
          <CardTitle className="text-base">SSG Data</CardTitle>
          <CardDescription>Security Guides & Tables</CardDescription>
        </CardHeader>
        <CardContent className="flex-1 flex flex-col min-h-0">
          {/* Job Control */}
          <div className="space-y-2 mb-4 pb-4 border-b">
            <div className="flex items-center gap-2 flex-wrap">
              <Button
                variant="default"
                size="sm"
                disabled={jobLoading || (jobStatus?.state === 'running' || jobStatus?.state === 'queued')}
                onClick={() => startJob.mutate({}, { onSuccess: () => refetchJobStatus() })}
              >
                <Play className="w-3 h-3 mr-1" />
                Start
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={jobLoading || !(jobStatus?.state === 'running' || jobStatus?.state === 'queued')}
                onClick={() => stopJob.mutate(undefined, { onSuccess: () => refetchJobStatus() })}
              >
                <Square className="w-3 h-3 mr-1" />
                Stop
              </Button>
              {(jobStatus?.state === 'running' || jobStatus?.state === 'queued') && (
                <Button
                  variant="outline"
                  size="sm"
                  disabled={jobLoading}
                  onClick={() => pauseJob.mutate(undefined, { onSuccess: () => refetchJobStatus() })}
                >
                  <Pause className="w-3 h-3 mr-1" />
                  Pause
                </Button>
              )}
              {jobStatus?.state === 'paused' && (
                <Button
                  variant="outline"
                  size="sm"
                  disabled={jobLoading}
                  onClick={() => resumeJob.mutate({ runId: jobStatus.id }, { onSuccess: () => refetchJobStatus() })}
                >
                  <RotateCcw className="w-3 h-3 mr-1" />
                  Resume
                </Button>
              )}
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  refetchGuides();
                  refetchTables();
                  refetchJobStatus();
                }}
              >
                <RotateCcw className="w-3 h-3 mr-1" />
                Refresh
              </Button>
            </div>
            {/* Job Status */}
            {jobStatus && (
              <div className="text-xs space-y-1">
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">Status:</span>
                  <Badge
                    variant={
                      jobStatus.state === 'running' || jobStatus.state === 'queued'
                        ? 'default'
                        : jobStatus.state === 'completed'
                          ? 'default'
                          : jobStatus.state === 'failed'
                            ? 'destructive'
                            : 'secondary'
                    }
                  >
                    {jobStatus.state}
                  </Badge>
                </div>
                {jobProgress && (
                  <>
                    {jobProgress.currentPhase && (
                      <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">Phase:</span>
                        <span className="capitalize">{jobProgress.currentPhase}</span>
                      </div>
                    )}
                    <div className="flex items-center justify-between">
                      <span className="text-muted-foreground">Tables:</span>
                      <span>
                        {jobProgress.processedTables} / {jobProgress.totalTables}
                      </span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-muted-foreground">Guides:</span>
                      <span>
                        {jobProgress.processedGuides} / {jobProgress.totalGuides}
                      </span>
                    </div>
                    {jobProgress.currentFile && (
                      <div className="text-muted-foreground truncate" title={jobProgress.currentFile}>
                        {jobProgress.currentFile}
                      </div>
                    )}
                    {(jobProgress.failedTables > 0 || jobProgress.failedGuides > 0) && (
                      <div className="text-destructive">
                        {jobProgress.failedTables + jobProgress.failedGuides} failed
                      </div>
                    )}
                  </>
                )}
              </div>
            )}
          </div>

          {/* Tabs for Guides and Tables */}
          <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as 'guides' | 'tables')} className="flex-1 flex flex-col min-h-0">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="guides">Guides</TabsTrigger>
              <TabsTrigger value="tables">Tables</TabsTrigger>
            </TabsList>
            
            <TabsContent value="guides" className="flex-1 overflow-auto mt-4">
              {guidesLoading ? (
                <div className="space-y-2">
                  <Skeleton className="h-12 w-full" />
                  <Skeleton className="h-12 w-full" />
                  <Skeleton className="h-12 w-full" />
                </div>
              ) : guidesError ? (
                <div className="text-sm text-destructive">Error loading guides</div>
              ) : guides.length === 0 ? (
                <div className="text-sm text-muted-foreground text-center py-8">
                  No guides available.<br />
                  Start import job to fetch guides.
                </div>
              ) : (
                <div className="space-y-1">
                  {guides.map((guide) => (
                    <div
                      key={guide.id}
                      className={`flex flex-col p-2 rounded cursor-pointer transition-colors ${
                        selectedGuide?.id === guide.id ? 'bg-muted' : 'hover:bg-muted/50'
                      }`}
                      onClick={() => {
                        setSelectedGuide(guide);
                        setSelectedTable(null);
                        setSelectedNode(null);
                      }}
                    >
                      <span className="text-sm font-medium truncate" title={guide.title}>
                        {guide.title}
                      </span>
                      <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <span>{guide.product}</span>
                        <span>•</span>
                        <span>{guide.profileId || guide.shortId}</span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </TabsContent>
            
            <TabsContent value="tables" className="flex-1 overflow-auto mt-4">
              {tablesLoading ? (
                <div className="space-y-2">
                  <Skeleton className="h-12 w-full" />
                  <Skeleton className="h-12 w-full" />
                  <Skeleton className="h-12 w-full" />
                </div>
              ) : tablesError ? (
                <div className="text-sm text-destructive">Error loading tables</div>
              ) : tables.length === 0 ? (
                <div className="text-sm text-muted-foreground text-center py-8">
                  No tables available.<br />
                  Start import job to fetch tables.
                </div>
              ) : (
                <div className="space-y-1">
                  {tables.map((table) => (
                    <div
                      key={table.id}
                      className={`flex flex-col p-2 rounded cursor-pointer transition-colors ${
                        selectedTable?.id === table.id ? 'bg-muted' : 'hover:bg-muted/50'
                      }`}
                      onClick={() => {
                        setSelectedTable(table);
                        setSelectedGuide(null);
                        setSelectedNode(null);
                      }}
                    >
                      <div className="flex items-center gap-2">
                        <Table className="w-4 h-4 text-purple-500 shrink-0" />
                        <span className="text-sm font-medium truncate flex-1" title={table.title}>
                          {table.title}
                        </span>
                      </div>
                      <div className="flex items-center gap-2 text-xs text-muted-foreground ml-6">
                        <span>{table.product}</span>
                        <span>•</span>
                        <span>{table.tableType}</span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>

      {/* Right Panel: Tree View or Table View or Detail */}
      <Card className="flex-1 flex flex-col">
        {selectedNode ? (
          <>
            <CardHeader className="pb-2">
              <Button variant="ghost" size="sm" className="w-fit" onClick={() => setSelectedNode(null)}>
                ← Back to tree
              </Button>
            </CardHeader>
            <CardContent className="flex-1 min-h-0">
              <DetailPanel selectedNode={selectedNode} onClose={() => setSelectedNode(null)} />
            </CardContent>
          </>
        ) : selectedTable ? (
          <SelectedTableView
            table={selectedTable}
            onBack={() => setSelectedTable(null)}
          />
        ) : selectedGuide ? (
          <SelectedGuideTree
            guideId={selectedGuide.id}
            guideTitle={selectedGuide.title}
            onNodeSelect={setSelectedNode}
          />
        ) : (
          <div className="flex items-center justify-center h-full text-muted-foreground">
            Select a guide or table to view details
          </div>
        )}
      </Card>
    </div>
  );
}

// Component for displaying selected table view
function SelectedTableView({
  table,
  onBack,
}: {
  table: any;
  onBack: () => void;
}) {
  const [page, setPage] = useState(0);
  const pageSize = 100;
  const { data, isLoading, error } = useSSGTableEntries(table.id, page * pageSize, pageSize);

  const entries: any[] = (data && Array.isArray(data.entries)) ? data.entries : [];
  const total: number = data?.total || 0;
  const totalPages = Math.ceil(total / pageSize);

  return (
    <>
      <CardHeader className="pb-2">
        <Button variant="ghost" size="sm" className="w-fit" onClick={onBack}>
          ← Back to list
        </Button>
        <CardTitle className="text-base">{table.title}</CardTitle>
        <CardDescription>
          {table.product} • {table.tableType}
        </CardDescription>
      </CardHeader>
      <CardContent className="flex-1 min-h-0 flex flex-col">
        {isLoading ? (
          <div className="space-y-2">
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-8 w-full" />
          </div>
        ) : error ? (
          <div className="text-sm text-destructive">Error loading table entries</div>
        ) : entries.length === 0 ? (
          <div className="text-sm text-muted-foreground text-center py-8">
            No entries available
          </div>
        ) : (
          <>
            <div className="flex-1 overflow-auto border rounded-md">
              <table className="w-full text-sm">
                <thead className="sticky top-0 bg-muted">
                  <tr>
                    <th className="p-2 text-left font-medium">Mapping</th>
                    <th className="p-2 text-left font-medium">Rule Title</th>
                    <th className="p-2 text-left font-medium">Description</th>
                  </tr>
                </thead>
                <tbody>
                  {entries.map((entry) => (
                    <tr key={entry.id} className="border-t hover:bg-muted/50">
                      <td className="p-2 font-mono text-xs align-top">{entry.mapping}</td>
                      <td className="p-2 align-top">{entry.ruleTitle}</td>
                      <td className="p-2 text-xs text-muted-foreground align-top">{entry.description.substring(0, 200)}{entry.description.length > 200 ? '...' : ''}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex items-center justify-between mt-4 pt-4 border-t">
                <div className="text-sm text-muted-foreground">
                  Page {page + 1} of {totalPages} ({total} entries)
                </div>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={page === 0}
                    onClick={() => setPage(page - 1)}
                  >
                    Previous
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={page >= totalPages - 1}
                    onClick={() => setPage(page + 1)}
                  >
                    Next
                  </Button>
                </div>
              </div>
            )}
          </>
        )}
      </CardContent>
    </>
  );
}

// Component for displaying selected guide tree
function SelectedGuideTree({
  guideId,
  guideTitle,
  onNodeSelect,
}: {
  guideId: string;
  guideTitle: string;
  onNodeSelect: (node: TreeNode) => void;
}) {
  const { data, isLoading, error } = useSSGTree(guideId);

  const treeNodes: TreeNode[] = data?.tree?.nodes || [];

  return (
    <>
      <CardHeader className="pb-2">
        <CardTitle className="text-base">{guideTitle}</CardTitle>
        <CardDescription>Tree structure of groups and rules</CardDescription>
      </CardHeader>
      <CardContent className="flex-1 min-h-0 overflow-auto">
        {isLoading ? (
          <div className="space-y-2">
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-8 w-full" />
          </div>
        ) : error ? (
          <div className="text-sm text-destructive">Error loading tree</div>
        ) : treeNodes.length === 0 ? (
          <div className="text-sm text-muted-foreground text-center py-8">
            No tree structure available
          </div>
        ) : (
          <div className="border rounded-md p-2">
            {treeNodes.map((node) => (
              <TreeViewNode key={node.id} node={node} onNodeClick={onNodeSelect} />
            ))}
          </div>
        )}
      </CardContent>
    </>
  );
}
