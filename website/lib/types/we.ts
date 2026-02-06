// CWE Data Types
export interface CWEItem {
	id: string;
	name: string;
	diagram?: string;
	abstraction: string;
	structure: string;
	status: string;
	description: string;
	extendedDescription?: string;
	likelihoodOfExploit?: string;
	relatedWeaknesses?: RelatedWeakness[];
	weaknessOrdinalities?: WeaknessOrdinality[];
	applicablePlatforms?: ApplicablePlatform[];
	backgroundDetails?: string[];
	detectionMethods?: DetectionMethod[];
	potentialMitigations?: Mitigation[];
	demonstrativeExamples?: DemonstrativeExample[];
	observedExamples?: ObservedExample[];
	functionalAreas?: string[];
	affectedResources?: string[];
	taxonomyMappings?: TaxonomyMapping[];
	relatedAttackPatterns?: string[];
	references?: Reference[];
	mappingNotes?: MappingNotes[];
	notes?: Note[];
	contentHistory?: ContentHistory[];
}

export interface Weakness {
	nature: string;
	cweId: string;
	viewId?: string;
	ordinal?: string;
	description?: string;
}

export interface WeaknessOrdinality {
	ordinality: string;
	description?: string;
}
