/// <reference path="../References.d.ts"/>
export const SYNC = 'pool.sync';
export const TRAVERSE = 'pool.traverse';
export const FILTER = 'pool.filter';
export const CHANGE = 'pool.change';

export interface Pool {
	id?: string;
	name?: string;
	comment?: string;
	delete_protection?: boolean;
	zone?: string;
	type?: string;
	vg_name?: string;
}

export interface Filter {
	id?: string;
	name?: string;
	comment?: string;
	vg_name?: string;
}

export type Pools = Pool[];

export type PoolRo = Readonly<Pool>;
export type PoolsRo = ReadonlyArray<PoolRo>;

export interface PoolDispatch {
	type: string;
	data?: {
		id?: string;
		pool?: Pool;
		pools?: Pools;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
