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
	aws_id?: string;
	aws_secret?: string;
}

export interface Filter {
	id?: string;
	name?: string;
	network_role?: string;
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
