/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as InstanceTypes from '../types/InstanceTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class InstancesStore extends EventEmitter {
	_instances: InstanceTypes.InstancesRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: InstanceTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._instances = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this.emitChange();
	}

	get instances(): InstanceTypes.InstancesRo {
		return this._instances;
	}

	get instancesM(): InstanceTypes.Instances {
		let instances: InstanceTypes.Instances = [];
		this._instances.forEach((instance: InstanceTypes.InstanceRo): void => {
			instances.push({
				...instance,
			});
		});
		return instances;
	}

	get page(): number {
		return this._page || 0;
	}

	get pageCount(): number {
		return this._pageCount || 20;
	}

	get pages(): number {
		return Math.ceil(this.count / this.pageCount);
	}

	get filter(): InstanceTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
	}

	instance(id: string): InstanceTypes.InstanceRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._instances[i];
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

	_traverse(page: number): void {
		this._page = Math.min(this.pages, page);
	}

	_filterCallback(filter: InstanceTypes.Filter): void {
		if ((this._filter !== null && filter === null) ||
			(this._filter === {} && filter !== null) || (
				filter && this._filter && (
					filter.name !== this._filter.name
				))) {
			this._traverse(0);
		}
		this._filter = filter;
		this.emitChange();
	}

	_sync(instances: InstanceTypes.Instance[], count: number): void {
		this._map = {};
		for (let i = 0; i < instances.length; i++) {
			instances[i] = Object.freeze(instances[i]);
			this._map[instances[i].id] = i;
		}

		this._count = count;
		this._instances = Object.freeze(instances);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_callback(action: InstanceTypes.InstanceDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case InstanceTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case InstanceTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case InstanceTypes.SYNC:
				this._sync(action.data.instances, action.data.count);
				break;
		}
	}
}

export default new InstancesStore();
