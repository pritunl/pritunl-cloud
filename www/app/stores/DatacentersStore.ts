/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class DatacentersStore extends EventEmitter {
	_datacenters: DatacenterTypes.DatacentersRo = Object.freeze([]);
	_datacentersName: DatacenterTypes.DatacentersRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: DatacenterTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_mapName: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._datacenters = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this.emitChange();
	}

	get datacenters(): DatacenterTypes.DatacentersRo {
		return this._datacenters;
	}

	get datacentersM(): DatacenterTypes.Datacenters {
		let datacenters: DatacenterTypes.Datacenters = [];
		this._datacenters.forEach((
			datacenter: DatacenterTypes.DatacenterRo): void => {

			datacenters.push({
				...datacenter,
			});
		});
		return datacenters;
	}

	get datacentersName(): DatacenterTypes.DatacentersRo {
		return this._datacentersName || [];
	}

	get datacentersNameM(): DatacenterTypes.Datacenters {
		let datacenters: DatacenterTypes.Datacenters = [];
		this._datacentersName.forEach((
			datacenter: DatacenterTypes.DatacenterRo): void => {

			datacenters.push({
				...datacenter,
			});
		});
		return datacenters;
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

	get filter(): DatacenterTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
	}

	datacenter(id: string): DatacenterTypes.DatacenterRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._datacenters[i];
	}

	datacenterName(id: string): DatacenterTypes.DatacenterRo {
		let i = this._mapName[id];
		if (i === undefined) {
			return null;
		}
		return this._datacentersName[i];
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

	_filterCallback(filter: DatacenterTypes.Filter): void {
		if ((this._filter !== null && filter === null) ||
			(!Object.keys(this._filter || {}).length && filter !== null) || (
				filter && this._filter && (
					filter.name !== this._filter.name
				))) {
			this._traverse(0);
		}
		this._filter = filter;
		this.emitChange();
	}

	_sync(datacenters: DatacenterTypes.Datacenter[], count: number): void {
		this._map = {};
		for (let i = 0; i < datacenters.length; i++) {
			datacenters[i] = Object.freeze(datacenters[i]);
			this._map[datacenters[i].id] = i;
		}

		this._count = count;
		this._datacenters = Object.freeze(datacenters);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_syncNames(datacenters: DatacenterTypes.Datacenter[]): void {
		this._mapName = {};
		for (let i = 0; i < datacenters.length; i++) {
			datacenters[i] = Object.freeze(datacenters[i]);
			this._mapName[datacenters[i].id] = i;
		}

		this._datacentersName = Object.freeze(datacenters);
		this.emitChange();
	}

	_callback(action: DatacenterTypes.DatacenterDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case DatacenterTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case DatacenterTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case DatacenterTypes.SYNC:
				this._sync(action.data.datacenters, action.data.count);
				break;

			case DatacenterTypes.SYNC_NAMES:
				this._syncNames(action.data.datacenters);
				break;
		}
	}
}

export default new DatacentersStore();
