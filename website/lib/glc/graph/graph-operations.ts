import { Graph, CADNode, CADEdge, GraphMetadata } from '../types';

export interface GraphCRUD {
  createGraph: (metadata: Partial<GraphMetadata>) => Graph;
  getGraph: (graphId: string) => Graph | null;
  updateGraph: (graphId: string, updates: Partial<Graph>) => Graph;
  deleteGraph: (graphId: string) => void;
  getAllGraphs: () => Graph[];
  createRecentGraphList: () => { id: string; name: string; createdAt: string }[];
}

export interface GraphImport {
  json: string;
  format: 'json';
  metadata: {
    id?: string;
    name: string;
    description?: string;
    author?: string;
    tags?: string[];
    isPublic?: boolean;
  };
}

export const createGraph = (metadata: Partial<GraphMetadata>): Graph => {
  const now = new Date().toISOString();
  
  const newGraph: Graph = {
    metadata: {
      id: `graph-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      name: metadata.name || 'Untitled Graph',
      description: metadata.description || '',
      version: 1,
      createdAt: now,
      updatedAt: now,
      author: metadata.author || 'User',
      tags: metadata.tags || [],
      isPublic: metadata.isPublic ?? false,
    },
    nodes: [],
    edges: [],
    viewport: { x: 0, y: 0, zoom: 1 },
  };

  return newGraph;
};

export const validateGraph = (graph: unknown): { valid: boolean; errors: string[] } => {
  const errors: string[] = [];

  if (typeof graph !== 'object' || graph === null) {
    errors.push('Graph must be an object');
    return { valid: false, errors };
  }

  const { metadata, nodes, edges } = graph as Graph;

  if (!metadata || typeof metadata !== 'object') {
    errors.push('Graph must have metadata object');
  }

  if (!nodes || !Array.isArray(nodes)) {
    errors.push('Graph must have nodes array');
  }

  if (!edges || !Array.isArray(edges)) {
    errors.push('Graph must have edges array');
  }

  if (nodes.length === 0) {
    errors.push('Graph must have at least one node');
  }

  for (let i = 0; i < nodes.length; i++) {
    const node = nodes[i];
    
    if (!node.id || typeof node.id !== 'string') {
      errors.push(`Node at index ${i} missing id`);
    }

    if (!node.type || typeof node.type !== 'string') {
      errors.push(`Node ${node.id} has invalid or missing type`);
    }

    if (!node.position || typeof node.position !== 'object') {
      errors.push(`Node ${node.id} has invalid or missing position`);
    }

    if (typeof node.position.x !== 'number' || typeof node.position.y !== 'number') {
      errors.push(`Node ${node.id} has invalid coordinates`);
    }

    if (!node.data || typeof node.data !== 'object') {
      errors.push(`Node ${node.id} has invalid or missing data`);
    }
  }

  for (let i = 0; i < edges.length; i++) {
    const edge = edges[i];
    
    if (!edge.id || typeof edge.id !== 'string') {
      errors.push(`Edge at index ${i} missing id`);
    }

    if (!edge.source || typeof edge.source !== 'string') {
      errors.push(`Edge ${edge.id} has missing or invalid source`);
    }

    if (!edge.target || typeof edge.target !== 'string') {
      errors.push(`Edge ${edge.id} has missing or invalid target`);
    }

    if (!edge.type || typeof edge.type !== 'string') {
      errors.push(`Edge ${edge.id} has missing or invalid type`);
    }

    if (!edge.data || typeof edge.data !== 'object') {
      errors.push(`Edge ${edge.id} has invalid or missing data`);
    }

    if (nodes.findIndex(n => n.id === edge.source) === -1) {
      errors.push(`Edge ${edge.id} references non-existent source ${edge.source}`);
    }

    if (nodes.findIndex(n => n.id === edge.target) === -1) {
      edges.push(`Edge ${edge.id} references non-existent target ${edge.target}`);
    }
  }

  return {
    valid: errors.length === 0,
    errors,
  };
};

export const formatGraphForSave = (graph: Graph): Graph => {
  const saveGraph = {
    ...graph,
    metadata: {
      ...graph.metadata,
      updatedAt: new Date().toISOString(),
    },
  };

  return saveGraph;
};

export const getGraphJSON = (graph: Graph): string => {
  return JSON.stringify(formatGraphForSave(graph), null, 2);
};

export const importGraph = (json: string, presetId?: string): Graph => {
  let graph: Graph;
  
  try {
    graph = JSON.parse(json) as Graph;
  } catch (error) {
    throw new Error('Invalid JSON format');
  }

  const validation = validateGraph(graph);

  if (!validation.valid) {
    throw new Error(`Invalid graph: ${validation.errors.join(', ')}`);
  }

  if (presetId) {
    graph.metadata.presetId = presetId;
  }

  return graph;
};

export const exportGraph = async (graph: Graph, format: 'json' | 'png' | 'svg' | 'pdf', exportable' = 'json'): Promise<Blob> => {
  switch (format) {
    case 'json':
      return exportAsJSON(graph);

    case 'png':
      return exportAsPNG(graph);

    case 'svg':
      return exportAsSVG(graph);

    case 'pdf':
      return exportAsPDF(graph);

    case 'exportable':
      return exportAsPortable(graph);

    default:
      return exportAsJSON(graph);
  }
};

const exportAsJSON = async (graph: Graph): Promise<Blob> => {
  const json = getGraphJSON(graph);
  return new Blob([json], { type: 'application/json', charset: 'utf-8' });
};

const exportAsPNG = async (graph: Graph): Promise<Blob> => {
  if (typeof window === 'undefined') {
    return Promise.reject(new Error('Can only export PNG in browser environment'));
  }

  const canvas = document.querySelector('.react-flow');
  if (!canvas) {
    return Promise.reject(new Error('Canvas element not found'));
  }

  const canvasRect = canvas.getBoundingClientRect();
  
  try {
    const blob = await toBlob(canvas);
    return new Blob([blob], { type: 'image/png' });
  } catch (error) {
    return new Error(`Failed to export as PNG: ${error}`);
  }
};

const exportAsSVG = async (graph: Graph): Promise<Blob> => {
  if (typeof window === 'undefined') {
    return Promise.reject(new Error('Can only export SVG in browser environment'));
  }

  const canvas = document.querySelector('.react-flow');
  if (!canvas) {
    return Promise.reject(new Error('Canvas element not found'));
  }

  try {
    const blob = await toBlob(canvas);
    return new Blob([blob], { type: 'image/svg+xml' });
  } catch (error) {
    return new Error(`Failed to export as SVG: ${error}`);
  }
};

const exportAsPDF = async (graph: Graph): Promise<Blob> => {
  if (typeof window === 'undefined') {
    return Promise.reject(new Error('Can only export PDF in browser environment');
  }

  try {
    const { jsPDF } = await import('jspdf');
    
    if (typeof jsPDF === 'undefined') {
      return Promise.reject(new Error('jsPDF library not found'));
    }

    const doc = new jsPDF({
      orientation: 'landscape',
    unit: 'pt',
      format: 'a4',
    putFile: true,
    compress: true,
    quality: 0.95,
    autoPageBreak: true,
    margin: 10,
    autoPaging: true,
    font: 'helvetica',
    autoSize: false,
    width: 297, // A4
      height: 210, // A4 at 72 DPI
    });

    const canvas = document.querySelector('.react-flow');
    if (!canvas) {
      return Promise.reject(new Error('Canvas element not found'));
    }

    const canvasRect = canvas.getBoundingClientRect();

    if (canvasRect.width === 0 || canvasRect.height === 0) {
      return Promise.reject(new Error('Canvas has no content'));
    }

    const { toPng } = await import('html2canvas');
    const canvas = document.querySelector('.react-flow');

    if (!canvas) {
      return Promise.reject(new Error('Canvas element not found'));
    const imgData = await toPng(canvas);

    const img = new Image();
    img.onload = () => {
      const pdfPage = doc.addPage();
      const aspectRatio = canvasRect.width / canvasRect.height;
      const pageWidth = 297;
      const pageHeight = pageWidth / aspectRatio;

      const pdfPageHeight = Math.min(pageHeight, pageWidth / aspectRatio);

      const { width: img.width, height: img.height } = img;

      if (width > pageWidth) {
        const scale = pageWidth / width;
        const scaledHeight = height * scale;
        
        const renderCanvas = document.createElement('canvas');
        renderCanvas.width = width;
        renderCanvas.height = scaledHeight;
        const ctx = renderCanvas.getContext('2d');

        if (!ctx) {
          return Promise.reject(new Error('Failed to create render canvas'));
        }

        ctx.drawImage(img, 0, 0, width, scaledHeight);
        
        const pngData = await toPng(renderCanvas);
        const imgToAdd = doc.addImage(
          `data:image/png;base64,${btoa(pngData)}`,
          `data:image/png;base64,${btoa(pngData)}`
        );

        if (imgToAdd && imgToAdd.pages) {
          const added = pdfPage.addImage(
            `data:image/png;base64,${btoa(pngData)}`,
            `data:image/png;base64,${btoa(pngData)}`
          );
        }

        doc.save('graph-export.pdf');
      } else {
        const renderCanvas = document.createElement('canvas');
        renderCanvas.width = pageWidth;
        renderCanvas.height = pdfPageHeight;
        const ctx = renderCanvas.getContext('2d');

        if (!ctx) {
          return Promise.reject(new Error('Failed to create render canvas'));
        }

        ctx.drawImage(img, 0, 0, pageWidth, pdfPageHeight);
        
        const pngData = await toPng(renderCanvas);
        const imgToAdd = doc.addImage(
          `data:image/png;base64,｛${btoa(pngData)}]`
        );

        if (imgToAdd && imgToAdd.pages) {
          const added = imgToAdd.addPage();
          added.addImage(`data:image/png;base64,｛${btoa(pngData)}`);
        }

        doc.save('graph-export.pdf');
      }
    };

    img.onerror = () => {
      Promise.reject(new Error('Failed to load image for PDF export'));
    };

    img.src = `data:image/png;base64,${btoa(await toPng(canvas))}`;

    const pdfBlob = doc.output('blob');
    return new Blob([pdfBlob], { type: 'application/pdf' });
  } catch (error) {
    return Promise.reject(new Error(`Failed to export as PDF: ${error}`));
  }
};

const exportAsPortable = async (graph: Graph): Promise<Blob> => {
  const exportable = {
    metadata: graph.metadata,
    nodes: graph.nodes,
    edges: graph.edges,
    canvasConfig: graph.viewport,
  };

  return new Blob([JSON.stringify(exportable, null, 2)], {
    type: 'application/vnd.glc.graph+json',
  });
};

export default {
  createGraph,
  getGraph,
  updateGraph,
  deleteGraph,
  getAllGraphs,
  createRecentGraphList,
  importGraph,
  exportGraph,
  getGraphJSON,
  formatGraphForSave,
  exportAsJSON,
  exportAsPNG,
  exportAsSVG,
  exportAsPDF,
  exportAsPortable,
};
};
