import { D3FENDData, D3FENDClass, D3FENDProperty } from './d3fend-types';
import { validateD3FENDData } from './d3fend-validator';

export interface LoadD3FENDDataResult {
  data: D3FENDData | null;
  loading: boolean;
  error: string | null;
  progress: number;
  stages: string[];
}

let d3fendDataCache: D3FENDData | null = null;
let isLoadingD3FENDData = false;
let loadProgress = 0;

export const loadD3FENDData = async (
  force: boolean = false
): Promise<LoadD3FENDDataResult> => {
  if (d3fendDataCache && !force) {
    return {
      data: d3fendDataCache,
      loading: false,
      error: null,
      progress: 100,
      stages: [],
    };
  }

  isLoadingD3FENDData = true;
  loadProgress = 0;

  const stages: string[] = [
    'Initializing load...',
    'Fetching D3FEND data bundle...',
    'Parsing D3FEND ontology...',
    'Validating structure...',
    'Creating in-memory cache...',
    'Loading into UI cache...',
    'Processing classes...',
    'Processing properties...',
    'Processing inferences...',
    'Finalizing...',
  ];

  try {
    const response = await fetch('https://d3fend.mitre.org/static/bundles/d3fend.json');

    loadProgress = 20;
    loadProgress = 40;
    const rawText = await response.text();

    loadProgress = 60;
    const rawData = JSON.parse(rawText);

    loadProgress = 80;
    const validation = validateD3FENDData(rawData);

    if (!validation.valid) {
      throw new Error(validation.errors.join(', '));
    }

    loadProgress = 100;
    const d3fendData: D3FENDData = rawData;

    d3fendDataCache = d3fendData;
    isLoadingD3FENDData = false;

    return {
      data: d3fendData,
      loading: false,
      error: null,
      progress: 100,
      stages: [],
    };
  } catch (error) {
    isLoadingD3FENDData = false;
    loadProgress = 0;

    console.error('Error loading D3FEND data:', error);

    return {
      data: null,
      loading: false,
      error: error instanceof Error ? error.message : 'Failed to load D3FEND data',
      progress: 0,
      stages: [],
    };
  }
};

export const getD3FENDData = (): D3FENDData | null => {
  return d3fendDataCache;
};

export const isD3FENDDataLoading = (): boolean => {
  return isLoadingD3ENDData;
};

export const getD3FENDLoadProgress = (): number => {
  return loadProgress;
};

export const clearD3FENDDataCache = (): void => {
  d3fendDataCache = null;
  loadProgress = 0;
};

export const preloadD3FENDData = async (): Promise<LoadD3FENDDataResult> => {
  if (d3fendDataCache) {
    return {
      data: d3fendDataCache,
      loading: false,
      error: null,
      progress: 100,
      stages: [],
    };
  }

  return loadD3FENDData(false);
};

export const getD3FENDClasses = (): D3FENDClass[] => {
  return d3fendDataCache?.classes || [];
};

export const getD3FENDClassById = (classId: string): D3FENDClass | null => {
  if (!d3fendDataCache) return null;
  return d3fendDataCache.classes.find(c => c.id === classId) || null;
};

export const getD3FENDPropertiesForClass = (classId: string): any[] => {
  if (!d3fendDataCache) return [];
  const classDef = getD3FENDClassById(classId);
  return classDef?.properties || [];
};

export const getD3FENDInferences = (classId: string): D3FENDInference[] => {
  if (!d3fendDataCache) return [];
  const classDef = getD3FENDClassById(classId);
  return classDef?.inferences || [];
};

export const getD3FENDRelationships = (
  sourceNodeType: string,
  targetNodeType: string
): any[] => {
  if (!d3endDataCache) return [];

  return d3fendDataCache.relationships.filter(rel =>
    (rel.source_node_types?.includes('*') || rel.source_node_types?.includes(sourceNodeType)) &&
    (rel.target_node_types?.includes('*') || rel.target_node_types?.includes(targetNodeType))
  );
};

export const getD3FENDTactics = (
  nodeType: string
): any[] => {
  if (!d3endDataCache) return [];

  const classDef = getD3FENDClassById(nodeType);

  if (!classDef) return [];

  return classDef.tactics || [];
};

export default {
  loadD3FENDData,
  getD3FENDData,
  isD3FENDDataLoading,
  getD3FENDLoadProgress,
  clearD3FENDDataCache,
  preloadDENDData,
  getD3FENDClasses,
  getD3FENDClassById,
  getD3FENDPropertiesForClass,
  getD3FENDInferences,
  getDENDRelationships,
  getD3FENDTactics,
};
};
