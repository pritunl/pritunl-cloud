/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as ZoneTypes from '../types/ZoneTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class ZonesStore extends EventEmitter {
	_zones: ZoneTypes.ZonesRo = Object.freeze([]);
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._zones = Object.freeze([]);
		this._map = {};
		this.emitChange();
	}

	get zones(): ZoneTypes.ZonesRo {
		return this._zones;
	}

	get zonesM(): ZoneTypes.Zones {
		let zones: ZoneTypes.Zones = [];
		this._zones.forEach((
				zone: ZoneTypes.ZoneRo): void => {
			zones.push({
				...zone,
			});
		});
		return zones;
	}

	zone(id: string): ZoneTypes.ZoneRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._zones[i];
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

	_sync(zones: ZoneTypes.Zone[]): void {
		this._map = {};
		for (let i = 0; i < zones.length; i++) {
			zones[i] = Object.freeze(zones[i]);
			this._map[zones[i].id] = i;
		}

		this._zones = Object.freeze(zones);
		this.emitChange();
	}

	_callback(action: ZoneTypes.ZoneDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case ZoneTypes.SYNC:
				this._sync(action.data.zones);
				break;
		}
	}
}

export default new ZonesStore();
