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
