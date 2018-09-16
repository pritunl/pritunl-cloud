/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as DomainTypes from '../types/DomainTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class DomainsNameStore extends EventEmitter {
	_domains: DomainTypes.DomainsRo = Object.freeze([]);
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	_reset(): void {
		this._domains = Object.freeze([]);
		this._map = {};
		this.emitChange();
	}

	get domains(): DomainTypes.DomainsRo {
		return this._domains;
	}

	get domainsM(): DomainTypes.Domains {
		let domains: DomainTypes.Domains = [];
		this._domains.forEach((
				domain: DomainTypes.DomainRo): void => {
			domains.push({
				...domain,
			});
		});
		return domains;
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

	_sync(domains: DomainTypes.Domain[]): void {
		this._map = {};
		for (let i = 0; i < domains.length; i++) {
			domains[i] = Object.freeze(domains[i]);
			this._map[domains[i].id] = i;
		}

		this._domains = Object.freeze(domains);
		this.emitChange();
	}

	_callback(action: DomainTypes.DomainDispatch): void {
		switch (action.type) {
			case GlobalTypes.RESET:
				this._reset();
				break;

			case DomainTypes.SYNC_NAME:
				this._sync(action.data.domains);
				break;
		}
	}
}

export default new DomainsNameStore();
