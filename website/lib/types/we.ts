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

export interface RelatedWeakness {
	nature: string;
	cweId: string;
	viewId?: string;
	ordinal?: string;
	description?: string;
}

export interface ApplicablePlatform {
	language?: string;
	technology?: string;
	class?: string;
	operatingSystem?: string;
	cweId?: string;
}

export interface DetectionMethod {
	method?: string;
	description?: string;
}

export interface Mitigation {
	phase?: string;
	strategy?: string;
	description?: string;
}

export interface DemonstrativeExample {
	heading?: string;
	content?: string;
	references?: string[];
}

export interface ObservedExample {
	reference?: string;
	description?: string;
	link?: string;
}

export interface TaxonomyMapping {
	taxonomyName: string;
	entryId?: string;
	entryName?: string;
	mappingType?: string;
}

export interface Reference {
	source?: string;
	url?: string;
}

export interface MappingNotes {
	usage?: string;
	type?: string;
	other?: string;
}

export interface Note {
	type?: string;
	title?: string;
	content?: string;
}

export interface ContentHistory {
	type: string;
	date: string;
	contributor?: string;
	comment?: string;
}
