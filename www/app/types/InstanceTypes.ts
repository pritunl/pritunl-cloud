/// <reference path="../References.d.ts"/>
export const SYNC = 'instance.sync';
export const SYNC_NODE = 'instance.sync_node';
export const TRAVERSE = 'instance.traverse';
export const FILTER = 'instance.filter';
export const CHANGE = 'instance.change';

export interface Instance {
	id: string;
	organization?: string;
	zone?: string;
	node?: string;
	image?: string;
	image_backing?: boolean;
	status?: string;
	state?: string;
	vm_state?: string;
	public_ips?: string[];
	public_ips6?: string[];
	private_ips?: string[];
	private_ips6?: string[];
	name?: string;
	init_disk_size?: number;
	memory?: number;
	processors?: number;
	network_roles?: string[];
	domain?: string;
	vpc?: string;
	count?: number;
	info?: Info;
}

export interface Filter {
	id?: string;
	name?: string;
	state?: string;
	network_role?: string;
	organization?: string;
	node?: string;
	zone?: string;
}

export interface Info {
	node?: string;
	firewall_rules?: string[];
	authorities?: string[];
	disks?: string[];
}

export type Instances = Instance[];
export type InstancesNode = Map<string, Instances>;

export type InstanceRo = Readonly<Instance>;
export type InstancesRo = ReadonlyArray<InstanceRo>;
export type InstancesNodeRo = Map<string, InstancesRo>;

export interface InstanceDispatch {
	type: string;
	data?: {
		id?: string;
		node?: string;
		instance?: Instance;
		instances?: Instances;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
