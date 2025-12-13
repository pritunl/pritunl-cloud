/// <reference path="../References.d.ts"/>
export const SYNC = 'domain.sync';
export const SYNC_NAME = 'domain.sync_name';
export const TRAVERSE = 'domain.traverse';
export const FILTER = 'domain.filter';
export const CHANGE = 'domain.change';

export interface Domain {
	id?: string;
	name?: string;
	comment?: string;
	organization?: string;
	type?: string;
	secret?: string;
	root_domain?: string;
	records?: Record[];
}

export interface Record {
	id?: string;
	domain?: string;
	timestamp?: string;
	sub_domain?: string;
	type?: string;
	value?: string;
	operation?: string;
}

export interface Filter {
	id?: string;
	name?: string;
	organization?: string;
}

export type Domains = Domain[];

export type DomainRo = Readonly<Domain>;
export type DomainsRo = ReadonlyArray<DomainRo>;

export interface DomainDispatch {
	type: string;
	data?: {
		id?: string;
		domain?: Domain;
		domains?: Domains;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
