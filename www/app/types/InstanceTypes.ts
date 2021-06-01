/// <reference path="../References.d.ts"/>
export const SYNC = 'instance.sync';
export const SYNC_NODE = 'instance.sync_node';
export const TRAVERSE = 'instance.traverse';
export const FILTER = 'instance.filter';
export const CHANGE = 'instance.change';

export interface Instance {
	id?: string;
	organization?: string;
	zone?: string;
	node?: string;
	image?: string;
	image_backing?: boolean;
	status?: string;
	uptime?: string;
	state?: string;
	vm_state?: string;
	vm_timestamp?: string;
	uefi?: boolean;
	secure_boot?: boolean;
	delete_protection?: boolean;
	public_ips?: string[];
	public_ips6?: string[];
	private_ips?: string[];
	private_ips6?: string[];
	gateway_ips?: string[];
	gateway_ips6?: string[];
	network_namespace?: string;
	host_ips?: string[];
	public_mac?: string;
	name?: string;
	comment?: string;
	init_disk_size?: number;
	memory?: number;
	processors?: number;
	network_roles?: string[];
	isos?: Iso[];
	usb_devices?: UsbDevice[];
	pci_devices?: PciDevice[];
	drive_devices?: DriveDevice[];
	iscsi_devices?: IscsiDevice[];
	vnc?: boolean;
	vnc_password?: string;
	vnc_display?: number;
	domain?: string;
	no_public_address?: boolean;
	no_host_address?: boolean;
	vpc?: string;
	subnet?: string;
	count?: number;
	info?: Info;
}

export interface Filter {
	id?: string;
	name?: string;
	state?: string;
	network_role?: string;
	network_namespace?: string;
	organization?: string;
	node?: string;
	zone?: string;
	vpc?: string;
	subnet?: string;
}

export interface Iso {
	name?: string;
}

export interface UsbDevice {
	name?: string;
	vendor?: string;
	product?: string;
	bus?: string;
	address?: string;
}

export interface PciDevice {
	slot?: string;
	class?: string;
	name?: string;
	driver?: string;
}

export interface IscsiDevice {
	host?: string;
	port?: number;
	iqn?: string;
	lun?: string;
	username?: string;
	password?: string;
	uri?: string;
}

export interface DriveDevice {
	id?: string;
}

export interface Info {
	node?: string;
	iscsi?: boolean;
	firewall_rules?: string[];
	authorities?: string[];
	disks?: string[];
	isos?: Iso[];
	usb_devices?: UsbDevice[];
	pci_devices?: PciDevice[];
	drive_devices?: DriveDevice[];
}

export type Instances = Instance[];
export type InstancesNode = Map<string, Instances>;

export type InstanceRo = Readonly<Instance>;
export type InstancesRo = ReadonlyArray<InstanceRo>;
export type InstancesNodeRo = Map<string, InstancesRo>;

export interface InstanceDispatch {
	type: string;
	data?: {
		id?: string;
		node?: string;
		instance?: Instance;
		instances?: Instances;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
