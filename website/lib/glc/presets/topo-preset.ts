/**
 * Topo-Graph Canvas Preset
 *
 * General-purpose graph and topology diagramming.
 * Light theme with 8 node types and 8 relationship types.
 */

import type { CanvasPreset } from '../types';

export const topoPreset: CanvasPreset = {
  meta: {
    id: 'topo',
    name: 'Topo-Graph',
    version: '1.0.0',
    description: 'General-purpose graph and topology diagramming',
    author: 'v2e',
    createdAt: '2026-02-10T00:00:00Z',
    updatedAt: '2026-02-10T00:00:00Z',
  },

  theme: {
    primary: '#3b82f6',
    background: '#ffffff',
    surface: '#f8fafc',
    text: '#1e293b',
    textMuted: '#64748b',
    border: '#e2e8f0',
    accent: '#8b5cf6',
    success: '#22c55e',
    warning: '#f59e0b',
    error: '#ef4444',
  },

  behavior: {
    snapToGrid: true,
    gridSize: 20,
    autoLayout: false,
    historyLimit: 100,
    autoSaveInterval: 5000,
    enableInference: false,
  },

  nodeTypes: [
    {
      id: 'node',
      label: 'Node',
      category: 'Basic',
      description: 'Generic node',
      icon: 'Circle',
      color: '#3b82f6',
      borderColor: '#2563eb',
      backgroundColor: '#dbeafe',
      defaultWidth: 140,
      defaultHeight: 60,
      properties: [
        { key: 'label', value: '', type: 'string', required: true },
        { key: 'description', value: '', type: 'string' },
      ],
    },
    {
      id: 'process',
      label: 'Process',
      category: 'Flow',
      description: 'A process step in a flowchart',
      icon: 'Play',
      color: '#22c55e',
      borderColor: '#16a34a',
      backgroundColor: '#dcfce7',
      defaultWidth: 140,
      defaultHeight: 60,
      properties: [
        { key: 'label', value: '', type: 'string', required: true },
        { key: 'duration', value: '', type: 'string' },
      ],
    },
    {
      id: 'decision',
      label: 'Decision',
      category: 'Flow',
      description: 'A decision point in a flowchart',
      icon: 'GitBranch',
      color: '#f59e0b',
      borderColor: '#d97706',
      backgroundColor: '#fef3c7',
      defaultWidth: 120,
      defaultHeight: 80,
      properties: [
        { key: 'condition', value: '', type: 'string', required: true },
      ],
    },
    {
      id: 'endpoint',
      label: 'Endpoint',
      category: 'Flow',
      description: 'Start or end point',
      icon: 'CircleDot',
      color: '#64748b',
      borderColor: '#475569',
      backgroundColor: '#f1f5f9',
      defaultWidth: 100,
      defaultHeight: 50,
      properties: [
        { key: 'type', value: 'start', type: 'string' },
      ],
    },
    {
      id: 'database',
      label: 'Database',
      category: 'Infrastructure',
      description: 'Database or data store',
      icon: 'Database',
      color: '#8b5cf6',
      borderColor: '#7c3aed',
      backgroundColor: '#ede9fe',
      defaultWidth: 140,
      defaultHeight: 70,
      properties: [
        { key: 'name', value: '', type: 'string', required: true },
        { key: 'type', value: '', type: 'string' },
      ],
    },
    {
      id: 'server',
      label: 'Server',
      category: 'Infrastructure',
      description: 'Server or compute resource',
      icon: 'Server',
      color: '#06b6d4',
      borderColor: '#0891b2',
      backgroundColor: '#cffafe',
      defaultWidth: 140,
      defaultHeight: 70,
      properties: [
        { key: 'hostname', value: '', type: 'string' },
        { key: 'ip', value: '', type: 'string' },
      ],
    },
    {
      id: 'client',
      label: 'Client',
      category: 'Infrastructure',
      description: 'Client device or application',
      icon: 'Monitor',
      color: '#ec4899',
      borderColor: '#db2777',
      backgroundColor: '#fce7f3',
      defaultWidth: 140,
      defaultHeight: 60,
      properties: [
        { key: 'name', value: '', type: 'string' },
        { key: 'type', value: '', type: 'string' },
      ],
    },
    {
      id: 'cloud',
      label: 'Cloud',
      category: 'Infrastructure',
      description: 'Cloud service or resource',
      icon: 'Cloud',
      color: '#0ea5e9',
      borderColor: '#0284c7',
      backgroundColor: '#e0f2fe',
      defaultWidth: 160,
      defaultHeight: 80,
      properties: [
        { key: 'provider', value: '', type: 'string' },
        { key: 'service', value: '', type: 'string' },
      ],
    },
  ],

  relations: [
    {
      id: 'connects',
      label: 'connects',
      sourceTypes: ['node', 'process', 'decision', 'endpoint', 'database', 'server', 'client', 'cloud'],
      targetTypes: ['node', 'process', 'decision', 'endpoint', 'database', 'server', 'client', 'cloud'],
      style: { strokeColor: '#64748b', strokeWidth: 2 },
    },
    {
      id: 'flows-to',
      label: 'flows to',
      sourceTypes: ['process', 'decision', 'endpoint'],
      targetTypes: ['process', 'decision', 'endpoint'],
      style: { strokeColor: '#3b82f6', strokeWidth: 2, markerEnd: true },
    },
    {
      id: 'yes',
      label: 'Yes',
      sourceTypes: ['decision'],
      targetTypes: ['process', 'decision', 'endpoint'],
      style: { strokeColor: '#22c55e', strokeWidth: 2, markerEnd: true },
    },
    {
      id: 'no',
      label: 'No',
      sourceTypes: ['decision'],
      targetTypes: ['process', 'decision', 'endpoint'],
      style: { strokeColor: '#ef4444', strokeWidth: 2, markerEnd: true },
    },
    {
      id: 'reads',
      label: 'reads',
      sourceTypes: ['server', 'client', 'process'],
      targetTypes: ['database'],
      style: { strokeColor: '#8b5cf6', strokeWidth: 2, markerEnd: true },
    },
    {
      id: 'writes',
      label: 'writes',
      sourceTypes: ['server', 'client', 'process'],
      targetTypes: ['database'],
      style: { strokeColor: '#8b5cf6', strokeWidth: 2, strokeStyle: 'dashed', markerEnd: true },
    },
    {
      id: 'calls',
      label: 'calls',
      sourceTypes: ['client', 'server', 'process'],
      targetTypes: ['server', 'cloud'],
      style: { strokeColor: '#06b6d4', strokeWidth: 2, markerEnd: true },
    },
    {
      id: 'depends-on',
      label: 'depends on',
      sourceTypes: ['server', 'client', 'cloud', 'database'],
      targetTypes: ['server', 'client', 'cloud', 'database'],
      style: { strokeColor: '#f59e0b', strokeWidth: 2, strokeStyle: 'dotted', markerEnd: true },
    },
  ],
};

export default topoPreset;
