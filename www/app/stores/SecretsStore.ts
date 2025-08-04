/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as SecretTypes from '../types/SecretTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class SecretsStore extends EventEmitter {
	_secrets: SecretTypes.SecretsRo = Object.freeze([]);
	_page: number;
	_pageCount: number;
	_filter: SecretTypes.Filter = null;
	_count: number;
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._secrets = Object.freeze([]);
		this._page = undefined;
		this._pageCount = undefined;
		this._filter = null;
		this._count = undefined;
		this._map = {};
		this.emitChange();
	}

	get secrets(): SecretTypes.SecretsRo {
		return this._secrets;
	}

	get secretsM(): SecretTypes.Secrets {
		let secrets: SecretTypes.Secrets = [];
		this._secrets.forEach((secret: SecretTypes.SecretRo): void => {
			secrets.push({
				...secret,
			});
		});
		return secrets;
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

	get filter(): SecretTypes.Filter {
		return this._filter;
	}

	get count(): number {
		return this._count || 0;
	}

	secret(id: string): SecretTypes.SecretRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._secrets[i];
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

	_filterCallback(filter: SecretTypes.Filter): void {
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

	_sync(secrets: SecretTypes.Secret[], count: number): void {
		this._map = {};
		for (let i = 0; i < secrets.length; i++) {
			secrets[i] = Object.freeze(secrets[i]);
			this._map[secrets[i].id] = i;
		}

		this._count = count;
		this._secrets = Object.freeze(secrets);
		this._page = Math.min(this.pages, this.page);

		this.emitChange();
	}

	_callback(action: SecretTypes.SecretDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case SecretTypes.TRAVERSE:
				this._traverse(action.data.page);
				break;

			case SecretTypes.FILTER:
				this._filterCallback(action.data.filter);
				break;

			case SecretTypes.SYNC:
				this._sync(action.data.secrets, action.data.count);
				break;
		}
	}
}

export default new SecretsStore();
