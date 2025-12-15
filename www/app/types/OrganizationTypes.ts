/// <reference path="../References.d.ts"/>
export const SYNC = 'organization.sync';
export const CHANGE = 'organization.change';
export const TRAVERSE = 'organization.traverse';
export const FILTER = 'organization.filter';
export const CURRENT = 'organization.current';

export interface Organization {
	id?: string;
	name?: string;
	comment?: string;
	roles?: string[];
}

export interface Filter {
	id?: string;
	name?: string;
	comment?: string;
}

export type Organizations = Organization[];

export type OrganizationRo = Readonly<Organization>;
export type OrganizationsRo = ReadonlyArray<OrganizationRo>;

export interface OrganizationDispatch {
	type: string;
	data?: {
		id?: string;
		organization?: Organization;
		organizations?: Organizations;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
