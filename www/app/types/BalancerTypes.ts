/// <reference path="../References.d.ts"/>
export const SYNC = 'balancer.sync';
export const TRAVERSE = 'balancer.traverse';
export const FILTER = 'balancer.filter';
export const CHANGE = 'balancer.change';

export interface Domain {
	domain?: string;
	host?: string;
}

export interface Backend {
	protocol?: string;
	hostname?: string;
	port?: number;
}

export interface State {
	timestamp?: string;
	requests?: number;
	retries?: number;
	online?: string[];
	unknown_high?: string[];
	unknown_mid?: string[];
	unknown_low?: string[];
	offline?: string[];
}

export interface Balancer {
	id?: string;
	name?: string;
	comment?: string;
	state?: boolean;
	type?: string;
	organization?: string;
	datacenter?: string;
	certificates?: string[];
	web_sockets?: boolean;
	domains?: Domain[];
	backends?: Backend[];
	check_path?: string;
	states?: {[key: string]: State};
}

export interface Filter {
	id?: string;
	name?: string;
	organization?: string;
	datacenter?: string;
}

export type Balancers = Balancer[];

export type BalancerRo = Readonly<Balancer>;
export type BalancersRo = ReadonlyArray<BalancerRo>;

export interface BalancerDispatch {
	type: string;
	data?: {
		id?: string;
		balancer?: Balancer;
		balancers?: Balancers;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
