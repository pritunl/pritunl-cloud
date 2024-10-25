/// <reference path="../References.d.ts"/>
export const SYNC = 'service.sync';
export const SYNC_UNIT = 'service.sync_unit';
export const TRAVERSE = 'service.traverse';
export const FILTER = 'service.filter';
export const CHANGE = 'service.change';

export interface Service {
	id?: string;
	name?: string;
	comment?: string;
	organization?: string;
	delete_protection?: boolean;
	units?: Unit[];
}

export interface Unit {
	id?: string;
	new?: boolean;
	name?: string;
	spec?: string;
	delete?: boolean;
}

export interface ServiceUnit {
	id?: string;
	service?: string;
	deployments?: Deployment[];
}

export interface Deployment {
	id?: string;
	service?: string;
	unit?: string;
	spec?: string;
	kind?: string;
	state?: string;
	node?: string;
	instance?: string;
	public_ips?: string[];
	public_ips6?: string[];
	private_ips?: string[];
	private_ips6?: string[];
	oracle_private_ips?: string[];
	oracle_public_ips?: string[];
	node_name?: string;
	instance_name?: string;
	instance_roles?: string[];
	instance_memory?: number;
	instance_processors?: number;
	instance_status?: string;
	instance_uptime?: string;
	instance_state?: string;
	instance_virt_state?: string;
}

export interface Filter {
	id?: string;
	name?: string;
	comment?: string;
	role?: string;
	organization?: string;
}

export type Services = Service[];

export type ServiceRo = Readonly<Service>;
export type ServicesRo = ReadonlyArray<ServiceRo>;

export interface ServiceDispatch {
	type: string;
	data?: {
		id?: string;
		service?: Service;
		services?: Services;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}

export interface ServiceUnitDispatch {
	type: string;
	data?: {
		unit_id?: string;
		unit?: ServiceUnit;
	};
}
