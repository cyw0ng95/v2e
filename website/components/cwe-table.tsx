import React, { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useCWEList } from "@/lib/hooks";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import BookmarkStar from "@/components/bookmark-star";

const PAGE_SIZE = 10;

interface CWETableProps {
  viewMode?: 'view' | 'learn';
}

export function CWETable({ viewMode = 'view' }: CWETableProps) {
  const [page, setPage] = useState(0);
  const [search, setSearch] = useState("");
  // Fixed page size: show 20 items per page
  const [pageSize, setPageSize] = useState(20);
  const [detailCWE, setDetailCWE] = useState<any | null>(null);
  const tableRef = React.useRef<HTMLDivElement>(null);

  const { data, isLoading } = useCWEList({ offset: page * pageSize, limit: pageSize, search });
  // Map backend CWE items to a plain object for table display
  const cweList = Array.isArray(data?.cwes)
    ? data.cwes.map((item: any) => ({
        id: item.id || item.ID || item.cweId || item.CWEID || item.CweId || '',
        name: item.name || item.Name || '',
        abstraction: item.abstraction || item.Abstraction || '',
        status: item.status || item.Status || '',
        description: item.description || item.Description || '',
      }))
    : [];
  const total = data?.total || 0;

  // Helper to find the original full CWE object from the backend response by various id fields
  const findOriginalCWE = (id: string) => {
    if (!Array.isArray(data?.cwes)) return null;
    return data.cwes.find((item: any) => {
      const candidate = (item.id || item.ID || item.cweId || item.CWEID || item.CweId || "").toString();
      return candidate === id?.toString();
    }) || null;
  };

  return (
    <Card className="h-full flex flex-col" ref={tableRef}>
      <CardHeader>
        <CardTitle>CWE Database</CardTitle>
        <CardDescription>Browse and manage CWE records in the local database</CardDescription>
        <div className="mt-3">
          <Input
            className="w-full"
            placeholder="Search CWE ID or name"
            value={search}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              setSearch(e.target.value);
              setPage(0);
            }}
          />
        </div>
      </CardHeader>
      <CardContent className="flex-1 min-h-0 overflow-auto">
        {isLoading ? (
          <Skeleton className="h-32 w-full" />
        ) : (
          <>
            <table className="min-w-full text-xs">
              <thead>
                <tr className="border-b">
                  <th className="text-left p-2">ID</th>
                  <th className="text-left p-2">Name</th>
                  <th className="text-left p-2">Abstraction</th>
                  <th className="text-left p-2">Status</th>
                  <th className="text-left p-2">Description</th>
                  <th className="text-left p-2">Action</th>
                </tr>
              </thead>
              <tbody>
                {/* Use the mapped plain object, not CWEItem type */}
                {cweList.map((cwe: any, idx: number) => (
                  <tr key={cwe.id || idx} className="border-b hover:bg-muted/30">
                    <td className="p-2 font-mono">{cwe.id}</td>
                    <td className="p-2">{cwe.name}</td>
                    <td className="p-2">{cwe.abstraction}</td>
                    <td className="p-2">{cwe.status}</td>
                    <td className="p-2 max-w-xs truncate" title={cwe.description}>{cwe.description}</td>
                    <td className="p-2">
                      <button
                        className="px-2 py-1 border rounded text-xs hover:bg-muted cursor-pointer disabled:cursor-not-allowed"
                        onClick={() => setDetailCWE(findOriginalCWE(cwe.id) || cwe)}
                      >
                        View Detail
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {/* Detail Modal */}
            {detailCWE && (
              <>
                <style>{`@keyframes v2e-fade-in { from { opacity: 0 } to { opacity: 1 } } @keyframes v2e-pop-in { from { opacity: 0; transform: translateY(-6px) scale(0.98); } to { opacity: 1; transform: translateY(0) scale(1); } }`}</style>
                <div
                  role="dialog"
                  aria-modal="true"
                  className="fixed inset-0 z-50 flex items-center justify-center p-4"
                  style={{ animation: 'v2e-fade-in 160ms ease-out both' }}
                >
                  {/* Mask underneath the modal content to capture/block clicks */}
                  <div
                    className="absolute inset-0 bg-black/40"
                    onClick={() => setDetailCWE(null)}
                    aria-hidden="true"
                  />

                  <div
                    className="relative max-w-4xl w-full bg-background rounded-lg shadow-lg p-6 overflow-auto max-h-[85vh]"
                    onClick={e => e.stopPropagation()}
                    style={{ animation: 'v2e-pop-in 180ms cubic-bezier(.2,.9,.2,1) both' }}
                  >
                    <Button
                      variant="ghost"
                      className="absolute top-4 right-4"
                      onClick={() => setDetailCWE(null)}
                    >
                      <span className="sr-only">Close</span>
                      &times;
                    </Button>
                    <div className="mb-4 flex items-start justify-between gap-4">
                      <div>
                        <h2 className="text-lg font-medium">CWE-{detailCWE.id || detailCWE.ID}: {detailCWE.name || detailCWE.Name}</h2>
                        <div className="mt-1 flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
                          <span><b>Abstraction:</b> {detailCWE.abstraction || detailCWE.Abstraction}</span>
                          <span><b>Status:</b> {detailCWE.status || detailCWE.Status}</span>
                          <span><b>Likelihood:</b> {detailCWE.likelihoodOfExploit || detailCWE.LikelihoodOfExploit}</span>
                          <span><b>Structure:</b> {detailCWE.structure || detailCWE.Structure}</span>
                          <span><b>ID:</b> {detailCWE.id || detailCWE.ID}</span>
                        </div>
                      </div>
                      <div className="text-sm text-muted-foreground flex flex-col items-end gap-2">
                        <div><b>Ordinalities:</b> {Array.isArray(detailCWE.weaknessOrdinalities || detailCWE.WeaknessOrdinalities) && (detailCWE.weaknessOrdinalities || detailCWE.WeaknessOrdinalities).map((wo: any, i: number) => <span key={i} className="inline-block mr-2 bg-muted px-2 py-0.5 rounded text-xs">{wo.ordinality || wo.Ordinality}</span>)}</div>
                        <BookmarkStar 
                          itemId={detailCWE.id || detailCWE.ID}
                          itemType="CWE"
                          itemTitle={`CWE-${detailCWE.id || detailCWE.ID}: ${detailCWE.name || detailCWE.Name}`}
                          itemDescription={detailCWE.description || detailCWE.Description || ''}
                          viewMode={viewMode}
                        />
                      </div>
                    </div>
                    <section className="mb-4">
                      <h3 className="font-semibold mb-2">Description</h3>
                      <p className="text-sm whitespace-pre-wrap">{detailCWE.description || detailCWE.Description}</p>
                    </section>
                    {detailCWE.extendedDescription || detailCWE.ExtendedDescription ? (
                      <section className="mb-4">
                        <h3 className="font-semibold mb-2">Extended Description</h3>
                        <p className="text-sm whitespace-pre-wrap">{detailCWE.extendedDescription || detailCWE.ExtendedDescription}</p>
                      </section>
                    ) : null}
                    {/* Demonstrative Examples */}
                    {Array.isArray(detailCWE.demonstrativeExamples || detailCWE.DemonstrativeExamples) && (detailCWE.demonstrativeExamples || detailCWE.DemonstrativeExamples).length > 0 && (
                      <details className="mb-4">
                        <summary className="font-semibold mb-2 cursor-pointer">Demonstrative Examples</summary>
                        <div className="mt-2">
                          {(detailCWE.demonstrativeExamples || detailCWE.DemonstrativeExamples).map((ex: any, i: number) => (
                            <div key={i} className="border rounded p-3 my-2 bg-muted/30">
                              {Array.isArray(ex.entries || ex.Entries) && (ex.entries || ex.Entries).map((entry: any, j: number) => (
                                <div key={j} className="mb-2">
                                  {(entry.introText || entry.IntroText) && <div className="text-xs mb-1 font-semibold">{entry.introText || entry.IntroText}</div>}
                                  {(entry.bodyText || entry.BodyText) && <div className="text-xs mb-1">{entry.bodyText || entry.BodyText}</div>}
                                  {(entry.exampleCode || entry.ExampleCode) && (
                                    <pre className="bg-zinc-100 dark:bg-zinc-800 rounded p-2 text-xs overflow-x-auto mb-1 whitespace-pre-wrap"><code>{entry.exampleCode || entry.ExampleCode}</code></pre>
                                  )}
                                  {(entry.language || entry.Language) && <span className="text-xs mr-2">Lang: {entry.language || entry.Language}</span>}
                                  {(entry.nature || entry.Nature) && <span className="text-xs">Type: {entry.nature || entry.Nature}</span>}
                                </div>
                              ))}
                            </div>
                          ))}
                        </div>
                      </details>
                    )}

                    {/* Observed Examples */}
                    {Array.isArray(detailCWE.observedExamples || detailCWE.ObservedExamples) && (detailCWE.observedExamples || detailCWE.ObservedExamples).length > 0 && (
                      <details className="mb-4">
                        <summary className="font-semibold mb-2 cursor-pointer">Observed Examples</summary>
                        <div className="mt-2">
                          <ul className="ml-4 list-disc text-sm">
                            {(detailCWE.observedExamples || detailCWE.ObservedExamples).map((ex: any, i: number) => (
                              <li key={i} className="mb-1">
                                {(ex.description || ex.Description)} {(ex.reference || ex.Reference) && (<a href={ex.link || ex.Link} target="_blank" rel="noopener noreferrer" className="underline text-blue-600">[{ex.reference || ex.Reference}]</a>)}
                              </li>
                            ))}
                          </ul>
                        </div>
                      </details>
                    )}

                    {/* Detection Methods */}
                    {Array.isArray(detailCWE.detectionMethods || detailCWE.DetectionMethods) && (detailCWE.detectionMethods || detailCWE.DetectionMethods).length > 0 && (
                      <details className="mb-4">
                        <summary className="font-semibold mb-2 cursor-pointer">Detection Methods</summary>
                        <div className="mt-2">
                          <ul className="ml-4 list-disc text-sm">
                            {(detailCWE.detectionMethods || detailCWE.DetectionMethods).map((dm: any, i: number) => (
                              <li key={i} className="mb-1">
                                <b>{dm.method || dm.Method}</b>: {dm.description || dm.Description} {(dm.effectiveness || dm.Effectiveness) && (<span className="ml-2">(Effectiveness: {dm.effectiveness || dm.Effectiveness})</span>)}
                              </li>
                            ))}
                          </ul>
                        </div>
                      </details>
                    )}

                    {/* Potential Mitigations */}
                    {Array.isArray(detailCWE.potentialMitigations || detailCWE.PotentialMitigations) && (detailCWE.potentialMitigations || detailCWE.PotentialMitigations).length > 0 && (
                      <details className="mb-4">
                        <summary className="font-semibold mb-2 cursor-pointer">Potential Mitigations</summary>
                        <div className="mt-2">
                          <ul className="ml-4 list-disc text-sm">
                            {(detailCWE.potentialMitigations || detailCWE.PotentialMitigations).map((mit: any, i: number) => (
                              <li key={i} className="mb-1">
                                {mit.description || mit.Description} {(mit.effectiveness || mit.Effectiveness) && (<span className="ml-2">(Effectiveness: {mit.effectiveness || mit.Effectiveness})</span>)}
                                {(mit.effectivenessNotes || mit.EffectivenessNotes) && (<div className="text-xs text-muted-foreground">{mit.effectivenessNotes || mit.EffectivenessNotes}</div>)}
                              </li>
                            ))}
                          </ul>
                        </div>
                      </details>
                    )}

                    {/* Content History */}
                    {(() => {
                      // Robustly handle all possible casings and aliases for ContentHistory
                      const contentHistory = detailCWE.contentHistory || detailCWE.ContentHistory || detailCWE.content_history || detailCWE['Content_History'] || [];
                      return Array.isArray(contentHistory) && contentHistory.length > 0 ? (
                        <details className="mb-4">
                          <summary className="font-semibold mb-2 cursor-pointer">Content History</summary>
                          <div className="mt-2">
                            <ul className="ml-4 list-disc text-sm">
                              {contentHistory.map((h: any, i: number) => (
                                <li key={i} className="mb-2">
                                  <div className="text-xs">
                                    <b>{h.type || h.Type}</b> {(h.submissionDate || h.SubmissionDate) && `on ${h.submissionDate || h.SubmissionDate}`} {(h.modificationDate || h.ModificationDate) && `on ${h.modificationDate || h.ModificationDate}`}
                                  </div>
                                  {(h.submissionName || h.SubmissionName) && <div className="text-xs">By: {h.submissionName || h.SubmissionName} ({h.submissionOrganization || h.SubmissionOrganization})</div>}
                                  {(h.modificationName || h.ModificationName) && <div className="text-xs">By: {h.modificationName || h.ModificationName} ({h.modificationOrganization || h.ModificationOrganization})</div>}
                                  {(h.modificationComment || h.ModificationComment) && <div className="text-xs italic">{h.modificationComment || h.ModificationComment}</div>}
                                </li>
                              ))}
                            </ul>
                          </div>
                        </details>
                      ) : null;
                    })()}

                    {/* Related Weaknesses */}
                    {Array.isArray(detailCWE.relatedWeaknesses || detailCWE.RelatedWeaknesses) && (detailCWE.relatedWeaknesses || detailCWE.RelatedWeaknesses).length > 0 && (
                      <details className="mb-4">
                        <summary className="font-semibold mb-2 cursor-pointer">Related Weaknesses</summary>
                        <div className="mt-2">
                          <ul className="ml-4 list-disc text-sm">
                            {(detailCWE.relatedWeaknesses || detailCWE.RelatedWeaknesses).map((rw: any, i: number) => (
                              <li key={i} className="mb-1">
                                {rw.cweID || rw.CweID || rw.cweId} ({rw.nature || rw.Nature}{(rw.ordinal || rw.Ordinal) ? `, ${rw.ordinal || rw.Ordinal}` : ''}{(rw.viewID || rw.ViewID) ? `, View: ${rw.viewID || rw.ViewID}` : ''})
                              </li>
                            ))}
                          </ul>
                        </div>
                      </details>
                    )}

                    {/* Taxonomy Mappings */}
                    {Array.isArray(detailCWE.taxonomyMappings || detailCWE.TaxonomyMappings) && (detailCWE.taxonomyMappings || detailCWE.TaxonomyMappings).length > 0 && (
                      <details className="mb-4">
                        <summary className="font-semibold mb-2 cursor-pointer">Taxonomy Mappings</summary>
                        <div className="mt-2">
                          <ul className="ml-4 list-disc text-sm">
                            {(detailCWE.taxonomyMappings || detailCWE.TaxonomyMappings).map((tm: any, i: number) => (
                              <li key={i} className="mb-1">
                                {tm.entryName || tm.EntryName} {(tm.entryID || tm.EntryID) && <span className="text-xs">({tm.entryID || tm.EntryID})</span>} {(tm.taxonomyName || tm.TaxonomyName) && <span className="text-xs">[{tm.taxonomyName || tm.TaxonomyName}]</span>}
                              </li>
                            ))}
                          </ul>
                        </div>
                      </details>
                    )}

                    {/* Notes */}
                    {Array.isArray(detailCWE.notes || detailCWE.Notes) && (detailCWE.notes || detailCWE.Notes).length > 0 && (
                      <details className="mb-4">
                        <summary className="font-semibold mb-2 cursor-pointer">Notes</summary>
                        <div className="mt-2">
                          <ul className="ml-4 list-disc text-sm">
                            {(detailCWE.notes || detailCWE.Notes).map((note: any, i: number) => (
                              <li key={i} className="mb-1">{note.note || note.Note}</li>
                            ))}
                          </ul>
                        </div>
                      </details>
                    )}
                    <details className="mt-4 text-sm">
                      <summary className="cursor-pointer">Raw JSON</summary>
                      <pre className="text-xs mt-2 overflow-x-auto bg-muted p-2 rounded">{JSON.stringify(detailCWE, null, 2)}</pre>
                    </details>
                  </div>
                </div>
              </>
            )}
            <div className="flex items-center justify-between mt-2 text-xs">
              <span>
                Showing {page * pageSize + 1}–{Math.min((page + 1) * pageSize, total)} of {total}
              </span>
              <div className="space-x-2">
                <button
                  className="px-2 py-1 border rounded disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
                  onClick={() => setPage(p => Math.max(0, p - 1))}
                  disabled={page === 0}
                >
                  Prev
                </button>
                <button
                  className="px-2 py-1 border rounded disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
                  onClick={() => setPage(p => (p + 1) * pageSize < total ? p + 1 : p)}
                  disabled={(page + 1) * pageSize >= total}
                >
                  Next
                </button>
              </div>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
}
