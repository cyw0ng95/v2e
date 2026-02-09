import { Edge } from '@xyflow/react';
import { DynamicEdge } from './dynamic-edge';
import { CADEdge } from '../../types';

export const edgeTypes = {
  'dynamic-edge': DynamicEdge,
};

export const createFlowEdge = (cadEdge: CADEdge): Edge => {
  return {
    id: cadEdge.id,
    type: 'dynamic-edge',
    source: cadEdge.source,
    target: cadEdge.target,
    data: cadEdge.data,
    animated: cadEdge.animated,
  };
};

export const createFlowEdges = (cadEdges: CADEdge[]): Edge[] => {
  return cadEdges.map(createFlowEdge);
};

export default {
  edgeTypes,
  createFlowEdge,
  createFlowEdges,
};
