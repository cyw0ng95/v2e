import { ExampleGraph, ExampleGraphsData, ExampleGraphSchema, ExampleGraphsDataSchema } from './example-types';

let examplesCache: ExampleGraph[] | null = null;

export async function loadExamples(): Promise<ExampleGraph[]> {
  if (examplesCache) {
    return examplesCache;
  }

  try {
    const response = await fetch('/glc/assets/examples/example-graphs.json');
    if (!response.ok) {
      throw new Error(`Failed to load examples: ${response.statusText}`);
    }

    const data = await response.json();
    const validated = ExampleGraphsDataSchema.parse(data);
    examplesCache = validated.examples;
    return examplesCache;
  } catch (error) {
    console.error('Error loading examples:', error);
    throw error;
  }
}

export async function getExampleById(id: string): Promise<ExampleGraph | null> {
  const examples = await loadExamples();
  return examples.find(example => example.id === id) || null;
}

export async function getExamplesByPreset(preset: string): Promise<ExampleGraph[]> {
  const examples = await loadExamples();
  return examples.filter(example => example.preset === preset);
}

export async function getExamplesByCategory(category: string): Promise<ExampleGraph[]> {
  const examples = await loadExamples();
  return examples.filter(example => example.category === category);
}

export async function searchExamples(query: string): Promise<ExampleGraph[]> {
  const examples = await loadExamples();
  const lowerQuery = query.toLowerCase();
  return examples.filter(example =>
    example.name.toLowerCase().includes(lowerQuery) ||
    example.description.toLowerCase().includes(lowerQuery)
  );
}

export async function getCategories(): Promise<string[]> {
  const examples = await loadExamples();
  const categories = new Set(examples.map(example => example.category));
  return Array.from(categories).sort();
}

export function validateExampleGraph(graph: unknown): ExampleGraph {
  return ExampleGraphSchema.parse(graph);
}

export function clearExamplesCache(): void {
  examplesCache = null;
}
