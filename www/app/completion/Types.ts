/// <reference path="../References.d.ts"/>
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
	datacenters: DatacenterTypes.DatacentersRo;
	nodes: NodeTypes.NodesRo;
	pools: PoolTypes.PoolsRo;
	zones: ZoneTypes.ZonesRo;
	shapes: ShapeTypes.ShapesRo;
	images: ImageTypes.ImagesRo;
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

export const Values: Record<string, string[]> = {
	"instance": [
		"id",
		"organization",
		"zone",
		"vpc",
		"subnet",
		"oracle_subnet",
		"oracle_vnic",
		"image",
		"state",
		"uefi",
		"secure_boot",
		"tpm",
		"dhcp_server",
		"cloud_type",
		"delete_protection",
		"skip_source_dest_check",
		"qemu_version",
		"public_ips",
		"public_ips6",
		"private_ips",
		"private_ips6",
		"gateway_ips",
		"gateway_ips6",
		"oracle_private_ips",
		"oracle_public_ips",
		"host_ips",
		"network_namespace",
		"no_public_address",
		"no_public_address6",
		"no_host_address",
		"node",
		"shape",
		"name",
		"root_enabled",
		"memory",
		"processors",
		"network_roles",
		"vnc",
		"spice",
		"gui",
		"deployment",
	],
	"vpc": [
		"id",
		"name",
		"vpc_id",
		"network",
		"network6",
	],
	"subnet": [
		"id",
		"name",
		"network",
	],
	"certificate": [
		"id",
		"name",
		"type",
		"key",
		"certificate",
	],
	"secret": [
		"id",
		"name",
		"type",
		"key",
		"value",
		"region",
		"public_key",
		"private_key",
	],
	"unit": [
		"id",
		"name",
		"kind",
		"count",
		"public_ips",
		"public_ips6",
		"healthy_public_ips",
		"healthy_public_ips6",
		"unhealthy_public_ips",
		"unhealthy_public_ips6",
		"private_ips",
		"private_ips6",
		"healthy_private_ips",
		"healthy_private_ips6",
		"unhealthy_private_ips",
		"unhealthy_private_ips6",
		"oracle_private_ips",
		"oracle_public_ips",
		"healthy_oracle_public_ips",
		"healthy_oracle_private_ips",
		"unhealthy_oracle_public_ips",
		"unhealthy_oracle_private_ips",
	],
}
