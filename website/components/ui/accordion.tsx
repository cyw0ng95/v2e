"use client";

import React from "react";
import { cn } from "@/lib/utils";

type AccordionProps = {
  children: React.ReactNode;
  multiple?: boolean;
  type?: 'single' | 'multiple';
  defaultValue?: string[];
};

export function Accordion({ children, type, defaultValue }: AccordionProps) {
  return <div className="space-y-2">{children}</div>;
}

export function AccordionItem({ children, value }: { children: React.ReactNode; value?: string }) {
  return <div className="border rounded-md overflow-hidden" data-value={value}>{children}</div>;
}

export function AccordionTrigger({ children, onClick, open, className }: { children: React.ReactNode; onClick?: () => void; open?: boolean; className?: string }) {
  return (
    <button
      onClick={onClick}
      className={cn("w-full text-left px-3 py-2 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 flex items-center justify-between", className)}
    >
      <span>{children}</span>
      <span className="ml-2 text-sm">{open ? "▾" : "▸"}</span>
    </button>
  );
}

export function AccordionContent({ children, open, className }: { children: React.ReactNode; open?: boolean; className?: string }) {
  return open ? <div className={cn("px-3 py-2", className)}>{children}</div> : null;
}

export default Accordion;
