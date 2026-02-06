// CVETag Types
export interface CVETag {
	sourceIdentifier: string;
	tags?: string[];
}

export interface Weakness {
	source: string;
	type: string;
	description: Description[];
}
