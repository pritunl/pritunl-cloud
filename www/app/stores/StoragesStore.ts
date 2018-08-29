/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as StorageTypes from '../types/StorageTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class StoragesStore extends EventEmitter {
	_storages: StorageTypes.StoragesRo = Object.freeze([]);
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._storages = Object.freeze([]);
		this._map = {};
		this.emitChange();
	}

	get storages(): StorageTypes.StoragesRo {
		return this._storages;
	}

	get storagesM(): StorageTypes.Storages {
		let storages: StorageTypes.Storages = [];
		this._storages.forEach((
				storage: StorageTypes.StorageRo): void => {
			storages.push({
				...storage,
			});
		});
		return storages;
	}

	storage(id: string): StorageTypes.StorageRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._storages[i];
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

	_sync(storages: StorageTypes.Storage[]): void {
		this._map = {};
		for (let i = 0; i < storages.length; i++) {
			storages[i] = Object.freeze(storages[i]);
			this._map[storages[i].id] = i;
		}

		this._storages = Object.freeze(storages);
		this.emitChange();
	}

	_callback(action: StorageTypes.StorageDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case StorageTypes.SYNC:
				this._sync(action.data.storages);
				break;
		}
	}
}

export default new StoragesStore();
