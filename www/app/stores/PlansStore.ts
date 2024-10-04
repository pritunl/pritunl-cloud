/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as PlanTypes from '../types/PlanTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class PlansStore extends EventEmitter {
	_plans: PlanTypes.PlansRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: PlanTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._plans = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this.emitChange();
	}

	get plans(): PlanTypes.PlansRo {
		return this._plans;
	}

	get plansM(): PlanTypes.Plans {
		let plans: PlanTypes.Plans = [];
		this._plans.forEach((plan: PlanTypes.PlanRo): void => {
			plans.push({
				...plan,
			});
		});
		return plans;
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

	get filter(): PlanTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
	}

	plan(id: string): PlanTypes.PlanRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._plans[i];
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

	_filterCallback(filter: PlanTypes.Filter): void {
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

	_sync(plans: PlanTypes.Plan[], count: number): void {
		this._map = {};
		for (let i = 0; i < plans.length; i++) {
			plans[i] = Object.freeze(plans[i]);
			this._map[plans[i].id] = i;
		}

		this._count = count;
		this._plans = Object.freeze(plans);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_callback(action: PlanTypes.PlanDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case PlanTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case PlanTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case PlanTypes.SYNC:
				this._sync(action.data.plans, action.data.count);
				break;
		}
	}
}

export default new PlansStore();
