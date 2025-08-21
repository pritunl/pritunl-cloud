/// <reference path="../References.d.ts"/>
import * as CompletionTypes from "../types/CompletionTypes";
import * as OrganizationTypes from "../types/OrganizationTypes";
import * as DomainTypes from "../types/DomainTypes";
import * as VpcTypes from "../types/VpcTypes";
import * as DatacenterTypes from "../types/DatacenterTypes";
import * as NodeTypes from "../types/NodeTypes";
import * as PoolTypes from "../types/PoolTypes";
import * as ZoneTypes from "../types/ZoneTypes";
import * as ShapeTypes from "../types/ShapeTypes";
import * as ImageTypes from "../types/ImageTypes";
import * as InstanceTypes from "../types/InstanceTypes";
import * as PlanTypes from "../types/PlanTypes";
import * as CertificateTypes from "../types/CertificateTypes";
import * as SecretTypes from "../types/SecretTypes";
import * as PodTypes from "../types/PodTypes";

export interface Resources {
	organizations: OrganizationTypes.OrganizationsRo;
	domains: DomainTypes.DomainsRo;
	vpcs: VpcTypes.VpcsRo;
	subnets: VpcTypes.Subnet[];
	datacenters: DatacenterTypes.DatacentersRo;
	nodes: NodeTypes.NodesRo;
	pools: PoolTypes.PoolsRo;
	zones: ZoneTypes.ZonesRo;
	shapes: ShapeTypes.ShapesRo;
	images: ImageTypes.ImagesRo;
	builds: CompletionTypes.Build[];
	instances: InstanceTypes.InstancesRo;
	plans: PlanTypes.PlansRo;
	certificates: CertificateTypes.CertificatesRo;
	secrets: SecretTypes.SecretsRo;
	pods: PodTypes.PodsRo;
	units: PodTypes.UnitsRo;
}

export interface Kind {
	name: string
	label: string
	title: string
}

export interface Resource {
	id: string
	name: string
	info: ResourceInfo[]
}

export interface ResourceInfo {
	label: string
	value: string | number
}

export interface Dispatch {
	type: string
}

export type SelectorInfo = {
	label: string;
	tooltip: string;
}

export const Selectors: Record<string, Record<string, SelectorInfo>> = {
	"instance": {
		"id": {
			label: "ID",
			tooltip: "Unique identifier of the instance"
		},
		"organization": {
			label: "Organization",
			tooltip: "Organization the instance belongs to"
		},
		"zone": {
			label: "Zone",
			tooltip: "Availability zone where the instance is deployed"
		},
		"vpc": {
			label: "VPC",
			tooltip: "Virtual Private Cloud network the instance is connected to"
		},
		"subnet": {
			label: "Subnet",
			tooltip: "Subnet within the VPC where the instance resides"
		},
		"cloud_subnet": {
			label: "Cloud Subnet",
			tooltip: "Cloud Cloud subnet configuration"
		},
		"cloud_vnic": {
			label: "Cloud VNIC",
			tooltip: "Cloud virtual network interface"
		},
		"image": {
			label: "Image",
			tooltip: "Base image used for the instance"
		},
		"state": {
			label: "State",
			tooltip: "Current operational state of the instance"
		},
		"uefi": {
			label: "UEFI",
			tooltip: "Unified Extensible Firmware Interface status"
		},
		"secure_boot": {
			label: "Secure Boot",
			tooltip: "Status of secure boot feature"
		},
		"tpm": {
			label: "TPM",
			tooltip: "Trusted Platform Module status"
		},
		"dhcp_server": {
			label: "DHCP Server",
			tooltip: "Dynamic Host Configuration Protocol server status"
		},
		"cloud_type": {
			label: "Cloud Type",
			tooltip: "Type of cloud infrastructure being used"
		},
		"delete_protection": {
			label: "Delete Protection",
			tooltip: "Status of deletion protection feature"
		},
		"skip_source_dest_check": {
			label: "Skip Source/Dest Check",
			tooltip: "Status of source/destination checking"
		},
		"qemu_version": {
			label: "QEMU Version",
			tooltip: "Version of QEMU virtualization software"
		},
		"public_ips": {
			label: "Public IPs",
			tooltip: "List of public IPv4 addresses"
		},
		"public_ips6": {
			label: "Public IPv6",
			tooltip: "List of public IPv6 addresses"
		},
		"private_ips": {
			label: "Private IPs",
			tooltip: "List of private IPv4 addresses"
		},
		"private_ips6": {
			label: "Private IPv6",
			tooltip: "List of private IPv6 addresses"
		},
		"gateway_ips": {
			label: "Gateway IPs",
			tooltip: "IPv4 gateway addresses"
		},
		"gateway_ips6": {
			label: "Gateway IPv6",
			tooltip: "IPv6 gateway addresses"
		},
		"cloud_private_ips": {
			label: "Cloud Private IPs",
			tooltip: "Cloud private IP addresses"
		},
		"cloud_public_ips": {
			label: "Cloud Public IPs",
			tooltip: "Cloud public IP addresses"
		},
		"host_ips": {
			label: "Host IPs",
			tooltip: "IP addresses of the host machine"
		},
		"network_namespace": {
			label: "Network Namespace",
			tooltip: "Network namespace configuration"
		},
		"no_public_address": {
			label: "No Public Address",
			tooltip: "Indicates if public IPv4 addressing is disabled"
		},
		"no_public_address6": {
			label: "No Public IPv6",
			tooltip: "Indicates if public IPv6 addressing is disabled"
		},
		"no_host_address": {
			label: "No Host Address",
			tooltip: "Indicates if host addressing is disabled"
		},
		"node": {
			label: "Node",
			tooltip: "Physical or virtual node where the instance runs"
		},
		"shape": {
			label: "Shape",
			tooltip: "Instance type and size configuration"
		},
		"name": {
			label: "Name",
			tooltip: "Display name of the instance"
		},
		"root_enabled": {
			label: "Root Enabled",
			tooltip: "Status of root access"
		},
		"memory": {
			label: "Memory",
			tooltip: "Allocated RAM"
		},
		"processors": {
			label: "Processors",
			tooltip: "Number of allocated CPU cores"
		},
		"roles": {
			label: "Roles",
			tooltip: "Access roles assigned to the instance"
		},
		"vnc": {
			label: "VNC",
			tooltip: "Virtual Network Computing status"
		},
		"spice": {
			label: "SPICE",
			tooltip: "Simple Protocol for Independent Computing Environments status"
		},
		"gui": {
			label: "GUI",
			tooltip: "Graphical User Interface status"
		},
		"deployment": {
			label: "Deployment",
			tooltip: "Deployment configuration details"
		}
	},
	"vpc": {
		"id": {
			label: "ID",
			tooltip: "Unique identifier of the VPC"
		},
		"name": {
			label: "Name",
			tooltip: "Display name of the VPC"
		},
		"vpc_id": {
			label: "VPC ID",
			tooltip: "Cloud provider's VPC identifier"
		},
		"network": {
			label: "Network",
			tooltip: "IPv4 network configuration"
		},
		"network6": {
			label: "Network IPv6",
			tooltip: "IPv6 network configuration"
		}
	},
	"subnet": {
		"id": {
			label: "ID",
			tooltip: "Unique identifier of the subnet"
		},
		"name": {
			label: "Name",
			tooltip: "Display name of the subnet"
		},
		"network": {
			label: "Network",
			tooltip: "Network address range of the subnet"
		}
	},
	"certificate": {
		"id": {
			label: "ID",
			tooltip: "Unique identifier of the certificate"
		},
		"name": {
			label: "Name",
			tooltip: "Display name of the certificate"
		},
		"type": {
			label: "Type",
			tooltip: "Type of certificate"
		},
		"key": {
			label: "Key",
			tooltip: "Certificate key information"
		},
		"certificate": {
			label: "Certificate",
			tooltip: "Certificate content"
		}
	},
	"secret": {
		"id": {
			label: "ID",
			tooltip: "Unique identifier of the secret"
		},
		"name": {
			label: "Name",
			tooltip: "Display name of the secret"
		},
		"type": {
			label: "Type",
			tooltip: "Type of secret"
		},
		"key": {
			label: "Key",
			tooltip: "Secret key identifier"
		},
		"value": {
			label: "Value",
			tooltip: "Protected secret value"
		},
		"region": {
			label: "Region",
			tooltip: "Region where the secret is stored"
		},
		"public_key": {
			label: "Public Key",
			tooltip: "Public key component"
		},
		"private_key": {
			label: "Private Key",
			tooltip: "Private key component"
		}
	},
	"unit": {
		"id": {
			label: "ID",
			tooltip: "Unique identifier of the unit"
		},
		"name": {
			label: "Name",
			tooltip: "Display name of the unit"
		},
		"kind": {
			label: "Kind",
			tooltip: "Type of unit"
		},
		"count": {
			label: "Count",
			tooltip: "Number of instances in the unit"
		},
		"public_ips": {
			label: "Public IPs",
			tooltip: "List of public IPv4 addresses"
		},
		"public_ips6": {
			label: "Public IPv6",
			tooltip: "List of public IPv6 addresses"
		},
		"healthy_public_ips": {
			label: "Healthy Public IPs",
			tooltip: "List of healthy public IPv4 addresses"
		},
		"healthy_public_ips6": {
			label: "Healthy Public IPv6",
			tooltip: "List of healthy public IPv6 addresses"
		},
		"unhealthy_public_ips": {
			label: "Unhealthy Public IPs",
			tooltip: "List of unhealthy public IPv4 addresses"
		},
		"unhealthy_public_ips6": {
			label: "Unhealthy Public IPv6",
			tooltip: "List of unhealthy public IPv6 addresses"
		},
		"private_ips": {
			label: "Private IPs",
			tooltip: "List of private IPv4 addresses"
		},
		"private_ips6": {
			label: "Private IPv6",
			tooltip: "List of private IPv6 addresses"
		},
		"healthy_private_ips": {
			label: "Healthy Private IPs",
			tooltip: "List of healthy private IPv4 addresses"
		},
		"healthy_private_ips6": {
			label: "Healthy Private IPv6",
			tooltip: "List of healthy private IPv6 addresses"
		},
		"unhealthy_private_ips": {
			label: "Unhealthy Private IPs",
			tooltip: "List of unhealthy private IPv4 addresses"
		},
		"unhealthy_private_ips6": {
			label: "Unhealthy Private IPv6",
			tooltip: "List of unhealthy private IPv6 addresses"
		},
		"cloud_private_ips": {
			label: "Cloud Private IPs",
			tooltip: "List of cloud private IP addresses"
		},
		"cloud_public_ips": {
			label: "Cloud Public IPs",
			tooltip: "List of cloud public IP addresses"
		},
		"healthy_cloud_public_ips": {
			label: "Healthy Cloud Public IPs",
			tooltip: "List of healthy cloud public IP addresses"
		},
		"healthy_cloud_private_ips": {
			label: "Healthy Cloud Private IPs",
			tooltip: "List of healthy cloud private IP addresses"
		},
		"unhealthy_cloud_public_ips": {
			label: "Unhealthy Cloud Public IPs",
			tooltip: "List of unhealthy cloud public IP addresses"
		},
		"unhealthy_cloud_private_ips": {
			label: "Unhealthy Cloud Private IPs",
			tooltip: "List of unhealthy cloud private IP addresses"
		}
	}
};
