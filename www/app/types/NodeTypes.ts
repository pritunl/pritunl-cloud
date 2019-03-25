/// <reference path="../References.d.ts"/>
export const SYNC = 'node.sync';
export const SYNC_ZONE = 'node.sync_zone';
export const TRAVERSE = 'node.traverse';
export const FILTER = 'node.filter';
export const CHANGE = 'node.change';

export interface Node {
	id?: string;
	types?: string[];
	zone?: string;
	name?: string;
	port?: number;
	no_redirect_server?: boolean;
	protocol?: string;
	hypervisor?: string;
	vga?: string;
	timestamp?: string;
	admin_domain?: string;
	user_domain?: string;
	certificates?: string[];
	network_mode?: string;
	external_interface?: string;
	internal_interface?: string;
	external_interfaces?: string[];
	internal_interfaces?: string[];
	available_interfaces?: string[];
	available_bridges?: string[];
	default_interface?: string;
	blocks?: BlockAttachment[];
	host_block?: string;
	host_nat?: boolean;
	host_nat_excludes?: string[];
	jumbo_frames?: boolean;
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
	hostname?: string;
	oracle_user?: string;
	oracle_public_key?: string;
	oracle_host_route?: boolean;
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

export interface BlockAttachment {
	interface?: string;
	block?: string;
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
