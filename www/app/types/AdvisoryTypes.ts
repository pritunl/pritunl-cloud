/// <reference path="../References.d.ts"/>
export const SYNC = 'advisory.sync';
export const TRAVERSE = 'advisory.traverse';
export const FILTER = 'advisory.filter';
export const CHANGE = 'advisory.change';

export interface Vulnerability {
	id?: string;
	source?: string;
	timestamp?: string;
	published?: string;
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
	analysis?: Analysis;
	sync?: string;
}

export interface Analysis {
	updated?: string;
	timestamp?: string;
	model?: string;
	summary?: string;
	impacted?: string;
	score?: number;
	universal?: boolean;
	processes?: string[];
	modules?: string[];
	ports?: AnalysisPort[];
	evidence?: AnalysisEvidence[];
	boundaries?: string[];
	mitigation?: string;
}

export interface AnalysisPort {
	name?: string;
	protocol?: string;
	port?: number;
}

export interface AnalysisEvidence {
	type?: string;
	url?: string;
	note?: string;
}

export interface InstanceInfo {
	id?: string;
	name?: string;
	state?: string;
	dismissed?: boolean;
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
	state?: string;
	dismissed?: boolean;
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
	vuxmls?: string[];
	vulnerabilities?: string[];
	pending?: number;
	complete?: boolean;
	instance_count?: number;
	node_count?: number;
}

export interface Resource {
	id?: string;
	organization?: string;
	reference?: string;
	resource?: string;
	kind?: string;
	state?: string;
	dismissed?: boolean;
	exclusions?: string[];
	timestamp?: string;
}

export interface ResourceAdvisory extends Advisory {
	vulnerability_docs?: Vulnerability[];
	resource?: Resource;
}

export interface DetailCounts {
	nodes?: number;
	unreachable_nodes?: number;
	dismissed_nodes?: number;
	instances?: number;
	unreachable_instances?: number;
	dismissed_instances?: number;
	active?: number;
}

export interface Detail {
	id?: string;
	vulnerabilities?: Vulnerability[];
	instances_info?: InstanceInfo[];
	nodes_info?: NodeInfo[];
	counts?: DetailCounts;
	exclusions?: {[cve: string]: number};
	page?: number;
	page_count?: number;
	count?: number;
}

export interface DismissData {
	dismiss?: boolean;
	restore?: boolean;
	dismissals?: string[];
	restores?: string[];
}

export interface MultiData {
	ids?: string[];
	dismiss?: boolean;
	restore?: boolean;
}

export interface Filter {
	id?: string;
	reference?: string;
	cve?: string;
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
