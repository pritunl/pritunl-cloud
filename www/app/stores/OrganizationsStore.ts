/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as Constants from '../Constants';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class OrganizationsStore extends EventEmitter {
	_organizations: OrganizationTypes.OrganizationsRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: OrganizationTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._organizations = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this.emitChange();
	}

	get organizations(): OrganizationTypes.OrganizationsRo {
		return this._organizations;
	}

	get organizationsM(): OrganizationTypes.Organizations {
		let organizations: OrganizationTypes.Organizations = [];
		this._organizations.forEach((organization: OrganizationTypes.OrganizationRo): void => {
			organizations.push({
				...organization,
			});
		});
		return organizations;
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

	get filter(): OrganizationTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
	}

	organization(id: string): OrganizationTypes.OrganizationRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._organizations[i];
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

	_filterCallback(filter: OrganizationTypes.Filter): void {
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

	_sync(organizations: OrganizationTypes.Organization[], count: number): void {
		this._map = {};
		for (let i = 0; i < organizations.length; i++) {
			organizations[i] = Object.freeze(organizations[i]);
			this._map[organizations[i].id] = i;
		}

		this._count = count;
		this._organizations = Object.freeze(organizations);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_callback(action: OrganizationTypes.OrganizationDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case OrganizationTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case OrganizationTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case OrganizationTypes.SYNC:
				this._sync(action.data.organizations, action.data.count);
				break;
		}
	}
}

export default new OrganizationsStore();
