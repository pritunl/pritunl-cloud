/// <reference path="../References.d.ts"/>
export const SYNC = 'vpc.sync';
export const SYNC_NAMES= 'vpc.sync_names';
export const TRAVERSE = 'vpc.traverse';
export const FILTER = 'vpc.filter';
export const CHANGE = 'vpc.change';

export interface Vpc {
	id?: string;
	name?: string;
	comment?: string;
	vpc_id?: number;
	network?: string;
	network6?: string;
	subnets?: Subnet[];
	organization?: string;
	datacenter?: string;
	routes?: Route[];
	maps?: Map[];
	arps?: Arp[];
}

export interface Subnet {
	id?: string;
	name?: string;
	network?: string;
}

export interface Route {
	destination?: string;
	target?: string;
}

export interface Map {
	destination?: string;
	target?: string;
}

export interface Arp {
	ip?: string;
	mac?: string;
}

export interface Filter {
	id?: string;
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
