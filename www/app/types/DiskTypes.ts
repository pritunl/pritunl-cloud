/// <reference path="../References.d.ts"/>
export const SYNC = 'disk.sync';
export const TRAVERSE = 'disk.traverse';
export const FILTER = 'disk.filter';
export const CHANGE = 'disk.change';

export interface Disk {
	id?: string;
	name?: string;
	comment?: string;
	type?: string;
	uuid?: string;
	node?: string;
	pool?: string;
	organization?: string;
	state?: string;
	action?: string;
	instance?: string;
	delete_protection?: boolean;
	file_system?: string;
	image?: string;
	restore_image?: string;
	backing?: boolean;
	backing_image?: string;
	index?: string;
	size?: number;
	new_size?: number;
	backup?: boolean;
	backups?: Backup[];
}

export interface Filter {
	id?: string;
	name?: string;
	organization?: string;
	datacenter?: string;
	instance?: string;
	node?: string;
}

export interface Backup {
	image?: string;
	name?: string;
}

export type Disks = Disk[];

export type DiskRo = Readonly<Disk>;
export type DisksRo = ReadonlyArray<DiskRo>;

export interface DiskDispatch {
	type: string;
	data?: {
		id?: string;
		disk?: Disk;
		disks?: Disks;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
