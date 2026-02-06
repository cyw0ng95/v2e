// CVETag Types

export interface Description {
	lang: string;
	value: string;
}
export interface CVETag {
	sourceIdentifier: string;
	tags?: string[];
}

export interface Weakness {
	source: string;
	type: string;
	description: Description[];
}
