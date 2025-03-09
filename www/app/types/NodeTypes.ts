/// <reference path="../References.d.ts"/>
export const SYNC = 'node.sync';
export const SYNC_ZONE = 'node.sync_zone';
export const TRAVERSE = 'node.traverse';
export const FILTER = 'node.filter';
export const CHANGE = 'node.change';

export interface Node {
	id?: string;
	types?: string[];
	datacenter?: string;
	zone?: string;
	name?: string;
	comment?: string;
	port?: number;
	no_redirect_server?: boolean;
	protocol?: string;
	hypervisor?: string;
	vga?: string;
	vga_render?: string;
	available_renders?: string[];
	gui?: boolean;
	gui_user?: string;
	gui_mode?: string;
	timestamp?: string;
	admin_domain?: string;
	user_domain?: string;
	webauthn_domain?: string;
	certificates?: string[];
	network_mode?: string;
	network_mode6?: string;
	external_interface?: string;
	internal_interface?: string;
	external_interfaces?: string[];
	internal_interfaces?: string[];
	external_interfaces6?: string[];
	available_interfaces?: string[];
	available_vpcs?: Vpc[];
	oracle_subnets?: string[];
	available_bridges?: string[];
	default_interface?: string;
	blocks?: BlockAttachment[];
	blocks6?: BlockAttachment[];
	pools?: string[];
	available_drives?: Drive[];
	instance_drives?: Drive[];
	no_host_network?: boolean;
	no_node_port_network?: boolean;
	host_nat?: boolean;
	default_no_public_address?: boolean;
	default_no_public_address6?: boolean;
	jumbo_frames?: boolean;
	jumbo_frames_internal?: boolean;
	iscsi?: boolean;
	usb_passthrough?: boolean;
	pci_passthrough?: boolean;
	hugepages?: boolean;
	hugepages_size?: number;
	firewall?: boolean;
	network_roles?: string[];
	requests_min?: number;
	cpu_units?: number;
	memory_units?: number;
	cpu_units_res?: number;
	memory_units_res?: number;
	memory?: number;
	hugepages_used?: number;
	load1?: number;
	load5?: number;
	load15?: number;
	public_ips?: string[];
	public_ips6?: string[];
	private_ips?: string[];
	forwarded_for_header?: string;
	forwarded_proto_header?: string;
	software_version?: string;
	hostname?: string;
	oracle_user?: string;
	oracle_public_key?: string;
}

export interface Vpc {
	id?: string;
	name?: string;
	network?: string;
	subnets?: Subnet[];
}

export interface Subnet {
	id?: string;
	vpc_id?: string;
	name?: string;
	network?: string;
}

export interface NodeInit {
	provider?: string;
	zone?: string;
	firewall?: boolean;
	internal_interface?: string;
	external_interface?: string;
	host_network?: string;
	block_gateway?: string;
	block_netmask?: string;
	block_subnets?: string[];
}

export interface Drive {
	id?: string;
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
