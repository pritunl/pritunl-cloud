/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as BalancerTypes from '../types/BalancerTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class BalancersStore extends EventEmitter {
	_balancers: BalancerTypes.BalancersRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: BalancerTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._balancers = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this.emitChange();
	}

	get balancers(): BalancerTypes.BalancersRo {
		return this._balancers;
	}

	get balancersM(): BalancerTypes.Balancers {
		let balancers: BalancerTypes.Balancers = [];
		this._balancers.forEach((balancer: BalancerTypes.BalancerRo): void => {
			balancers.push({
				...balancer,
			});
		});
		return balancers;
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

	get filter(): BalancerTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
	}

	balancer(id: string): BalancerTypes.BalancerRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._balancers[i];
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

	_filterCallback(filter: BalancerTypes.Filter): void {
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

	_sync(balancers: BalancerTypes.Balancer[], count: number): void {
		this._map = {};
		for (let i = 0; i < balancers.length; i++) {
			balancers[i] = Object.freeze(balancers[i]);
			this._map[balancers[i].id] = i;
		}

		this._count = count;
		this._balancers = Object.freeze(balancers);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_callback(action: BalancerTypes.BalancerDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case BalancerTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case BalancerTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case BalancerTypes.SYNC:
				this._sync(action.data.balancers, action.data.count);
				break;
		}
	}
}

export default new BalancersStore();
