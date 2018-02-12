/// <reference path="../References.d.ts"/>
export const SYNC = 'instance.sync';
export const TRAVERSE = 'user.traverse';
export const FILTER = 'user.filter';
export const CHANGE = 'instance.change';

export interface Instance {
	id: string;
	organization?: string;
	zone?: string;
	name?: number;
}

export interface Filter {
	name?: string;
}

export type Instances = Instance[];

export type InstanceRo = Readonly<Instance>;
export type InstancesRo = ReadonlyArray<InstanceRo>;

export interface InstanceDispatch {
	type: string;
	data?: {
		id?: string;
		instance?: Instance;
		instances?: Instances;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
