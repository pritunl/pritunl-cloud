/// <reference path="../References.d.ts"/>
export const SYNC = 'instance.sync';
export const TRAVERSE = 'instance.traverse';
export const FILTER = 'instance.filter';
export const CHANGE = 'instance.change';

export interface Instance {
	id: string;
	organization?: string;
	zone?: string;
	node?: string;
	image?: string;
	status?: string;
	state?: string;
	vm_state?: string;
	public_ip?: string;
	public_ip6?: string;
	name?: string;
	memory?: number;
	processors?: number;
	count?: number;
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
