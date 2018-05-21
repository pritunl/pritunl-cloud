/// <reference path="../References.d.ts"/>
export const SYNC = 'vpc.sync';
export const SYNC_NAMES= 'vpc.sync_names';
export const TRAVERSE = 'vpc.traverse';
export const FILTER = 'vpc.filter';
export const CHANGE = 'vpc.change';

export interface Vpc {
	id?: string;
	name?: string;
	network?: string;
	network6?: string;
	organization?: string;
	datacenter?: string;
	routes?: Route[];
}

export interface Route {
	destination?: string;
	target?: string;
}

export interface Filter {
	name?: string;
	network?: string;
	organization?: string;
	datacenter?: string;
}

export type Vpcs = Vpc[];

export type VpcRo = Readonly<Vpc>;
export type VpcsRo = ReadonlyArray<VpcRo>;

export interface VpcDispatch {
	type: string;
	data?: {
		id?: string;
		vpc?: Vpc;
		vpcs?: Vpcs;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
