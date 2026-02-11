// CPE Match Types
export interface CPEMatch {
	vulnerable: boolean;
	criteria: string;
	matchCriteriaId: string;
	versionStartExcluding?: string;
	versionStartIncluding?: string;
	versionEndExcluding?: string;
	versionEndIncluding?: string;
}
