"use client";

import React from "react";

type AccordionProps = {
  children: React.ReactNode;
  multiple?: boolean;
};

export function Accordion({ children }: AccordionProps) {
  return <div className="space-y-2">{children}</div>;
}

export function AccordionItem({ children }: { children: React.ReactNode }) {
  return <div className="border rounded-md overflow-hidden">{children}</div>;
}

export function AccordionTrigger({ children, onClick, open }: { children: React.ReactNode; onClick?: () => void; open?: boolean }) {
  return (
    <button
      onClick={onClick}
      className="w-full text-left px-3 py-2 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 flex items-center justify-between"
    >
      <span>{children}</span>
      <span className="ml-2 text-sm">{open ? "▾" : "▸"}</span>
    </button>
  );
}

export function AccordionContent({ children, open }: { children: React.ReactNode; open?: boolean }) {
  return <div className={`px-3 py-2 ${open ? "block" : "hidden"}`}>{children}</div>;
}

export default Accordion;
