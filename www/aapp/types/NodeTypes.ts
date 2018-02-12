/// <reference path="../References.d.ts"/>
export const SYNC = 'node.sync';
export const TRAVERSE = 'node.traverse';
export const FILTER = 'node.filter';
export const CHANGE = 'node.change';

export interface Node {
	id: string;
	type?: string;
	name?: string;
	port?: number;
	protocol?: string;
	timestamp?: string;
	admin_domain?: string;
	user_domain?: string;
	certificates?: string[];
	requests_min?: number;
	memory?: number;
	load1?: number;
	load5?: number;
	load15?: number;
	forwarded_for_header?: string;
}

export interface Filter {
	name?: string;
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
