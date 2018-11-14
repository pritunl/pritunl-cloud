/// <reference path="../References.d.ts"/>
export const SYNC = 'node.sync';
export const SYNC_ZONE = 'node.sync_zone';
export const TRAVERSE = 'node.traverse';
export const FILTER = 'node.filter';
export const CHANGE = 'node.change';

export interface Node {
	id: string;
	types?: string[];
	zone?: string;
	name?: string;
	port?: number;
	protocol?: string;
	hypervisor?: string;
	timestamp?: string;
	admin_domain?: string;
	user_domain?: string;
	certificates?: string[];
	external_interface?: string;
	internal_interface?: string;
	firewall?: boolean;
	network_roles?: string[];
	requests_min?: number;
	cpu_units?: number;
	memory_units?: number;
	memory?: number;
	load1?: number;
	load5?: number;
	load15?: number;
	public_ips?: string[];
	public_ips6?: string[];
	forwarded_for_header?: string;
	forwarded_proto_header?: string;
	software_version?: string;
}

export interface Filter {
	id?: string;
	name?: string;
	zone?: string;
	network_role?: string;
	admin?: boolean;
	user?: boolean;
	hypervisor?: boolean;
}

export type Nodes = Node[];

export type NodeRo = Readonly<Node>;
export type NodesRo = ReadonlyArray<NodeRo>;

export interface NodeDispatch {
	type: string;
	data?: {
		id?: string;
		node?: Node;
		nodes?: Nodes;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
