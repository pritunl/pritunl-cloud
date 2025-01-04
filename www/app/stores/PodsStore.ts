/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as PodTypes from '../types/PodTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class PodsStore extends EventEmitter {
	_pods: PodTypes.PodsRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: PodTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._pods = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this.emitChange();
	}

	get pods(): PodTypes.PodsRo {
		return this._pods;
	}

	get podsM(): PodTypes.Pods {
		let pods: PodTypes.Pods = [];
		this._pods.forEach((pod: PodTypes.PodRo): void => {
			pods.push({
				...pod,
			});
		});
		return pods;
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

	get filter(): PodTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
	}

	pod(id: string): PodTypes.PodRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._pods[i];
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

	addChangeListen(callback: () => void): void {
		this.once(GlobalTypes.CHANGE, callback);
	}

	_traverse(page: number): void {
		this._page = Math.min(this.pages, page);
	}

	_filterCallback(filter: PodTypes.Filter): void {
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

	_sync(pods: PodTypes.Pod[], count: number): void {
		this._map = {};
		for (let i = 0; i < pods.length; i++) {
			pods[i] = Object.freeze(pods[i]);
			this._map[pods[i].id] = i;
		}

		this._count = count;
		this._pods = Object.freeze(pods);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_callback(action: PodTypes.PodDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case PodTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case PodTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case PodTypes.SYNC:
				this._sync(action.data.pods, action.data.count);
				break;
		}
	}
}

export default new PodsStore();
