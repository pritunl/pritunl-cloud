/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as DomainTypes from '../types/DomainTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class DomainsStore extends EventEmitter {
	_domains: DomainTypes.DomainsRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: DomainTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._domains = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this.emitChange();
	}

	get domains(): DomainTypes.DomainsRo {
		return this._domains;
	}

	get domainsM(): DomainTypes.Domains {
		let domains: DomainTypes.Domains = [];
		this._domains.forEach((domain: DomainTypes.DomainRo): void => {
			domains.push({
				...domain,
			});
		});
		return domains;
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

	get filter(): DomainTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
	}

	domain(id: string): DomainTypes.DomainRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._domains[i];
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

	_filterCallback(filter: DomainTypes.Filter): void {
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

	_sync(domains: DomainTypes.Domain[], count: number): void {
		this._map = {};
		for (let i = 0; i < domains.length; i++) {
			domains[i] = Object.freeze(domains[i]);
			this._map[domains[i].id] = i;
		}

		this._count = count;
		this._domains = Object.freeze(domains);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_callback(action: DomainTypes.DomainDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case DomainTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case DomainTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case DomainTypes.SYNC:
				this._sync(action.data.domains, action.data.count);
				break;
		}
	}
}

export default new DomainsStore();
