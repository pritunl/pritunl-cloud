/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class DatacentersStore extends EventEmitter {
	_datacenters: DatacenterTypes.DatacentersRo = Object.freeze([]);
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

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

	datacenter(id: string): DatacenterTypes.DatacenterRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._datacenters[i];
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

	_sync(datacenters: DatacenterTypes.Datacenter[]): void {
		this._map = {};
		for (let i = 0; i < datacenters.length; i++) {
			datacenters[i] = Object.freeze(datacenters[i]);
			this._map[datacenters[i].id] = i;
		}

		this._datacenters = Object.freeze(datacenters);
		this.emitChange();
	}

	_callback(action: DatacenterTypes.DatacenterDispatch): void {
		switch (action.type) {
			case DatacenterTypes.SYNC:
				this._sync(action.data.datacenters);
				break;
		}
	}
}

export default new DatacentersStore();
