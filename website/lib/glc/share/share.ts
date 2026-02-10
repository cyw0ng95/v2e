/**
 * GLC Share & Embed Utilities
 *
 * URL encoding with compression for sharing graphs
 */

import { compressToEncodedURIComponent, decompressFromEncodedURIComponent } from 'lz-string';
import type { Graph } from '../types';

interface ShareData {
  graph: Graph;
  presetId: string;
  timestamp: number;
}

/**
 * Encode graph to URL-safe string
 */
export function encodeGraphToUrl(graph: Graph, presetId: string): string {
  const data: ShareData = {
    graph,
    presetId,
    timestamp: Date.now(),
  };

  const json = JSON.stringify(data);
  return compressToEncodedURIComponent(json);
}

/**
 * Decode graph from URL string
 */
export function decodeGraphFromUrl(encoded: string): ShareData | null {
  try {
    const json = decompressFromEncodedURIComponent(encoded);
    if (!json) return null;

    return JSON.parse(json) as ShareData;
  } catch {
    return null;
  }
}

/**
 * Generate share URL for graph
 */
export function generateShareUrl(graph: Graph, presetId: string, baseUrl?: string): string {
  const encoded = encodeGraphToUrl(graph, presetId);
  const base = baseUrl || (typeof window !== 'undefined' ? window.location.origin : '');

  return `${base}/glc/${presetId}?share=${encoded}`;
}

/**
 * Parse share URL and extract graph data
 */
export function parseShareUrl(url: string): ShareData | null {
  try {
    const urlObj = new URL(url);
    const shareParam = urlObj.searchParams.get('share');

    if (!shareParam) return null;

    return decodeGraphFromUrl(shareParam);
  } catch {
    return null;
  }
}

/**
 * Generate embed iframe code
 */
export function generateEmbedCode(graph: Graph, presetId: string, options: {
  width?: string;
  height?: string;
  baseUrl?: string;
} = {}): string {
  const { width = '800', height = '600', baseUrl } = options;
  const shareUrl = generateShareUrl(graph, presetId, baseUrl);

  return `<iframe src="${shareUrl}" width="${width}" height="${height}" frameborder="0" style="border: 1px solid #ccc; border-radius: 8px;"></iframe>`;
}

/**
 * Copy text to clipboard
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    // Fallback for older browsers
    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';
    textarea.style.opacity = '0';
    document.body.appendChild(textarea);
    textarea.select();

    try {
      document.execCommand('copy');
      return true;
    } catch {
      return false;
    } finally {
      document.body.removeChild(textarea);
    }
  }
}

/**
 * Share dialog component data
 */
export interface ShareDialogData {
  shareUrl: string;
  embedCode: string;
  graphName: string;
}
