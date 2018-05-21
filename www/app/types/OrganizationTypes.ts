/// <reference path="../References.d.ts"/>
export const SYNC = 'organization.sync';
export const CHANGE = 'organization.change';
export const CURRENT = 'organization.current';

export interface Organization {
	id: string;
	name?: string;
	roles?: string[];
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
		current?: string;
	};
}
