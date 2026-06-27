/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as AdvisoryTypes from '../types/AdvisoryTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class AdvisoriesStore extends EventEmitter {
	_advisories: AdvisoryTypes.AdvisoriesRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: AdvisoryTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._advisories = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this.emitChange();
	}

	get advisories(): AdvisoryTypes.AdvisoriesRo {
		return this._advisories;
	}

	get advisoriesM(): AdvisoryTypes.Advisories {
		let advisories: AdvisoryTypes.Advisories = [];
		this._advisories.forEach((
				advisory: AdvisoryTypes.AdvisoryRo): void => {
			advisories.push({
				...advisory,
			});
		});
		return advisories;
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

	get filter(): AdvisoryTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
	}

	advisory(id: string): AdvisoryTypes.AdvisoryRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._advisories[i];
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

	_filterCallback(filter: AdvisoryTypes.Filter): void {
		if ((this._filter !== null && filter === null) ||
			(!Object.keys(this._filter || {}).length && filter !== null) || (
				filter && this._filter && (
					filter.reference !== this._filter.reference
				))) {
			this._traverse(0);
		}
		this._filter = filter;
		this.emitChange();
	}

	_sync(advisories: AdvisoryTypes.Advisory[], count: number): void {
		this._map = {};
		for (let i = 0; i < advisories.length; i++) {
			advisories[i] = Object.freeze(advisories[i]);
			this._map[advisories[i].id] = i;
		}

		this._count = count;
		this._advisories = Object.freeze(advisories);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_callback(action: AdvisoryTypes.AdvisoryDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case AdvisoryTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case AdvisoryTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case AdvisoryTypes.SYNC:
				this._sync(action.data.advisories, action.data.count);
				break;
		}
	}
}

export default new AdvisoriesStore();
