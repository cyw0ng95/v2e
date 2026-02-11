import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { CCEDetailModal } from "@/components/cce-detail-modal";

interface CCE {
  id: string;
  title: string;
  description: string;
  owner: string;
  status: string;
  type: string;
  reference: string;
  metadata: string;
}

interface CCEListResponse {
  cces: CCE[];
  offset: number;
  limit: number;
  total: number;
}

const PAGE_SIZE = 20;

export function CCETable() {
  const [page, setPage] = useState(0);
  const [search, setSearch] = useState("");
  const [data, setData] = useState<CCEListResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [detailCCE, setDetailCCE] = useState<CCE | null>(null);

  const fetchCCEs = async () => {
    setIsLoading(true);
    try {
      const response = await fetch("/restful/rpc", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          method: "RPCListCCEs",
          target: "local",
          params: {
            offset: page * PAGE_SIZE,
            limit: PAGE_SIZE,
            search: search || undefined
          }
        })
      });

      const result = await response.json();
      if (result.success) {
        setData(result.payload);
      }
    } catch (error) {
      console.error("Failed to fetch CCEs:", error);
    } finally {
      setIsLoading(false);
    }
  };

  React.useEffect(() => {
    fetchCCEs();
  }, [page, search]);

  const totalPages = data ? Math.ceil(data.total / PAGE_SIZE) : 0;
  const cceList = data?.cces || [];

  return (
    <>
      <Card className="h-full flex flex-col">
        <CardHeader>
          <CardTitle>CCE Database</CardTitle>
          <CardDescription>Browse and manage CCE records in local database</CardDescription>
          <div className="mt-3">
            <Input
              className="w-full"
              placeholder="Search CCE ID or title"
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
                    <th className="text-left p-2">Title</th>
                    <th className="text-left p-2">Owner</th>
                    <th className="text-left p-2">Status</th>
                    <th className="text-left p-2">Type</th>
                    <th className="text-left p-2">Action</th>
                  </tr>
                </thead>
                <tbody>
                  {cceList.map((cce: CCE, idx: number) => (
                    <tr key={cce.id || idx} className="border-b hover:bg-muted/30">
                      <td className="p-2 font-mono">{cce.id}</td>
                      <td className="p-2">{cce.title}</td>
                      <td className="p-2">{cce.owner}</td>
                      <td className="p-2">{cce.status}</td>
                      <td className="p-2">{cce.type}</td>
                      <td className="p-2">
                        <Button
                          size="sm"
                          variant="outline"
                          className="text-xs"
                          onClick={() => setDetailCCE(cce)}
                        >
                          View Detail
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>

              {/* Pagination */}
              <div className="flex items-center justify-between mt-4 pt-4 border-t">
                <div className="text-xs text-muted-foreground">
                  Showing {Math.min(page * PAGE_SIZE + 1, data?.total || 0)} to{" "}
                  {Math.min((page + 1) * PAGE_SIZE, data?.total || 0)} of{" "}
                  {data?.total || 0}
                </div>
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => setPage(Math.max(0, page - 1))}
                    disabled={page === 0}
                  >
                    <ChevronLeft className="h-4 w-4" />
                    Previous
                  </Button>
                  <span className="text-sm flex items-center">
                    Page {page + 1} of {totalPages}
                  </span>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => setPage(Math.min(totalPages - 1, page + 1))}
                    disabled={page >= totalPages - 1}
                  >
                    Next
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* Detail Modal */}
      {detailCCE && (
        <CCEDetailModal
          cce={detailCCE}
          open={!!detailCCE}
          onOpenChange={(open) => !open && setDetailCCE(null)}
        />
      )}
    </>
  );
}
