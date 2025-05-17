/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher'
import EventEmitter from "../EventEmitter"
import * as CompletionTypes from "../types/CompletionTypes"
import * as GlobalTypes from "../types/GlobalTypes"

export interface Kind {
	name: string
	label: string
	title: string
}

export interface Resource {
	id: string
	name: string
	label?: string
	info: ResourceInfo[]
	tags?: Tag[]
}

export interface ResourceInfo {
	label: string
	value: string | number
}

export interface Tag {
	name: string
	label: string
}

export interface Dispatch {
	type: string
}

class CompletionCache extends EventEmitter {
	_kindMap: Record<string, number> = {}
	_kinds: Kind[] = []
	_resourceMap: Record<string, Record<string, number>> = {}
	_resources: Record<string, Resource[]> = {}
	_token = Dispatcher.register((this._callback).bind(this))

	constructor() {
		super()
	}

	get kinds(): Kind[] {
		return this._kinds
	}

	kind(name: string): Kind {
		const i = this._kindMap[name]
		if (i === undefined) {
			return null
		}

		return this._kinds[i]
	}

	resource(kindName: string, name: string): Resource {
		const kindResourceMap = this._resourceMap[kindName]
		if (!kindResourceMap) {
			return null
		}

		const i = kindResourceMap[name]
		if (i === undefined) {
			return null
		}

		return this._resources[kindName][i]
	}

	resources(kind: string): Resource[] {
		return (this._resources[kind] || [])
	}

	addChangeListener(callback: () => void): void {
		this.on(GlobalTypes.CHANGE, callback)
	}

	removeChangeListener(callback: () => void): void {
		this.removeListener(GlobalTypes.CHANGE, callback)
	}

	_reset(): void {
		this._kinds = []
		this._kindMap = {}
		this._resources = {}
		this._resourceMap = {}
	}

	_callback(action: Dispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset()
				break
		}
	}

	update(resources: CompletionTypes.Completion): void {
		this._kinds = []
		this._resources = {}
		let resourceList: Resource[]
		let subResourceList: Resource[]

		this._kinds.push({
			name: "organization",
			label: "Organization",
			title: "**Organization**",
		})
		resourceList = []
		for (let item of (resources.organizations ||[])) {
			resourceList.push({
				id: item.id,
				name: item.name,
				info: [
					{
						label: "**Name**",
						value: item.name,
					},
				],
			})
		}
		this._resources["organization"] = resourceList

		this._kinds.push({
			name: "domain",
			label: "Domain",
			title: "**Domain**",
		})
		resourceList = []
		for (let item of (resources.domains ||[])) {
			resourceList.push({
				id: item.id,
				name: item.name,
				info: [
					{
						label: "**Name**",
						value: item.name,
					},
				],
			})
		}
		this._resources["domain"] = resourceList

		this._kinds.push({
			name: "vpc",
			label: "VPC",
			title: "**VPC**",
		})
		this._kinds.push({
			name: "subnet",
			label: "Subnet",
			title: "**Subnet**",
		})
		resourceList = []
		subResourceList = []
		for (let item of (resources.vpcs ||[])) {
			resourceList.push({
				id: item.id,
				name: item.name,
				info: [
					{
						label: "**Name**",
						value: item.name,
					},
				],
			})

			for (let subItem of (item.subnets || [])) {
				subResourceList.push({
					id: subItem.id,
					name: subItem.name,
					label: "[" + item.name + "] " + subItem.name,
					info: [
						{
							label: "**Name**",
							value: subItem.name,
						},
						{
							label: "**VPC**",
							value: item.name,
						},
					],
				})
			}
		}
		this._resources["vpc"] = resourceList
		this._resources["subnet"] = subResourceList

		this._kinds.push({
			name: "datacenter",
			label: "Datacenter",
			title: "**Datacenter**",
		})
		resourceList = []
		for (let item of (resources.datacenters ||[])) {
			resourceList.push({
				id: item.id,
				name: item.name,
				info: [
					{
						label: "**Name**",
						value: item.name,
					},
				],
			})
		}
		this._resources["datacenter"] = resourceList

		this._kinds.push({
			name: "node",
			label: "Node",
			title: "**Node**",
		})
		resourceList = []
		for (let item of (resources.nodes ||[])) {
			resourceList.push({
				id: item.id,
				name: item.name,
				info: [
					{
						label: "**Name**",
						value: item.name,
					},
				],
			})
		}
		this._resources["node"] = resourceList

		this._kinds.push({
			name: "pool",
			label: "Pool",
			title: "**Pool**",
		})
		resourceList = []
		for (let item of (resources.pools ||[])) {
			resourceList.push({
				id: item.id,
				name: item.name,
				info: [
					{
						label: "**Name**",
						value: item.name,
					},
				],
			})
		}
		this._resources["pool"] = resourceList

		this._kinds.push({
			name: "zone",
			label: "Zone",
			title: "**Zone**",
		})
		resourceList = []
		for (let item of (resources.zones ||[])) {
			resourceList.push({
				id: item.id,
				name: item.name,
				info: [
					{
						label: "**Name**",
						value: item.name,
					},
				],
			})
		}
		this._resources["zone"] = resourceList

		this._kinds.push({
			name: "shape",
			label: "Shapes",
			title: "**Shapes**",
		})
		resourceList = []
		for (let item of (resources.shapes ||[])) {
			resourceList.push({
				id: item.id,
				name: item.name,
				info: [
					{
						label: "**Name**",
						value: item.name,
					},
				],
			})
		}
		this._resources["shape"] = resourceList

		this._kinds.push({
			name: "image",
			label: "Image",
			title: "**Image**",
		})
		resourceList = []
		for (let item of (resources.images ||[])) {
			resourceList.push({
				id: item.id,
				name: item.name,
				info: [
					{
						label: "**Name**",
						value: item.name,
					},
				],
			})
		}
		this._resources["image"] = resourceList

		this._kinds.push({
			name: "build",
			label: "Build",
			title: "**Build**",
		})
		resourceList = []
		for (let item of (resources.builds ||[])) {
			let tags: Tag[] = []
			for (let tag of (item.tags || [])) {
				tags.push({
					name: tag.tag,
					label: tag.tag,
				})
			}

			resourceList.push({
				id: item.id,
				name: item.name,
				info: [
					{
						label: "**Name**",
						value: item.name,
					},
				],
				tags: tags,
			})
		}
		this._resources["build"] = resourceList

		this._kinds.push({
			name: "instance",
			label: "Instance",
			title: "**Instance**",
		})
		resourceList = []
		for (let item of (resources.instances ||[])) {
			resourceList.push({
				id: item.id,
				name: item.name,
				info: [
					{
						label: "**Name**",
						value: item.name,
					},
					{
						label: "**Memory**",
						value: item.memory,
					},
					{
						label: "**Processors**",
						value: item.processors,
					},
				],
			})
		}
		this._resources["instance"] = resourceList

		this._kinds.push({
			name: "plan",
			label: "Plan",
			title: "**Plan**",
		})
		resourceList = []
		for (let item of (resources.plans ||[])) {
			resourceList.push({
				id: item.id,
				name: item.name,
				info: [
					{
						label: "**Name**",
						value: item.name,
					},
				],
			})
		}
		this._resources["plan"] = resourceList

		this._kinds.push({
			name: "certificate",
			label: "Certificate",
			title: "**Certificate**",
		})
		resourceList = []
		for (let item of (resources.certificates ||[])) {
			resourceList.push({
				id: item.id,
				name: item.name,
				info: [
					{
						label: "**Name**",
						value: item.name,
					},
				],
			})
		}
		this._resources["certificate"] = resourceList

		this._kinds.push({
			name: "secret",
			label: "Secret",
			title: "**Secret**",
		})
		resourceList = []
		for (let item of (resources.secrets ||[])) {
			resourceList.push({
				id: item.id,
				name: item.name,
				info: [
					{
						label: "**Name**",
						value: item.name,
					},
				],
			})
		}
		this._resources["secret"] = resourceList

		this._kinds.push({
			name: "pod",
			label: "Pod",
			title: "**Pod**",
		})
		resourceList = []
		for (let item of (resources.pods ||[])) {
			resourceList.push({
				id: item.id,
				name: item.name,
				info: [
					{
						label: "**Name**",
						value: item.name,
					},
				],
			})
		}
		this._resources["pod"] = resourceList

		this._kinds.push({
			name: "unit",
			label: "Unit",
			title: "**Unit**",
		})
		resourceList = []
		for (let item of (resources.units ||[])) {
			resourceList.push({
				id: item.id,
				name: item.name,
				info: [
					{
						label: "**Name**",
						value: item.name,
					},
				],
			})
		}
		this._resources["unit"] = resourceList

		this._kindMap = {}
		for (let i = 0; i < this._kinds.length; i++) {
			this._kindMap[this._kinds[i].name] = i
		}

		this._resourceMap = {}
		Object.entries(this._resources).forEach(([kindName, resources]) => {
			let kindResourceMap: Record<string, number> = {}
			for (let i = 0; i < resources.length; i++) {
				kindResourceMap[resources[i].name] = i
			}
			this._resourceMap[kindName] = kindResourceMap
		})
	}
}

export default new CompletionCache()
