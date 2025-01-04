/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as PodTypes from '../types/PodTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class PodsUnitStore extends EventEmitter {
	_units: {[key: string]: PodTypes.PodUnit} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._units = {};
		this.emitChange();
	}

	unit(unitId: string): PodTypes.PodUnit {
		return this._units[unitId];
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

	_sync(unit: PodTypes.PodUnit): void {
		this._units[unit.id] = Object.freeze(unit);
		this.emitChange();
	}

	_callback(action: PodTypes.PodUnitDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case PodTypes.SYNC_UNIT:
				this._sync(action.data.unit);
				break;
		}
	}
}

export default new PodsUnitStore();
