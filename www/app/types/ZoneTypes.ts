/// <reference path="../References.d.ts"/>
export const SYNC = 'zone.sync';
export const CHANGE = 'zone.change';

export interface Zone {
	id: string;
	datacenter?: string;
	name?: string;
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
	};
}
