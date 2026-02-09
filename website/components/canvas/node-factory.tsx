import { Node } from '@xyflow/react';
import { DynamicNode } from './dynamic-node';
import { CADNode } from "@/lib/glc/types"

export const nodeTypes = {
  'dynamic-node': DynamicNode,
};

export const createFlowNode = (cadNode: CADNode): Node => {
  return {
    id: cadNode.id,
    type: 'dynamic-node',
    position: cadNode.position,
    data: cadNode.data,
  };
};

export const createFlowNodes = (cadNodes: CADNode[]): Node[] => {
  return cadNodes.map(createFlowNode);
};

export default {
  nodeTypes,
  createFlowNode,
  createFlowNodes,
};
