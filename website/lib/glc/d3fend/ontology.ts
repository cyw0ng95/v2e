/**
 * D3FEND Ontology Data
 *
 * Simplified D3FEND ontology for the GLC canvas.
 * Full ontology can be loaded from assets/d3fend/d3fend.json
 */

export interface D3FENDClass {
  id: string;
  label: string;
  description?: string;
  parent?: string;
  children?: string[];
  techniques?: string[];
}

// Top-level D3FEND classes
export const D3FEND_CLASSES: D3FENDClass[] = [
  {
    id: 'd3f:DefensiveTechnique',
    label: 'Defensive Technique',
    description: 'Techniques to defend against cyber attacks',
    children: [
      'd3f:Hardening',
      'd3f:Detection',
      'd3f:Isolation',
      'd3f:Restoration',
    ],
  },
  {
    id: 'd3f:Hardening',
    label: 'Hardening',
    description: 'Strengthening systems against attacks',
    parent: 'd3f:DefensiveTechnique',
    children: [
      'd3f:ApplicationHardening',
      'd3f:PlatformHardening',
    ],
  },
  {
    id: 'd3f:Detection',
    label: 'Detection',
    description: 'Identifying malicious activity',
    parent: 'd3f:DefensiveTechnique',
    children: [
      'd3f:NetworkTrafficAnalysis',
      'd3f:FileAnalysis',
      'd3f:ProcessAnalysis',
    ],
  },
  {
    id: 'd3f:Isolation',
    label: 'Isolation',
    description: 'Containing threats',
    parent: 'd3f:DefensiveTechnique',
  },
  {
    id: 'd3f:Restoration',
    label: 'Restoration',
    description: 'Recovering from attacks',
    parent: 'd3f:DefensiveTechnique',
  },
  {
    id: 'd3f:ApplicationHardening',
    label: 'Application Hardening',
    parent: 'd3f:Hardening',
  },
  {
    id: 'd3f:PlatformHardening',
    label: 'Platform Hardening',
    parent: 'd3f:Hardening',
  },
  {
    id: 'd3f:NetworkTrafficAnalysis',
    label: 'Network Traffic Analysis',
    parent: 'd3f:Detection',
  },
  {
    id: 'd3f:FileAnalysis',
    label: 'File Analysis',
    parent: 'd3f:Detection',
  },
  {
    id: 'd3f:ProcessAnalysis',
    label: 'Process Analysis',
    parent: 'd3f:Detection',
  },
  {
    id: 'd3f:DigitalSecurity',
    label: 'Digital Security',
    description: 'Security-related digital objects',
  },
  {
    id: 'd3f:Analytic',
    label: 'Analytic',
    description: 'Analysis techniques and methods',
  },
];

// Get class by ID
export function getD3FENDClass(id: string): D3FENDClass | undefined {
  return D3FEND_CLASSES.find((c) => c.id === id);
}

// Get children of a class
export function getD3FENDChildren(parentId: string): D3FENDClass[] {
  const parent = getD3FENDClass(parentId);
  if (!parent?.children) return [];

  return parent.children
    .map((id) => getD3FENDClass(id))
    .filter((c): c is D3FENDClass => c !== undefined);
}

// Get ancestors of a class
export function getD3FENDAncestors(classId: string): D3FENDClass[] {
  const ancestors: D3FENDClass[] = [];
  let current = getD3FENDClass(classId);

  while (current?.parent) {
    const parent = getD3FENDClass(current.parent);
    if (parent) {
      ancestors.push(parent);
      current = parent;
    } else {
      break;
    }
  }

  return ancestors;
}

// Search classes
export function searchD3FENDClasses(query: string): D3FENDClass[] {
  const lower = query.toLowerCase();
  return D3FEND_CLASSES.filter(
    (c) =>
      c.label.toLowerCase().includes(lower) ||
      c.description?.toLowerCase().includes(lower) ||
      c.id.toLowerCase().includes(lower)
  );
}

// Get all class paths (for breadcrumbs)
export function getClassPath(classId: string): D3FENDClass[] {
  const ancestors = getD3FENDAncestors(classId);
  const current = getD3FENDClass(classId);
  if (current) {
    return [...ancestors.reverse(), current];
  }
  return ancestors.reverse();
}
