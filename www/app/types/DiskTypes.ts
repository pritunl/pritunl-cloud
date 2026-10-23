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
	deployment?: string;
	delete_protection?: boolean;
	file_system?: string;
	image?: string;
	restore_image?: string;
	index?: string;
	size?: number;
	lv_size?: number;
	new_size?: number;
	backup?: boolean;
	backups?: Backup[];
	instance_info?: InstanceInfo;
	metadata?: Metadata;
	action_data?: ActionData;
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

export interface Metadata {
	timestamp?: string;
	size?: number;
	table?: string;
	partitions?: Partition[];
}

export interface Partition {
	number?: number;
	start?: number;
	size?: number;
	file_system?: string;
	uuid?: string;
	label?: string;
	volume_group?: VolumeGroup;
}

export interface VolumeGroup {
	name?: string;
	size?: number;
	free?: number;
	extent_size?: number;
	volumes?: LogicalVolume[];
}

export interface LogicalVolume {
	name?: string;
	size?: number;
	file_system?: string;
	uuid?: string;
	label?: string;
}

export interface ActionData {
	partitions?: ResizePartition[];
	volumes?: ResizeVolume[];
}

export interface ResizePartition {
	number?: number;
	size?: number;
	fill?: boolean;
}

export interface ResizeVolume {
	volume_group?: string;
	name?: string;
	size?: number;
	fill?: boolean;
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
