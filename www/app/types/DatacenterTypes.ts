/// <reference path="../References.d.ts"/>
export const SYNC = 'datacenter.sync';
export const SYNC_NAMES = 'datacenter.sync_names';
export const TRAVERSE = 'datacenter.traverse';
export const FILTER = 'datacenter.filter';
export const CHANGE = 'datacenter.change';

export interface Datacenter {
	id?: string;
	name?: string;
	comment?: string;
	match_organizations?: boolean;
	organizations?: string[];
	public_storages?: string[];
	private_storage?: string;
	private_storage_class?: string;
	backup_storage?: string;
	backup_storage_class?: string;
	network_mode?: string;
}

export interface Filter {
	id?: string;
	name?: string;
	comment?: string;
}

export type Datacenters = Datacenter[];

export type DatacenterRo = Readonly<Datacenter>;
export type DatacentersRo = ReadonlyArray<DatacenterRo>;

export interface DatacenterDispatch {
	type: string;
	data?: {
		id?: string;
		datacenter?: Datacenter;
		datacenters?: Datacenters;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
