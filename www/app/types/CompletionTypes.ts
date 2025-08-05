/// <reference path="../References.d.ts"/>

import * as OrganizationTypes from "./OrganizationTypes"
import * as AuthorityTypes from "./AuthorityTypes"
import * as PolicyTypes from "./PolicyTypes"
import * as DomainTypes from "./DomainTypes"
import * as BalancerTypes from "./BalancerTypes"
import * as VpcTypes from "./VpcTypes"
import * as DatacenterTypes from "./DatacenterTypes"
import * as BlockTypes from "./BlockTypes"
import * as NodeTypes from "./NodeTypes"
import * as DiskTypes from "./DiskTypes"
import * as PoolTypes from "./PoolTypes"
import * as ZoneTypes from "./ZoneTypes"
import * as ShapeTypes from "./ShapeTypes"
import * as ImageTypes from "./ImageTypes"
import * as InstanceTypes from "./InstanceTypes"
import * as FirewallTypes from "./FirewallTypes"
import * as PlanTypes from "./PlanTypes"
import * as CertificateTypes from "./CertificateTypes"
import * as SecretTypes from "./SecretTypes"
import * as PodTypes from "./PodTypes"

export const SYNC = "completion.sync"
export const FILTER = "completion.filter"
export const CHANGE = "completion.change"

export interface Completion {
	organizations?: OrganizationTypes.Organization[]
	authorities?: AuthorityTypes.Authority[]
	policies?: PolicyTypes.Policy[]
	domains?: DomainTypes.Domain[]
	balancers?: BalancerTypes.Balancer[]
	vpcs?: VpcTypes.Vpc[]
	subnets?: VpcTypes.Subnet[]
	datacenters?: DatacenterTypes.Datacenter[]
	blocks?: BlockTypes.Block[]
	nodes?: NodeTypes.Node[]
	disks?: DiskTypes.Disk[]
	pools?: PoolTypes.Pool[]
	zones?: ZoneTypes.Zone[]
	shapes?: ShapeTypes.Shape[]
	images?: ImageTypes.Image[]
	builds?: Build[];
	instances?: InstanceTypes.Instance[]
	firewalls?: FirewallTypes.Firewall[]
	plans?: PlanTypes.Plan[]
	certificates?: CertificateTypes.Certificate[]
	secrets?: SecretTypes.Secret[]
	pods?: PodTypes.Pod[]
	units?: PodTypes.Unit[]
}

export interface CompletionMap {
	organizations?: {[key: string]: number}
	authorities?: {[key: string]: number}
	policies?: {[key: string]: number}
	domains?: {[key: string]: number}
	balancers?: {[key: string]: number}
	vpcs?: {[key: string]: number}
	subnets?: {[key: string]: number}
	datacenters?: {[key: string]: number}
	blocks?: {[key: string]: number}
	nodes?: {[key: string]: number}
	disks?: {[key: string]: number}
	pools?: {[key: string]: number}
	zones?: {[key: string]: number}
	shapes?: {[key: string]: number}
	images?: {[key: string]: number}
	builds?: {[key: string]: number}
	instances?: {[key: string]: number}
	firewalls?: {[key: string]: number}
	plans?: {[key: string]: number}
	certificates?: {[key: string]: number}
	secrets?: {[key: string]: number}
	pods?: {[key: string]: number}
	units?: {[key: string]: number}
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
