/// <reference path="../References.d.ts"/>

import * as OrganizationTypes from "./OrganizationTypes"
import * as DomainTypes from "./DomainTypes"
import * as VpcTypes from "./VpcTypes"
import * as DatacenterTypes from "./DatacenterTypes"
import * as NodeTypes from "./NodeTypes"
import * as PoolTypes from "./PoolTypes"
import * as ZoneTypes from "./ZoneTypes"
import * as ShapeTypes from "./ShapeTypes"
import * as ImageTypes from "./ImageTypes"
import * as InstanceTypes from "./InstanceTypes"
import * as PlanTypes from "./PlanTypes"
import * as CertificateTypes from "./CertificateTypes"
import * as SecretTypes from "./SecretTypes"
import * as PodTypes from "./PodTypes"

export const SYNC = "completion.sync"
export const FILTER = "completion.filter"
export const CHANGE = "completion.change"

export interface Completion {
	organizations?: OrganizationTypes.Organization[]
	domains?: DomainTypes.Domain[]
	vpcs?: VpcTypes.Vpc[]
	subnets?: VpcTypes.Subnet[]
	datacenters?: DatacenterTypes.Datacenter[]
	nodes?: NodeTypes.Node[]
	pools?: PoolTypes.Pool[]
	zones?: ZoneTypes.Zone[]
	shapes?: ShapeTypes.Shape[]
	images?: ImageTypes.Image[]
	builds?: Build[];
	instances?: InstanceTypes.Instance[]
	plans?: PlanTypes.Plan[]
	certificates?: CertificateTypes.Certificate[]
	secrets?: SecretTypes.Secret[]
	pods?: PodTypes.Pod[]
	units?: PodTypes.Unit[]
}

export interface Build {
	id?: string
	name?: string
	pod?: string
	organization?: string
	tags?: BuildTag[]
}

export interface BuildTag {
	tag?: string
	timestamp?: string
}

export interface Filter {
}

export interface CompletionDispatch {
	type: string
	data?: {
		completion?: Completion
		filter?: Filter
	}
}
