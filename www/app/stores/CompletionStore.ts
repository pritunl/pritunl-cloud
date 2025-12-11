/// <reference path="../References.d.ts"/>
import Dispatcher from "../dispatcher/Dispatcher"
import EventEmitter from "../EventEmitter"
import * as Router from '../Router';
import * as Constants from '../Constants';
import * as CompletionTypes from "../types/CompletionTypes"
import * as OrganizationTypes from "../types/OrganizationTypes"
import * as AuthorityTypes from "../types/AuthorityTypes"
import * as PolicyTypes from "../types/PolicyTypes"
import * as DomainTypes from "../types/DomainTypes"
import * as BalancerTypes from "../types/BalancerTypes"
import * as VpcTypes from "../types/VpcTypes"
import * as DatacenterTypes from "../types/DatacenterTypes"
import * as BlockTypes from "../types/BlockTypes"
import * as NodeTypes from "../types/NodeTypes"
import * as DiskTypes from "../types/DiskTypes"
import * as PoolTypes from "../types/PoolTypes"
import * as ZoneTypes from "../types/ZoneTypes"
import * as ShapeTypes from "../types/ShapeTypes"
import * as ImageTypes from "../types/ImageTypes"
import * as StorageTypes from "../types/StorageTypes"
import * as InstanceTypes from "../types/InstanceTypes"
import * as FirewallTypes from "../types/FirewallTypes"
import * as PlanTypes from "../types/PlanTypes"
import * as CertificateTypes from "../types/CertificateTypes"
import * as SecretTypes from "../types/SecretTypes"
import * as PodTypes from "../types/PodTypes"
import * as GlobalTypes from "../types/GlobalTypes"

class CompletionStore extends EventEmitter {
	_userOrg: string;
	_data: CompletionTypes.Completion = Object.freeze({})
	_map: CompletionTypes.CompletionMap = Object.freeze({})
	_filter: CompletionTypes.Filter = null;
	_token = Dispatcher.register((this._callback).bind(this))

	_reset(userOrg: string): void {
		this._userOrg = userOrg;
		this._data = Object.freeze({})
		this._map = Object.freeze({})
		this._filter = null
		this.emitChange()
	}

	get userOrganization(): string {
		return this._userOrg;
	}

	get completion(): CompletionTypes.Completion {
		return this._data
	}

	get filter(): CompletionTypes.Filter {
		return this._filter;
	}

	get organizations(): OrganizationTypes.OrganizationsRo {
		return this._data.organizations || [];
	}

	organization(id: string): OrganizationTypes.OrganizationRo {
		let index = this._map?.organizations?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.organizations[index];
	}

	get authorities(): AuthorityTypes.AuthoritiesRo {
		return this._data.authorities || [];
	}

	authority(id: string): AuthorityTypes.AuthorityRo {
		let index = this._map?.authorities?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.authorities[index];
	}

	get policies(): PolicyTypes.PoliciesRo {
		return this._data.policies || [];
	}

	policy(id: string): PolicyTypes.PolicyRo {
		let index = this._map?.policies?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.policies[index];
	}

	get domains(): DomainTypes.DomainsRo {
		return this._data.domains || [];
	}

	domain(id: string): DomainTypes.DomainRo {
		let index = this._map?.domains?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.domains[index];
	}

	get balancers(): BalancerTypes.BalancersRo {
		return this._data.balancers || [];
	}

	balancer(id: string): BalancerTypes.BalancerRo {
		let index = this._map?.balancers?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.balancers[index];
	}

	get vpcs(): VpcTypes.VpcsRo {
		return this._data.vpcs || [];
	}

	vpc(id: string): VpcTypes.VpcRo {
		let index = this._map?.vpcs?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.vpcs[index];
	}

	get subnets(): VpcTypes.Subnet[] {
		return this._data.subnets || [];
	}

	subnet(id: string): VpcTypes.Subnet {
		let index = this._map?.subnets?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.subnets[index];
	}

	get datacenters(): DatacenterTypes.DatacentersRo {
		return this._data.datacenters || [];
	}

	datacenter(id: string): DatacenterTypes.DatacenterRo {
		let index = this._map?.datacenters?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.datacenters[index];
	}

	get blocks(): BlockTypes.BlocksRo {
		return this._data.blocks || [];
	}

	block(id: string): BlockTypes.BlockRo {
		let index = this._map?.blocks?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.blocks[index];
	}

	get nodes(): NodeTypes.NodesRo {
		return this._data.nodes || [];
	}

	node(id: string): NodeTypes.NodeRo {
		let index = this._map?.nodes?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.nodes[index];
	}

	get disks(): DiskTypes.DisksRo {
		return this._data.disks || [];
	}

	disk(id: string): DiskTypes.DiskRo {
		let index = this._map?.disks?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.disks[index];
	}

	get pools(): PoolTypes.PoolsRo {
		return this._data.pools || [];
	}

	pool(id: string): PoolTypes.PoolRo {
		let index = this._map?.pools?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.pools[index];
	}

	get zones(): ZoneTypes.ZonesRo {
		return this._data.zones || [];
	}

	zone(id: string): ZoneTypes.ZoneRo {
		let index = this._map?.zones?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.zones[index];
	}

	get shapes(): ShapeTypes.ShapesRo {
		return this._data.shapes || [];
	}

	shape(id: string): ShapeTypes.ShapeRo {
		let index = this._map?.shapes?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.shapes[index];
	}

	get images(): ImageTypes.ImagesRo {
		return this._data.images || [];
	}

	image(id: string): ImageTypes.ImageRo {
		let index = this._map?.images?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.images[index];
	}

	get storages(): StorageTypes.StoragesRo {
		return this._data.storages || [];
	}

	storage(id: string): StorageTypes.StorageRo {
		let index = this._map?.storages?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.storages[index];
	}

	get builds(): CompletionTypes.Build[] {
		return this._data.builds || [];
	}

	build(id: string): CompletionTypes.Build {
		let index = this._map?.builds?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.builds[index];
	}

	get instances(): InstanceTypes.InstancesRo {
		return this._data.instances || [];
	}

	instance(id: string): InstanceTypes.InstanceRo {
		let index = this._map?.instances?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.instances[index];
	}

	get firewalls(): FirewallTypes.FirewallsRo {
		return this._data.firewalls || [];
	}

	firewall(id: string): FirewallTypes.FirewallRo {
		let index = this._map?.firewalls?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.firewalls[index];
	}

	get plans(): PlanTypes.PlansRo {
		return this._data.plans || [];
	}

	plan(id: string): PlanTypes.PlanRo {
		let index = this._map?.plans?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.plans[index];
	}

	get certificates(): CertificateTypes.CertificatesRo {
		return this._data.certificates || [];
	}

	certificate(id: string): CertificateTypes.CertificateRo {
		let index = this._map?.certificates?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.certificates[index];
	}

	get secrets(): SecretTypes.SecretsRo {
		return this._data.secrets || [];
	}

	secret(id: string): SecretTypes.SecretRo {
		let index = this._map?.secrets?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.secrets[index];
	}

	get pods(): PodTypes.PodsRo {
		return this._data.pods || [];
	}

	pod(id: string): PodTypes.PodRo {
		let index = this._map?.pods?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.pods[index];
	}

	get units(): PodTypes.UnitsRo {
		return this._data.units || [];
	}

	unit(id: string): PodTypes.UnitRo {
		let index = this._map?.units?.[id];
		if (index === undefined) {
			return null;
		}
		return this._data.units[index];
	}

	emitChange(): void {
		this.emitDefer(GlobalTypes.CHANGE)
	}

	addChangeListener(callback: () => void): void {
		this.on(GlobalTypes.CHANGE, callback)
	}

	removeChangeListener(callback: () => void): void {
		this.removeListener(GlobalTypes.CHANGE, callback)
	}

	_filterCallback(filter: CompletionTypes.Filter): void {
		this._filter = filter
		this.emitChange()
	}

	_sync(completion: CompletionTypes.Completion): void {
		let dataMap: {[key: string]: any} = {}
		let subnets: VpcTypes.Subnet[] = []

		Object.entries(completion).forEach(([resource, items]) => {
			let itemsMap: {[key: string]: number} = {}
			if (items) {
				if (resource === "vpc") {
					for (let i = 0; i < items.length; i++) {
						itemsMap[items[i].id] = i
					}
				} else {
					for (let i = 0; i < items.length; i++) {
						subnets.push(...(items[i].subnets || []) as VpcTypes.Subnet[])
						itemsMap[items[i].id] = i
					}
				}
			}
			dataMap[resource] = itemsMap
		})

		let subnetsMap: {[key: string]: any} = {}
		for (let i = 0; i < subnets.length; i++) {
			subnetsMap[subnets[i].id] = i
		}
		completion.subnets = subnets
		dataMap["subnets"] = subnetsMap

		this._data = Object.freeze(completion)
		this._map = dataMap as CompletionTypes.CompletionMap

		if (Constants.user && !this._userOrg) {
			this._userOrg = this._data?.organizations?.[0]?.id
		}

		this.emitChange()
	}

	_callback(action: CompletionTypes.CompletionDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset(action.data.organization)
				break

			case GlobalTypes.RELOAD:
				Router.refresh()
				break;

			case CompletionTypes.FILTER:
				this._filterCallback(action.data.filter)
				break

			case CompletionTypes.SYNC:
				this._sync(action.data.completion)
				break
		}
	}
}

export default new CompletionStore()
