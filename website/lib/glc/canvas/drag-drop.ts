import { NodeTypeDefinition } from '../types';

export interface DraggedNodeType {
  id: string;
  data: NodeTypeDefinition;
}

export const onDragStart = (
  event: React.DragEvent,
  nodeType: NodeTypeDefinition
): void => {
  try {
    event.dataTransfer.effectAllowed = 'copy';
    event.dataTransfer.setData('nodeType', nodeType.id);
    event.dataTransfer.setData('nodeTypeData', JSON.stringify(nodeType));
    
    const data: DraggedNodeType = {
      id: nodeType.id,
      data: nodeType,
    };
    event.dataTransfer.setData('application/json', JSON.stringify(data));
  } catch (error) {
    console.error('Error in onDragStart:', error);
  }
};

export const onDragOver = (event: React.DragEvent): void => {
  event.preventDefault();
  event.dataTransfer.dropEffect = 'copy';
};

export const calculateDropPosition = (
  event: React.DragEvent,
  canvasBounds: DOMRect
): { x: number; y: number } => {
  const scaleX = canvasBounds.width / event.dataTransfer.files.length || 1;
  const scaleY = canvasBounds.height / event.dataTransfer.files.length || 1;
  
  let x = event.clientX - canvasBounds.left;
  let y = event.clientY - canvasBounds.top;
  
  return { x, y };
};

export const onDrop = (
  event: React.DragEvent,
  canvasBounds: DOMRect,
  onCreateNode: (nodeType: NodeTypeDefinition, position: { x: number; y: number }) => void
): void => {
  event.preventDefault();
  event.stopPropagation();

  try {
    const nodeTypeId = event.dataTransfer.getData('nodeType');
    
    if (!nodeTypeId) {
      console.warn('No node type data found in drag event');
      return;
    }

    const nodeTypeDataStr = event.dataTransfer.getData('nodeTypeData');
    
    if (!nodeTypeDataStr) {
      console.warn('No node type data JSON found in drag event');
      return;
    }

    const nodeTypeData: NodeTypeDefinition = JSON.parse(nodeTypeDataStr);
    const position = calculateDropPosition(event, canvasBounds);
    
    onCreateNode(nodeTypeData, position);
  } catch (error) {
    console.error('Error in onDrop:', error);
  }
};

export const onDragLeave = (event: React.DragEvent): void => {
  event.preventDefault();
};

export const isDragEvent = (event: DragEvent): boolean => {
  return event.dataTransfer.types.length > 0;
};

export const isValidDropTarget = (event: DragEvent): boolean => {
  const nodeTypeId = event.dataTransfer.getData('nodeType');
  return !!nodeTypeId;
};

export default {
  onDragStart,
  onDragOver,
  calculateDropPosition,
  onDrop,
  onDragLeave,
  isDragEvent,
  isValidDropTarget,
};
