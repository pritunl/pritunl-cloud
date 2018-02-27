/// <reference path="../References.d.ts"/>
export const SYNC = 'authority.sync';
export const TRAVERSE = 'authority.traverse';
export const FILTER = 'authority.filter';
export const CHANGE = 'authority.change';

export interface Authority {
	id?: string;
	name?: string;
	organization?: string;
	network_roles?: string[];
}

export interface Filter {
	name?: string;
}

export type Authorities = Authority[];

export type AuthorityRo = Readonly<Authority>;
export type AuthoritiesRo = ReadonlyArray<AuthorityRo>;

export interface AuthorityDispatch {
	type: string;
	data?: {
		id?: string;
		authority?: Authority;
		authorities?: Authorities;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
