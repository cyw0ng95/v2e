import { CADNode, CADEdge, CanvasPreset, NodeTypeDefinition, RelationshipDefinition } from '../types';

export const generateId = (): string => {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
};

export const findNodeTypeById = (preset: CanvasPreset, typeId: string): NodeTypeDefinition | undefined => {
  return preset.nodeTypes.find(type => type.id === typeId);
};

export const findRelationshipTypeById = (preset: CanvasPreset, typeId: string): RelationshipDefinition | undefined => {
  return preset.relationshipTypes.find(type => type.id === typeId);
};

export const getValidRelationshipTypes = (
  preset: CanvasPreset,
  sourceNodeType: string,
  targetNodeType: string
): RelationshipDefinition[] => {
  return preset.relationshipTypes.filter(rel => 
    (rel.sourceNodeTypes.includes('*') || rel.sourceNodeTypes.includes(sourceNodeType)) &&
    (rel.targetNodeTypes.includes('*') || rel.targetNodeTypes.includes(targetNodeType))
  );
};

export const validateNodePosition = (
  nodes: CADNode[],
  position: { x: number; y: number },
  threshold: number = 20
): boolean => {
  return !nodes.some(node => 
    Math.abs(node.position.x - position.x) < threshold &&
    Math.abs(node.position.y - position.y) < threshold
  );
};

export const getDefaultNodePosition = (nodes: CADNode[]): { x: number; y: number } => {
  if (nodes.length === 0) {
    return { x: 400, y: 300 };
  }
  
  const lastNode = nodes[nodes.length - 1];
  return {
    x: lastNode.position.x + 150,
    y: lastNode.position.y,
  };
};

export const calculateEdgeStyle = (relationshipType: RelationshipDefinition, isAnimated?: boolean): Record<string, any> => {
  return {
    stroke: relationshipType.style.strokeColor,
    strokeWidth: relationshipType.style.strokeWidth,
    strokeDasharray: relationshipType.style.strokeStyle === 'dashed' ? '5,5' : 
                     relationshipType.style.strokeStyle === 'dotted' ? '2,2' : 
                     undefined,
    animated: isAnimated ?? relationshipType.style.animated,
  };
};

export const calculateNodeStyle = (nodeType: NodeTypeDefinition, isSelected?: boolean): Record<string, any> => {
  const baseStyle = {
    backgroundColor: nodeType.style.backgroundColor,
    borderColor: isSelected ? '#ffffff' : nodeType.style.borderColor,
    color: nodeType.style.textColor,
    borderWidth: nodeType.style.borderWidth ?? 2,
    borderRadius: nodeType.style.borderRadius ?? 8,
    padding: nodeType.style.padding ?? '12px',
  };
  
  return baseStyle;
};

export const formatTimestamp = (timestamp: string): string => {
  try {
    const date = new Date(timestamp);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return timestamp;
  }
};

export const debounce = <T extends (...args: any[]) => any>(
  func: T,
  wait: number
): ((...args: Parameters<T>) => void) => {
  let timeout: NodeJS.Timeout | null = null;
  
  return (...args: Parameters<T>) => {
    if (timeout) clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
};

export const throttle = <T extends (...args: any[]) => any>(
  func: T,
  limit: number
): ((...args: Parameters<T>) => void) => {
  let inThrottle: boolean;
  
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args);
      inThrottle = true;
      setTimeout(() => inThrottle = false, limit);
    }
  };
};

export const cloneDeep = <T>(obj: T): T => {
  return JSON.parse(JSON.stringify(obj));
};

export const isEmpty = (value: any): boolean => {
  if (value === null || value === undefined) return true;
  if (typeof value === 'string') return value.trim() === '';
  if (Array.isArray(value)) return value.length === 0;
  if (typeof value === 'object') return Object.keys(value).length === 0;
  return false;
};

export const getErrorMessage = (error: any): string => {
  if (error?.message) return error.message;
  if (typeof error === 'string') return error;
  return 'An unknown error occurred';
};

export const downloadJSON = (data: any, filename: string): void => {
  const json = JSON.stringify(data, null, 2);
  const blob = new Blob([json], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
};

export const readJSONFile = (file: File): Promise<any> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const json = JSON.parse(e.target?.result as string);
        resolve(json);
      } catch (error) {
        reject(new Error('Invalid JSON file'));
      }
    };
    reader.onerror = () => reject(new Error('Failed to read file'));
    reader.readAsText(file);
  });
};
