import React, { useState } from "react";
import { useCWEViews, useCWEJobStatus, useStartCWEViewJob, useStopCWEViewJob } from "../lib/hooks";
import { Button } from "./ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Skeleton } from "./ui/skeleton";

import { Badge } from "./ui/badge";
import { LucideEye } from "lucide-react";


export function CWEViews() {
  const { data, isLoading, error, refetch } = useCWEViews();
  const [selectedView, setSelectedView] = useState<any | null>(null);
  const [modalOpen, setModalOpen] = useState(false);

  // Job control hooks
  const { data: jobStatus, isLoading: jobLoading } = useCWEJobStatus();
  const startJob = useStartCWEViewJob();
  const stopJob = useStopCWEViewJob();

  // Defensive: extract views from payload shape
  let views: any[] = [];
  if (data && Array.isArray(data.views)) {
    views = data.views;
  }

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>CWE Views</CardTitle>
      </CardHeader>
      <CardContent>
        {/* Job Control Buttons */}
        <div className="flex items-center gap-2 mb-4">
          <Button
            variant="default"
            size="sm"
            disabled={jobLoading || jobStatus?.running}
            onClick={() => startJob.mutate({})}
          >
            Start Import
          </Button>
          <Button
            variant="secondary"
            size="sm"
            disabled={jobLoading || !jobStatus?.running}
            onClick={() => stopJob.mutate()}
          >
            Stop Import
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetch()}
          >
            Refresh
          </Button>
          {jobLoading ? (
            <span className="text-xs text-muted-foreground ml-2">Loading job status...</span>
          ) : jobStatus?.running ? (
            <span className="text-xs text-green-600 ml-2">Job running</span>
          ) : (
            <span className="text-xs text-muted-foreground ml-2">Idle</span>
          )}
        </div>
        {/* CWE View List */}
        <div className="h-96 w-full pr-2 overflow-auto">
          <div className="flex flex-col gap-2">
            {views.map((view: any) => (
              <div
                key={view.ID || view.id}
                className="flex items-center justify-between border rounded px-3 py-2 hover:bg-muted transition cursor-pointer"
                onClick={() => {
                  setSelectedView(view);
                  setModalOpen(true);
                }}
              >
                <div className="flex flex-col">
                  <span className="font-medium text-base">
                    {view.Name || view.name} <Badge variant="outline">{view.ID || view.id}</Badge>
                  </span>
                  <span className="text-muted-foreground text-xs">
                    {view.Objective || view.objective}
                  </span>
                  <span className="text-muted-foreground text-xs">
                    Status: {view.Status || view.status} | Type: {view.Type || view.type}
                  </span>
                </div>
                <Button variant="ghost" size="icon" tabIndex={-1}>
                  <LucideEye className="w-5 h-5" />
                </Button>
              </div>
            ))}
          </div>
        </div>
        {/* Integrated modal/detail view for CWE View */}
        {modalOpen && selectedView && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
            <div className="bg-background rounded-lg shadow-lg w-full max-w-2xl p-6 relative">
              <button
                className="absolute top-2 right-2 text-muted-foreground hover:text-foreground"
                onClick={() => setModalOpen(false)}
                aria-label="Close"
              >
                <span aria-hidden>×</span>
              </button>
              <h2 className="text-xl font-bold mb-2 flex items-center gap-2">
                {selectedView.Name || selectedView.name} <Badge variant="outline">{selectedView.ID || selectedView.id}</Badge>
              </h2>
              <div className="mb-2 text-muted-foreground text-sm">
                {selectedView.Objective || selectedView.objective}
              </div>
              <div className="mb-2 text-muted-foreground text-xs">
                Status: {selectedView.Status || selectedView.status} | Type: {selectedView.Type || selectedView.type}
              </div>
              {/* Members, etc. can be added here if present */}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
