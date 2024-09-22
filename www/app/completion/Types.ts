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
