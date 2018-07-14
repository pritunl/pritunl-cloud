/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as InstanceTypes from '../types/InstanceTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class InstancesNodeStore extends EventEmitter {
	_instances: InstanceTypes.InstancesNodeRo = new Map<string, InstanceTypes.InstancesRo>(Object.freeze([]));
	_map: {[key: string]: [string, number]} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._instances = new Map<string, InstanceTypes.InstancesRo>(
			Object.freeze([]));
		this._map = {};
		this.emitChange();
	}

	instances(node: string): InstanceTypes.InstancesRo {
		return this._instances.get(node) || [];
	}

	instance(id: string): InstanceTypes.InstanceRo {
		let x = this._map[id];
		if (x === undefined) {
			return null;
		}
		return this._instances.get(x[0])[x[1]];
	}

	emitChange(): void {
		this.emitDefer(GlobalTypes.CHANGE);
	}

	addChangeListener(callback: () => void): void {
		this.on(GlobalTypes.CHANGE, callback);
	}

	removeChangeListener(callback: () => void): void {
		this.removeListener(GlobalTypes.CHANGE, callback);
	}

	_sync(node: string, instances: InstanceTypes.Instance[]): void {
		for (let i = 0; i < instances.length; i++) {
			instances[i] = Object.freeze(instances[i]);
		}
		this._instances.set(node, Object.freeze(instances));

		this._map = {};
		for (let item of this._instances.entries()) {
			let insts = item[1];

			for (let i = 0; i < insts.length; i++) {
				this._map[insts[i].id] = [item[0], i];
			}
		}

		this.emitChange();
	}

	_callback(action: InstanceTypes.InstanceDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case InstanceTypes.SYNC_NODE:
				this._sync(action.data.node, action.data.instances);
				break;
		}
	}
}

export default new InstancesNodeStore();
