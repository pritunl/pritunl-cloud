/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher'
import EventEmitter from "../EventEmitter"
import * as CompletionTypes from "./Types"
import * as GlobalTypes from "../types/GlobalTypes"

class CompletionStore extends EventEmitter {
	_kindMap: Record<string, number> = {}
	_kinds: CompletionTypes.Kind[] = []
	_resourceMap: Record<string, Record<string, number>> = {}
	_resources: Record<string, CompletionTypes.Resource[]> = {}
	_token = Dispatcher.register((this._callback).bind(this))

	constructor() {
		super()

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

	get kinds(): CompletionTypes.Kind[] {
		return this._kinds
	}

	kind(name: string): CompletionTypes.Kind {
		const i = this._kindMap[name]
		if (i === undefined) {
			return null
		}

		return this._kinds[i]
	}

	resource(kindName: string, name: string): CompletionTypes.Resource {
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

	resources(kind: string): CompletionTypes.Resource[] {
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
		this._resources = {}
	}

	_callback(action: CompletionTypes.Dispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset()
				break
		}
	}
}

export default new CompletionStore()
