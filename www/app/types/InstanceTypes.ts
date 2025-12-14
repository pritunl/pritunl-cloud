/// <reference path="../References.d.ts"/>
export const SYNC = 'instance.sync';
export const SYNC_NODE = 'instance.sync_node';
export const TRAVERSE = 'instance.traverse';
export const FILTER = 'instance.filter';
export const CHANGE = 'instance.change';

import * as PageInfos from '../components/PageInfo';

export interface Instance {
	id?: string;
	organization?: string;
	zone?: string;
	node?: string;
	shape?: string;
	image?: string;
	image_backing?: boolean;
	disk_type?: string;
	disk_pool?: string;
	status?: string;
	status_info?: StatusInfo;
	uptime?: string;
	state?: string;
	action?: string;
	timestamp?: string;
	uefi?: boolean;
	secure_boot?: boolean;
	tpm?: boolean;
	dhcp_server?: boolean;
	cloud_type?: string;
	cloud_script?: string;
	delete_protection?: boolean;
	skip_source_dest_check?: boolean;
	qemu_version?: string;
	public_ips?: string[];
	public_ips6?: string[];
	private_ips?: string[];
	private_ips6?: string[];
	gateway_ips?: string[];
	gateway_ips6?: string[];
	cloud_private_ips?: string[];
	cloud_public_ips?: string[];
	cloud_public_ips6?: string[];
	network_namespace?: string;
	host_ips?: string[];
	node_port_ips?: string[];
	public_mac?: string;
	name?: string;
	comment?: string;
	init_disk_size?: number;
	memory?: number;
	processors?: number;
	roles?: string[];
	isos?: Iso[];
	usb_devices?: UsbDevice[];
	pci_devices?: PciDevice[];
	drive_devices?: DriveDevice[];
	iscsi_devices?: IscsiDevice[];
	mounts?: Mount[];
	root_enabled?: boolean;
	root_passwd?: string;
	vnc?: boolean;
	vnc_password?: string;
	vnc_display?: number;
	spice?: boolean;
	spice_password?: string;
	spice_port?: number;
	gui?: boolean;
	no_public_address?: boolean;
	no_public_address6?: boolean;
	no_host_address?: boolean;
	vpc?: string;
	subnet?: string;
	cloud_subnet?: string;
	count?: number;
	guest?: Guest;
	info?: Info;
}

export interface Filter {
	id?: string;
	name?: string;
	comment?: string;
	state?: string;
	role?: string;
	network_namespace?: string;
	organization?: string;
	node?: string;
	zone?: string;
	vpc?: string;
	subnet?: string;
}

export interface StatusInfo {
	download_progress: number;
	download_speed: number;
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

export interface Mount {
	name?: string;
	type?: string;
	host_path?: string;
}

export interface DriveDevice {
	id?: string;
}

export interface CloudSubnet {
	id?: string;
	name?: string;
}

export interface Guest {
	timestamp?: string;
	heartbeat?: string;
	memory?: number;
	hugepages?: number;
	load1?: number;
	load5?: number;
	load15?: number;
}

export interface Info {
	node?: string;
	node_public_ip?: string;
	mtu?: number;
	iscsi?: boolean;
	firewall_rules?: Record<string, string>;
	authorities?: string[];
	disks?: string[];
	isos?: Iso[];
	usb_devices?: UsbDevice[];
	pci_devices?: PciDevice[];
	drive_devices?: DriveDevice[];
	cloud_subnets?: CloudSubnet[];
}

export function FirewallFields(info: Info): PageInfos.Field[] {
	if (!info.firewall_rules) {
		return [];
	}

	return Object.entries(info.firewall_rules).map(([key, value]) => ({
		label: key,
		value: value,
	}));
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
		scope?: string;
		instance?: Instance;
		instances?: Instances;
		page?: number;
		pageCount?: number;
		filter?: Filter;
		count?: number;
	};
}
