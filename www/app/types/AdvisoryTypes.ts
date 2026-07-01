/// <reference path="../References.d.ts"/>
export const SYNC = 'advisory.sync';
export const TRAVERSE = 'advisory.traverse';
export const FILTER = 'advisory.filter';
export const CHANGE = 'advisory.change';

export interface Vulnerability {
	id?: string;
	timestamp?: string;
	status?: string;
	description?: string;
	statement?: string;
	score?: number;
	severity?: string;
	vector?: string;
	complexity?: string;
	privileges?: string;
	interaction?: string;
	scope?: string;
	confidentiality?: string;
	integrity?: string;
	availability?: string;
}

export interface InstanceInfo {
	id?: string;
	name?: string;
	status?: string;
	timestamp?: string;
	uptime?: string;
	public_ips?: string[];
	public_ips6?: string[];
	private_ips?: string[];
	private_ips6?: string[];
	cloud_public_ips?: string[];
	cloud_public_ips6?: string[];
}

export interface NodeInfo {
	id?: string;
	name?: string;
	timestamp?: string;
	public_ips?: string[];
	public_ips6?: string[];
	private_ips?: string[];
}

export interface Advisory {
	id?: string;
	organization?: string;
	reference?: string;
	dismissed?: boolean;
	type?: string;
	updated?: string;
	severity?: string;
	description?: string;
	score?: number;
	packages?: string[];
	vulnerabilities?: Vulnerability[];
	instances?: string[];
	nodes?: string[];
	dismissed_resources?: string[];
	instances_info?: InstanceInfo[];
	nodes_info?: NodeInfo[];
}

export interface DismissData {
	dismiss?: boolean;
	restore?: boolean;
	dismissals?: string[];
	restores?: string[];
}

export interface Filter {
	id?: string;
	reference?: string;
	type?: string;
	severity?: string;
	organization?: string;
	dismissed?: boolean;
}

export type Advisories = Advisory[];

export type AdvisoryRo = Readonly<Advisory>;
export type AdvisoriesRo = ReadonlyArray<AdvisoryRo>;

export interface AdvisoryDispatch {
	type: string;
	data?: {
		id?: string;
		advisory?: Advisory;
		advisories?: Advisories;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
