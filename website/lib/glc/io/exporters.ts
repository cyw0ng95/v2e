/**
 * GLC Graph Exporters
 *
 * Export graphs to various formats: PNG, SVG, JSON
 */

import { toSvg, toPng } from 'html-to-image';
import type { Graph } from '../types';

interface ExportOptions {
  backgroundColor?: string;
  pixelRatio?: number;
  quality?: number;
}

/**
 * Export canvas to PNG
 */
export async function exportToPng(
  element: HTMLElement,
  options: ExportOptions = {}
): Promise<Blob> {
  const { backgroundColor = '#ffffff', pixelRatio = 2, quality = 0.95 } = options;

  const dataUrl = await toPng(element, {
    backgroundColor,
    pixelRatio,
    quality,
    filter: (node: Node) => {
      // Skip controls, minimap, and other UI elements
      if (node instanceof HTMLElement) {
        const classes = node.className;
        if (typeof classes === 'string') {
          return !classes.includes('react-flow__controls') &&
                 !classes.includes('react-flow__minimap') &&
                 !classes.includes('react-flow__panel');
        }
      }
      return true;
    },
  });

  const response = await fetch(dataUrl);
  return response.blob();
}

/**
 * Export canvas to SVG
 */
export async function exportToSvg(
  element: HTMLElement,
  options: ExportOptions = {}
): Promise<Blob> {
  const { backgroundColor = '#ffffff' } = options;

  const svgDataUrl = await toSvg(element, {
    backgroundColor,
    filter: (node: Node) => {
      if (node instanceof HTMLElement) {
        const classes = node.className;
        if (typeof classes === 'string') {
          return !classes.includes('react-flow__controls') &&
                 !classes.includes('react-flow__minimap') &&
                 !classes.includes('react-flow__panel');
        }
      }
      return true;
    },
  });

  // Convert data URL to blob
  const response = await fetch(svgDataUrl);
  return response.blob();
}

/**
 * Export graph to JSON
 */
export function exportToJson(graph: Graph): string {
  return JSON.stringify(graph, null, 2);
}

/**
 * Download blob as file
 */
export function downloadBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

/**
 * Download graph as JSON file
 */
export function downloadGraphJson(graph: Graph): void {
  const json = exportToJson(graph);
  const blob = new Blob([json], { type: 'application/json' });
  downloadBlob(blob, `${graph.metadata.name || 'graph'}.json`);
}

/**
 * Download canvas as PNG
 */
export async function downloadCanvasPng(
  element: HTMLElement,
  filename: string,
  options?: ExportOptions
): Promise<void> {
  const blob = await exportToPng(element, options);
  downloadBlob(blob, filename);
}

/**
 * Download canvas as SVG
 */
export async function downloadCanvasSvg(
  element: HTMLElement,
  filename: string,
  options?: ExportOptions
): Promise<void> {
  const blob = await exportToSvg(element, options);
  downloadBlob(blob, filename);
}
