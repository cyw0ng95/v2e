"use client";

import * as React from "react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

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
        "inline-flex rounded-md shadow-xs bg-muted p-1 w-full",
        className
      )}
      role="radiogroup"
      aria-label="View mode toggle"
    >
      <Button
        variant={value === "view" ? "default" : "ghost"}
        size="sm"
        className={cn(
          "flex-1 h-8 text-sm font-medium transition-all duration-200",
          value === "view" 
            ? "shadow-sm" 
            : "hover:bg-accent hover:text-accent-foreground"
        )}
        onClick={() => onValueChange("view")}
        aria-checked={value === "view"}
        role="radio"
        tabIndex={value === "view" ? 0 : -1}
      >
        View
      </Button>
      <Button
        variant={value === "learn" ? "default" : "ghost"}
        size="sm"
        className={cn(
          "flex-1 h-8 text-sm font-medium transition-all duration-200",
          value === "learn" 
            ? "shadow-sm" 
            : "hover:bg-accent hover:text-accent-foreground"
        )}
        onClick={() => onValueChange("learn")}
        aria-checked={value === "learn"}
        role="radio"
        tabIndex={value === "learn" ? 0 : -1}
      >
        Learn
      </Button>
    </div>
  );
}