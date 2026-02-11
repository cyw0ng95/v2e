// ASVS Data Types
export interface ASVSItem {
	requirementID: string;
	chapter: string;
	section: string;
	description: string;
	level1: boolean;
	level2: boolean;
	level3: boolean;
	cwe?: string;
}

// AT&CK Data Types
export interface AttackTechnique {
	id: string;
	name: string;
	description?: string;
	domain?: string;
	platform?: string;
	created?: string;
	modified?: string;
	revoked?: boolean;
	deprecated?: boolean;
}

export interface AttackTactic {
	id: string;
	name?: string;
	description?: string;
	domain?: string;
	created?: string;
	modified?: string;
}

export interface AttackMitigation {
	id: string;
	name: string;
	description?: string;
	domain?: string;
	created?: string;
	modified?: string;
}

export interface AttackSoftware {
	id: string;
	name: string;
	description?: string;
	type?: string;
	domain?: string;
	created?: string;
	modified?: string;
}

export interface AttackGroup {
	id: string;
	name?: string;
	description?: string;
	domain?: string;
	created?: string;
	modified?: string;
}
