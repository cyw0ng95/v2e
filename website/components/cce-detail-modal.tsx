import React from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { CopyIcon, ExternalLinkIcon } from "lucide-react";

interface CCEDetailModalProps {
  cce: {
    id: string;
    title: string;
    description: string;
    owner: string;
    status: string;
    type: string;
    reference: string;
    metadata: string;
  };
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CCEDetailModal({ cce, open, onOpenChange }: CCEDetailModalProps) {
  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl max-h-[90vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center justify-between">
            <span>{cce.id} - {cce.title}</span>
            <div className="flex gap-2">
              <Button
                size="sm"
                variant="outline"
                onClick={() => copyToClipboard(cce.id, "CCE ID")}
              >
                <CopyIcon className="h-4 w-4" />
                Copy ID
              </Button>
              {cce.reference && (
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => window.open(cce.reference, "_blank")}
                >
                  <ExternalLinkIcon className="h-4 w-4" />
                  Reference
                </Button>
              )}
            </div>
          </DialogTitle>
          <DialogDescription>
            CCE (Common Configuration Enumeration) details and metadata
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-auto px-4 py-4 max-h-[60vh]">
          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-semibold mb-2">Basic Information</h3>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="text-muted-foreground">ID:</span>
                  <span className="ml-2 font-mono">{cce.id}</span>
                </div>
                <div>
                  <span className="text-muted-foreground">Owner:</span>
                  <span className="ml-2">{cce.owner}</span>
                </div>
                <div>
                  <span className="text-muted-foreground">Status:</span>
                  <span className={`ml-2 ${cce.status === "ACTIVE" ? "text-green-600" : "text-orange-600"}`}>
                    {cce.status}
                  </span>
                </div>
                <div>
                  <span className="text-muted-foreground">Type:</span>
                  <span className="ml-2">{cce.type}</span>
                </div>
              </div>
            </div>

            <Separator />

            <div>
              <h3 className="text-sm font-semibold mb-2">Description</h3>
              <p className="text-sm text-muted-foreground whitespace-pre-wrap">
                {cce.description || "No description available"}
              </p>
            </div>

            {cce.reference && (
              <>
                <Separator />
                <div>
                  <h3 className="text-sm font-semibold mb-2">Reference</h3>
                  <a
                    href={cce.reference}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm text-blue-600 hover:underline break-all"
                  >
                    {cce.reference}
                  </a>
                </div>
              </>
            )}

            {cce.metadata && (
              <>
                <Separator />
                <div>
                  <h3 className="text-sm font-semibold mb-2">Metadata</h3>
                  <div className="bg-muted/50 p-3 rounded text-xs font-mono overflow-x-auto">
                    <pre>{cce.metadata}</pre>
                  </div>
                </div>
              </>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
