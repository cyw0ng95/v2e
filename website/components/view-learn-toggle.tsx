"use client";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Eye, BookOpen } from 'lucide-react';

interface ViewLearnToggleProps {
  value: "view" | "learn";
  onValueChange: (value: "view" | "learn") => void;
  className?: string;
}

export function ViewLearnToggle({
  value,
  onValueChange,
  className
}: ViewLearnToggleProps) {
  return (
    <div
      className={cn(
        "relative inline-flex rounded-lg bg-muted/40 p-0.5 border border-border/40 w-full",
        className
      )}
      role="radiogroup"
      aria-label="View mode toggle"
    >
      {/* Sliding indicator */}
      <div
        className={cn(
          "absolute top-0.5 bottom-0.5 left-0.5 rounded-md transition-all duration-300 ease-out",
          value === "view"
            ? "w-[calc(50%-2px)]"
            : "w-[calc(50%-2px)] translate-x-full"
        )}
        style={{
          backgroundColor: value === "view"
            ? "var(--primary)"
            : "var(--primary)",
          opacity: 0.08,
        }}
      />

      <Button
        variant="ghost"
        size="sm"
        className={cn(
          "flex-1 h-8 text-sm font-medium transition-all duration-200 relative z-10",
          value === "view"
            ? "text-foreground"
            : "text-muted-foreground/80 hover:text-foreground"
        )}
        onClick={() => onValueChange("view")}
        aria-checked={value === "view"}
        role="radio"
        tabIndex={value === "view" ? 0 : -1}
      >
        <Eye className="h-3.5 w-3.5 mr-1.5" />
        View
      </Button>

      <Button
        variant="ghost"
        size="sm"
        className={cn(
          "flex-1 h-8 text-sm font-medium transition-all duration-200 relative z-10",
          value === "learn"
            ? "text-foreground"
            : "text-muted-foreground/80 hover:text-foreground"
        )}
        onClick={() => onValueChange("learn")}
        aria-checked={value === "learn"}
        role="radio"
        tabIndex={value === "learn" ? 0 : -1}
      >
        <BookOpen className="h-3.5 w-3.5 mr-1.5" />
        Learn
      </Button>
    </div>
  );
}
