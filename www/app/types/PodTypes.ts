/// <reference path="../References.d.ts"/>
export const SYNC = 'pod.sync';
export const SYNC_UNIT = 'pod.sync_unit';
export const TRAVERSE = 'pod.traverse';
export const FILTER = 'pod.filter';
export const CHANGE = 'pod.change';

export interface Pod {
	id?: string;
	name?: string;
	comment?: string;
	organization?: string;
	delete_protection?: boolean;
	units?: Unit[];
	drafts?: Unit[];
}

export interface Unit {
	id?: string;
	name?: string;
	kind?: string;
	spec?: string;
	spec_index?: number
	last_spec?: string;
	deploy_spec?: string;
	delete?: boolean;
	new?: boolean;
}

export interface PodUnit {
	id?: string;
	kind?: string;
	pod?: string;
	commits?: Commit[]
	deployments?: Deployment[];
}

export interface Commit {
	id?: string
	pod?: string
	unit?: string
	index?: number
	timestamp?: string
	name?: string
	kind?: string
	count?: number
	hash?: string
	data?: string
}

export interface CommitData {
	unit?: string;
	specs?: Commit[];
	count?: number;
	page?: number;
	page_count?: number;
}

export interface Deployment {
	id?: string;
	pod?: string;
	unit?: string;
	timestamp?: string;
	tags?: string[];
	spec?: string;
	kind?: string;
	state?: string;
	action?: string;
	status?: string;
	node?: string;
	instance?: string;
	instance_data?: InstanceData;
	domain_data?: DomainData;
	zone_name?: string;
	node_name?: string;
	spec_offset?: number;
	spec_index?: number;
	spec_timestamp?: string;
	instance_name?: string;
	instance_roles?: string[];
	instance_memory?: number;
	instance_processors?: number;
	instance_status?: string;
	instance_uptime?: string;
	instance_state?: string;
	instance_guest_status?: string;
	instance_timestamp?: string;
	instance_heartbeat?: string;
	instance_memory_usage?: number;
	instance_hugepages?: number;
	instance_load1?: number;
	instance_load5?: number;
	instance_load15?: number;
	image_id?: string;
	image_name?: string;
}

export interface InstanceData {
	public_ips?: string[];
	public_ips6?: string[];
	private_ips?: string[];
	private_ips6?: string[];
	cloud_private_ips?: string[];
	cloud_public_ips?: string[];
}

export interface DomainData {
	records?: RecordData[];
}

export interface RecordData {
	domain?: string;
	value?: string;
}

export interface Filter {
	id?: string;
	name?: string;
	comment?: string;
	role?: string;
	organization?: string;
}

export interface Build {
	id?: string;
	name?: string;
	pod?: string;
	unit?: string;
	tags?: string;
}

export type Pods = Pod[];

export type PodRo = Readonly<Pod>;
export type PodsRo = ReadonlyArray<PodRo>;

export type Units = Unit[];

export type UnitRo = Readonly<Unit>;
export type UnitsRo = ReadonlyArray<UnitRo>;

export interface PodDispatch {
	type: string;
	data?: {
		id?: string;
		pod?: Pod;
		pods?: Pods;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}

export interface PodUnitDispatch {
	type: string;
	data?: {
		unit_id?: string;
		unit?: PodUnit;
	};
}
