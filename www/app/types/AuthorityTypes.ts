/// <reference path="../References.d.ts"/>
export const SYNC = 'authority.sync';
export const SYNC_NAMES = 'authority.sync_names';
export const TRAVERSE = 'authority.traverse';
export const FILTER = 'authority.filter';
export const CHANGE = 'authority.change';

export interface Authority {
	id?: string;
	name?: string;
	comment?: string;
	type?: string;
	organization?: string;
	network_roles?: string[];
	key?: string;
	roles?: string[];
	certificate?: string;
}

export interface Filter {
	id?: string;
	name?: string;
	role?: string;
	network_role?: string;
	organization?: string;
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
