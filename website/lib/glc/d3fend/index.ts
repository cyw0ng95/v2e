/**
 * GLC D3FEND Module
 *
 * Provides D3FEND ontology types, loader, and utilities.
 */

// Types
export * from './types';

// Simplified ontology (static fallback)
export * from './ontology';

// Lazy loader for full ontology
export {
  loadD3FENDOntology,
  getD3FENDOntology,
  getOntologyLoadState,
  getClassById,
  getClassChildren,
  getClassAncestors,
  buildTreeNodes,
  searchClasses,
  resetLoader,
} from './loader';
