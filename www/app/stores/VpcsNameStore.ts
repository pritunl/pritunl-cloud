/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as VpcTypes from '../types/VpcTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class VpcsZoneStore extends EventEmitter {
	_vpcs: VpcTypes.VpcsRo = Object.freeze([]);
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._vpcs = Object.freeze([]);
		this._map = {};
		this.emitChange();
	}

	get vpcs(): VpcTypes.VpcsRo {
		return this._vpcs;
	}

	get vpcsM(): VpcTypes.Vpcs {
		let vpcs: VpcTypes.Vpcs = [];
		this._vpcs.forEach((
				vpc: VpcTypes.VpcRo): void => {
			vpcs.push({
				...vpc,
			});
		});
		return vpcs;
	}

	vpc(id: string): VpcTypes.VpcRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._vpcs[i];
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

	_sync(vpcs: VpcTypes.Vpc[]): void {
		this._map = {};
		for (let i = 0; i < vpcs.length; i++) {
			vpcs[i] = Object.freeze(vpcs[i]);
			this._map[vpcs[i].id] = i;
		}

		this._vpcs = Object.freeze(vpcs);
		this.emitChange();
	}

	_callback(action: VpcTypes.VpcDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case VpcTypes.SYNC_NAMES:
				this._sync(action.data.vpcs);
				break;
		}
	}
}

export default new VpcsZoneStore();
