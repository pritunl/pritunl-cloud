/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as StorageTypes from '../types/StorageTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class StoragesStore extends EventEmitter {
	_storages: StorageTypes.StoragesRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: StorageTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._storages = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this.emitChange();
	}

	get storages(): StorageTypes.StoragesRo {
		return this._storages;
	}

	get storagesM(): StorageTypes.Storages {
		let storages: StorageTypes.Storages = [];
		this._storages.forEach((storage: StorageTypes.StorageRo): void => {
			storages.push({
				...storage,
			});
		});
		return storages;
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

	get filter(): StorageTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
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

	_traverse(page: number): void {
		this._page = Math.min(this.pages, page);
	}

	_filterCallback(filter: StorageTypes.Filter): void {
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

	_sync(storages: StorageTypes.Storage[], count: number): void {
		this._map = {};
		for (let i = 0; i < storages.length; i++) {
			storages[i] = Object.freeze(storages[i]);
			this._map[storages[i].id] = i;
		}

		this._count = count;
		this._storages = Object.freeze(storages);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_callback(action: StorageTypes.StorageDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case StorageTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case StorageTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case StorageTypes.SYNC:
				this._sync(action.data.storages, action.data.count);
				break;
		}
	}
}

export default new StoragesStore();
