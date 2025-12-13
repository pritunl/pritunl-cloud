/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as ZoneTypes from '../types/ZoneTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class ZonesStore extends EventEmitter {
	_zones: ZoneTypes.ZonesRo = Object.freeze([]);
	_zonesName: ZoneTypes.ZonesRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: ZoneTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_mapName: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._zones = Object.freeze([]);
		this._zonesName = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this._mapName = {};
		this.emitChange();
	}

	get zones(): ZoneTypes.ZonesRo {
		return this._zones;
	}

	get zonesM(): ZoneTypes.Zones {
		let zones: ZoneTypes.Zones = [];
		this._zones.forEach((zone: ZoneTypes.ZoneRo): void => {
			zones.push({
				...zone,
			});
		});
		return zones;
	}

	get zonesName(): ZoneTypes.ZonesRo {
		return this._zonesName || [];
	}

	get zonesNameM(): ZoneTypes.Zones {
		let zones: ZoneTypes.Zones = [];
		this._zonesName.forEach((
			zone: ZoneTypes.ZoneRo): void => {

			zones.push({
				...zone,
			});
		});
		return zones;
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

	get filter(): ZoneTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
	}

	zone(id: string): ZoneTypes.ZoneRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._zones[i];
	}

	zoneName(id: string): ZoneTypes.ZoneRo {
		let i = this._mapName[id];
		if (i === undefined) {
			return null;
		}
		return this._zonesName[i];
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

	_filterCallback(filter: ZoneTypes.Filter): void {
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

	_sync(zones: ZoneTypes.Zone[], count: number): void {
		this._map = {};
		for (let i = 0; i < zones.length; i++) {
			zones[i] = Object.freeze(zones[i]);
			this._map[zones[i].id] = i;
		}

		this._count = count;
		this._zones = Object.freeze(zones);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_syncNames(zones: ZoneTypes.Zone[]): void {
		this._mapName = {};
		for (let i = 0; i < zones.length; i++) {
			zones[i] = Object.freeze(zones[i]);
			this._mapName[zones[i].id] = i;
		}

		this._zonesName = Object.freeze(zones);
		this.emitChange();
	}

	_callback(action: ZoneTypes.ZoneDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case ZoneTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case ZoneTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case ZoneTypes.SYNC:
				this._sync(action.data.zones, action.data.count);
				break;

			case ZoneTypes.SYNC_NAMES:
				this._syncNames(action.data.zones);
				break;
		}
	}
}

export default new ZonesStore();
