/// <reference path="../References.d.ts"/>
export const SYNC = 'pod.sync';
export const TRAVERSE = 'pod.traverse';
export const FILTER = 'pod.filter';
export const CHANGE = 'pod.change';

export interface Pod {
	id?: string;
	name?: string;
	comment?: string;
	organization?: string;
	type?: string;
	delete_protection?: boolean;
	zone?: string;
	roles?: string[];
	spec?: string;
}

export interface Filter {
	id?: string;
	name?: string;
	comment?: string;
	role?: string;
	organization?: string;
}

export type Pods = Pod[];

export type PodRo = Readonly<Pod>;
export type PodsRo = ReadonlyArray<PodRo>;

export interface PodDispatch {
	type: string;
	data?: {
		id?: string;
		pod?: Pod;
		pods?: Pods;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
