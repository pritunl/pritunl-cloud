/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as DiskTypes from '../types/DiskTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class DisksStore extends EventEmitter {
	_disks: DiskTypes.DisksRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: DiskTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._disks = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this.emitChange();
	}

	get disks(): DiskTypes.DisksRo {
		return this._disks;
	}

	get disksM(): DiskTypes.Disks {
		let disks: DiskTypes.Disks = [];
		this._disks.forEach((disk: DiskTypes.DiskRo): void => {
			disks.push({
				...disk,
			});
		});
		return disks;
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

	get filter(): DiskTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
	}

	disk(id: string): DiskTypes.DiskRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._disks[i];
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

	_filterCallback(filter: DiskTypes.Filter): void {
		if ((this._filter !== null && filter === null) ||
			(!Object.keys(this._filter).length && filter !== null) || (
				filter && this._filter && (
					filter.name !== this._filter.name
				))) {
			this._traverse(0);
		}
		this._filter = filter;
		this.emitChange();
	}

	_sync(disks: DiskTypes.Disk[], count: number): void {
		this._map = {};
		for (let i = 0; i < disks.length; i++) {
			disks[i] = Object.freeze(disks[i]);
			this._map[disks[i].id] = i;
		}

		this._count = count;
		this._disks = Object.freeze(disks);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_callback(action: DiskTypes.DiskDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case DiskTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case DiskTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case DiskTypes.SYNC:
				this._sync(action.data.disks, action.data.count);
				break;
		}
	}
}

export default new DisksStore();
