/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as VpcTypes from '../types/VpcTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class VpcsStore extends EventEmitter {
	_vpcs: VpcTypes.VpcsRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: VpcTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._vpcs = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this.emitChange();
	}

	get vpcs(): VpcTypes.VpcsRo {
		return this._vpcs;
	}

	get vpcsM(): VpcTypes.Vpcs {
		let vpcs: VpcTypes.Vpcs = [];
		this._vpcs.forEach((vpc: VpcTypes.VpcRo): void => {
			vpcs.push({
				...vpc,
			});
		});
		return vpcs;
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

	get filter(): VpcTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
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

	_traverse(page: number): void {
		this._page = Math.min(this.pages, page);
	}

	_filterCallback(filter: VpcTypes.Filter): void {
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

	_sync(vpcs: VpcTypes.Vpc[], count: number): void {
		this._map = {};
		for (let i = 0; i < vpcs.length; i++) {
			vpcs[i] = Object.freeze(vpcs[i]);
			this._map[vpcs[i].id] = i;
		}

		this._count = count;
		this._vpcs = Object.freeze(vpcs);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_callback(action: VpcTypes.VpcDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case VpcTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case VpcTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case VpcTypes.SYNC:
				this._sync(action.data.vpcs, action.data.count);
				break;
		}
	}
}

export default new VpcsStore();
