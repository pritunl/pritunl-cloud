/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as FirewallTypes from '../types/FirewallTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class FirewallsStore extends EventEmitter {
	_firewalls: FirewallTypes.FirewallsRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: FirewallTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._firewalls = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this.emitChange();
	}

	get firewalls(): FirewallTypes.FirewallsRo {
		return this._firewalls;
	}

	get firewallsM(): FirewallTypes.Firewalls {
		let firewalls: FirewallTypes.Firewalls = [];
		this._firewalls.forEach((firewall: FirewallTypes.FirewallRo): void => {
			firewalls.push({
				...firewall,
			});
		});
		return firewalls;
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

	get filter(): FirewallTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
	}

	firewall(id: string): FirewallTypes.FirewallRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._firewalls[i];
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

	_filterCallback(filter: FirewallTypes.Filter): void {
		if ((this._filter !== null && filter === null) ||
			(this._filter === {} && filter !== null) || (
				filter && this._filter && (
					filter.name !== this._filter.name
				))) {
			this._traverse(0);
		}
		this._filter = filter;
		this.emitChange();
	}

	_sync(firewalls: FirewallTypes.Firewall[], count: number): void {
		this._map = {};
		for (let i = 0; i < firewalls.length; i++) {
			firewalls[i] = Object.freeze(firewalls[i]);
			this._map[firewalls[i].id] = i;
		}

		this._count = count;
		this._firewalls = Object.freeze(firewalls);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_callback(action: FirewallTypes.FirewallDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case FirewallTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case FirewallTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case FirewallTypes.SYNC:
				this._sync(action.data.firewalls, action.data.count);
				break;
		}
	}
}

export default new FirewallsStore();
