/// <reference path="../References.d.ts"/>
export const SYNC = 'zone.sync';
export const TRAVERSE = 'zone.traverse';
export const FILTER = 'zone.filter';
export const CHANGE = 'zone.change';

export interface Zone {
	id?: string;
	datacenter?: string;
	name?: string;
	comment?: string;
	network_mode?: string;
}

export interface Filter {
	id?: string;
	name?: string;
	comment?: string;
}

export type Zones = Zone[];

export type ZoneRo = Readonly<Zone>;
export type ZonesRo = ReadonlyArray<ZoneRo>;

export interface ZoneDispatch {
	type: string;
	data?: {
		id?: string;
		zone?: Zone;
		zones?: Zones;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
