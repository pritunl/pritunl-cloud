/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as PoolTypes from '../types/PoolTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class PoolsStore extends EventEmitter {
	_pools: PoolTypes.PoolsRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: PoolTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._pools = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this.emitChange();
	}

	get pools(): PoolTypes.PoolsRo {
		return this._pools;
	}

	get poolsM(): PoolTypes.Pools {
		let pools: PoolTypes.Pools = [];
		this._pools.forEach((pool: PoolTypes.PoolRo): void => {
			pools.push({
				...pool,
			});
		});
		return pools;
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

	get filter(): PoolTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
	}

	pool(id: string): PoolTypes.PoolRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._pools[i];
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

	_filterCallback(filter: PoolTypes.Filter): void {
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

	_sync(pools: PoolTypes.Pool[], count: number): void {
		this._map = {};
		for (let i = 0; i < pools.length; i++) {
			pools[i] = Object.freeze(pools[i]);
			this._map[pools[i].id] = i;
		}

		this._count = count;
		this._pools = Object.freeze(pools);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_callback(action: PoolTypes.PoolDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case PoolTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case PoolTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case PoolTypes.SYNC:
				this._sync(action.data.pools, action.data.count);
				break;
		}
	}
}

export default new PoolsStore();
