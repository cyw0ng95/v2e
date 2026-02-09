import { CanvasPreset } from '../types';

export interface CanvasConfig {
  defaultViewport: { x: number; y: number; zoom: number };
  minZoom: number;
  maxZoom: number;
  snapToGrid: boolean;
  snapGrid: [number, number];
  nodesDraggable: boolean;
  nodesConnectable: boolean;
  elementsSelectable: boolean;
  panOnDrag: boolean;
  panOnScroll: boolean;
  zoomOnScroll: boolean;
  zoomOnPinch: boolean;
  zoomOnDoubleClick: boolean;
  preventScrolling: boolean;
  fitViewOnInit: boolean;
  deleteKeyCode: string;
  selectionKeyCode: string;
  multiSelectionKeyCode: string;
}

export const getCanvasConfig = (preset: CanvasPreset): CanvasConfig => {
  const { behavior } = preset;
  
  return {
    defaultViewport: { x: 0, y: 0, zoom: 1 },
    minZoom: 0.1,
    maxZoom: 4,
    snapToGrid: behavior.snapToGrid,
    snapGrid: [behavior.gridSize, behavior.gridSize],
    nodesDraggable: true,
    nodesConnectable: true,
    elementsSelectable: true,
    panOnDrag: true,
    panOnScroll: false,
    zoomOnScroll: true,
    zoomOnPinch: true,
    zoomOnDoubleClick: false,
    preventScrolling: false,
    fitViewOnInit: false,
    deleteKeyCode: 'Delete',
    selectionKeyCode: 'Shift',
    multiSelectionKeyCode: 'Meta',
  };
};

export const getCanvasBackground = (preset: CanvasPreset): string => {
  const { styling, behavior } = preset;
  
  if (behavior.snapToGrid) {
    return `radial-gradient(${styling.gridColor} 1px, transparent 1px) ${styling.gridSize}px`;
  }
  
  return styling.backgroundColor;
};

export const getCanvasStyles = (preset: CanvasPreset): Record<string, string> => {
  const { styling } = preset;
  
  return {
    '--canvas-bg': styling.backgroundColor,
    '--canvas-grid': styling.gridColor,
    '--canvas-primary': styling.primaryColor,
    '--canvas-font': styling.fontFamily,
  };
};

export const getNodeStyle = (preset: CanvasPreset, nodeType: string): Record<string, any> => {
  const nodeTypeDefinition = preset.nodeTypes.find(nt => nt.id === nodeType);
  
  if (!nodeTypeDefinition) {
    return {};
  }
  
  return {
    background: nodeTypeDefinition.style.backgroundColor,
    border: `${nodeTypeDefinition.style.borderWidth || 2}px solid ${nodeTypeDefinition.style.borderColor}`,
    color: nodeTypeDefinition.style.textColor,
    borderRadius: nodeTypeDefinition.style.borderRadius || 8,
    padding: nodeTypeDefinition.style.padding || '12px',
    fontFamily: preset.styling.fontFamily,
  };
};

export const getEdgeStyle = (preset: CanvasPreset, edgeType: string): Record<string, any> => {
  const edgeTypeDefinition = preset.relationshipTypes.find(rt => rt.id === edgeType);
  
  if (!edgeTypeDefinition) {
    return {};
  }
  
  const style: Record<string, any> = {
    stroke: edgeTypeDefinition.style.strokeColor,
    strokeWidth: edgeTypeDefinition.style.strokeWidth,
  };
  
  if (edgeTypeDefinition.style.strokeStyle === 'dashed') {
    style.strokeDasharray = '5,5';
  } else if (edgeTypeDefinition.style.strokeStyle === 'dotted') {
    style.strokeDasharray = '2,2';
  }
  
  return style;
};

export const applyPresetTheme = (preset: CanvasPreset): void => {
  const styles = getCanvasStyles(preset);
  
  for (const [property, value] of Object.entries(styles)) {
    document.documentElement.style.setProperty(property, value);
  }
};

export const removePresetTheme = (): void => {
  const properties = ['--canvas-bg', '--canvas-grid', '--canvas-primary', '--canvas-font'];
  
  for (const property of properties) {
    document.documentElement.style.removeProperty(property);
  }
};

export default {
  getCanvasConfig,
  getCanvasBackground,
  getCanvasStyles,
  getNodeStyle,
  getEdgeStyle,
  applyPresetTheme,
  removePresetTheme,
};
